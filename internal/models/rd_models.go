package models

import "encoding/json"

// ============================================================================
// REAL-DEBRID MODELS
// ============================================================================

// UnrestrictResponse représente la réponse de l'API Real-Debrid pour le débridage
// d'un lien (endpoint /unrestrict/link). Elle contient les informations sur le fichier
// débridé, y compris le lien de téléchargement direct généré par Real-Debrid.
type UnrestrictResponse struct {
	ID         string `json:"id"`         // Identifiant unique Real-Debrid du lien débridé
	Filename   string `json:"filename"`   // Nom du fichier débridé
	MimeType   string `json:"mimeType"`   // Type MIME du fichier (par exemple "video/mp4")
	Filesize   int64  `json:"filesize"`   // Taille du fichier en octets
	Link       string `json:"link"`       // Lien original fourni à Real-Debrid
	Host       string `json:"host"`       // Hébergeur du fichier (par exemple "uptobox.com")
	Chunks     int    `json:"chunks"`     // Nombre de segments pour le téléchargement
	CRC        int    `json:"crc"`        // Somme de contrôle CRC du fichier
	Download   string `json:"download"`   // Lien de téléchargement direct HD généré
	Streamable int    `json:"streamable"` // Indique si le fichier est streamable (1 = oui, 0 = non)
}

// RDError représente une réponse d'erreur de l'API Real-Debrid.
// Elle est retournée lorsqu'une requête échoue avec un code d'erreur spécifique.
type RDError struct {
	Error     string `json:"error"`      // Message d'erreur descriptif
	ErrorCode int    `json:"error_code"` // Code d'erreur numérique Real-Debrid
}

// TranscodeResponse représente la réponse de l'API Real-Debrid pour le transcodage
// d'un fichier vidéo (endpoint /streaming/transcode/{id}). Elle contient les URLs
// des différents formats de streaming disponibles (Apple HLS, DASH, MP4, WebM).
type TranscodeResponse struct {
	Apple struct {
		Full string `json:"full"` // URL du flux Apple HLS complet
	} `json:"apple"` // Format Apple HLS (HTTP Live Streaming)
	Dash struct {
		Full string `json:"full"` // URL du flux DASH complet
	} `json:"dash"` // Format DASH (Dynamic Adaptive Streaming over HTTP)
	LiveMP4 struct {
		Full string `json:"full"` // URL du flux MP4 en direct
	} `json:"liveMP4"` // Format MP4 en streaming progressif
	H264WebM struct {
		Full string `json:"full"` // URL du flux H264/WebM complet
	} `json:"h264WebM"` // Format H264 dans un conteneur WebM
}

// MediaInfoResponse représente la réponse de l'API Real-Debrid pour les informations
// média d'un fichier (endpoint /streaming/mediaInfos/{id}). Elle fournit les métadonnées
// complètes du fichier vidéo : durée, bitrate, pistes vidéo/audio/sous-titres,
// qualités disponibles et formats de transcodage.
type MediaInfoResponse struct {
	Filename           string            `json:"filename"`           // Nom du fichier source
	Hoster             string            `json:"hoster"`             // Nom de l'hébergeur
	Link               string            `json:"link"`               // Lien original du fichier
	Type               string            `json:"type"`               // Type de contenu (par exemple "video")
	Season             *string           `json:"season"`             // Numéro de saison (nil si film)
	Episode            *string           `json:"episode"`            // Numéro d'épisode (nil si film)
	Year               *int              `json:"year"`               // Année de sortie (nil si inconnu)
	Duration           float64           `json:"duration"`           // Durée en secondes
	Bitrate            int               `json:"bitrate"`            // Débit binaire en bits/s
	Size               int64             `json:"size"`               // Taille en octets
	Details            Details           `json:"details"`            // Détails des pistes vidéo, audio et sous-titres
	BaseURL            string            `json:"baseUrl"`            // URL de base pour construire les liens de streaming
	AvailableFormats   map[string]string `json:"availableFormats"`   // Formats de transcodage disponibles (clé : format, valeur : URL)
	AvailableQualities map[string]string `json:"availableQualities"` // Qualités disponibles (clé : résolution, valeur : URL)
	ModelURL           string            `json:"modelUrl"`           // URL modèle pour reconstruire les liens de qualité
	Host               string            `json:"host"`               // Hébergeur du fichier
	PosterPath         string            `json:"poster_path"`        // Chemin de l'affiche (ajouté côté serveur)
	AudioImage         string            `json:"audio_image"`        // Image associée à la piste audio
	BackdropPath       string            `json:"backdrop_path"`      // Chemin de l'image de fond (ajouté côté serveur)
}

