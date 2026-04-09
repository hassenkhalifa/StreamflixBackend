package models

import (
	"sync"
	"time"
)

// ============================================================================
// CACHE
// ============================================================================

// Variables globales pour le cache de recherche. SearchCache stocke les résultats
// de recherche TMDB en mémoire avec une durée de vie définie par CacheTTL.
// CacheMutex protège l'accès concurrent à SearchCache.
var (
	// SearchCache est une map en mémoire associant une clé de recherche (chaîne)
	// à un résultat mis en cache. Utilisé pour éviter les appels répétés à l'API TMDB
	// lors de recherches identiques dans un court intervalle.
	SearchCache = make(map[string]CachedSearch)
	// CacheMutex est un verrou lecture/écriture protégeant l'accès concurrent à SearchCache.
	CacheMutex sync.RWMutex
	// CacheTTL définit la durée de vie des entrées dans SearchCache (5 minutes par défaut).
	CacheTTL = 5 * time.Minute
)

// MovieGenreCacheKey est la clé de cache pour les requêtes de films par genre.
// Elle combine l'identifiant du genre, le numéro de page et la langue
// pour identifier de manière unique chaque requête paginée.
type MovieGenreCacheKey struct {
	GenreID  int    // Identifiant TMDB du genre
	Page     int    // Numéro de page pour la pagination
	Language string // Code de langue (par exemple "fr-FR")
}

// popularMoviesCacheKey est la clé de cache pour les films populaires.
// Elle combine la page et la langue pour identifier chaque requête paginée.
type popularMoviesCacheKey struct {
	page     string // Numéro de page sous forme de chaîne
	language string // Code de langue pour la requête
}

// topRatedMoviesCacheKey est la clé de cache pour les films les mieux notés.
type topRatedMoviesCacheKey struct {
	page int // Numéro de page pour la pagination
}

// trendingMoviesCacheKey est la clé de cache pour les films tendance.
// Elle combine la fenêtre temporelle (day/week), la page et la langue.
type trendingMoviesCacheKey struct {
	timeWindow string // Fenêtre temporelle : "day" ou "week"
	page       int    // Numéro de page pour la pagination
	language   string // Code de langue pour la requête
}

// ContentDetailsCacheKey est la clé de cache pour les détails d'un contenu spécifique.
type ContentDetailsCacheKey struct {
	movieID int // Identifiant TMDB du film ou de la série
}

// similarMoviesCacheKey est la clé de cache pour les films similaires à un film donné.
type similarMoviesCacheKey struct {
	movieID int    // Identifiant TMDB du film source
	page    string // Numéro de page sous forme de chaîne
}

// movieCreditsCacheKey est la clé de cache pour les crédits (casting, équipe) d'un film.
type movieCreditsCacheKey struct {
	movieID int // Identifiant TMDB du film
}

// movieImdbIDCacheKey est la clé de cache pour l'identifiant IMDB d'un film.
type movieImdbIDCacheKey struct {
	movieID int // Identifiant TMDB du film
}

// genreCategoriesCacheKey est la clé de cache pour la liste des catégories de genres.
type genreCategoriesCacheKey struct {
	language string // Code de langue pour les noms de genres localisés
}

// Cache est une structure de cache générique thread-safe avec expiration.
// Elle stocke des entrées de type V indexées par des clés de type K,
// avec un TTL (Time To Live) configurable pour chaque entrée.
type Cache[K comparable, V any] struct {
	Mu    sync.RWMutex   // Verrou lecture/écriture pour l'accès concurrent
	Store map[K]Entry[V] // Map de stockage associant clés et entrées
	Ttl   time.Duration  // Durée de vie des entrées du cache
}

// Entry représente une entrée individuelle dans le cache générique.
// Elle contient les données stockées ainsi que leur date d'expiration.
type Entry[V any] struct {
	Data      V         // Données mises en cache
	ExpiresAt time.Time // Date et heure d'expiration de cette entrée
}

// ============================================================================
// CONTENT & MEDIA MODELS
// ============================================================================

// Movie représente un film avec ses informations de base.
// Cette structure est utilisée pour les réponses simplifiées de l'API interne.
// Les données sont sérialisées en JSON pour le frontend.
type Movie struct {
	ID       int      `json:"id"`     // Identifiant unique du film
	Title    string   `json:"title"`  // Titre du film
	Year     int      `json:"year"`   // Année de sortie
	Rating   float32  `json:"rating"` // Note moyenne (sur 10)
	Genre    []string `json:"genre"`  // Liste des genres associés
	ImageURL string   `json:"image"`  // URL de l'affiche du film
}

