package models

import (
	"sync"
	"time"
)

// ZtParser représente le parser pour le site Zone Téléchargement.
// Il encapsule la configuration du scraper (URL de base, catégories supportées),
// le rate limiting (délai minimum entre les requêtes) et un mode développement
// pour le débogage. Le mutex protège l'accès concurrent au timestamp de la dernière requête.
type ZtParser struct {
	BaseUrl              string        // URL de base du site Zone Téléchargement (peut changer dynamiquement)
	AllCategories        []string      // Liste des catégories de contenu supportées par le scraper
	LastRequestTimestamp time.Time     // Horodatage de la dernière requête envoyée (pour le rate limiting)
	RequestTimeInBetween time.Duration // Délai minimum à respecter entre deux requêtes consécutives
	DevMode              bool          // Active le mode développement avec des logs supplémentaires
	MoviesDbKey          string        // Clé API pour accéder à la base de données de films (TMDB/TVMaze)
	Mu                   sync.Mutex    // Mutex protégeant l'accès concurrent à LastRequestTimestamp
}

// SearchResult représente un résultat de recherche provenant de Zone Téléchargement.
// Chaque résultat correspond à un contenu trouvé lors du scraping de la page de résultats.
// La structure contient les informations visibles sur la page de recherche
// ainsi que les métadonnées temporelles de publication.
type SearchResult struct {
	Title              string    `json:"title"`               // Titre du contenu tel qu'affiché sur Zone Téléchargement
	Url                string    `json:"url"`                 // URL complète de la page de détail du contenu
	Id                 string    `json:"id"`                  // Identifiant unique du contenu sur Zone Téléchargement
	Image              string    `json:"image"`               // URL de l'image d'aperçu du contenu
	Quality            string    `json:"quality"`             // Qualité du contenu (par exemple "1080p", "720p", "HDRip")
	Language           string    `json:"language"`            // Langue du contenu (par exemple "FRENCH", "MULTI", "VOSTFR")
	PublishedOn        time.Time `json:"published_on"`        // Date et heure de publication sur le site
	PublishedTimestamp int64     `json:"published_timestamp"` // Timestamp Unix de publication pour le tri
}

// MovieDetails représente les détails complets d'un film scrappé depuis Zone Téléchargement.
// Cette structure combine les informations extraites de la page Zone Téléchargement
// avec les métadonnées enrichies provenant de l'API TMDB (note, genres, synopsis).
// Les liens de téléchargement sont organisés par hébergeur.
type MovieDetails struct {
	Title        string            `json:"title"`         // Titre du film (potentiellement avec qualité et langue)
	OriginalName string            `json:"original_name"` // Titre original du film (souvent en anglais)
	Language     string            `json:"language"`      // Langue du contenu (par exemple "FRENCH", "MULTI")
	Quality      string            `json:"quality"`       // Qualité vidéo (par exemple "1080p", "BDRip")
	Size         string            `json:"size"`          // Taille du fichier (par exemple "4.2 Go")
	Description  string            `json:"description"`   // Synopsis du film (enrichi via TMDB)
	Poster       string            `json:"poster"`        // URL de l'affiche du film (depuis TMDB)
	ReleaseDate  string            `json:"release_date"`  // Date de sortie au format "YYYY-MM-DD"
	Genres       []string          `json:"genres"`        // Liste des genres (depuis TMDB)
	VoteAverage  float64           `json:"vote_average"`  // Note moyenne TMDB (sur 10)
	Directors    []string          `json:"directors"`     // Liste des réalisateurs (depuis TMDB)
	Actors       []string          `json:"actors"`        // Liste des acteurs principaux (depuis TMDB)
	DownloadLink map[string]string `json:"download_link"` // Liens de téléchargement par hébergeur (clé : nom, valeur : URL)
}

