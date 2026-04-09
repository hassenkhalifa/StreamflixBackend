// services/torrentio.go
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// TorrentioClient est le client pour l'API Torrentio (addon Stremio).
// Torrentio agrège les résultats de plusieurs indexeurs de torrents et permet
// de découvrir les sources de streaming disponibles pour un contenu donné.
// La configuration inclut :
//   - BaseURL       : URL de base de l'instance Torrentio
//   - RealDebridKey : clé API Real-Debrid pour le filtrage des résultats par cache
//   - Providers     : liste des indexeurs activés (yts, eztv, 1337x, rarbg, etc.)
//   - Sort          : critère de tri des résultats (ex: "qualitysize")
//   - QualityFilter : qualités à exclure des résultats (ex: "scr", "cam")
type TorrentioClient struct {
	BaseURL       string
	RealDebridKey string
	Providers     []string
	Sort          string
	QualityFilter []string
}

// TorrentioStream représente un flux torrent individuel retourné par l'API Torrentio.
// Chaque stream contient les métadonnées du torrent (nom, titre), son identifiant
// unique (InfoHash), l'index du fichier vidéo dans le torrent (FileIdx),
// une URL optionnelle pour le débridage, et la liste des sources/trackers associés.
type TorrentioStream struct {
	Name     string   `json:"name"`
	Title    string   `json:"title"`
	InfoHash string   `json:"infoHash"`
	FileIdx  int      `json:"fileIdx"`
	URL      string   `json:"url,omitempty"` // Pour debrid
	Sources  []string `json:"sources,omitempty"`
}

// TorrentioResponse est la structure de réponse de l'API Torrentio.
// Elle encapsule la liste des streams disponibles pour un contenu donné.
type TorrentioResponse struct {
	Streams []TorrentioStream `json:"streams"`
}

// NewTorrentioClient crée un nouveau client Torrentio avec une configuration par
// défaut. Les indexeurs activés sont yts, eztv, 1337x, rarbg et thepiratebay.
// Le tri est par qualité et taille ("qualitysize"), et les qualités "scr" (screener)
// et "cam" (caméra) sont exclues par défaut. Le paramètre rdKey active le filtrage
// par disponibilité en cache Real-Debrid dans les résultats Torrentio.
func NewTorrentioClient(rdKey string) *TorrentioClient {
	return &TorrentioClient{
		BaseURL:       "https://torrentio.strem.fun",
		RealDebridKey: rdKey,
		Providers:     []string{"yts", "eztv", "1337x", "rarbg", "thepiratebay"},
		Sort:          "qualitysize",
		QualityFilter: []string{"scr", "cam"},
	}
}

// buildOptions construit la chaîne d'options de configuration pour l'URL Torrentio.
// Les options sont séparées par le caractère pipe (|) et incluent :
//   - providers     : indexeurs de torrents à interroger (séparés par des virgules)
//   - sort          : critère de tri des résultats
//   - qualityfilter : qualités à exclure (séparées par des virgules)
//   - realdebrid    : clé API pour le filtrage par cache Real-Debrid
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

// GetMovieStreams recherche les streams torrent disponibles pour un film identifié
// par son identifiant IMDb (ex: "tt1234567"). L'URL construite suit le format
// Stremio addon : {baseURL}/{options}/stream/movie/{imdbID}.json
// Les résultats sont filtrés et triés selon la configuration du client.
// Retourne la liste des TorrentioStream trouvés ou une erreur.
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

// GetSeriesStreams recherche les streams torrent disponibles pour un épisode
// spécifique d'une série. L'identifiant IMDb de la série est complété par le
// numéro de saison et d'épisode dans l'URL : {baseURL}/{options}/stream/series/{imdbID}:{season}:{episode}.json
// Retourne la liste des TorrentioStream correspondant à cet épisode.
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

// ExtractMagnets construit les liens magnet à partir d'une liste de TorrentioStream.
// Pour chaque stream possédant un InfoHash, un lien magnet est construit au format
// "magnet:?xt=urn:btih:{infoHash}". Les trackers présents dans le champ Sources
// (préfixés par "tracker:") sont ajoutés en tant que paramètres &tr= au lien magnet
// pour faciliter la découverte des pairs. Retourne la liste des liens magnet.
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