// Categories représente une catégorie de contenu affichée dans l'interface.
// Chaque catégorie regroupe des films ou séries par genre et possède
// un style visuel (couleur) et des aperçus d'images.
type Categories struct {
	ID           int      `json:"id"`           // Identifiant unique de la catégorie
	CategoryName string   `json:"categoryName"` // Nom affiché de la catégorie (par exemple "Action")
	Description  string   `json:"description"`  // Description courte de la catégorie
	Href         string   `json:"href"`         // Lien de navigation vers la page de la catégorie
	Color        string   `json:"color"`        // Classe CSS de couleur Tailwind pour le style
	Previews     []string `json:"previews"`     // URLs des images d'aperçu pour la catégorie
}

// Cast représente un membre du casting dans les détails d'un contenu.
type Cast struct {
	Name  string `json:"name"`  // Nom complet de l'acteur
	Role  string `json:"role"`  // Rôle ou personnage interprété
	Image string `json:"image"` // URL de la photo de l'acteur
}

// ContentDetails représente les détails complets d'un contenu (film ou série).
// Cette structure contient toutes les informations nécessaires à l'affichage
// de la page de détail d'un contenu dans le frontend.
type ContentDetails struct {
	ID             int      `json:"id"`             // Identifiant unique du contenu
	Title          string   `json:"title"`          // Titre du contenu
	Image          string   `json:"image"`          // URL de l'affiche principale
	BackdropImage  string   `json:"backdropImage"`  // URL de l'image de fond (backdrop)
	Year           int      `json:"year"`           // Année de sortie ou de première diffusion
	Genres         []string `json:"genres"`         // Liste des genres associés
	Rating         float32  `json:"rating"`         // Note moyenne (sur 10)
	Duration       string   `json:"duration"`       // Durée formatée (par exemple "2h15")
	Synopsis       string   `json:"synopsis"`       // Résumé de l'intrigue
	Director       string   `json:"director"`       // Nom du réalisateur
	Producer       string   `json:"producer"`       // Nom du producteur ou de la société de production
	Languages      string   `json:"languages"`      // Langues parlées dans le contenu
	Classification string   `json:"classification"` // Classification d'âge (par exemple "PG-13")
	Cast           []Cast   `json:"cast"`           // Liste des acteurs principaux
}

// MovieDTO est le Data Transfer Object principal pour les films.
// Il contient les informations essentielles d'un film, transformées depuis
// les données brutes TMDB et prêtes à être envoyées au frontend.
type MovieDTO struct {
	ID     int      `json:"id"`     // Identifiant TMDB du film
	Title  string   `json:"title"`  // Titre du film (localisé)
	Image  string   `json:"image"`  // URL complète de l'affiche (construite depuis poster_path)
	Year   int      `json:"year"`   // Année de sortie extraite de release_date
	Genre  []string `json:"genre"`  // Noms des genres traduits via MovieGenreMap
	Rating float64  `json:"rating"` // Note moyenne TMDB (sur 10)
}

// TmdbPopularResponse représente la réponse de l'API TMDB pour l'endpoint des films populaires.
// Elle contient un tableau de films bruts à transformer en MovieDTO.
type TmdbPopularResponse struct {
	Results []tmdbMovie `json:"results"` // Liste des films populaires retournés par TMDB
}

// tmdbMovie représente un film tel que retourné par l'API TMDB dans les listes
// (populaires, tendances, etc.). Cette structure non exportée sert de modèle
// intermédiaire avant transformation en MovieDTO.
type tmdbMovie struct {
	ID           int     `json:"id"`            // Identifiant unique TMDB
	Title        string  `json:"title"`         // Titre du film
	PosterPath   string  `json:"poster_path"`   // Chemin relatif de l'affiche sur TMDB
	BackdropPath string  `json:"backdrop_path"` // Chemin relatif de l'image de fond sur TMDB
	ReleaseDate  string  `json:"release_date"`  // Date de sortie au format "YYYY-MM-DD"
	GenreIDs     []int   `json:"genre_ids"`     // Identifiants numériques des genres
	VoteAverage  float64 `json:"vote_average"`  // Note moyenne des votes
}

// ContentDetailsDTO est le Data Transfer Object pour les détails complets d'un contenu.
// Il est retourné par l'API StreamFlix et combine les informations provenant
// de plusieurs endpoints TMDB (détails, crédits). Les champs optionnels sont
// omis du JSON s'ils sont vides (omitempty).
type ContentDetailsDTO struct {
	ID             int      `json:"id"`                        // Identifiant TMDB du contenu
	Title          string   `json:"title"`                     // Titre du contenu
	Image          string   `json:"image"`                     // URL complète de l'affiche
	Imdbid         string   `json:"imdb_id"`                   // Identifiant IMDB (par exemple "tt1234567")
	BackdropImage  string   `json:"backdropImage,omitempty"`   // URL de l'image de fond (optionnelle)
	Year           int      `json:"year"`                      // Année de sortie
	Genres         []string `json:"genres"`                    // Noms des genres
	Rating         float64  `json:"rating"`                    // Note moyenne
	Duration       string   `json:"duration,omitempty"`        // Durée formatée (optionnelle)
	Synopsis       string   `json:"synopsis,omitempty"`        // Résumé de l'intrigue (optionnel)
	Director       string   `json:"director,omitempty"`        // Réalisateur (optionnel)
	Producer       string   `json:"producer,omitempty"`        // Producteur (optionnel)
	Languages      string   `json:"languages,omitempty"`       // Langues parlées (optionnel)
	Classification string   `json:"classification,omitempty"`  // Classification d'âge (optionnelle)
}

