package handlers

import (
	"StreamflixBackend/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func ConvertToDASHLive(inputURL, outDir string) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}

	manifestPath := filepath.Join(outDir, "manifest.mpd")
	ctx := context.Background()

	args := []string{
		"-y",
		"-i", inputURL, // ❌ RETIRE -re pour ne pas limiter la vitesse

		"-map", "0:v:0",
		"-map", "0:a:0",

		// Vidéo
		"-c:v", "libx264",
		"-preset", "veryfast", // Plus rapide mais meilleure qualité qu'ultrafast
		"-profile:v", "main",
		"-level", "4.0",
		"-crf", "23",
		"-pix_fmt", "yuv420p",
		"-g", "48",
		"-sc_threshold", "0",

		// Audio
		"-c:a", "aac",
		"-ac", "2",
		"-b:a", "128k",

		// DASH avec tous les segments gardés
		"-f", "dash",
		"-seg_duration", "4", // Segments de 4s
		"-use_template", "1",
		"-use_timeline", "1",
		"-adaptation_sets", "id=0,streams=v id=1,streams=a",
		// ❌ RETIRE window_size, streaming, ldash pour garder tous les segments

		manifestPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// ✅ Lance en arrière-plan
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("ffmpeg start error: %w", err)
	}

	// ✅ Attends que le manifest soit créé
	for i := 0; i < 30; i++ {
		if _, err := os.Stat(manifestPath); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// ✅ Continue en arrière-plan
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("FFmpeg finished with error: %v | %s", err, stderr.String())
		} else {
			log.Println("FFmpeg finished successfully")
		}
	}()

	return manifestPath, nil
}

func ConvertToDASHVOD(inputURL, outDir string) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}

	manifestPath := filepath.Join(outDir, "manifest.mpd")
	ctx := context.Background()

	args := []string{
		"-y",
		"-i", inputURL,

		// Map les flux vidéo et audio
		"-map", "0:v:0",
		"-map", "0:a:0",

		// ✅ Encodage vidéo H.264
		"-c:v", "libx264",
		"-preset", "veryfast", // Bon compromis qualité/vitesse
		"-profile:v", "main",
		"-level", "4.0",
		"-crf", "23", // Qualité constante (18-28, plus bas = meilleure qualité)
		"-pix_fmt", "yuv420p",
		"-g", "48", // GOP de 2 secondes à 24fps
		"-keyint_min", "48",
		"-sc_threshold", "0", // Désactive la détection de scène

		// ✅ Encodage audio AAC
		"-c:a", "aac",
		"-ac", "2", // Stéréo
		"-b:a", "128k",
		"-ar", "48000", // Sample rate

		// ✅ DASH VOD (pas de streaming live)
		"-f", "dash",
		"-seg_duration", "4", // Segments de 4 secondes
		"-init_seg_name", "init-$RepresentationID$.m4s",
		"-media_seg_name", "chunk-$RepresentationID$-$Number%05d$.m4s",
		"-use_template", "1",
		"-use_timeline", "1",
		"-adaptation_sets", "id=0,streams=v id=1,streams=a",
		"-single_file", "0", // Segments séparés (pas un seul fichier)

		manifestPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	// Capture les logs FFmpeg
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Printf("Starting FFmpeg transcoding: %s", inputURL)

	// ✅ Exécution synchrone (attend la fin du transcodage)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %w\nLogs: %s", err, stderr.String())
	}

	log.Printf("FFmpeg transcoding completed successfully")

	// Vérifie que le manifest existe
	if _, err := os.Stat(manifestPath); err != nil {
		return "", fmt.Errorf("manifest not found: %w", err)
	}

	return manifestPath, nil
}

func GetVideoResolution(filePath string) (*models.VideoResolution, error) {
	// Commande ffprobe pour obtenir les métadonnées en JSON
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		filePath,
	)

	// Exécute la commande
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %w", err)
	}

	// Parse le JSON
	var result models.FFProbeResultVideo
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Cherche le stream vidéo
	for _, stream := range result.Streams {
		if stream.CodecType == "video" && stream.Width > 0 && stream.Height > 0 {
			return &models.VideoResolution{
				Width:  stream.Width,
				Height: stream.Height,
			}, nil
		}
	}

	return nil, fmt.Errorf("no video stream found")
}

func GetAudioTrackDetails(filePath string) ([]models.AudioTrackDetail, error) {
	// Commande ffprobe pour obtenir les métadonnées en JSON
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		filePath,
	)

	// Exécute la commande
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %w", err)
	}

	// Parse le JSON
	var result models.FFProbeResultAudio
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Extrait les détails des pistes audio
	var audioTrackDetails []models.AudioTrackDetail
	for _, stream := range result.Streams {
		if stream.CodecType == "audio" {
			language := getLanguageFromTags(stream.Tags, stream.Index)
			audioTrackDetails = append(audioTrackDetails, models.AudioTrackDetail{
				Index:    stream.Index,
				Language: language,
			})
		}
	}

	return audioTrackDetails, nil
}

// getLanguageFromTags extrait la langue des tags ou retourne une valeur par défaut
func getLanguageFromTags(tags map[string]interface{}, index int) string {
	if tags == nil {
		return fmt.Sprintf("Unknown language for stream %d", index)
	}

	// Essaye différentes clés possibles pour la langue
	if lang, ok := tags["language"].(string); ok && lang != "" {
		return lang
	}

	if lang, ok := tags["LANGUAGE"].(string); ok && lang != "" {
		return lang
	}

	if lang, ok := tags["Language"].(string); ok && lang != "" {
		return lang
	}

	return fmt.Sprintf("Unknown language for stream %d", index)
}