// SeriesDetails représente les détails complets d'une série scrappée depuis Zone Téléchargement.
// Similaire à MovieDetails, mais les liens de téléchargement sont organisés en deux niveaux :
// par épisode puis par hébergeur, reflétant la structure multi-épisodes d'une série.
type SeriesDetails struct {
	Title        string                       `json:"title"`         // Titre de la série (avec saison, qualité et langue)
	OriginalName string                       `json:"original_name"` // Titre original de la série
	Language     string                       `json:"language"`      // Langue du contenu
	Quality      string                       `json:"quality"`       // Qualité vidéo
	Size         string                       `json:"size"`          // Taille totale de la saison
	Description  string                       `json:"description"`   // Synopsis de la série (enrichi via TVMaze/TMDB)
	Poster       string                       `json:"poster"`        // URL de l'affiche
	ReleaseDate  string                       `json:"release_date"`  // Date de première diffusion
	Genres       []string                     `json:"genres"`        // Liste des genres
	VoteAverage  float64                      `json:"vote_average"`  // Note moyenne
	Directors    []string                     `json:"directors"`     // Liste des créateurs/réalisateurs
	Actors       []string                     `json:"actors"`        // Liste des acteurs principaux
	DownloadLink map[string]map[string]string `json:"download_link"` // Liens par épisode puis par hébergeur (épisode -> hébergeur -> URL)
}

// ZtBasicInfo représente les informations de base extraites d'une page de contenu
// Zone Téléchargement avant l'enrichissement par les API externes (TMDB, TVMaze).
// Le champ Links utilise une interface{} car sa structure diffère selon le type de contenu :
// map[string]string pour les films (hébergeur -> URL) ou
// map[string]map[string]string pour les séries (épisode -> hébergeur -> URL).
type ZtBasicInfo struct {
	Name         string      `json:"name"`          // Nom du contenu extrait de la page
	OriginalName string      `json:"original_name"` // Titre original (souvent en anglais)
	Language     string      `json:"language"`      // Langue du contenu
	Quality      string      `json:"quality"`       // Qualité vidéo
	Size         string      `json:"size"`          // Taille du fichier ou de la saison
	Links        interface{} `json:"links"`         // Liens de téléchargement (structure polymorphe selon le type de contenu)
}

// TmdbMovieResponse représente la réponse de l'API TMDB pour une recherche de films
// dans le contexte de Zone Téléchargement. Utilisée pour enrichir les résultats
// du scraping avec les métadonnées TMDB (affiche, note, genres).
type TmdbMovieResponse struct {
	Results []TmdbMovie `json:"results"` // Liste des films correspondant à la recherche
}

// TmdbMovie représente un film retourné par l'API TMDB dans le contexte
// de Zone Téléchargement. Cette structure contient les champs nécessaires
// pour enrichir un résultat de scraping avec des métadonnées de qualité.
type TmdbMovie struct {
	Id            int     `json:"id"`             // Identifiant unique TMDB
	Title         string  `json:"title"`          // Titre localisé du film
	OriginalTitle string  `json:"original_title"` // Titre original du film
	Overview      string  `json:"overview"`       // Synopsis du film
	PosterPath    string  `json:"poster_path"`    // Chemin relatif de l'affiche sur TMDB
	ReleaseDate   string  `json:"release_date"`   // Date de sortie au format "YYYY-MM-DD"
	VoteAverage   float64 `json:"vote_average"`   // Note moyenne des votes TMDB
	GenreIds      []int   `json:"genre_ids"`      // Identifiants numériques des genres
}

// TmdbGenresResponse représente la réponse de l'API TMDB pour l'endpoint /genre/movie/list.
// Elle contient la liste complète des genres disponibles, utilisée pour convertir
// les identifiants de genre en noms lisibles dans le contexte Zone Téléchargement.
type TmdbGenresResponse struct {
	Genres []TmdbGenre `json:"genres"` // Liste de tous les genres de films disponibles
}

// TmdbGenre représente un genre individuel retourné par l'API TMDB.
// Utilisé dans le contexte Zone Téléchargement pour la résolution des noms de genre.
type TmdbGenre struct {
	Id   int    `json:"id"`   // Identifiant numérique du genre
	Name string `json:"name"` // Nom du genre (localisé selon la langue de la requête)
}

// TmdbCreditsResponse représente les crédits d'un film retournés par l'API TMDB
// (endpoint /movie/{id}/credits). Utilisée dans le contexte Zone Téléchargement
// pour enrichir les détails d'un film avec le casting et l'équipe technique.
type TmdbCreditsResponse struct {
	Cast []TmdbCastMember `json:"cast"` // Liste des acteurs
	Crew []TmdbCrewMember `json:"crew"` // Liste de l'équipe technique
}