// Details contient les pistes média (vidéo, audio, sous-titres) d'un fichier.
// Les champs Video et Subtitles utilisent des types personnalisés (VideoTracks et SubtitleTracks)
// car l'API Real-Debrid retourne un tableau vide [] lorsqu'aucune piste n'existe,
// mais un objet {} lorsqu'il y a des pistes. Le désérialiseur personnalisé gère ce polymorphisme.
type Details struct {
	Video     VideoTracks           `json:"video"`     // Pistes vidéo (type personnalisé gérant le polymorphisme [] vs {})
	Audio     map[string]AudioTrack `json:"audio"`     // Pistes audio indexées par identifiant (par exemple "1", "2")
	Subtitles SubtitleTracks        `json:"subtitles"` // Pistes de sous-titres (type personnalisé gérant le polymorphisme [] vs {})
}

// VideoTracks est un type personnalisé représentant les pistes vidéo d'un fichier média.
// Il est défini comme une map de chaînes vers VideoTrack car l'API Real-Debrid retourne
// les pistes sous forme d'objet JSON avec des clés numériques (par exemple {"1": {...}}).
// Le type implémente un UnmarshalJSON personnalisé pour gérer le polymorphisme JSON.
type VideoTracks map[string]VideoTrack

// UnmarshalJSON gère la désérialisation polymorphe du champ "video" de l'API Real-Debrid.
// L'API retourne soit un tableau vide [] lorsqu'il n'y a aucune piste vidéo,
// soit un objet JSON {} contenant les pistes vidéo indexées par clé numérique.
// Cette méthode détecte le type JSON (tableau ou objet) et désérialise en conséquence,
// initialisant une map vide dans le cas d'un tableau.
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

// VideoTrack représente une piste vidéo individuelle dans les informations média Real-Debrid.
type VideoTrack struct {
	Stream     string `json:"stream"`     // Identifiant du flux vidéo
	Lang       string `json:"lang"`       // Langue de la piste vidéo
	LangISO    string `json:"lang_iso"`   // Code ISO de la langue
	Codec      string `json:"codec"`      // Codec vidéo utilisé (par exemple "h264", "hevc")
	Colorspace string `json:"colorspace"` // Espace colorimétrique (par exemple "bt709")
	Width      int    `json:"width"`      // Largeur en pixels
	Height     int    `json:"height"`     // Hauteur en pixels
}

// AudioTrack représente une piste audio individuelle dans les informations média Real-Debrid.
type AudioTrack struct {
	Stream   string  `json:"stream"`   // Identifiant du flux audio
	Lang     string  `json:"lang"`     // Langue de la piste audio
	LangISO  string  `json:"lang_iso"` // Code ISO de la langue
	Codec    string  `json:"codec"`    // Codec audio utilisé (par exemple "aac", "eac3", "dts")
	Sampling int     `json:"sampling"` // Fréquence d'échantillonnage en Hz
	Channels float64 `json:"channels"` // Nombre de canaux audio (par exemple 2.0, 5.1, 7.1)
}

// SubtitleTracks est un type personnalisé représentant les pistes de sous-titres d'un fichier média.
// Comme VideoTracks, il gère le polymorphisme JSON de l'API Real-Debrid qui retourne
// soit un tableau vide [] soit un objet {} pour les sous-titres.
type SubtitleTracks map[string]SubtitleTrack

// UnmarshalJSON gère la désérialisation polymorphe du champ "subtitles" de l'API Real-Debrid.
// L'API retourne soit un tableau vide [] lorsqu'il n'y a aucun sous-titre,
// soit un objet JSON {} contenant les pistes de sous-titres indexées par clé numérique.
// Cette méthode détecte le type JSON (tableau ou objet) et désérialise en conséquence,
// initialisant une map vide dans le cas d'un tableau.
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

