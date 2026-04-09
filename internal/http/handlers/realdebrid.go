package handlers

import (
	"StreamflixBackend/internal/models"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

// buildQualitiesList convertit une map de qualités disponibles en un slice ordonné de [models.Quality].
//
// Les qualités sont triées selon un ordre de priorité prédéfini : "Original", "1080", "720",
// "480", "360". Seules les clés présentes dans la map sont incluses dans le résultat.
// Contrairement à [buildQualitiesListWithURLs] (dans player.go), cette fonction ne calcule
// pas d'URL pour chaque qualité ; elle se contente de mapper clé/valeur.
//
// Paramètre :
//   - qualities : map[string]string où la clé est le label de qualité (ex. "1080")
//     et la valeur est l'identifiant interne correspondant.
//
// Retourne un slice de [models.Quality] trié par résolution décroissante.
func buildQualitiesList(qualities map[string]string) []models.Quality {
	result := []models.Quality{}

	// Ordre de priorité pour l'affichage
	order := []string{"Original", "1080", "720", "480", "360"}

	for _, key := range order {
		if value, ok := qualities[key]; ok {
			result = append(result, models.Quality{
				Label: key,
				Value: value,
			})
		}
	}

	return result
}

// buildAudioList convertit la map des pistes audio du média en un slice de [models.PlayerAudioTrack]
// exploitable par le lecteur vidéo du frontend.
//
// Pour chaque piste audio, un label descriptif est généré au format :
//
//	"<langue> (<CODEC> <canaux>)"
//
// Par exemple : "English (AAC Stereo)" ou "French (DTS 5.1)".
// Le nombre de canaux est affiché en notation numérique (ex. "5.1") si >= 6,
// sinon "Stereo" est utilisé.
//
// Paramètre :
//   - audioTracks : map[string][models.AudioTrack] provenant de MediaInfoResponse.Details.Audio,
//     où la clé est l'identifiant de la piste (ex. "eng1", "fre1").
//
// Retourne un slice de [models.PlayerAudioTrack] avec label, valeur et codec pour chaque piste.
func buildAudioList(audioTracks map[string]models.AudioTrack) []models.PlayerAudioTrack {
	audioList := []models.PlayerAudioTrack{}

	for key, track := range audioTracks {
		channels := ""
		if track.Channels >= 6 {
			channels = fmt.Sprintf(" %.1f", track.Channels)
		} else {
			channels = " Stereo"
		}

		label := fmt.Sprintf("%s (%s%s)", track.Lang, strings.ToUpper(track.Codec), channels)

		audioList = append(audioList, models.PlayerAudioTrack{
			Label: label,
			Value: key,
			Codec: strings.ToLower(track.Codec),
		})
	}

	return audioList
}

// buildSubtitlesList convertit la map des pistes de sous-titres en un slice de
// [models.PlayerSubtitleTrack] pour le lecteur vidéo.
//
// Le premier élément du slice retourné est toujours l'option "Aucun" (valeur "none"),
// permettant à l'utilisateur de désactiver les sous-titres. Les pistes suivantes sont
// ajoutées à partir de la map fournie, chaque entrée utilisant la langue comme label.
//
// Paramètre :
//   - subtitleTracks : [models.SubtitleTracks] provenant de MediaInfoResponse.Details.Subtitles,
//     où la clé est l'identifiant de la piste (ex. "fre1") et la valeur contient la langue.
//
// Retourne un slice de [models.PlayerSubtitleTrack] commençant par "Aucun".
func buildSubtitlesList(subtitleTracks models.SubtitleTracks) []models.PlayerSubtitleTrack {
	subtitleList := []models.PlayerSubtitleTrack{
		{
			Label: "Aucun",
			Value: "none",
		},
	}

	for key, track := range subtitleTracks {
		subtitleList = append(subtitleList, models.PlayerSubtitleTrack{
			Label: track.Lang,
			Value: key,
		})
	}

	return subtitleList
}

// UnrestrictAndGetMPD déverrouille un lien hébergeur via l'API Real-Debrid et retourne
// l'identifiant du fichier déverrouillé.
//
// Cette fonction effectue un appel POST à l'endpoint /unrestrict/link de l'API Real-Debrid.
// Le lien original (ex. lien de téléchargement d'un hébergeur premium) est envoyé en tant
// que paramètre de formulaire. L'API retourne un objet [models.UnrestrictResponse] contenant
// l'identifiant du fichier déverrouillé.
//
// Appel API externe :
//   - POST https://api.real-debrid.com/rest/1.0/unrestrict/link
//   - En-têtes : Authorization Bearer, Content-Type application/x-www-form-urlencoded
//   - Timeout : 30 secondes
//
// Paramètres :
//   - token : jeton d'authentification Real-Debrid (OAuth Bearer token).
//   - originalLink : URL du lien hébergeur à déverrouiller.
//
// Retourne l'identifiant (ID) du fichier déverrouillé, ou une erreur dans les cas suivants :
//   - échec de création de la requête HTTP,
//   - erreur réseau ou timeout,
//   - code HTTP non 2xx (l'erreur Real-Debrid est parsée si possible),
//   - réponse JSON invalide ou identifiant vide dans la réponse.
func UnrestrictAndGetMPD(token, originalLink string) (string, error) {
	endpoint := "https://api.real-debrid.com/rest/1.0/unrestrict/link"

	form := url.Values{}
	form.Set("link", originalLink)

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var rdErr models.RDError
		if err := json.Unmarshal(body, &rdErr); err != nil {
			return "", fmt.Errorf("real-debrid http %d: %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("real-debrid http %d: %s (code=%d)", resp.StatusCode, rdErr.Error, rdErr.ErrorCode)
	}

	var out models.UnrestrictResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("cannot parse unrestrict response: %w; raw=%s", err, string(body))
	}

	if out.ID == "" {
		return "", fmt.Errorf("unrestrict response has empty id; raw=%s", string(body))
	}

	fmt.Println("out : id : ", out.ID)
	return out.ID, nil
}

// GetMPDFromTranscode récupère l'URL du manifeste DASH (MPD) pour un fichier donné
// via l'endpoint de transcodage de l'API Real-Debrid.
//
// Cette fonction effectue un appel GET à l'endpoint /streaming/transcode/{fileID}.
// L'API retourne un objet [models.TranscodeResponse] contenant les URLs de streaming
// dans différents formats. Seule l'URL DASH complète (Dash.Full) est extraite et retournée.
//
// Appel API externe :
//   - GET https://api.real-debrid.com/rest/1.0/streaming/transcode/{fileID}
//   - En-tête : Authorization Bearer
//   - Timeout : 30 secondes
//
// Paramètres :
//   - token : jeton d'authentification Real-Debrid (OAuth Bearer token).
//   - fileID : identifiant du fichier sur Real-Debrid (obtenu via [UnrestrictAndGetMPD]).
//
// Retourne l'URL complète du manifeste DASH (MPD), ou une erreur dans les cas suivants :
//   - échec de création de la requête HTTP,
//   - erreur réseau ou timeout,
//   - code HTTP non 2xx,
//   - réponse JSON invalide ou URL DASH absente de la réponse.
func GetMPDFromTranscode(token, fileID string) (string, error) {
	endpoint := fmt.Sprintf("https://api.real-debrid.com/rest/1.0/streaming/transcode/%s", fileID)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("real-debrid http %d: %s", resp.StatusCode, string(body))
	}

	var out models.TranscodeResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("cannot parse transcode response: %w; raw=%s", err, string(body))
	}

	if out.Dash.Full == "" {
		return "", fmt.Errorf("no DASH (mpd) url returned; raw=%s", string(body))
	}

	return out.Dash.Full, nil
}