// TmdbMovieDetails représente la réponse détaillée de l'API TMDB pour un film spécifique
// (endpoint /movie/{id}). Cette structure contient toutes les informations disponibles
// sur un film, y compris les genres complets, les langues parlées et les sociétés de production.
type TmdbMovieDetails struct {
	ID           int    `json:"id"`            // Identifiant unique TMDB
	Title        string `json:"title"`         // Titre du film
	PosterPath   string `json:"poster_path"`   // Chemin relatif de l'affiche
	BackdropPath string `json:"backdrop_path"` // Chemin relatif de l'image de fond
	ReleaseDate  string `json:"release_date"`  // Date de sortie au format "YYYY-MM-DD"
	ImdbId       string `json:"imdb_id"`       // Identifiant IMDB
	Genres       []struct {
		ID   int    `json:"id"`   // Identifiant du genre
		Name string `json:"name"` // Nom du genre
	} `json:"genres"` // Liste complète des genres (avec ID et nom)
	VoteAverage     float64 `json:"vote_average"` // Note moyenne des votes
	Runtime         int     `json:"runtime"`      // Durée en minutes
	Overview        string  `json:"overview"`     // Synopsis du film
	SpokenLanguages []struct {
		Name string `json:"name"` // Nom de la langue
	} `json:"spoken_languages"` // Langues parlées dans le film
	ProductionCompanies []struct {
		Name string `json:"name"` // Nom de la société de production
	} `json:"production_companies"` // Sociétés de production
}

// CastMemberDTO représente un membre du casting dans la réponse de l'API StreamFlix.
// Il est utilisé pour afficher la fiche individuelle d'un acteur.
type CastMemberDTO struct {
	ID    int    `json:"id"`    // Identifiant unique de l'acteur
	Name  string `json:"name"`  // Nom complet de l'acteur
	Role  string `json:"role"`  // Rôle ou personnage interprété
	Image string `json:"image"` // URL de la photo de profil
}

// CastMemberMoviesDTO représente un membre du casting dans le contexte des crédits d'un film.
// Contrairement à CastMemberDTO, cette structure n'inclut pas l'identifiant de l'acteur
// et est utilisée spécifiquement dans MovieCreditsDTO.
type CastMemberMoviesDTO struct {
	Name  string `json:"name"`  // Nom complet de l'acteur
	Role  string `json:"role"`  // Rôle ou personnage interprété
	Image string `json:"image"` // URL de la photo de profil
}

// MovieCreditsDTO regroupe les principaux crédits d'un film : réalisateur, producteur,
// scénariste et casting. Cette structure est retournée par l'API StreamFlix
// et est construite à partir des données de l'endpoint TMDB /movie/{id}/credits.
type MovieCreditsDTO struct {
	Director string                `json:"director"` // Nom du réalisateur principal
	Producer string                `json:"producer"`  // Nom du producteur principal
	Writer   string                `json:"writer"`    // Nom du scénariste principal
	Cast     []CastMemberMoviesDTO `json:"cast"`      // Liste des acteurs principaux
}

// TmdbMovieCredits représente la réponse brute de l'API TMDB pour les crédits
// d'un film (endpoint /movie/{id}/credits). Elle contient le casting complet
// et l'équipe technique sous forme de tableaux de structures anonymes.
type TmdbMovieCredits struct {
	Cast []struct {
		Name        string `json:"name"`         // Nom de l'acteur
		Character   string `json:"character"`    // Personnage interprété
		ProfilePath string `json:"profile_path"` // Chemin relatif de la photo de profil
	} `json:"cast"` // Liste des acteurs
	Crew []struct {
		Name        string `json:"name"`         // Nom du membre de l'équipe
		Job         string `json:"job"`          // Poste occupé (Director, Producer, etc.)
		ProfilePath string `json:"profile_path"` // Chemin relatif de la photo de profil
	} `json:"crew"` // Liste de l'équipe technique
}

// TmdbSearchResult représente la réponse de l'API TMDB pour une recherche de films.
// Elle contient les résultats de recherche avec les informations minimales
// nécessaires pour identifier un film.
type TmdbSearchResult struct {
	Results []struct {
		Id     int    `json:"id"`      // Identifiant TMDB du film
		Title  string `json:"title"`   // Titre du film
		ImdbId string `json:"imdb_id"` // Identifiant IMDB
	} `json:"results"` // Liste des résultats de recherche
}

