package handlers

import (
	"StreamflixBackend/internal/models"
	"fmt"
	"strings"
)

// GetVideoPlayer construit un objet [models.VideoPlayer] complet à partir des informations
// réelles du média hébergé sur Real-Debrid.
//
// Cette fonction orchestre plusieurs étapes :
//  1. Récupération des métadonnées du média via [GetMediaInfos] (appel API Real-Debrid).
//  2. Détermination des pistes par défaut : première piste audio trouvée, pas de sous-titres,
//     et qualité "UHD" si disponible, sinon "Original".
//  3. Construction de l'URL de streaming par défaut via [BuildStreamURL].
//  4. Génération de la liste des qualités disponibles avec leurs URLs pré-calculées
//     via [buildQualitiesListWithURLs].
//  5. Construction des listes de pistes audio et de sous-titres disponibles
//     via [buildAudioList] et [buildSubtitlesList].
//  6. Sélection de l'image poster (backdrop en priorité, sinon poster vertical).
//
// Paramètres :
//   - token : jeton d'authentification Real-Debrid (OAuth Bearer token).
//   - fileID : identifiant du fichier sur Real-Debrid, obtenu après unrestrict.
//
// Retourne un [models.VideoPlayer] entièrement configuré pour le frontend, ou une erreur
// si la récupération des infos média ou la construction de l'URL échoue.
func GetVideoPlayer(token, fileID string) (models.VideoPlayer, error) {
	// 1. Récupérer les infos média depuis Real-Debrid
	mediaInfo, err := GetMediaInfos(token, fileID)
	if err != nil {
		return models.VideoPlayer{}, fmt.Errorf("cannot get media infos: %w", err)
	}

	// 2. Déterminer les pistes par défaut
	defaultAudio := "eng1"
	defaultCodec := "aac"
	if len(mediaInfo.Details.Audio) > 0 {
		for key, track := range mediaInfo.Details.Audio {
			defaultAudio = key
			defaultCodec = strings.ToLower(track.Codec)
			break
		}
	}
	defaultSubtitles := "none"

	defaultQuality := "Original"
	if _, ok := mediaInfo.AvailableQualities["UHD"]; ok {
		defaultQuality = "UHD"
	}
	// 3. Construire l'URL du stream par défaut (Qualité Originale)
	streamURL, err := BuildStreamURL(mediaInfo, defaultAudio, defaultSubtitles, defaultCodec, defaultQuality)
	if err != nil {
		return models.VideoPlayer{}, fmt.Errorf("cannot build default stream url: %w", err)
	}

	// 4. Construire la liste des qualités avec leurs URLs respectives
	availableQualities := buildQualitiesListWithURLs(mediaInfo, defaultAudio, defaultSubtitles, defaultCodec)

	// 5. Construire les listes d'audio et sous-titres
	availableAudio := buildAudioList(mediaInfo.Details.Audio)
	availableSubtitles := buildSubtitlesList(mediaInfo.Details.Subtitles)

	// 6. Gérer le poster (Backdrop est souvent mieux pour le lecteur que le Poster vertical)
	posterURL := mediaInfo.BackdropPath
	if posterURL == "" {
		posterURL = mediaInfo.PosterPath
	}

	videoPlayer := models.VideoPlayer{
		Src:                streamURL,
		Title:              mediaInfo.Filename, // ✅ Vrai nom du fichier/film
		Poster:             posterURL,          // ✅ Vrai poster/backdrop
		Quality:            defaultQuality,
		AudioFormat:        defaultCodec,
		Autoplay:           false,
		ModelURL:           mediaInfo.ModelURL,
		AvailableQualities: availableQualities,
		AvailableAudio:     availableAudio,
		AvailableSubtitles: availableSubtitles,
	}

	return videoPlayer, nil
}

// buildQualitiesListWithURLs génère la liste ordonnée des qualités vidéo disponibles,
// chacune associée à son URL de streaming pré-calculée.
//
// Les qualités sont triées par ordre décroissant de résolution, de "Original" (qualité native)
// jusqu'à "360P". Seules les qualités présentes dans mediaInfo.AvailableQualities sont incluses.
// Pour chaque qualité, l'URL est construite via [BuildStreamURL] avec les paramètres audio,
// sous-titres et codec fournis. Si la construction de l'URL échoue pour une qualité donnée,
// celle-ci est silencieusement ignorée.
//
// Le label affiché est formaté via [formatQualityLabel] (ex. "FHD (1080p) - Standard Bitrate").
//
// Paramètres :
//   - mediaInfo : métadonnées du média issues de l'API Real-Debrid.
//   - audio : identifiant de la piste audio sélectionnée (ex. "eng1").
//   - subs : identifiant de la piste de sous-titres sélectionnée (ex. "none").
//   - codec : codec audio souhaité en minuscules (ex. "aac").
//
// Retourne un slice de [models.Quality] trié par résolution décroissante.
func buildQualitiesListWithURLs(mediaInfo *models.MediaInfoResponse, audio, subs, codec string) []models.Quality {
	result := []models.Quality{}

	order := []string{
		"Original",
		"UHD",
		"1440P",
		"1080P High",
		"1080P",
		"720P High",
		"720P",
		"480P High",
		"480P",
		"360P",
	}

	for _, key := range order {
		if value, ok := mediaInfo.AvailableQualities[key]; ok {
			qualityURL, err := BuildStreamURL(mediaInfo, audio, subs, codec, key)
			if err != nil {
				continue
			}

			label := formatQualityLabel(key)

			result = append(result, models.Quality{
				Label: label,
				Value: value,
				URL:   qualityURL,
			})
		}
	}

	return result
}

// formatQualityLabel convertit une clé de qualité interne en label lisible pour l'utilisateur.
//
// La correspondance est définie dans un dictionnaire statique, par exemple :
//   - "Original" ou "UHD" -> "UHD (2160p)"
//   - "1080P High" -> "FHD (1080p) - High Bitrate"
//   - "720P" -> "HD (720p) - Standard Bitrate"
//
// Si la clé n'est pas trouvée dans le dictionnaire, un fallback générique est utilisé
// en remplaçant simplement "P" par "p" dans la clé (ex. "540P" -> "540p").
//
// Paramètre :
//   - key : clé de qualité telle que retournée par l'API Real-Debrid (ex. "1080P", "720P High").
//
// Retourne le label formaté correspondant.
func formatQualityLabel(key string) string {

	labelMap := map[string]string{
		"Original": "UHD (2160p)",
		"UHD":      "UHD (2160p)",
		"1440P":    "QHD (1440p)",

		"1080P High": "FHD (1080p) - High Bitrate",
		"1080P":      "FHD (1080p) - Standard Bitrate",

		"720P High": "HD (720p) - High Bitrate",
		"720P":      "HD (720p) - Standard Bitrate",

		"480P High": "SD (480p) - High Bitrate",
		"480P":      "SD (480p) - Standard Bitrate",

		"360P": "LD (360p)",
	}

	// Si la clé existe, on retourne le label
	if label, ok := labelMap[key]; ok {
		return label
	}

	// fallback générique
	return strings.Replace(key, "P", "p", 1)
}
