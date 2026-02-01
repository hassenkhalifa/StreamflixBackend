// services/streaming.go
package handlers

import (
	"fmt"
	"strings"
	"time"
)

type StreamingService struct {
	TorrentioClient  *TorrentioClient
	RealDebridClient *RealDebridClient
}

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

func NewStreamingService(torrentioClient *TorrentioClient, rdClient *RealDebridClient) *StreamingService {
	return &StreamingService{
		TorrentioClient:  torrentioClient,
		RealDebridClient: rdClient,
	}
}

// Workflow complet : Magnet → Real-Debrid → Stream
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

// Obtenir les URLs de streaming pour un torrent cached
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

// Sélectionner le meilleur fichier vidéo (le plus gros)
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

// Helpers
func extractQuality(name string) string {
	qualities := []string{"2160p", "1080p", "720p", "480p"}
	for _, q := range qualities {
		if strings.Contains(name, q) {
			return q
		}
	}
	return "Unknown"
}

func extractSize(name string) string {
	// Extraire "💾 2.1 GB" du nom
	parts := strings.Split(name, "💾")
	if len(parts) > 1 {
		return strings.TrimSpace(strings.Split(parts[1], "👤")[0])
	}
	return ""
}

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

func extractFileIDFromURL(downloadURL string) string {
	// Extraire l'ID depuis une URL comme:
	// https://download.real-debrid.com/d/ABCD1234/filename.mkv
	parts := strings.Split(downloadURL, "/d/")
	if len(parts) > 1 {
		return strings.Split(parts[1], "/")[0]
	}
	return ""
}
