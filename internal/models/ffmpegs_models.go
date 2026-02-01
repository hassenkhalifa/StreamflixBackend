package models

// ============================================================================
// FFPROBE MODELS
// ============================================================================

// FFProbeResultVideo structure pour parser le JSON de ffprobe (vidéo)
type FFProbeResultVideo struct {
	Streams []StreamSpecVideo `json:"streams"`
}

// FFProbeResultAudio structure pour parser le JSON de ffprobe (audio)
type FFProbeResultAudio struct {
	Streams []StreamSpecAudio `json:"streams"`
}

// StreamSpecVideo représente un stream vidéo
type StreamSpecVideo struct {
	CodecType string `json:"codec_type"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

// StreamSpecAudio représente un stream audio
type StreamSpecAudio struct {
	Index     int                    `json:"index"`
	CodecType string                 `json:"codec_type"`
	Tags      map[string]interface{} `json:"tags,omitempty"`
}

// VideoResolution représente la résolution d'une vidéo
type VideoResolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// AudioTrackDetail représente les détails d'une piste audio (ffprobe)
type AudioTrackDetail struct {
	Index    int    `json:"index"`
	Language string `json:"language"`
}
