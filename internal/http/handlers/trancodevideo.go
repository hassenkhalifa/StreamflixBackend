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

// ConvertToDASHLive convertit une source vidéo en flux MPEG-DASH en mode "live"
// (non-bloquant). La commande FFmpeg est lancée en arrière-plan et la fonction
// retourne dès que le fichier manifeste (manifest.mpd) est créé sur le disque.
//
// Paramètres FFmpeg utilisés :
//   - -y                     : écrase les fichiers existants sans confirmation
//   - -i inputURL            : URL source de la vidéo à transcoder
//   - -map 0:v:0 / -map 0:a:0 : sélectionne le premier flux vidéo et audio
//   - -c:v libx264           : encodeur vidéo H.264 logiciel
//   - -preset veryfast       : compromis vitesse/qualité (plus rapide qu'ultrafast en qualité)
//   - -profile:v main        : profil H.264 Main (compatibilité large)
//   - -level 4.0             : niveau H.264 4.0 (supporte 1080p@30fps)
//   - -crf 23                : qualité constante (échelle 0-51, 23 = défaut équilibré)
//   - -pix_fmt yuv420p       : format pixel standard pour la compatibilité maximale
//   - -g 48                  : taille du GOP (Group of Pictures) = 48 images
//   - -sc_threshold 0        : désactive la détection de changement de scène pour un GOP régulier
//   - -c:a aac               : encodeur audio AAC
//   - -ac 2                  : mixage stéréo (2 canaux)
//   - -b:a 128k              : débit audio 128 kbps
//   - -f dash                : format de sortie MPEG-DASH
//   - -seg_duration 4        : durée des segments de 4 secondes
//   - -use_template 1        : utilise un modèle pour les noms de segments
//   - -use_timeline 1        : utilise une timeline dans le manifeste MPD
//   - -adaptation_sets       : définit les ensembles d'adaptation (vidéo=0, audio=1)
//
// Le processus FFmpeg continue en arrière-plan via une goroutine. La fonction
// attend jusqu'à 15 secondes (30 x 500ms) que le manifeste soit créé avant de retourner.
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

// ConvertToDASHVOD convertit une source vidéo en flux MPEG-DASH en mode VOD
// (Video On Demand). Contrairement à ConvertToDASHLive, cette fonction est
// synchrone et attend la fin complète du transcodage avant de retourner.
//
// Paramètres FFmpeg utilisés (en plus de ceux de ConvertToDASHLive) :
//   - -keyint_min 48          : intervalle minimum entre les keyframes (identique à -g)
//   - -ar 48000               : fréquence d'échantillonnage audio de 48 kHz
//   - -init_seg_name          : modèle de nommage du segment d'initialisation (init-$RepresentationID$.m4s)
//   - -media_seg_name         : modèle de nommage des segments média (chunk-$RepresentationID$-$Number%05d$.m4s)
//   - -single_file 0          : produit des segments séparés (pas un fichier unique)
//
// Le mode VOD est adapté aux fichiers complets déjà téléchargés, car il produit
// un manifeste DASH statique avec tous les segments référencés dès le départ.
// Vérifie l'existence du manifeste après le transcodage et retourne une erreur si absent.
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

// GetVideoResolution utilise ffprobe pour extraire la résolution (largeur x hauteur)
// du premier flux vidéo trouvé dans le fichier spécifié. La commande ffprobe est
// exécutée avec une sortie JSON silencieuse (-v quiet -print_format json -show_streams).
// Parcourt les flux retournés et retourne la résolution du premier flux de type "video"
// ayant une largeur et une hauteur positives. Retourne une erreur si aucun flux
// vidéo valide n'est trouvé dans le fichier.
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

// GetAudioTrackDetails utilise ffprobe pour extraire les détails de toutes les
// pistes audio présentes dans le fichier spécifié. Pour chaque flux audio trouvé,
// la fonction extrait l'index du flux et la langue (via les tags de métadonnées).
// La commande ffprobe est exécutée avec les mêmes paramètres que GetVideoResolution.
// Retourne une liste d'AudioTrackDetail contenant l'index et la langue de chaque piste.
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

// getLanguageFromTags extrait la langue d'un flux audio à partir de ses tags
// de métadonnées ffprobe. La fonction essaie plusieurs variantes de casse pour
// la clé "language" (minuscules, majuscules, capitalisée) afin de gérer les
// différents encodeurs et formats de conteneurs. Si aucune clé n'est trouvée
// ou si les tags sont nil, retourne un message par défaut incluant l'index du flux.
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