// TmdbMovieImdbId représente la réponse partielle de l'API TMDB contenant
// uniquement l'identifiant IMDB d'un film. Utilisé pour les requêtes
// qui ne nécessitent que l'ID IMDB (par exemple pour Torrentio).
type TmdbMovieImdbId struct {
	ImdbId string `json:"imdb_id"` // Identifiant IMDB au format "tt" suivi de chiffres
}

// TMDBMovieRaw représente un film brut tel que retourné par les endpoints
// de liste de l'API TMDB (discover, search, etc.). Cette structure contient
// les champs essentiels avant transformation en MovieDTO.
type TMDBMovieRaw struct {
	ID          int     `json:"id"`           // Identifiant unique TMDB
	Title       string  `json:"title"`        // Titre du film
	PosterPath  string  `json:"poster_path"`  // Chemin relatif de l'affiche
	ReleaseDate string  `json:"release_date"` // Date de sortie au format "YYYY-MM-DD"
	GenreIDs    []int   `json:"genre_ids"`    // Identifiants numériques des genres
	VoteAverage float64 `json:"vote_average"` // Note moyenne des votes
}

// TMDBResponse représente une réponse paginée générique de l'API TMDB
// contenant une liste de films bruts. Utilisée pour les endpoints
// qui retournent uniquement la liste des résultats sans métadonnées de pagination.
type TMDBResponse struct {
	Results []TMDBMovieRaw `json:"results"` // Liste des films retournés
}

// TMDBTrendingMovieRaw représente un film brut provenant de l'endpoint
// des tendances TMDB (/trending/movie/{time_window}). La structure est
// identique à TMDBMovieRaw mais distincte pour la clarté sémantique.
type TMDBTrendingMovieRaw struct {
	ID          int     `json:"id"`           // Identifiant unique TMDB
	Title       string  `json:"title"`        // Titre du film
	PosterPath  string  `json:"poster_path"`  // Chemin relatif de l'affiche
	ReleaseDate string  `json:"release_date"` // Date de sortie au format "YYYY-MM-DD"
	GenreIDs    []int   `json:"genre_ids"`    // Identifiants numériques des genres
	VoteAverage float64 `json:"vote_average"` // Note moyenne des votes
}

// TMDBTrendingResponse représente la réponse paginée de l'endpoint TMDB
// des films tendance. Elle inclut les métadonnées de pagination
// (page courante, nombre total de pages et de résultats).
type TMDBTrendingResponse struct {
	Page         int                    `json:"page"`          // Numéro de la page courante
	Results      []TMDBTrendingMovieRaw `json:"results"`       // Liste des films tendance
	TotalPages   int                    `json:"total_pages"`   // Nombre total de pages disponibles
	TotalResults int                    `json:"total_results"` // Nombre total de résultats
}

// TMDBDiscoverResponse représente la réponse paginée de l'endpoint TMDB
// discover/movie. Elle permet de parcourir les films par genre, année,
// note, etc. avec pagination complète.
type TMDBDiscoverResponse struct {
	Page         int            `json:"page"`          // Numéro de la page courante
	Results      []TMDBMovieRaw `json:"results"`       // Liste des films découverts
	TotalPages   int            `json:"total_pages"`   // Nombre total de pages disponibles
	TotalResults int            `json:"total_results"` // Nombre total de résultats
}

// SearchMoviesParams regroupe les paramètres de recherche avancée de films.
// Cette structure est utilisée par le service de recherche pour construire
// la requête vers l'API TMDB avec tous les filtres possibles.
type SearchMoviesParams struct {
	BearerToken string  // Token d'authentification pour l'API TMDB
	Query       string  // Texte de recherche saisi par l'utilisateur
	GenresCSV   string  // Liste de genres séparés par des virgules (par exemple "28,35")
	YearsCSV    string  // Liste d'années séparées par des virgules
	SortBy      string  // Critère de tri (par exemple "popularity.desc")
	Page        int     // Numéro de page pour la pagination
	Language    string  // Code de langue pour la localisation (par exemple "fr-FR")
	Rating      float64 // Note minimale pour filtrer les résultats
}

// FetchParams regroupe les paramètres internes pour une requête de récupération
// de films. Cette structure est utilisée après le traitement des SearchMoviesParams
// pour déterminer quel endpoint TMDB appeler (search vs discover).
type FetchParams struct {
	BearerToken string // Token d'authentification pour l'API TMDB
	Query       string // Texte de recherche
	Genre       string // Identifiant du genre unique
	Year        int    // Année de sortie
	SortBy      string // Critère de tri
	Page        int    // Numéro de page
	Language    string // Code de langue
	HasQuery    bool   // Indique si une recherche textuelle est active
	HasGenres   bool   // Indique si un filtre par genre est actif
}

