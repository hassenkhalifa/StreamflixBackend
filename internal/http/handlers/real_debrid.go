package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RealDebridClient est le client HTTP pour l'API Real-Debrid (v1.0).
// Real-Debrid est un service de débridage qui convertit des liens magnet/torrent
// en liens de téléchargement direct à haute vitesse. Ce client gère l'ensemble
// du cycle de vie d'un torrent : ajout du magnet, sélection des fichiers,
// attente du téléchargement, dé-restriction des liens et transcodage en streaming.
// L'authentification se fait via un jeton Bearer (APIKey).
type RealDebridClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// AddMagnetResponse représente la réponse de l'API lors de l'ajout d'un lien magnet.
// L'ID retourné est utilisé pour toutes les opérations ultérieures sur ce torrent
// (consultation des infos, sélection des fichiers, suivi du téléchargement).
type AddMagnetResponse struct {
	ID  string `json:"id"`
	URI string `json:"uri"`
}

// TorrentInfo contient les informations détaillées d'un torrent sur Real-Debrid.
// Le champ Status reflète l'état du torrent dans le pipeline de traitement :
//   - "magnet_error"             : erreur lors de la résolution du magnet
//   - "waiting_files_selection"  : en attente de la sélection des fichiers
//   - "downloading"              : téléchargement en cours (voir Progress)
//   - "downloaded"               : téléchargement terminé, liens disponibles
//
// Le champ Links contient les URLs des fichiers sélectionnés une fois téléchargés.
type TorrentInfo struct {
	ID       string        `json:"id"`
	Filename string        `json:"filename"`
	Hash     string        `json:"hash"`
	Bytes    int64         `json:"bytes"`
	Status   string        `json:"status"` // "magnet_error", "waiting_files_selection", "downloading", "downloaded"
	Progress float64       `json:"progress"`
	Links    []string      `json:"links"`
	Files    []TorrentFile `json:"files"`
}

// TorrentFile représente un fichier individuel à l'intérieur d'un torrent.
// Le champ Selected indique si le fichier a été sélectionné pour le téléchargement
// (0 = non sélectionné, 1 = sélectionné). Le champ Path contient le chemin
// relatif du fichier dans l'arborescence du torrent.
type TorrentFile struct {
	ID       int    `json:"id"`
	Path     string `json:"path"`
	Bytes    int64  `json:"bytes"`
	Selected int    `json:"selected"` // 0 ou 1
}

// UnrestrictResponse représente la réponse de l'API de dé-restriction de lien.
// Elle contient le lien de téléchargement direct (Download), les informations
// sur le fichier (nom, taille, hébergeur), et éventuellement des alternatives
// de streaming (HLS, DASH) avec différentes qualités disponibles.
// Le champ Streamable vaut 1 si le fichier peut être lu en streaming.
type UnrestrictResponse struct {
	ID          string            `json:"id"`
	Filename    string            `json:"filename"`
	Filesize    int64             `json:"filesize"`
	Link        string            `json:"link"`
	Host        string            `json:"host"`
	Download    string            `json:"download"`
	Streamable  int               `json:"streamable"` // 0 ou 1
	Quality     map[string]string `json:"quality,omitempty"`
	Alternative []Alternative     `json:"alternative,omitempty"`
}

// Alternative représente une variante de streaming disponible pour un fichier
// dé-restreint. Chaque alternative propose un format de streaming (HLS ou DASH)
// avec ses qualités disponibles et un lien de téléchargement spécifique.
type Alternative struct {
	Type     string            `json:"type"` // "hls", "dash"
	Quality  map[string]string `json:"quality"`
	Filename string            `json:"filename"`
	Download string            `json:"download"`
}

// StreamingTranscode contient les URLs de transcodage fournies par Real-Debrid
// pour la lecture en streaming. Quatre formats sont disponibles :
//   - Apple : transcodage HLS compatible avec les appareils Apple
//   - Dash  : transcodage MPEG-DASH pour les lecteurs web adaptatifs
//   - Livemp4 : transcodage MP4 progressif pour la lecture directe
//   - H264WebM : transcodage H.264 au format WebM
//
// Chaque format expose un champ Full contenant l'URL de la qualité maximale.
type StreamingTranscode struct {
	Apple struct {
		Full string `json:"full"`
	} `json:"apple"`
	Dash struct {
		Full string `json:"full"`
	} `json:"dash"`
	Livemp4 struct {
		Full string `json:"full"`
	} `json:"livemp4"`
	H264WebM struct {
		Full string `json:"full"`
	} `json:"h264_webm"`
}

// NewRealDebridClient crée un nouveau client Real-Debrid configuré avec la clé API
// fournie. Le client utilise l'URL de base de l'API REST v1.0 et un timeout HTTP
// de 30 secondes par défaut.
func NewRealDebridClient(apiKey string) *RealDebridClient {
	return &RealDebridClient{
		APIKey:  apiKey,
		BaseURL: "https://api.real-debrid.com/rest/1.0",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest est la méthode utilitaire interne pour exécuter des requêtes HTTP
// vers l'API Real-Debrid. Elle construit l'URL complète à partir de l'endpoint,
// ajoute l'en-tête d'authentification Bearer, et configure le Content-Type
// en "application/x-www-form-urlencoded" pour les méthodes POST et PUT.
func (r *RealDebridClient) doRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	url := r.BaseURL + endpoint

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.APIKey)

	if method == "POST" || method == "PUT" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return r.Client.Do(req)
}

