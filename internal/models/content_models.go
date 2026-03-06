package models

import (
	"sync"
	"time"
)

// ============================================================================
// CACHE
// ============================================================================

var (
	SearchCache = make(map[string]CachedSearch)
	CacheMutex  sync.RWMutex
	CacheTTL    = 5 * time.Minute
)

type MovieGenreCacheKey struct {
	GenreID  int
	Page     int
	Language string
}
type popularMoviesCacheKey struct {
	page     string
	language string
}

type topRatedMoviesCacheKey struct {
	page int
}

type trendingMoviesCacheKey struct {
	timeWindow string
	page       int
	language   string
}

type ContentDetailsCacheKey struct {
	movieID int
}

type similarMoviesCacheKey struct {
	movieID int
	page    string
}

type movieCreditsCacheKey struct {
	movieID int
}

type movieImdbIDCacheKey struct {
	movieID int
}

type genreCategoriesCacheKey struct {
	language string
}

type Cache[K comparable, V any] struct {
	Mu    sync.RWMutex
	Store map[K]Entry[V]
	Ttl   time.Duration
}
type Entry[V any] struct {
	Data      V
	ExpiresAt time.Time
}

// ============================================================================
// CONTENT & MEDIA MODELS
// ============================================================================

type Movie struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Year     int      `json:"year"`
	Rating   float32  `json:"rating"`
	Genre    []string `json:"genre"`
	ImageURL string   `json:"image"`
}

type Categories struct {
	ID           int      `json:"id"`
	CategoryName string   `json:"categoryName"`
	Description  string   `json:"description"`
	Href         string   `json:"href"`
	Color        string   `json:"color"`
	Previews     []string `json:"previews"`
}

type Cast struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Image string `json:"image"`
}

type ContentDetails struct {
	ID             int      `json:"id"`
	Title          string   `json:"title"`
	Image          string   `json:"image"`
	BackdropImage  string   `json:"backdropImage"`
	Year           int      `json:"year"`
	Genres         []string `json:"genres"`
	Rating         float32  `json:"rating"`
	Duration       string   `json:"duration"`
	Synopsis       string   `json:"synopsis"`
	Director       string   `json:"director"`
	Producer       string   `json:"producer"`
	Languages      string   `json:"languages"`
	Classification string   `json:"classification"`
	Cast           []Cast   `json:"cast"`
}

type MovieDTO struct {
	ID     int      `json:"id"`
	Title  string   `json:"title"`
	Image  string   `json:"image"`
	Year   int      `json:"year"`
	Genre  []string `json:"genre"`
	Rating float64  `json:"rating"`
}

type TmdbPopularResponse struct {
	Results []tmdbMovie `json:"results"`
}

type tmdbMovie struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	ReleaseDate  string  `json:"release_date"`
	GenreIDs     []int   `json:"genre_ids"`
	VoteAverage  float64 `json:"vote_average"`
}

type ContentDetailsDTO struct {
	ID             int      `json:"id"`
	Title          string   `json:"title"`
	Image          string   `json:"image"`
	Imdbid         string   `json:"imdb_id"`
	BackdropImage  string   `json:"backdropImage,omitempty"`
	Year           int      `json:"year"`
	Genres         []string `json:"genres"`
	Rating         float64  `json:"rating"`
	Duration       string   `json:"duration,omitempty"`
	Synopsis       string   `json:"synopsis,omitempty"`
	Director       string   `json:"director,omitempty"`
	Producer       string   `json:"producer,omitempty"`
	Languages      string   `json:"languages,omitempty"`
	Classification string   `json:"classification,omitempty"`
}

type TmdbMovieDetails struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
	ReleaseDate  string `json:"release_date"`
	ImdbId       string `json:"imdb_id"`
	Genres       []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	VoteAverage     float64 `json:"vote_average"`
	Runtime         int     `json:"runtime"`
	Overview        string  `json:"overview"`
	SpokenLanguages []struct {
		Name string `json:"name"`
	} `json:"spoken_languages"`
	ProductionCompanies []struct {
		Name string `json:"name"`
	} `json:"production_companies"`
}

type CastMemberDTO struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Role  string `json:"role"`
	Image string `json:"image"`
}

type CastMemberMoviesDTO struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Image string `json:"image"`
}

type MovieCreditsDTO struct {
	Director string                `json:"director"`
	Producer string                `json:"producer"`
	Writer   string                `json:"writer"`
	Cast     []CastMemberMoviesDTO `json:"cast"`
}

type TmdbMovieCredits struct {
	Cast []struct {
		Name        string `json:"name"`
		Character   string `json:"character"`
		ProfilePath string `json:"profile_path"`
	} `json:"cast"`
	Crew []struct {
		Name        string `json:"name"`
		Job         string `json:"job"`
		ProfilePath string `json:"profile_path"`
	} `json:"crew"`
}