// CachedSearch représente un résultat de recherche mis en cache.
// Il associe les résultats bruts TMDB à une date d'expiration
// pour gérer automatiquement l'invalidation du cache.
type CachedSearch struct {
	Results   []TMDBMovieRaw // Résultats de recherche TMDB mis en cache
	ExpiresAt time.Time      // Date et heure d'expiration de cette entrée
}

// TMDBGenre représente un genre tel que retourné par l'API TMDB
// dans l'endpoint /genre/movie/list. Il contient l'identifiant
// numérique et le nom localisé du genre.
type TMDBGenre struct {
	ID   int    `json:"id"`   // Identifiant numérique du genre
	Name string `json:"name"` // Nom du genre (localisé selon la langue de la requête)
}

// TMDBGenreMovieListResponse représente la réponse de l'API TMDB
// pour l'endpoint /genre/movie/list. Elle contient la liste complète
// des genres de films disponibles sur TMDB.
type TMDBGenreMovieListResponse struct {
	Genres []TMDBGenre `json:"genres"` // Liste de tous les genres de films
}

// CategoryDTO est le Data Transfer Object pour une catégorie de contenu.
// Il est retourné par l'API StreamFlix et représente une catégorie
// navigable dans l'interface (par exemple "Action", "Comédie").
type CategoryDTO struct {
	ID           int      `json:"id"`                    // Identifiant TMDB du genre
	CategoryName string   `json:"categoryName"`          // Nom de la catégorie affiché
	Description  string   `json:"description"`           // Description de la catégorie
	Href         string   `json:"href"`                  // Lien de navigation
	Color        string   `json:"color"`                 // Classe CSS de couleur Tailwind
	Previews     []string `json:"previews,omitempty"`    // URLs des images d'aperçu (optionnel)
}

// ============================================================================
// TV MODELS
// ============================================================================

// TMDBTrendingTVResponse représente la réponse paginée de l'endpoint TMDB
// des séries TV tendance (/trending/tv/{time_window}).
type TMDBTrendingTVResponse struct {
	Page    int             `json:"page"`    // Numéro de la page courante
	Results []TMDBTVRawItem `json:"results"` // Liste des séries tendance
}

// TMDBTVRawItem représente une série TV brute provenant de l'endpoint
// des tendances TMDB. Elle inclut des informations supplémentaires
// comme la langue originale et le pays d'origine.
type TMDBTVRawItem struct {
	ID               int      `json:"id"`                // Identifiant unique TMDB
	Name             string   `json:"name"`              // Nom de la série
	FirstAirDate     string   `json:"first_air_date"`    // Date de première diffusion au format "YYYY-MM-DD"
	GenreIDs         []int    `json:"genre_ids"`         // Identifiants numériques des genres
	PosterPath       string   `json:"poster_path"`       // Chemin relatif de l'affiche
	VoteAverage      float64  `json:"vote_average"`      // Note moyenne des votes
	OriginalLanguage string   `json:"original_language"` // Code ISO de la langue originale
	OriginCountry    []string `json:"origin_country"`    // Codes ISO des pays d'origine
}

// TVDTO est le Data Transfer Object pour les séries TV avec informations de langue et pays.
// Il est utilisé pour les listes de séries tendance où les métadonnées régionales sont pertinentes.
type TVDTO struct {
	ID       int      `json:"id"`       // Identifiant TMDB de la série
	Name     string   `json:"name"`     // Nom de la série
	Image    string   `json:"image"`    // URL complète de l'affiche
	Year     int      `json:"year"`     // Année de première diffusion
	Genres   []string `json:"genres"`   // Noms des genres traduits
	Rating   float64  `json:"rating"`   // Note moyenne
	Language string   `json:"language"` // Langue originale
	Country  []string `json:"country"`  // Pays d'origine
}

// TMDBTVRaw représente une série TV brute provenant de l'endpoint discover/tv de TMDB.
// Cette structure est utilisée pour les requêtes de découverte de séries TV.
type TMDBTVRaw struct {
	ID           int     `json:"id"`             // Identifiant unique TMDB
	Name         string  `json:"name"`           // Nom de la série
	PosterPath   string  `json:"poster_path"`    // Chemin relatif de l'affiche
	FirstAirDate string  `json:"first_air_date"` // Date de première diffusion
	GenreIDs     []int   `json:"genre_ids"`      // Identifiants numériques des genres
	VoteAverage  float64 `json:"vote_average"`   // Note moyenne
	Overview     string  `json:"overview"`       // Synopsis de la série
}

