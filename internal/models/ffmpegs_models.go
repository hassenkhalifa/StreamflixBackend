package models

// ============================================================================
// FFPROBE MODELS
// ============================================================================

// FFProbeResultVideo représente la sortie JSON de ffprobe lors de l'analyse des flux vidéo
// d'un fichier multimédia. La commande ffprobe est exécutée avec les options
// -show_streams -select_streams v pour ne récupérer que les pistes vidéo.
type FFProbeResultVideo struct {
	Streams []StreamSpecVideo `json:"streams"` // Liste des pistes vidéo détectées dans le fichier
}

// FFProbeResultAudio représente la sortie JSON de ffprobe lors de l'analyse des flux audio
// d'un fichier multimédia. La commande ffprobe est exécutée avec les options
// -show_streams -select_streams a pour ne récupérer que les pistes audio.
type FFProbeResultAudio struct {
	Streams []StreamSpecAudio `json:"streams"` // Liste des pistes audio détectées dans le fichier
}

// StreamSpecVideo représente les spécifications d'une piste vidéo individuelle
// telle que retournée par ffprobe. Elle contient le type de codec ainsi que
// la résolution (largeur et hauteur en pixels) nécessaires pour déterminer
// les qualités de transcodage disponibles.
type StreamSpecVideo struct {
	CodecType string `json:"codec_type"`    // Type de codec (toujours "video" pour cette structure)
	Width     int    `json:"width,omitempty"`  // Largeur de la vidéo en pixels (omis si non disponible)
	Height    int    `json:"height,omitempty"` // Hauteur de la vidéo en pixels (omis si non disponible)
}

// StreamSpecAudio représente les spécifications d'une piste audio individuelle
// telle que retournée par ffprobe. Elle contient l'index de la piste, le type de codec
// et les tags associés (notamment la langue via le tag "language").
type StreamSpecAudio struct {
	Index     int                    `json:"index"`          // Index de la piste audio dans le fichier (commence à 0)
	CodecType string                 `json:"codec_type"`     // Type de codec (toujours "audio" pour cette structure)
	Tags      map[string]interface{} `json:"tags,omitempty"` // Tags de métadonnées (clé "language" pour la langue, etc.)
}

// VideoResolution représente la résolution d'une vidéo en pixels.
// Cette structure est utilisée après l'analyse ffprobe pour stocker
// et comparer les résolutions, notamment pour déterminer les qualités
// de transcodage disponibles (1080p, 720p, 480p, etc.).
type VideoResolution struct {
	Width  int `json:"width"`  // Largeur en pixels
	Height int `json:"height"` // Hauteur en pixels
}

// AudioTrackDetail représente les détails simplifiés d'une piste audio
// extraite par ffprobe. Cette structure est le résultat du traitement
// de StreamSpecAudio et ne conserve que l'index et la langue de la piste,
// informations nécessaires pour proposer le choix de langue au lecteur vidéo.
type AudioTrackDetail struct {
	Index    int    `json:"index"`    // Index de la piste audio dans le fichier
	Language string `json:"language"` // Code de langue ISO (par exemple "fre", "eng", "und" pour indéterminé)
}
