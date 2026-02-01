package models

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

// --- TMDB shapes (champs courants de /3/movie/popular) ---

type TmdbPopularResponse struct {
	Results []tmdbMovie `json:"results"`
}

type tmdbMovie struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	ReleaseDate  string  `json:"release_date"` // "YYYY-MM-DD"
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
	Runtime         int     `json:"runtime"` // en minutes
	Overview        string  `json:"overview"`
	SpokenLanguages []struct {
		Name string `json:"name"`
	} `json:"spoken_languages"`
	ProductionCompanies []struct {
		Name string `json:"name"`
	} `json:"production_companies"`
}

type CastMemberDTO struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Image string `json:"image"`
}

type MovieCreditsDTO struct {
	Director string          `json:"director"`
	Producer string          `json:"producer"`
	Writer   string          `json:"writer"`
	Cast     []CastMemberDTO `json:"cast"`
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

// TmdbSearchResult représente un résultat de recherche TMDB
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
