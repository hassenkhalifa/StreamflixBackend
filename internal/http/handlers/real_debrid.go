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

type RealDebridClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// Structures de réponse

type AddMagnetResponse struct {
	ID  string `json:"id"`
	URI string `json:"uri"`
}

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

type TorrentFile struct {
	ID       int    `json:"id"`
	Path     string `json:"path"`
	Bytes    int64  `json:"bytes"`
	Selected int    `json:"selected"` // 0 ou 1
}

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

type Alternative struct {
	Type     string            `json:"type"` // "hls", "dash"
	Quality  map[string]string `json:"quality"`
	Filename string            `json:"filename"`
	Download string            `json:"download"`
}

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

func NewRealDebridClient(apiKey string) *RealDebridClient {
	return &RealDebridClient{
		APIKey:  apiKey,
		BaseURL: "https://api.real-debrid.com/rest/1.0",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Helper pour faire des requêtes
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

// 1. Ajouter un magnet
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

// 2. Récupérer les infos du torrent
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

// 3. Sélectionner les fichiers (tous ou spécifiques)
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

// 4. Attendre que le torrent soit téléchargé
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

// 5. Unrestrict un lien pour obtenir le stream
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

// 6. Récupérer les liens de streaming (HLS, DASH, etc.)
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

// Vérifier la disponibilité instantanée (cache)
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
