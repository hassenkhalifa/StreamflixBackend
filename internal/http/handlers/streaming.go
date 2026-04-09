// services/streaming.go
package handlers

import (
	"fmt"
	"strings"
	"time"
)

// StreamingService orchestre le workflow complet de streaming :
// découverte de torrents via Torrentio, débridage via Real-Debrid,
// et récupération des URLs de lecture (téléchargement direct, DASH, HLS).
// Il relie le TorrentioClient (recherche de sources) au RealDebridClient
// (conversion magnet → lien de streaming haute vitesse).
type StreamingService struct {
	TorrentioClient  *TorrentioClient
	RealDebridClient *RealDebridClient
}

// StreamResult représente un flux de streaming résolu pour un contenu donné.
// Il contient les métadonnées du torrent (titre, qualité, taille), l'état
// de cache Real-Debrid, ainsi que les différentes URLs de lecture disponibles :
//   - StreamURL : lien de téléchargement direct (non restreint)
//   - DashURL   : lien de transcodage MPEG-DASH
//   - HlsURL    : lien de transcodage HLS (Apple)
//   - Magnet    : lien magnet original avec les trackers
type StreamResult struct {
	Title       string `json:"title"`
	Quality     string `json:"quality"`
	Size        string `json:"size"`
	IsCached    bool   `json:"is_cached"`
	StreamURL   string `json:"stream_url"`
	DashURL     string `json:"dash_url,omitempty"`
	HlsURL      string `json:"hls_url,omitempty"`
	DownloadURL string `json:"download_url"`
	Magnet      string `json:"magnet"`
}

// NewStreamingService crée une nouvelle instance de StreamingService
// en injectant les clients Torrentio et Real-Debrid nécessaires
// au workflow de résolution de streams.
func NewStreamingService(torrentioClient *TorrentioClient, rdClient *RealDebridClient) *StreamingService {
	return &StreamingService{
		TorrentioClient:  torrentioClient,
		RealDebridClient: rdClient,
	}
}

