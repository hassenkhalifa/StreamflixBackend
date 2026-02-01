// services/torrentio.go
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type TorrentioClient struct {
	BaseURL       string
	RealDebridKey string
	Providers     []string
	Sort          string
	QualityFilter []string
}

type TorrentioStream struct {
	Name     string   `json:"name"`
	Title    string   `json:"title"`
	InfoHash string   `json:"infoHash"`
	FileIdx  int      `json:"fileIdx"`
	URL      string   `json:"url,omitempty"` // Pour debrid
	Sources  []string `json:"sources,omitempty"`
}

type TorrentioResponse struct {
	Streams []TorrentioStream `json:"streams"`
}

func NewTorrentioClient(rdKey string) *TorrentioClient {
	return &TorrentioClient{
		BaseURL:       "https://torrentio.strem.fun",
		RealDebridKey: rdKey,
		Providers:     []string{"yts", "eztv", "1337x", "rarbg", "thepiratebay"},
		Sort:          "qualitysize",
		QualityFilter: []string{"scr", "cam"},
	}
}

// Construire les options de configuration
func (t *TorrentioClient) buildOptions() string {
	var opts []string

	// Providers
	if len(t.Providers) > 0 {
		opts = append(opts, fmt.Sprintf("providers=%s",
			strings.Join(t.Providers, ",")))
	}

	// Sort
	if t.Sort != "" {
		opts = append(opts, fmt.Sprintf("sort=%s", t.Sort))
	}

	// Quality filter
	if len(t.QualityFilter) > 0 {
		opts = append(opts, fmt.Sprintf("qualityfilter=%s",
			strings.Join(t.QualityFilter, ",")))
	}

	// Real-Debrid
	if t.RealDebridKey != "" {
		opts = append(opts, fmt.Sprintf("realdebrid=%s", t.RealDebridKey))
	}

	return strings.Join(opts, "|")
}

// Rechercher des streams pour un film
func (t *TorrentioClient) GetMovieStreams(imdbID string) ([]TorrentioStream, error) {
	options := t.buildOptions()
	url := fmt.Sprintf("%s/%s/stream/movie/%s.json",
		t.BaseURL, options, imdbID)

	fmt.Println("Torrentio URL:", url) // DEBUG

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode) // DEBUG

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Body:", string(body)) // DEBUG

	var result TorrentioResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode torrentio response: %w", err)
	}

	return result.Streams, nil
}

// Rechercher des streams pour une série
func (t *TorrentioClient) GetSeriesStreams(imdbID string, season, episode int) ([]TorrentioStream, error) {
	options := t.buildOptions()
	url := fmt.Sprintf("%s/%s/stream/series/%s:%d:%d.json",
		t.BaseURL, options, imdbID, season, episode)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TorrentioResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Streams, nil
}

// Extraire les magnets des streams
func (t *TorrentioClient) ExtractMagnets(streams []TorrentioStream) []string {
	var magnets []string

	for _, stream := range streams {
		if stream.InfoHash != "" {
			// Construire le magnet link
			magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s", stream.InfoHash)

			// Ajouter les trackers si disponibles
			if len(stream.Sources) > 0 {
				for _, tracker := range stream.Sources {
					if strings.HasPrefix(tracker, "tracker:") {
						trackerURL := strings.TrimPrefix(tracker, "tracker:")
						magnet += fmt.Sprintf("&tr=%s", trackerURL)
					}
				}
			}

			magnets = append(magnets, magnet)
		}
	}

	return magnets
}
