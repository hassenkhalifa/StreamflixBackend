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

// buildQualitiesList convertit les qualités disponibles en slice de Quality
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

// UnrestrictAndGetMPD unrestrict un lien et retourne l'URL MPD
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

// GetMPDFromTranscode récupère l'URL MPD depuis l'endpoint transcode
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

// GetMediaInfos récupère les informations détaillées d'un média depuis Real-Debrid
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

// BuildStreamURL construit une URL de streaming à partir du modelUrl et des paramètres
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
