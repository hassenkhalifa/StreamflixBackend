package handlers

import (
	"StreamflixBackend/internal/models"
	"fmt"
	"strings"
)

// GetVideoPlayer construit un VideoPlayer complet avec toutes les infos réelles
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

// buildQualitiesListWithURLs génère la liste des qualités avec l'URL pré-calculée pour chacune
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

// formatQualityLabel formate les labels qualité + résolution + bitrate
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
