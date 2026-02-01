package models

import (
	"sync"
	"time"
)

// ZtParser représente le parser Zone Téléchargement
type ZtParser struct {
	BaseUrl              string
	AllCategories        []string
	LastRequestTimestamp time.Time
	RequestTimeInBetween time.Duration
	DevMode              bool
	MoviesDbKey          string
	Mu                   sync.Mutex
}

// SearchResult représente un résultat de recherche
type SearchResult struct {
	Title              string    `json:"title"`
	Url                string    `json:"url"`
	Id                 string    `json:"id"`
	Image              string    `json:"image"`
	Quality            string    `json:"quality"`
	Language           string    `json:"language"`
	PublishedOn        time.Time `json:"published_on"`
	PublishedTimestamp int64     `json:"published_timestamp"`
}

// MovieDetails représente les détails complets d'un film
type MovieDetails struct {
	Title        string            `json:"title"`
	OriginalName string            `json:"original_name"`
	Language     string            `json:"language"`
	Quality      string            `json:"quality"`
	Size         string            `json:"size"`
	Description  string            `json:"description"`
	Poster       string            `json:"poster"`
	ReleaseDate  string            `json:"release_date"`
	Genres       []string          `json:"genres"`
	VoteAverage  float64           `json:"vote_average"`
	Directors    []string          `json:"directors"`
	Actors       []string          `json:"actors"`
	DownloadLink map[string]string `json:"download_link"`
}

// SeriesDetails représente les détails complets d'une série
type SeriesDetails struct {
	Title        string                       `json:"title"`
	OriginalName string                       `json:"original_name"`
	Language     string                       `json:"language"`
	Quality      string                       `json:"quality"`
	Size         string                       `json:"size"`
	Description  string                       `json:"description"`
	Poster       string                       `json:"poster"`
	ReleaseDate  string                       `json:"release_date"`
	Genres       []string                     `json:"genres"`
	VoteAverage  float64                      `json:"vote_average"`
	Directors    []string                     `json:"directors"`
	Actors       []string                     `json:"actors"`
	DownloadLink map[string]map[string]string `json:"download_link"` // episode -> host -> url
}

// ZtBasicInfo représente les infos de base extraites de ZT
type ZtBasicInfo struct {
	Name         string      `json:"name"`
	OriginalName string      `json:"original_name"`
	Language     string      `json:"language"`
	Quality      string      `json:"quality"`
	Size         string      `json:"size"`
	Links        interface{} `json:"links"` // map[string]string pour films, map[string]map[string]string pour séries
}

// TmdbMovieResponse représente la réponse de l'API TMDB pour la recherche
type TmdbMovieResponse struct {
	Results []TmdbMovie `json:"results"`
}

// TmdbMovie représente un film dans TMDB
type TmdbMovie struct {
	Id            int     `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	Overview      string  `json:"overview"`
	PosterPath    string  `json:"poster_path"`
	ReleaseDate   string  `json:"release_date"`
	VoteAverage   float64 `json:"vote_average"`
	GenreIds      []int   `json:"genre_ids"`
}

// TmdbGenresResponse représente la liste des genres TMDB
type TmdbGenresResponse struct {
	Genres []TmdbGenre `json:"genres"`
}

// TmdbGenre représente un genre TMDB
type TmdbGenre struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// TmdbCreditsResponse représente les crédits d'un film
type TmdbCreditsResponse struct {
	Cast []TmdbCastMember `json:"cast"`
	Crew []TmdbCrewMember `json:"crew"`
}

// TmdbCastMember représente un acteur
type TmdbCastMember struct {
	Name string `json:"name"`
}

// TmdbCrewMember représente un membre de l'équipe
type TmdbCrewMember struct {
	Name string `json:"name"`
	Job  string `json:"job"`
}

// TvmazeSearchResponse représente la réponse de recherche TVMaze
type TvmazeSearchResponse struct {
	Show TvmazeShow `json:"show"`
}

// TvmazeShow représente une série sur TVMaze
type TvmazeShow struct {
	Id        int          `json:"id"`
	Name      string       `json:"name"`
	Summary   string       `json:"summary"`
	Premiered string       `json:"premiered"`
	Genres    []string     `json:"genres"`
	Rating    TvmazeRating `json:"rating"`
	Image     TvmazeImage  `json:"image"`
}

// TvmazeRating représente la note d'une série
type TvmazeRating struct {
	Average float64 `json:"average"`
}

// TvmazeImage représente l'image d'une série
type TvmazeImage struct {
	Original string `json:"original"`
}

// TvmazeCastMember représente un acteur sur TVMaze
type TvmazeCastMember struct {
	Person TvmazePerson `json:"person"`
}

// TvmazeCrewMember représente un membre de l'équipe sur TVMaze
type TvmazeCrewMember struct {
	Type   string       `json:"type"`
	Person TvmazePerson `json:"person"`
}

// TvmazePerson représente une personne sur TVMaze
type TvmazePerson struct {
	Name string `json:"name"`
}

// ZtUrlResponse représente la réponse de l'API d'URL
type ZtUrlResponse struct {
	Url       string `json:"url"`
	Timestamp string `json:"timestamp"`
}

// ErrorResponse représente une réponse d'erreur
type ErrorResponse struct {
	Status bool     `json:"status"`
	Error  string   `json:"error"`
	Stack  []string `json:"stack,omitempty"`
}