// AddMagnet soumet un lien magnet à Real-Debrid pour initier le processus de
// téléchargement. C'est la première étape du workflow torrent.
// L'API retourne un identifiant unique (ID) et une URI qui permettent de suivre
// et gérer le torrent ajouté. Attend un code HTTP 201 (Created) en réponse.
func (r *RealDebridClient) AddMagnet(magnetLink string) (*AddMagnetResponse, error) {
	data := url.Values{}
	data.Set("magnet", magnetLink)

	resp, err := r.doRequest("POST", "/torrents/addMagnet", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to add magnet: %s", string(body))
	}

	var result AddMagnetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTorrentInfo récupère les informations complètes d'un torrent identifié par
// son ID Real-Debrid. Les informations incluent le statut du téléchargement,
// la progression, la liste des fichiers contenus et les liens de téléchargement
// disponibles. Cette méthode est utilisée à la fois pour inspecter les fichiers
// avant sélection et pour vérifier l'état d'avancement du téléchargement.
func (r *RealDebridClient) GetTorrentInfo(torrentID string) (*TorrentInfo, error) {
	resp, err := r.doRequest("GET", "/torrents/info/"+torrentID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get torrent info: %s", string(body))
	}

	var info TorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

// SelectFiles indique à Real-Debrid quels fichiers du torrent doivent être
// téléchargés. Si fileIDs est vide, tous les fichiers sont sélectionnés ("all").
// Sinon, seuls les fichiers dont les identifiants sont fournis seront téléchargés.
// Les identifiants sont transmis sous forme de chaîne séparée par des virgules
// (ex: "1,3,5"). Cette étape est obligatoire avant que le téléchargement ne démarre.
func (r *RealDebridClient) SelectFiles(torrentID string, fileIDs []int) error {
	data := url.Values{}

	if len(fileIDs) == 0 {
		data.Set("files", "all")
	} else {
		// Convertir []int en string "1,2,3"
		fileIDsStr := make([]string, len(fileIDs))
		for i, id := range fileIDs {
			fileIDsStr[i] = fmt.Sprintf("%d", id)
		}
		data.Set("files", strings.Join(fileIDsStr, ","))
	}

	resp, err := r.doRequest("POST", "/torrents/selectFiles/"+torrentID, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to select files: %s", string(body))
	}

	return nil
}

// WaitForDownload attend que le torrent passe au statut "downloaded" en interrogeant
// périodiquement l'API (toutes les 2 secondes). Pour les torrents déjà en cache
// Real-Debrid, le retour est quasi-instantané. Le paramètre maxWait définit le
// délai maximum d'attente avant de retourner une erreur de timeout.
// Retourne une erreur immédiate si le torrent est en erreur (magnet_error, error,
// virus, dead) ou si les fichiers n'ont pas encore été sélectionnés.
func (r *RealDebridClient) WaitForDownload(torrentID string, maxWait time.Duration) (*TorrentInfo, error) {
	timeout := time.After(maxWait)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for download")
		case <-ticker.C:
			info, err := r.GetTorrentInfo(torrentID)
			if err != nil {
				return nil, err
			}

			// Vérifier le statut
			switch info.Status {
			case "downloaded":
				return info, nil
			case "magnet_error", "error", "virus", "dead":
				return nil, fmt.Errorf("torrent error: %s", info.Status)
			case "downloading":
				fmt.Printf("Downloading... %.2f%%\n", info.Progress)
			case "waiting_files_selection":
				return nil, fmt.Errorf("files not selected")
			}
		}
	}
}

// UnrestrictLink dé-restreint un lien hébergeur pour obtenir un lien de
// téléchargement direct à haute vitesse. Cette opération transforme un lien
// Real-Debrid interne en un lien directement accessible, avec les informations
// sur le fichier (nom, taille, possibilité de streaming) et éventuellement
// des alternatives de streaming en différentes qualités.
func (r *RealDebridClient) UnrestrictLink(link string) (*UnrestrictResponse, error) {
	data := url.Values{}
	data.Set("link", link)

	resp, err := r.doRequest("POST", "/unrestrict/link", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to unrestrict link: %s", string(body))
	}

	var result UnrestrictResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStreamingTranscode récupère les URLs de transcodage en streaming pour un
// fichier identifié par son ID Real-Debrid. L'API fournit plusieurs formats :
// Apple (HLS), DASH (MPEG-DASH), Livemp4 (MP4 progressif) et H264WebM.
// Ces URLs permettent la lecture adaptative directement dans un lecteur web
// sans nécessiter de téléchargement préalable du fichier complet.
func (r *RealDebridClient) GetStreamingTranscode(fileID string) (*StreamingTranscode, error) {
	resp, err := r.doRequest("GET", "/streaming/transcode/"+fileID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get streaming transcode: %s", string(body))
	}

	var result StreamingTranscode
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CheckInstantAvailability vérifie si un torrent identifié par son hash est déjà
// disponible dans le cache de Real-Debrid. Un torrent en cache peut être converti
// en lien de téléchargement direct instantanément, sans attendre le téléchargement
// complet. La vérification se fait en recherchant la présence du hash (en minuscules)
// dans la réponse JSON de l'API. Retourne true si le torrent est en cache.
func (r *RealDebridClient) CheckInstantAvailability(hash string) (bool, error) {
	resp, err := r.doRequest("GET", "/torrents/instantAvailability/"+hash, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	// Si le hash existe dans la réponse, c'est cached
	if _, exists := result[strings.ToLower(hash)]; exists {
		return true, nil
	}

	return false, nil
}
