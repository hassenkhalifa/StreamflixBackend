package models

// ============================================================================
// PLAYER MODELS (modèles retournés au frontend pour le lecteur vidéo)
// ============================================================================

// VideoPlayer représente la configuration complète du lecteur vidéo envoyée au frontend.
// Cette structure contient toutes les informations nécessaires pour initialiser le lecteur :
// la source vidéo, les métadonnées d'affichage (titre, affiche), les options de qualité,
// les pistes audio disponibles et les sous-titres. Le champ ModelURL permet au frontend
// de reconstruire dynamiquement les URLs lors d'un changement de qualité.
type VideoPlayer struct {
	Src                string                `json:"src"`                // URL de la source vidéo principale (qualité par défaut)
	Title              string                `json:"title"`             // Titre du contenu affiché dans le lecteur
	Poster             string                `json:"poster"`            // URL de l'image d'affiche affichée avant la lecture
	Quality            string                `json:"quality"`           // Qualité actuellement sélectionnée (par exemple "1080p")
	AudioFormat        string                `json:"audioFormat"`       // Format audio du flux actuel (par exemple "AAC", "EAC3")
	Autoplay           bool                  `json:"autoplay"`         // Indique si la lecture doit démarrer automatiquement
	ModelURL           string                `json:"modelUrl"`          // URL modèle pour reconstruire les liens selon la qualité choisie
	AvailableQualities []Quality             `json:"availableQualities"` // Liste des qualités vidéo disponibles
	AvailableAudio     []PlayerAudioTrack    `json:"availableAudio"`    // Liste des pistes audio disponibles
	AvailableSubtitles []PlayerSubtitleTrack `json:"availableSubtitles"` // Liste des pistes de sous-titres disponibles
}

// Quality représente une option de qualité vidéo disponible dans le lecteur.
// Chaque qualité possède un libellé lisible pour l'interface, une valeur technique
// pour la logique de sélection et une URL complète pour le streaming.
type Quality struct {
	Label string `json:"label"` // Libellé affiché dans le sélecteur de qualité (par exemple "1080p", "720p", "Original")
	Value string `json:"value"` // Valeur technique de la qualité (par exemple "1080", "720", "full")
	URL   string `json:"url"`   // URL complète du flux vidéo pour cette qualité
}

// PlayerAudioTrack représente une piste audio disponible dans le lecteur vidéo.
// Elle est dérivée des informations média de Real-Debrid (AudioTrack/AudioTrackInfo)
// et formatée pour l'affichage dans l'interface du lecteur.
type PlayerAudioTrack struct {
	Label string `json:"label"` // Libellé affiché (par exemple "French (AAC)", "English (EAC3 5.1)")
	Value string `json:"value"` // Identifiant technique de la piste (par exemple "fre1", "eng1")
	Codec string `json:"codec"` // Codec audio (par exemple "aac", "eac3", "dts")
}

// PlayerSubtitleTrack représente une piste de sous-titres disponible dans le lecteur vidéo.
// Elle est dérivée des informations média de Real-Debrid (SubtitleTrack/SubtitleInfo)
// et inclut toujours une option "Aucun" (value: "none") pour désactiver les sous-titres.
type PlayerSubtitleTrack struct {
	Label string `json:"label"` // Libellé affiché (par exemple "French", "English", "Aucun")
	Value string `json:"value"` // Identifiant technique de la piste (par exemple "fre1", "eng1", "none")
}