// SubtitleTrack représente une piste de sous-titres individuelle dans les informations média Real-Debrid.
type SubtitleTrack struct {
	Stream  string `json:"stream"`   // Identifiant du flux de sous-titres
	Lang    string `json:"lang"`     // Langue des sous-titres
	LangISO string `json:"lang_iso"` // Code ISO de la langue
	Type    string `json:"type"`     // Type de sous-titres (par exemple "srt", "ass")
}

// MediaDetails contient les pistes média analysées d'un fichier, utilisée comme structure
// intermédiaire après traitement des données brutes de Real-Debrid. Contrairement à Details,
// cette structure utilise des maps standard car le polymorphisme JSON a déjà été résolu.
type MediaDetails struct {
	Video     map[string]VideoTrackInfo `json:"video"`     // Pistes vidéo indexées par identifiant
	Audio     map[string]AudioTrackInfo `json:"audio"`     // Pistes audio indexées par identifiant
	Subtitles map[string]SubtitleInfo   `json:"subtitles"` // Sous-titres indexés par identifiant
}

// VideoTrackInfo représente une piste vidéo dans MediaDetails.
// Structure renommée par rapport à VideoTrack pour éviter la confusion
// avec les types utilisés dans le contexte du lecteur vidéo (player).
type VideoTrackInfo struct {
	Stream     string `json:"stream"`     // Identifiant du flux vidéo
	Lang       string `json:"lang"`       // Langue de la piste
	LangISO    string `json:"lang_iso"`   // Code ISO de la langue
	Codec      string `json:"codec"`      // Codec vidéo (par exemple "h264", "hevc")
	Colorspace string `json:"colorspace"` // Espace colorimétrique
	Width      int    `json:"width"`      // Largeur en pixels
	Height     int    `json:"height"`     // Hauteur en pixels
}

// AudioTrackInfo représente une piste audio dans MediaDetails.
// Structure renommée pour éviter la confusion avec AudioTrack utilisé
// dans le contexte du désérialiseur Real-Debrid.
type AudioTrackInfo struct {
	Stream   string  `json:"stream"`   // Identifiant du flux audio
	Lang     string  `json:"lang"`     // Langue de la piste
	LangISO  string  `json:"lang_iso"` // Code ISO de la langue
	Codec    string  `json:"codec"`    // Codec audio (par exemple "aac", "eac3")
	Sampling int     `json:"sampling"` // Fréquence d'échantillonnage en Hz
	Channels float64 `json:"channels"` // Nombre de canaux (par exemple 5.1)
}

// SubtitleInfo représente une piste de sous-titres dans MediaDetails.
// Structure renommée pour éviter la confusion avec SubtitleTrack.
type SubtitleInfo struct {
	Stream  string `json:"stream"`   // Identifiant du flux de sous-titres
	Lang    string `json:"lang"`     // Langue des sous-titres
	LangISO string `json:"lang_iso"` // Code ISO de la langue
	Type    string `json:"type"`     // Type de sous-titres
}

// TorrentioResponse représente la réponse principale de l'API Torrentio (addon Stremio).
// Elle contient la liste des flux torrent disponibles pour un contenu donné,
// ainsi que les paramètres de cache HTTP pour optimiser les requêtes.
type TorrentioResponse struct {
	Streams         []TorrentioStream `json:"streams"`                    // Liste des flux torrent disponibles
	CacheMaxAge     int               `json:"cacheMaxAge,omitempty"`      // Durée maximale de mise en cache HTTP en secondes
	StaleRevalidate int               `json:"staleRevalidate,omitempty"`  // Durée de revalidation stale-while-revalidate
	StaleError      int               `json:"staleError,omitempty"`       // Durée de tolérance stale-if-error
}