// GetStreamForMovie exécute le workflow complet magnet → Real-Debrid → stream
// pour un film identifié par son identifiant IMDb.
//
// Étapes du workflow :
//  1. Récupération des streams disponibles depuis Torrentio
//  2. Vérification du cache Real-Debrid pour chaque torrent
//  3. Pour les torrents en cache, résolution immédiate des URLs de lecture
//
// Retourne une liste de StreamResult triée avec les métadonnées et les URLs.
func (s *StreamingService) GetStreamForMovie(imdbID string) ([]StreamResult, error) {
	// 1. Récupérer les streams depuis Torrentio
	streams, err := s.TorrentioClient.GetMovieStreams(imdbID)
	if err != nil {
		return nil, fmt.Errorf("failed to get streams from Torrentio: %w", err)
	}

	var results []StreamResult

	for _, stream := range streams {
		// 2. Vérifier si c'est cached sur Real-Debrid
		isCached, _ := s.RealDebridClient.CheckInstantAvailability(stream.InfoHash)

		result := StreamResult{
			Title:    stream.Title,
			Quality:  extractQuality(stream.Name),
			Size:     extractSize(stream.Name),
			IsCached: isCached,
			Magnet:   buildMagnetLink(stream.InfoHash, stream.Sources),
		}

		// Si c'est cached, on peut obtenir le stream immédiatement
		if isCached {
			streamURL, dashURL, hlsURL, err := s.getStreamURLs(stream.InfoHash)
			if err == nil {
				result.StreamURL = streamURL
				result.DashURL = dashURL
				result.HlsURL = hlsURL
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// getStreamURLs résout les URLs de streaming (téléchargement, DASH, HLS) pour un
// torrent identifié par son infoHash. Cette méthode exécute le pipeline complet
// Real-Debrid en 7 étapes :
//  1. Ajout du magnet à Real-Debrid
//  2. Récupération des informations du torrent (liste des fichiers)
//  3. Sélection du meilleur fichier vidéo (le plus volumineux)
//  4. Demande de sélection des fichiers auprès de l'API
//  5. Attente du téléchargement (instantané si le torrent est en cache)
//  6. Dé-restriction du lien pour obtenir l'URL de téléchargement direct
//  7. Récupération des variantes de transcodage (DASH et HLS)
//
// Retourne dans l'ordre : URL de téléchargement, URL DASH, URL HLS, ou une erreur.
func (s *StreamingService) getStreamURLs(infoHash string) (string, string, string, error) {
	// Construire le magnet
	magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s", infoHash)

	// 1. Ajouter le magnet à Real-Debrid
	addResp, err := s.RealDebridClient.AddMagnet(magnet)
	if err != nil {
		return "", "", "", err
	}

	// 2. Récupérer les infos du torrent
	info, err := s.RealDebridClient.GetTorrentInfo(addResp.ID)
	if err != nil {
		return "", "", "", err
	}

	// 3. Sélectionner le plus gros fichier vidéo
	videoFileID := s.selectBestVideoFile(info.Files)
	if videoFileID == -1 {
		return "", "", "", fmt.Errorf("no video file found")
	}

	// 4. Sélectionner les fichiers
	err = s.RealDebridClient.SelectFiles(addResp.ID, []int{videoFileID})
	if err != nil {
		return "", "", "", err
	}

	// 5. Attendre le téléchargement (si cached, c'est instantané)
	info, err = s.RealDebridClient.WaitForDownload(addResp.ID, 30*time.Second)
	if err != nil {
		return "", "", "", err
	}

	// 6. Unrestrict le premier lien
	if len(info.Links) == 0 {
		return "", "", "", fmt.Errorf("no links available")
	}

	unrestrictResp, err := s.RealDebridClient.UnrestrictLink(info.Links[0])
	if err != nil {
		return "", "", "", err
	}

	// 7. Récupérer les liens de streaming (DASH, HLS)
	var dashURL, hlsURL string

	// Extraire l'ID du fichier depuis l'URL de téléchargement
	fileID := extractFileIDFromURL(unrestrictResp.Download)
	if fileID != "" {
		transcode, err := s.RealDebridClient.GetStreamingTranscode(fileID)
		if err == nil {
			dashURL = transcode.Dash.Full
			hlsURL = transcode.Apple.Full
		}
	}

	return unrestrictResp.Download, dashURL, hlsURL, nil
}

// selectBestVideoFile parcourt la liste des fichiers d'un torrent et retourne
// l'identifiant du fichier vidéo le plus volumineux. La détection se fait par
// extension de fichier parmi les formats courants : .mkv, .mp4, .avi, .mov,
// .wmv, .flv, .webm. Retourne -1 si aucun fichier vidéo n'est trouvé.
func (s *StreamingService) selectBestVideoFile(files []TorrentFile) int {
	var bestFile TorrentFile
	bestFileID := -1

	videoExtensions := []string{".mkv", ".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm"}

	for _, file := range files {
		// Vérifier si c'est un fichier vidéo
		isVideo := false
		for _, ext := range videoExtensions {
			if strings.HasSuffix(strings.ToLower(file.Path), ext) {
				isVideo = true
				break
			}
		}

		if isVideo && file.Bytes > bestFile.Bytes {
			bestFile = file
			bestFileID = file.ID
		}
	}

	return bestFileID
}

// extractQuality extrait la résolution vidéo à partir du nom d'un stream Torrentio.
// Recherche les résolutions courantes dans l'ordre décroissant : 2160p, 1080p,
// 720p, 480p. Retourne "Unknown" si aucune résolution n'est détectée.
func extractQuality(name string) string {
	qualities := []string{"2160p", "1080p", "720p", "480p"}
	for _, q := range qualities {
		if strings.Contains(name, q) {
			return q
		}
	}
	return "Unknown"
}

// extractSize extrait la taille du fichier depuis le nom d'un stream Torrentio.
// Le format attendu utilise l'emoji disquette (ex: "... 💾 2.1 GB 👤 ...").
// La taille est extraite entre le séparateur 💾 et le séparateur 👤.
// Retourne une chaîne vide si le format n'est pas reconnu.
func extractSize(name string) string {
	// Extraire "💾 2.1 GB" du nom
	parts := strings.Split(name, "💾")
	if len(parts) > 1 {
		return strings.TrimSpace(strings.Split(parts[1], "👤")[0])
	}
	return ""
}

// buildMagnetLink construit un lien magnet complet à partir d'un infoHash et
// d'une liste de sources Torrentio. Les sources préfixées par "tracker:" sont
// ajoutées en tant que paramètres tracker (&tr=) dans l'URI magnet.
func buildMagnetLink(infoHash string, sources []string) string {
	magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s", infoHash)

	for _, source := range sources {
		if strings.HasPrefix(source, "tracker:") {
			tracker := strings.TrimPrefix(source, "tracker:")
			magnet += fmt.Sprintf("&tr=%s", tracker)
		}
	}

	return magnet
}

// extractFileIDFromURL extrait l'identifiant de fichier Real-Debrid depuis une URL
// de téléchargement. Le format attendu est :
// https://download.real-debrid.com/d/{ID}/filename.ext
// L'identifiant est nécessaire pour appeler l'API de transcodage streaming.
// Retourne une chaîne vide si le format de l'URL n'est pas reconnu.
func extractFileIDFromURL(downloadURL string) string {
	// Extraire l'ID depuis une URL comme:
	// https://download.real-debrid.com/d/ABCD1234/filename.mkv
	parts := strings.Split(downloadURL, "/d/")
	if len(parts) > 1 {
		return strings.Split(parts[1], "/")[0]
	}
	return ""
}