// GetMediaInfos récupère les métadonnées détaillées d'un fichier média depuis l'API Real-Debrid.
//
// Cette fonction effectue un appel GET à l'endpoint /streaming/mediaInfos/{fileID}.
// L'API retourne un objet [models.MediaInfoResponse] riche contenant :
//   - le nom du fichier, le type de média, la durée,
//   - les pistes audio et sous-titres disponibles (Details.Audio, Details.Subtitles),
//   - les qualités de transcodage disponibles (AvailableQualities),
//   - le modèle d'URL de streaming (ModelURL) avec des placeholders à remplacer,
//   - les chemins vers le poster et le backdrop.
//
// Des logs sont émis à chaque étape pour faciliter le débogage (token tronqué, URL, status,
// corps brut de la réponse, résultat final).
//
// Appel API externe :
//   - GET https://api.real-debrid.com/rest/1.0/streaming/mediaInfos/{fileID}
//   - En-têtes : Authorization Bearer, User-Agent "curl/8.0"
//   - Timeout : 30 secondes
//
// Paramètres :
//   - token : jeton d'authentification Real-Debrid (OAuth Bearer token).
//     Les 8 premiers caractères sont loggés pour le débogage.
//   - fileID : identifiant du fichier sur Real-Debrid.
//
// Retourne un pointeur vers [models.MediaInfoResponse], ou une erreur dans les cas suivants :
//   - échec de création de la requête HTTP,
//   - erreur réseau ou timeout,
//   - code HTTP non 2xx (l'erreur Real-Debrid est parsée via [models.RDError] si possible),
//   - réponse JSON invalide.
func GetMediaInfos(token, fileID string) (*models.MediaInfoResponse, error) {
	log.Printf("→ GetMediaInfos: token=%s..., fileID=%s", token[:8], fileID)

	endpoint := fmt.Sprintf("https://api.real-debrid.com/rest/1.0/streaming/mediaInfos/%s", fileID)
	log.Printf("→ URL: %s", endpoint)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "curl/8.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("→ StatusCode: %d", resp.StatusCode)
	log.Printf("→ Raw Body: %s", string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var rdErr models.RDError
		if err := json.Unmarshal(body, &rdErr); err != nil {
			return nil, fmt.Errorf("real-debrid http %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("real-debrid http %d: %s (code=%d)", resp.StatusCode, rdErr.Error, rdErr.ErrorCode)
	}

	var out models.MediaInfoResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("cannot parse mediaInfos response: %w; raw=%s", err, string(body))
	}

	log.Printf("✅ MediaInfos OK: filename=%s, type=%s, duration=%.2f", out.Filename, out.Type, out.Duration)

	return &out, nil
}

// BuildStreamURL construit une URL de streaming complète en remplaçant les placeholders
// du modèle d'URL (ModelURL) par les paramètres de lecture souhaités.
//
// Le ModelURL provenant de l'API Real-Debrid contient des placeholders entre accolades :
//
//	{audio}, {subtitles}, {audioCodec}, {quality}, {format}
//
// Cette fonction remplace chacun de ces placeholders par les valeurs fournies en paramètres.
// Le format est toujours fixé à "mpd" (DASH). La qualité est d'abord résolue depuis la map
// mediaInfo.AvailableQualities à l'aide de qualityKey.
//
// Paramètres :
//   - mediaInfo : métadonnées du média contenant ModelURL et AvailableQualities.
//   - audio : identifiant de la piste audio (ex. "eng1", "fre1").
//   - subtitles : identifiant de la piste de sous-titres (ex. "fre1", "none").
//   - audioCodec : codec audio en minuscules (ex. "aac", "dts").
//   - qualityKey : clé de qualité (ex. "1080P", "720P", "Original").
//
// Retourne l'URL de streaming complète, ou une erreur si qualityKey n'est pas trouvée
// dans la map AvailableQualities du média.
func BuildStreamURL(mediaInfo *models.MediaInfoResponse, audio, subtitles, audioCodec, qualityKey string) (string, error) {
	// Récupérer la valeur interne de la qualité
	qualityValue, ok := mediaInfo.AvailableQualities[qualityKey]
	if !ok {
		return "", fmt.Errorf("quality key '%s' not found in availableQualities", qualityKey)
	}

	// Remplacer les placeholders dans modelUrl
	url := mediaInfo.ModelURL
	url = strings.Replace(url, "{audio}", audio, 1)
	url = strings.Replace(url, "{subtitles}", subtitles, 1)
	url = strings.Replace(url, "{audioCodec}", audioCodec, 1)
	url = strings.Replace(url, "{quality}", qualityValue, 1)
	url = strings.Replace(url, "{format}", "mpd", 1)

	return url, nil
}