// TMDBDiscoverTVResponse représente la réponse paginée de l'endpoint TMDB
// discover/tv. Elle permet de parcourir les séries TV avec filtres.
type TMDBDiscoverTVResponse struct {
	Page    int         `json:"page"`    // Numéro de la page courante
	Results []TMDBTVRaw `json:"results"` // Liste des séries découvertes
}

// Tvdto est un Data Transfer Object simplifié pour les séries TV.
// Il contient les informations essentielles sans les métadonnées régionales,
// utilisé dans les contextes où la langue et le pays ne sont pas nécessaires.
type Tvdto struct {
	ID     int      `json:"id"`     // Identifiant TMDB de la série
	Name   string   `json:"name"`   // Nom de la série
	Image  string   `json:"image"`  // URL complète de l'affiche
	Year   int      `json:"year"`   // Année de première diffusion
	Genres []string `json:"genres"` // Noms des genres traduits
	Rating float64  `json:"rating"` // Note moyenne
}

// TMDBTVShowRaw représente une série TV brute avec des champs étendus
// incluant l'image de fond et la langue originale. Cette structure est utilisée
// pour les listes de séries nécessitant plus de détails visuels.
type TMDBTVShowRaw struct {
	ID               int     `json:"id"`                // Identifiant unique TMDB
	Name             string  `json:"name"`              // Nom de la série
	PosterPath       string  `json:"poster_path"`       // Chemin relatif de l'affiche
	BackdropPath     string  `json:"backdrop_path"`     // Chemin relatif de l'image de fond
	FirstAirDate     string  `json:"first_air_date"`    // Date de première diffusion
	GenreIDs         []int   `json:"genre_ids"`         // Identifiants numériques des genres
	VoteAverage      float64 `json:"vote_average"`      // Note moyenne
	Overview         string  `json:"overview"`          // Synopsis de la série
	OriginalLanguage string  `json:"original_language"` // Code ISO de la langue originale
}

// TMDBTVShowResponse représente la réponse paginée de l'API TMDB pour les listes
// de séries TV avec métadonnées de pagination complètes.
type TMDBTVShowResponse struct {
	Page         int             `json:"page"`          // Numéro de la page courante
	Results      []TMDBTVShowRaw `json:"results"`       // Liste des séries TV
	TotalPages   int             `json:"total_pages"`   // Nombre total de pages
	TotalResults int             `json:"total_results"` // Nombre total de résultats
}

// TVShowDTO est le Data Transfer Object enrichi pour les séries TV.
// Il est utilisé pour les listes de séries où l'image de fond, le synopsis
// et la langue sont affichés (par exemple la page d'accueil).
type TVShowDTO struct {
	ID       int      `json:"id"`       // Identifiant TMDB de la série
	Title    string   `json:"title"`    // Nom de la série
	Image    string   `json:"image"`    // URL complète de l'affiche
	Backdrop string   `json:"backdrop"` // URL complète de l'image de fond
	Year     string   `json:"year"`     // Année de première diffusion (sous forme de chaîne)
	Genres   []string `json:"genres"`   // Noms des genres traduits
	Rating   float64  `json:"rating"`   // Note moyenne
	Overview string   `json:"overview"` // Synopsis de la série
	Language string   `json:"language"` // Langue originale
}

// TMDBTVDetailsRaw représente la réponse détaillée de l'API TMDB pour une série TV
// spécifique (endpoint /tv/{id} avec append_to_response). Cette structure contient
// toutes les informations disponibles : saisons, crédits, séries similaires,
// classifications et réseaux de diffusion.
type TMDBTVDetailsRaw struct {
	ID              int                  `json:"id"`               // Identifiant unique TMDB
	Name            string               `json:"name"`             // Nom de la série
	Overview        string               `json:"overview"`         // Synopsis
	PosterPath      string               `json:"poster_path"`      // Chemin relatif de l'affiche
	BackdropPath    string               `json:"backdrop_path"`    // Chemin relatif de l'image de fond
	FirstAirDate    string               `json:"first_air_date"`   // Date de première diffusion
	VoteAverage     float64              `json:"vote_average"`     // Note moyenne
	EpisodeRunTime  []int                `json:"episode_run_time"` // Durées typiques d'un épisode en minutes
	Genres          []TMDBGenre          `json:"genres"`           // Liste des genres complets
	CreatedBy       []TMDBCreatedBy      `json:"created_by"`       // Créateurs de la série
	Networks        []TMDBNetwork        `json:"networks"`         // Réseaux de diffusion
	SpokenLanguages []TMDBSpokenLanguage `json:"spoken_languages"` // Langues parlées
	ContentRatings  TMDBContentRatings   `json:"content_ratings"`  // Classifications par pays
	Seasons         []TMDBSeasonRaw      `json:"seasons"`          // Liste des saisons
	Credits         TMDBCredits          `json:"credits"`          // Casting de la série
	Similar         TMDBSimilarTV        `json:"similar"`          // Séries similaires
}

