package models

import "encoding/json"

// ============================================================================
// REAL-DEBRID MODELS
// ============================================================================

type UnrestrictResponse struct {
	ID         string `json:"id"`
	Filename   string `json:"filename"`
	MimeType   string `json:"mimeType"`
	Filesize   int64  `json:"filesize"`
	Link       string `json:"link"`
	Host       string `json:"host"`
	Chunks     int    `json:"chunks"`
	CRC        int    `json:"crc"`
	Download   string `json:"download"`
	Streamable int    `json:"streamable"`
}

type RDError struct {
	Error     string `json:"error"`
	ErrorCode int    `json:"error_code"`
}

type TranscodeResponse struct {
	Apple struct {
		Full string `json:"full"`
	} `json:"apple"`
	Dash struct {
		Full string `json:"full"`
	} `json:"dash"`
	LiveMP4 struct {
		Full string `json:"full"`
	} `json:"liveMP4"`
	H264WebM struct {
		Full string `json:"full"`
	} `json:"h264WebM"`
}

// MediaInfoResponse représente la réponse de /streaming/mediaInfos/{id}
type MediaInfoResponse struct {
	Filename           string            `json:"filename"`
	Hoster             string            `json:"hoster"`
	Link               string            `json:"link"`
	Type               string            `json:"type"`
	Season             *string           `json:"season"`
	Episode            *string           `json:"episode"`
	Year               *int              `json:"year"`
	Duration           float64           `json:"duration"`
	Bitrate            int               `json:"bitrate"`
	Size               int64             `json:"size"`
	Details            Details           `json:"details"`
	BaseURL            string            `json:"baseUrl"`
	AvailableFormats   map[string]string `json:"availableFormats"`
	AvailableQualities map[string]string `json:"availableQualities"`
	ModelURL           string            `json:"modelUrl"`
	Host               string            `json:"host"`
	PosterPath         string            `json:"poster_path"`
	AudioImage         string            `json:"audio_image"`
	BackdropPath       string            `json:"backdrop_path"`
}

type Details struct {
	Video     VideoTracks           `json:"video"` // ✅ Type personnalisé pour gérer [] ou {}
	Audio     map[string]AudioTrack `json:"audio"`
	Subtitles SubtitleTracks        `json:"subtitles"` // ✅ Type personnalisé pour gérer [] ou {}
}

// ✅ Type personnalisé pour gérer video qui peut être [] ou {}
type VideoTracks map[string]VideoTrack

func (v *VideoTracks) UnmarshalJSON(data []byte) error {
	// Si c'est un array vide []
	if len(data) > 0 && data[0] == '[' {
		*v = make(map[string]VideoTrack)
		return nil
	}
	// Sinon c'est un objet {}
	var tracks map[string]VideoTrack
	if err := json.Unmarshal(data, &tracks); err != nil {
		return err
	}
	*v = tracks
	return nil
}

type VideoTrack struct {
	Stream     string `json:"stream"`
	Lang       string `json:"lang"`
	LangISO    string `json:"lang_iso"`
	Codec      string `json:"codec"`
	Colorspace string `json:"colorspace"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

type AudioTrack struct {
	Stream   string  `json:"stream"`
	Lang     string  `json:"lang"`
	LangISO  string  `json:"lang_iso"`
	Codec    string  `json:"codec"`
	Sampling int     `json:"sampling"`
	Channels float64 `json:"channels"`
}

// ✅ Type personnalisé pour gérer subtitles qui peut être [] ou {}
type SubtitleTracks map[string]SubtitleTrack

func (s *SubtitleTracks) UnmarshalJSON(data []byte) error {
	// Si c'est un array vide []
	if len(data) > 0 && data[0] == '[' {
		*s = make(map[string]SubtitleTrack)
		return nil
	}
	// Sinon c'est un objet {}
	var tracks map[string]SubtitleTrack
	if err := json.Unmarshal(data, &tracks); err != nil {
		return err
	}
	*s = tracks
	return nil
}

type SubtitleTrack struct {
	Stream  string `json:"stream"`
	Lang    string `json:"lang"`
	LangISO string `json:"lang_iso"`
	Type    string `json:"type"`
}

type MediaDetails struct {
	Video     map[string]VideoTrackInfo `json:"video"`
	Audio     map[string]AudioTrackInfo `json:"audio"`
	Subtitles map[string]SubtitleInfo   `json:"subtitles"`
}

// ✅ Renommé pour éviter confusion avec les types Player
type VideoTrackInfo struct {
	Stream     string `json:"stream"`
	Lang       string `json:"lang"`
	LangISO    string `json:"lang_iso"`
	Codec      string `json:"codec"`
	Colorspace string `json:"colorspace"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

type AudioTrackInfo struct {
	Stream   string  `json:"stream"`
	Lang     string  `json:"lang"`
	LangISO  string  `json:"lang_iso"`
	Codec    string  `json:"codec"`
	Sampling int     `json:"sampling"`
	Channels float64 `json:"channels"`
}

type SubtitleInfo struct {
	Stream  string `json:"stream"`
	Lang    string `json:"lang"`
	LangISO string `json:"lang_iso"`
	Type    string `json:"type"`
}

// TorrentioResponse - Réponse principale de Torrentio
type TorrentioResponse struct {
	Streams         []TorrentioStream `json:"streams"`
	CacheMaxAge     int               `json:"cacheMaxAge,omitempty"`
	StaleRevalidate int               `json:"staleRevalidate,omitempty"`
	StaleError      int               `json:"staleError,omitempty"`
}

// TorrentioStream - Un stream torrent individuel
type TorrentioStream struct {
	Name          string                 `json:"name"`
	Title         string                 `json:"title"`
	InfoHash      string                 `json:"infoHash"`
	FileIdx       *int                   `json:"fileIdx,omitempty"` // Pointeur car peut être absent
	BehaviorHints TorrentioBehaviorHints `json:"behaviorHints,omitempty"`
	Sources       []string               `json:"sources,omitempty"`
}

// TorrentioBehaviorHints - Métadonnées du stream
type TorrentioBehaviorHints struct {
	BingeGroup string `json:"bingeGroup,omitempty"`
	Filename   string `json:"filename,omitempty"`
}

type RdAddMagnetResponse struct {
	Id  string `json:"id"`
	Uri string `json:"uri"`
}

type RdTorrentInfo struct {
	Id               string       `json:"id"`
	Filename         string       `json:"filename"`
	OriginalFilename string       `json:"original_filename"`
	Hash             string       `json:"hash"`
	Bytes            int64        `json:"bytes"`
	OriginalBytes    int64        `json:"original_bytes"`
	Host             string       `json:"host"`
	Split            int          `json:"split"`
	Progress         float64      `json:"progress"`
	Status           string       `json:"status"`
	Added            string       `json:"added"`
	Files            []RdFileInfo `json:"files"`
	Links            []string     `json:"links"`
	Ended            string       `json:"ended"`
}

type RdFileInfo struct {
	Id       int    `json:"id"`
	Path     string `json:"path"`
	Bytes    int64  `json:"bytes"`
	Selected int    `json:"selected"`
}

type RdUnrestrictResponse struct {
	Id         string `json:"id"`
	Filename   string `json:"filename"`
	MimeType   string `json:"mimeType"`
	Filesize   int64  `json:"filesize"`
	Link       string `json:"link"` // Lien original fourni
	Host       string `json:"host"`
	Chunks     int    `json:"chunks"`
	Download   string `json:"download"` // Lien direct HD
	Streamable int    `json:"streamable"`
}