type TmdbSearchResult struct {
	Results []struct {
		Id     int    `json:"id"`
		Title  string `json:"title"`
		ImdbId string `json:"imdb_id"`
	} `json:"results"`
}

type TmdbMovieImdbId struct {
	ImdbId string `json:"imdb_id"`
}

type TMDBMovieRaw struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	GenreIDs    []int   `json:"genre_ids"`
	VoteAverage float64 `json:"vote_average"`
}

type TMDBResponse struct {
	Results []TMDBMovieRaw `json:"results"`
}

type TMDBTrendingMovieRaw struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	GenreIDs    []int   `json:"genre_ids"`
	VoteAverage float64 `json:"vote_average"`
}

type TMDBTrendingResponse struct {
	Page         int                    `json:"page"`
	Results      []TMDBTrendingMovieRaw `json:"results"`
	TotalPages   int                    `json:"total_pages"`
	TotalResults int                    `json:"total_results"`
}

type TMDBDiscoverResponse struct {
	Page         int            `json:"page"`
	Results      []TMDBMovieRaw `json:"results"`
	TotalPages   int            `json:"total_pages"`
	TotalResults int            `json:"total_results"`
}

type SearchMoviesParams struct {
	BearerToken string
	Query       string
	GenresCSV   string
	YearsCSV    string
	SortBy      string
	Page        int
	Language    string
	Rating      float64
}

type FetchParams struct {
	BearerToken string
	Query       string
	Genre       string
	Year        int
	SortBy      string
	Page        int
	Language    string
	HasQuery    bool
	HasGenres   bool
}

type CachedSearch struct {
	Results   []TMDBMovieRaw
	ExpiresAt time.Time
}

type TMDBGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TMDBGenreMovieListResponse struct {
	Genres []TMDBGenre `json:"genres"`
}

type CategoryDTO struct {
	ID           int      `json:"id"`
	CategoryName string   `json:"categoryName"`
	Description  string   `json:"description"`
	Href         string   `json:"href"`
	Color        string   `json:"color"`
	Previews     []string `json:"previews,omitempty"`
}

// ============================================================================
// TV MODELS
// ============================================================================

type TMDBTrendingTVResponse struct {
	Page    int             `json:"page"`
	Results []TMDBTVRawItem `json:"results"`
}

type TMDBTVRawItem struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	FirstAirDate     string   `json:"first_air_date"`
	GenreIDs         []int    `json:"genre_ids"`
	PosterPath       string   `json:"poster_path"`
	VoteAverage      float64  `json:"vote_average"`
	OriginalLanguage string   `json:"original_language"`
	OriginCountry    []string `json:"origin_country"`
}

type TVDTO struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Image    string   `json:"image"`
	Year     int      `json:"year"`
	Genres   []string `json:"genres"`
	Rating   float64  `json:"rating"`
	Language string   `json:"language"`
	Country  []string `json:"country"`
}

type TMDBTVRaw struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	PosterPath   string  `json:"poster_path"`
	FirstAirDate string  `json:"first_air_date"`
	GenreIDs     []int   `json:"genre_ids"`
	VoteAverage  float64 `json:"vote_average"`
	Overview     string  `json:"overview"`
}

type TMDBDiscoverTVResponse struct {
	Page    int         `json:"page"`
	Results []TMDBTVRaw `json:"results"`
}

type Tvdto struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Image  string   `json:"image"`
	Year   int      `json:"year"`
	Genres []string `json:"genres"`
	Rating float64  `json:"rating"`
}

type TMDBTVShowRaw struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	FirstAirDate     string  `json:"first_air_date"`
	GenreIDs         []int   `json:"genre_ids"`
	VoteAverage      float64 `json:"vote_average"`
	Overview         string  `json:"overview"`
	OriginalLanguage string  `json:"original_language"`
}

type TMDBTVShowResponse struct {
	Page         int             `json:"page"`
	Results      []TMDBTVShowRaw `json:"results"`
	TotalPages   int             `json:"total_pages"`
	TotalResults int             `json:"total_results"`
}

type TVShowDTO struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Image    string   `json:"image"`
	Backdrop string   `json:"backdrop"`
	Year     string   `json:"year"`
	Genres   []string `json:"genres"`
	Rating   float64  `json:"rating"`
	Overview string   `json:"overview"`
	Language string   `json:"language"`
}