// TMDBCreatedBy représente un créateur de série TV tel que retourné par l'API TMDB.
type TMDBCreatedBy struct {
	Name string `json:"name"` // Nom du créateur
}

// TMDBNetwork représente un réseau de diffusion (chaîne TV ou plateforme de streaming)
// tel que retourné par l'API TMDB.
type TMDBNetwork struct {
	Name string `json:"name"` // Nom du réseau (par exemple "Netflix", "HBO")
}

// TMDBSpokenLanguage représente une langue parlée dans une série TV.
type TMDBSpokenLanguage struct {
	EnglishName string `json:"english_name"` // Nom de la langue en anglais
}

// TMDBContentRatings contient les classifications de contenu par pays pour une série TV.
type TMDBContentRatings struct {
	Results []TMDBContentRating `json:"results"` // Liste des classifications par pays
}

// TMDBContentRating représente la classification de contenu d'une série pour un pays donné.
type TMDBContentRating struct {
	ISO31661 string `json:"iso_3166_1"` // Code ISO 3166-1 du pays (par exemple "FR", "US")
	Rating   string `json:"rating"`     // Classification (par exemple "TV-MA", "-16")
}

// TMDBSeasonRaw représente une saison brute d'une série TV retournée par l'API TMDB.
type TMDBSeasonRaw struct {
	SeasonNumber int    `json:"season_number"` // Numéro de la saison (0 pour les spéciaux)
	EpisodeCount int    `json:"episode_count"` // Nombre d'épisodes dans la saison
	AirDate      string `json:"air_date"`      // Date de diffusion de la saison
	PosterPath   string `json:"poster_path"`   // Chemin relatif de l'affiche de la saison
}

// TMDBCredits contient le casting d'une série TV tel que retourné par l'API TMDB.
type TMDBCredits struct {
	Cast []TMDBCastMember `json:"cast"` // Liste des acteurs de la série
}

// TMDBCastMember représente un acteur dans le casting d'une série ou d'un film TMDB.
type TMDBCastMember struct {
	ID          int    `json:"id"`           // Identifiant unique de l'acteur
	Name        string `json:"name"`         // Nom complet de l'acteur
	Character   string `json:"character"`    // Personnage interprété
	ProfilePath string `json:"profile_path"` // Chemin relatif de la photo de profil
	Order       int    `json:"order"`        // Ordre d'apparition dans le casting
}

// TMDBSimilarTV contient la liste des séries TV similaires retournées par l'API TMDB.
type TMDBSimilarTV struct {
	Results []TMDBSimilarTVItem `json:"results"` // Liste des séries similaires
}

// TMDBSimilarTVItem représente une série TV similaire dans les résultats TMDB.
type TMDBSimilarTVItem struct {
	ID           int     `json:"id"`             // Identifiant unique TMDB
	Name         string  `json:"name"`           // Nom de la série
	PosterPath   string  `json:"poster_path"`    // Chemin relatif de l'affiche
	BackdropPath string  `json:"backdrop_path"`  // Chemin relatif de l'image de fond
	FirstAirDate string  `json:"first_air_date"` // Date de première diffusion
	GenreIDs     []int   `json:"genre_ids"`      // Identifiants numériques des genres
	VoteAverage  float64 `json:"vote_average"`   // Note moyenne
}

// TMDBSeasonDetailsRaw représente les détails d'une saison spécifique d'une série TV,
// incluant la liste complète de ses épisodes. Retourné par l'endpoint /tv/{id}/season/{num}.
type TMDBSeasonDetailsRaw struct {
	SeasonNumber int              `json:"season_number"` // Numéro de la saison
	Episodes     []TMDBEpisodeRaw `json:"episodes"`      // Liste des épisodes de la saison
}

// TMDBEpisodeRaw représente un épisode brut d'une série TV tel que retourné par l'API TMDB.
type TMDBEpisodeRaw struct {
	EpisodeNumber int     `json:"episode_number"` // Numéro de l'épisode dans la saison
	Name          string  `json:"name"`           // Titre de l'épisode
	Overview      string  `json:"overview"`       // Synopsis de l'épisode
	AirDate       string  `json:"air_date"`       // Date de diffusion
	Runtime       int     `json:"runtime"`        // Durée en minutes
	StillPath     string  `json:"still_path"`     // Chemin relatif de l'image de l'épisode
	VoteAverage   float64 `json:"vote_average"`   // Note moyenne de l'épisode
}