// TorrentioStream représente un flux torrent individuel retourné par Torrentio.
// Chaque flux correspond à une version spécifique du contenu (qualité, langue, codec)
// identifiée par son hash torrent.
type TorrentioStream struct {
	Name          string                 `json:"name"`                    // Nom du flux (contient généralement la qualité et le codec)
	Title         string                 `json:"title"`                   // Titre détaillé avec métadonnées (taille, seeders, etc.)
	InfoHash      string                 `json:"infoHash"`                // Hash unique du torrent pour l'ajout via magnet
	FileIdx       *int                   `json:"fileIdx,omitempty"`       // Index du fichier dans le torrent (pointeur car peut être absent)
	BehaviorHints TorrentioBehaviorHints `json:"behaviorHints,omitempty"` // Métadonnées supplémentaires sur le flux
	Sources       []string               `json:"sources,omitempty"`       // Sources du torrent (trackers, DHT)
}

// TorrentioBehaviorHints contient les métadonnées comportementales d'un flux Torrentio.
// Ces informations permettent au client de regrouper les épisodes pour le binge-watching
// et d'identifier le nom exact du fichier dans le torrent.
type TorrentioBehaviorHints struct {
	BingeGroup string `json:"bingeGroup,omitempty"` // Groupe de binge-watching pour enchaîner les épisodes
	Filename   string `json:"filename,omitempty"`   // Nom du fichier dans le torrent
}

// RdAddMagnetResponse représente la réponse de l'API Real-Debrid lors de l'ajout
// d'un lien magnet (endpoint /torrents/addMagnet). Elle retourne l'identifiant
// du torrent créé et l'URI pour suivre son état.
type RdAddMagnetResponse struct {
	Id  string `json:"id"`  // Identifiant unique du torrent dans Real-Debrid
	Uri string `json:"uri"` // URI pour consulter les informations du torrent
}

// RdTorrentInfo représente les informations détaillées d'un torrent dans Real-Debrid
// (endpoint /torrents/info/{id}). Elle contient l'état du téléchargement,
// la liste des fichiers et les liens de téléchargement générés.
type RdTorrentInfo struct {
	Id               string       `json:"id"`                // Identifiant unique du torrent
	Filename         string       `json:"filename"`          // Nom du fichier principal
	OriginalFilename string       `json:"original_filename"` // Nom original du fichier
	Hash             string       `json:"hash"`              // Hash du torrent
	Bytes            int64        `json:"bytes"`             // Taille totale en octets
	OriginalBytes    int64        `json:"original_bytes"`    // Taille originale en octets
	Host             string       `json:"host"`              // Hébergeur
	Split            int          `json:"split"`             // Nombre de parties du split
	Progress         float64      `json:"progress"`          // Progression du téléchargement (0-100)
	Status           string       `json:"status"`            // Statut du torrent (magnet_error, waiting_files_selection, downloaded, etc.)
	Added            string       `json:"added"`             // Date d'ajout
	Files            []RdFileInfo `json:"files"`             // Liste des fichiers dans le torrent
	Links            []string     `json:"links"`             // Liens de téléchargement générés par Real-Debrid
	Ended            string       `json:"ended"`             // Date de fin du téléchargement
}

// RdFileInfo représente un fichier individuel à l'intérieur d'un torrent Real-Debrid.
type RdFileInfo struct {
	Id       int    `json:"id"`       // Identifiant du fichier dans le torrent
	Path     string `json:"path"`     // Chemin relatif du fichier dans le torrent
	Bytes    int64  `json:"bytes"`    // Taille du fichier en octets
	Selected int    `json:"selected"` // Indique si le fichier est sélectionné pour le téléchargement (1 = oui)
}

// RdUnrestrictResponse représente la réponse de débridage Real-Debrid pour un lien spécifique.
// Elle est similaire à UnrestrictResponse mais utilise des conventions de nommage différentes
// pour les champs JSON (Id vs ID). Utilisée dans le flux de traitement des torrents.
type RdUnrestrictResponse struct {
	Id         string `json:"id"`         // Identifiant unique du lien débridé
	Filename   string `json:"filename"`   // Nom du fichier débridé
	MimeType   string `json:"mimeType"`   // Type MIME du fichier
	Filesize   int64  `json:"filesize"`   // Taille en octets
	Link       string `json:"link"`       // Lien original fourni
	Host       string `json:"host"`       // Hébergeur du fichier
	Chunks     int    `json:"chunks"`     // Nombre de segments de téléchargement
	Download   string `json:"download"`   // Lien direct HD généré par Real-Debrid
	Streamable int    `json:"streamable"` // Indique si le fichier est streamable (1 = oui)
}