type TMDBTVDetailsRaw struct {
	ID              int                  `json:"id"`
	Name            string               `json:"name"`
	Overview        string               `json:"overview"`
	PosterPath      string               `json:"poster_path"`
	BackdropPath    string               `json:"backdrop_path"`
	FirstAirDate    string               `json:"first_air_date"`
	VoteAverage     float64              `json:"vote_average"`
	EpisodeRunTime  []int                `json:"episode_run_time"`
	Genres          []TMDBGenre          `json:"genres"`
	CreatedBy       []TMDBCreatedBy      `json:"created_by"`
	Networks        []TMDBNetwork        `json:"networks"`
	SpokenLanguages []TMDBSpokenLanguage `json:"spoken_languages"`
	ContentRatings  TMDBContentRatings   `json:"content_ratings"`
	Seasons         []TMDBSeasonRaw      `json:"seasons"`
	Credits         TMDBCredits          `json:"credits"`
	Similar         TMDBSimilarTV        `json:"similar"`
}

type TMDBCreatedBy struct {
	Name string `json:"name"`
}

type TMDBNetwork struct {
	Name string `json:"name"`
}

type TMDBSpokenLanguage struct {
	EnglishName string `json:"english_name"`
}

type TMDBContentRatings struct {
	Results []TMDBContentRating `json:"results"`
}

type TMDBContentRating struct {
	ISO31661 string `json:"iso_3166_1"`
	Rating   string `json:"rating"`
}

type TMDBSeasonRaw struct {
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	AirDate      string `json:"air_date"`
	PosterPath   string `json:"poster_path"`
}

type TMDBCredits struct {
	Cast []TMDBCastMember `json:"cast"`
}

type TMDBCastMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
	Order       int    `json:"order"`
}

type TMDBSimilarTV struct {
	Results []TMDBSimilarTVItem `json:"results"`
}

type TMDBSimilarTVItem struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	FirstAirDate string  `json:"first_air_date"`
	GenreIDs     []int   `json:"genre_ids"`
	VoteAverage  float64 `json:"vote_average"`
}

type TMDBSeasonDetailsRaw struct {
	SeasonNumber int              `json:"season_number"`
	Episodes     []TMDBEpisodeRaw `json:"episodes"`
}

type TMDBEpisodeRaw struct {
	EpisodeNumber int     `json:"episode_number"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	AirDate       string  `json:"air_date"`
	Runtime       int     `json:"runtime"`
	StillPath     string  `json:"still_path"`
	VoteAverage   float64 `json:"vote_average"`
}

type TVDetailsDTO struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	Image          string          `json:"image"`
	BackdropImage  string          `json:"backdropImage"`
	Year           int             `json:"year"`
	Rating         float64         `json:"rating"`
	EpisodeRuntime int             `json:"episodeRuntime"`
	Genres         []string        `json:"genres"`
	Synopsis       string          `json:"synopsis"`
	CreatedBy      string          `json:"createdBy"`
	Networks       string          `json:"networks"`
	Languages      string          `json:"languages"`
	Classification string          `json:"classification"`
	Cast           []CastMemberDTO `json:"cast"`
}

type SeasonDTO struct {
	SeasonNumber int          `json:"seasonNumber"`
	EpisodeCount int          `json:"episodeCount"`
	Year         int          `json:"year"`
	Episodes     []EpisodeDTO `json:"episodes"`
}

type EpisodeDTO struct {
	EpisodeNumber int    `json:"episodeNumber"`
	Name          string `json:"name"`
	Overview      string `json:"overview"`
	AirDate       string `json:"airDate"`
	Runtime       int    `json:"runtime"`
	Still         string `json:"still"`
}

type SimilarTVDTO struct {
	ID     int      `json:"id"`
	Title  string   `json:"title"`
	Image  string   `json:"image"`
	Poster string   `json:"poster"`
	Year   int      `json:"year"`
	Genres []string `json:"genres"`
	Rating float64  `json:"rating"`
}

type TVInfoResponse struct {
	ContentData  TVDetailsDTO    `json:"contentData"`
	Seasons      []SeasonDTO     `json:"seasons"`
	Credits      []CastMemberDTO `json:"credits"`
	SimilarItems []SimilarTVDTO  `json:"similarItems"`
}

type TVSearchResult struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	PosterPath   string  `json:"poster_path"`
	FirstAirDate string  `json:"first_air_date"`
	GenreIDs     []int   `json:"genre_ids"`
	VoteAverage  float64 `json:"vote_average"`
}

type TVSearchResponse struct {
	Results []TVSearchResult `json:"results"`
}

type TVSearchDTO struct {
	ID     int      `json:"id"`
	Title  string   `json:"title"`
	Image  string   `json:"image"`
	Year   int      `json:"year"`
	Genre  []string `json:"genre"`
	Rating float64  `json:"rating"`
}