// TVDetailsDTO est le Data Transfer Object pour les détails complets d'une série TV.
// Il est retourné par l'API StreamFlix et contient les informations principales
// de la série, transformées depuis les données brutes TMDB.
type TVDetailsDTO struct {
	ID             int             `json:"id"`             // Identifiant TMDB de la série
	Name           string          `json:"name"`           // Nom de la série
	Image          string          `json:"image"`          // URL complète de l'affiche
	BackdropImage  string          `json:"backdropImage"`  // URL complète de l'image de fond
	Year           int             `json:"year"`           // Année de première diffusion
	Rating         float64         `json:"rating"`         // Note moyenne
	EpisodeRuntime int             `json:"episodeRuntime"` // Durée typique d'un épisode en minutes
	Genres         []string        `json:"genres"`         // Noms des genres
	Synopsis       string          `json:"synopsis"`       // Synopsis de la série
	CreatedBy      string          `json:"createdBy"`      // Noms des créateurs (séparés par des virgules)
	Networks       string          `json:"networks"`       // Noms des réseaux de diffusion
	Languages      string          `json:"languages"`      // Langues parlées dans la série
	Classification string          `json:"classification"` // Classification d'âge
	Cast           []CastMemberDTO `json:"cast"`           // Casting principal
}

// SeasonDTO est le Data Transfer Object pour une saison de série TV.
// Il contient les métadonnées de la saison et la liste de ses épisodes.
type SeasonDTO struct {
	SeasonNumber int          `json:"seasonNumber"` // Numéro de la saison
	EpisodeCount int          `json:"episodeCount"` // Nombre total d'épisodes
	Year         int          `json:"year"`         // Année de diffusion de la saison
	Episodes     []EpisodeDTO `json:"episodes"`     // Liste des épisodes
}

// EpisodeDTO est le Data Transfer Object pour un épisode de série TV.
type EpisodeDTO struct {
	EpisodeNumber int    `json:"episodeNumber"` // Numéro de l'épisode
	Name          string `json:"name"`          // Titre de l'épisode
	Overview      string `json:"overview"`      // Synopsis de l'épisode
	AirDate       string `json:"airDate"`       // Date de diffusion
	Runtime       int    `json:"runtime"`       // Durée en minutes
	Still         string `json:"still"`         // URL de l'image de l'épisode
}

// SimilarTVDTO est le Data Transfer Object pour une série TV similaire.
// Il est utilisé dans la section "Séries similaires" de la page de détail.
type SimilarTVDTO struct {
	ID     int      `json:"id"`     // Identifiant TMDB de la série
	Title  string   `json:"title"`  // Nom de la série
	Image  string   `json:"image"`  // URL complète de l'affiche
	Poster string   `json:"poster"` // URL alternative de l'affiche
	Year   int      `json:"year"`   // Année de première diffusion
	Genres []string `json:"genres"` // Noms des genres
	Rating float64  `json:"rating"` // Note moyenne
}

// TVInfoResponse est la réponse complète de l'API StreamFlix pour les détails d'une série TV.
// Elle regroupe toutes les informations nécessaires à l'affichage de la page de détail :
// les données principales de la série, les saisons avec leurs épisodes,
// le casting complet et les séries similaires.
type TVInfoResponse struct {
	ContentData  TVDetailsDTO    `json:"contentData"`  // Détails principaux de la série
	Seasons      []SeasonDTO     `json:"seasons"`      // Liste des saisons avec épisodes
	Credits      []CastMemberDTO `json:"credits"`      // Casting complet
	SimilarItems []SimilarTVDTO  `json:"similarItems"` // Séries similaires recommandées
}

// TVSearchResult représente un résultat de recherche de série TV
// provenant de l'API TMDB. Cette structure contient les champs bruts
// nécessaires pour construire un TVSearchDTO.
type TVSearchResult struct {
	ID           int     `json:"id"`             // Identifiant unique TMDB
	Name         string  `json:"name"`           // Nom de la série
	PosterPath   string  `json:"poster_path"`    // Chemin relatif de l'affiche
	FirstAirDate string  `json:"first_air_date"` // Date de première diffusion
	GenreIDs     []int   `json:"genre_ids"`      // Identifiants numériques des genres
	VoteAverage  float64 `json:"vote_average"`   // Note moyenne
}

// TVSearchResponse représente la réponse de l'API TMDB pour une recherche de séries TV.
type TVSearchResponse struct {
	Results []TVSearchResult `json:"results"` // Liste des résultats de recherche
}

// TVSearchDTO est le Data Transfer Object pour un résultat de recherche de série TV.
// Il est retourné par l'API StreamFlix après transformation des données brutes TMDB.
type TVSearchDTO struct {
	ID     int      `json:"id"`     // Identifiant TMDB de la série
	Title  string   `json:"title"`  // Nom de la série
	Image  string   `json:"image"`  // URL complète de l'affiche
	Year   int      `json:"year"`   // Année de première diffusion
	Genre  []string `json:"genre"`  // Noms des genres traduits
	Rating float64  `json:"rating"` // Note moyenne
}