// TmdbCastMember représente un acteur dans les crédits TMDB.
// Utilisé pour extraire les noms des acteurs principaux d'un film.
type TmdbCastMember struct {
	Name string `json:"name"` // Nom complet de l'acteur
}

// TmdbCrewMember représente un membre de l'équipe technique dans les crédits TMDB.
// Le champ Job permet d'identifier le rôle (Director, Producer, Writer, etc.)
// pour extraire les informations pertinentes.
type TmdbCrewMember struct {
	Name string `json:"name"` // Nom complet du membre de l'équipe
	Job  string `json:"job"`  // Poste occupé (par exemple "Director", "Producer", "Screenplay")
}

// TvmazeSearchResponse représente la réponse de l'API TVMaze pour une recherche de séries.
// Chaque résultat de recherche encapsule un objet TvmazeShow contenant
// les informations détaillées de la série trouvée.
type TvmazeSearchResponse struct {
	Show TvmazeShow `json:"show"` // Informations détaillées de la série trouvée
}

// TvmazeShow représente une série TV telle que retournée par l'API TVMaze.
// Cette structure est utilisée pour enrichir les résultats de scraping
// Zone Téléchargement avec des métadonnées de qualité (affiche, note, genres, synopsis).
type TvmazeShow struct {
	Id        int          `json:"id"`        // Identifiant unique TVMaze
	Name      string       `json:"name"`      // Nom de la série
	Summary   string       `json:"summary"`   // Résumé HTML de la série (contient des balises <p>, <b>, etc.)
	Premiered string       `json:"premiered"` // Date de première diffusion au format "YYYY-MM-DD"
	Genres    []string     `json:"genres"`    // Liste des genres en anglais
	Rating    TvmazeRating `json:"rating"`    // Note de la série
	Image     TvmazeImage  `json:"image"`     // Images de la série (affiche)
}

// TvmazeRating représente la note d'une série sur TVMaze.
// La note est une moyenne sur 10 calculée à partir des votes des utilisateurs.
type TvmazeRating struct {
	Average float64 `json:"average"` // Note moyenne sur 10 (peut être 0 si non notée)
}

// TvmazeImage contient les URLs des images d'une série sur TVMaze.
// Seule l'image originale (haute résolution) est utilisée dans StreamFlix.
type TvmazeImage struct {
	Original string `json:"original"` // URL de l'image en résolution originale
}

// TvmazeCastMember représente un membre du casting d'une série sur TVMaze.
// La structure encapsule un objet TvmazePerson contenant le nom de l'acteur.
type TvmazeCastMember struct {
	Person TvmazePerson `json:"person"` // Informations sur l'acteur
}

// TvmazeCrewMember représente un membre de l'équipe technique d'une série sur TVMaze.
// Le champ Type identifie le rôle (Creator, Director, Writer, etc.)
// tandis que Person contient les informations de la personne.
type TvmazeCrewMember struct {
	Type   string       `json:"type"`   // Rôle dans l'équipe (par exemple "Creator", "Executive Producer")
	Person TvmazePerson `json:"person"` // Informations sur la personne
}

// TvmazePerson représente une personne (acteur ou membre d'équipe) sur TVMaze.
type TvmazePerson struct {
	Name string `json:"name"` // Nom complet de la personne
}

// ZtUrlResponse représente la réponse de l'API interne qui fournit l'URL actuelle
// du site Zone Téléchargement. Le site changeant régulièrement de domaine,
// cette structure permet de récupérer dynamiquement l'URL valide avec son horodatage.
type ZtUrlResponse struct {
	Url       string `json:"url"`       // URL actuelle du site Zone Téléchargement
	Timestamp string `json:"timestamp"` // Horodatage de la dernière vérification de l'URL
}

// ErrorResponse représente une réponse d'erreur standardisée de l'API StreamFlix.
// Elle est utilisée pour retourner des erreurs structurées au client,
// avec un statut booléen, un message d'erreur et une pile d'appels optionnelle
// (uniquement en mode développement).
type ErrorResponse struct {
	Status bool     `json:"status"`          // Toujours false pour les erreurs
	Error  string   `json:"error"`           // Message d'erreur descriptif
	Stack  []string `json:"stack,omitempty"` // Pile d'appels pour le débogage (optionnelle, mode dev uniquement)
}
