package models

// ============================================================================
// PLAYER MODELS (ce que tu renvoies à ton front)
// ============================================================================

type VideoPlayer struct {
	Src                string                `json:"src"`
	Title              string                `json:"title"`
	Poster             string                `json:"poster"`
	Quality            string                `json:"quality"`
	AudioFormat        string                `json:"audioFormat"`
	Autoplay           bool                  `json:"autoplay"`
	ModelURL           string                `json:"modelUrl"` // ✅ Ajouté pour reconstruire les URLs
	AvailableQualities []Quality             `json:"availableQualities"`
	AvailableAudio     []PlayerAudioTrack    `json:"availableAudio"`
	AvailableSubtitles []PlayerSubtitleTrack `json:"availableSubtitles"`
}

type Quality struct {
	Label string `json:"label"` // ex: "1080p", "720p", "Original"
	Value string `json:"value"` // ex: "1080", "720", "full"
	URL   string `json:"url"`   // ✅ URL complète pour cette qualité
}

// Track audio pour le player (dérivé de MediaAudioTrack)
type PlayerAudioTrack struct {
	Label string `json:"label"` // ex: "French (AAC)"
	Value string `json:"value"` // ex: "fre1", "eng1"
	Codec string `json:"codec"` // ex: "aac", "eac3"
}

// Track sous-titres pour le player (dérivé de MediaSubtitleTrack)
type PlayerSubtitleTrack struct {
	Label string `json:"label"` // ex: "French", "English", "None"
	Value string `json:"value"` // ex: "fre1", "eng1", "none"
}
