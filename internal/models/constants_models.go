package models

import (
	"StreamflixBackend/internal/cache"
	"time"
)

var Gradients = []string{
	"from-purple-500 to-pink-500",
	"from-blue-500 to-cyan-500",
	"from-green-500 to-teal-500",
	"from-yellow-500 to-orange-500",
	"from-red-500 to-pink-500",
	"from-indigo-500 to-purple-500",
	"from-gray-500 to-slate-500",
}

var MovieGenres = struct {
	ACTION          int
	ADVENTURE       int
	ANIMATION       int
	COMEDY          int
	CRIME           int
	DOCUMENTARY     int
	DRAMA           int
	FAMILY          int
	FANTASY         int
	HISTORY         int
	HORROR          int
	MUSIC           int
	MYSTERY         int
	ROMANCE         int
	SCIENCE_FICTION int
	TV_MOVIE        int
	THRILLER        int
	WAR             int
	WESTERN         int
}{
	ACTION:          28,
	ADVENTURE:       12,
	ANIMATION:       16,
	COMEDY:          35,
	CRIME:           80,
	DOCUMENTARY:     99,
	DRAMA:           18,
	FAMILY:          10751,
	FANTASY:         14,
	HISTORY:         36,
	HORROR:          27,
	MUSIC:           10402,
	MYSTERY:         9648,
	ROMANCE:         10749,
	SCIENCE_FICTION: 878,
	TV_MOVIE:        10770,
	THRILLER:        53,
	WAR:             10752,
	WESTERN:         37,
}

var TVGenres = struct {
	ACTION_ADVENTURE int
	ANIMATION        int
	COMEDY           int
	CRIME            int
	DOCUMENTARY      int
	DRAMA            int
	FAMILY           int
	KIDS             int
	MYSTERY          int
	NEWS             int
	REALITY          int
	SCI_FI_FANTASY   int
	SOAP             int
	TALK             int
	WAR_POLITICS     int
	WESTERN          int
}{
	ACTION_ADVENTURE: 10759,
	ANIMATION:        16,
	COMEDY:           35,
	CRIME:            80,
	DOCUMENTARY:      99,
	DRAMA:            18,
	FAMILY:           10751,
	KIDS:             10762,
	MYSTERY:          9648,
	NEWS:             10763,
	REALITY:          10764,
	SCI_FI_FANTASY:   10765,
	SOAP:             10766,
	TALK:             10767,
	WAR_POLITICS:     10768,
	WESTERN:          37,
}

var MovieGenreMap = map[int]string{
	28:    "Action",
	12:    "Aventure",
	16:    "Animation",
	35:    "Comédie",
	80:    "Crime",
	99:    "Documentaire",
	18:    "Drame",
	10751: "Familial",
	14:    "Fantastique",
	36:    "Histoire",
	27:    "Horreur",
	10402: "Musique",
	9648:  "Mystère",
	10749: "Romance",
	878:   "Science-Fiction",
	10770: "Téléfilm",
	53:    "Thriller",
	10752: "Guerre",
	37:    "Western",
}

var TVGenreMap = map[int]string{
	10759: "Action & Aventure",
	16:    "Animation",
	35:    "Comédie",
	80:    "Crime",
	99:    "Documentaire",
	18:    "Drame",
	10751: "Familial",
	10762: "Enfants",
	9648:  "Mystère",
	10763: "Actualités",
	10764: "Téléréalité",
	10765: "Science-Fiction & Fantastique",
	10766: "Feuilleton",
	10767: "Talk-show",
	10768: "Guerre & Politique",
	37:    "Western",
}

var TvGenreMap = map[int]string{
	10759: "Action & Adventure",
	16:    "Animation",
	35:    "Comédie",
	80:    "Crime",
	99:    "Documentaire",
	18:    "Drame",
	10751: "Famille",
	10762: "Kids",
	9648:  "Mystery",
	10763: "News",
	10764: "Reality",
	10765: "Sci-Fi & Fantasy",
	10766: "Soap",
	10767: "Talk",
	10768: "War & Politics",
	37:    "Western",
}

var GenreCategoryColor = map[string]string{
	"Action":          "from-red-600 to-red-800",
	"Aventure":        "from-orange-500 to-amber-600",
	"Animation":       "from-pink-500 to-fuchsia-600",
	"Comédie":         "from-yellow-400 to-amber-500",
	"Crime":           "from-slate-700 to-slate-900",
	"Documentaire":    "from-blue-500 to-sky-600",
	"Drame":           "from-indigo-600 to-indigo-800",
	"Familial":        "from-lime-400 to-green-500",
	"Fantastique":     "from-purple-500 to-violet-700",
	"Histoire":        "from-amber-600 to-yellow-800",
	"Horreur":         "from-red-900 to-black",
	"Musique":         "from-fuchsia-500 to-pink-600",
	"Mystère":         "from-violet-800 to-slate-900",
	"Romance":         "from-rose-400 to-pink-600",
	"Science-Fiction": "from-cyan-500 to-blue-700",
	"Téléfilm":        "from-gray-500 to-gray-700",
	"Thriller":        "from-orange-700 to-red-900",
	"Guerre":          "from-stone-600 to-stone-800",
	"Western":         "from-orange-300 to-amber-500",
}

var (
	PopularMoviesCache   = cache.New[popularMoviesCacheKey, []MovieDTO](30 * time.Minute)
	TopRatedMoviesCache  = cache.New[topRatedMoviesCacheKey, []MovieDTO](30 * time.Minute)
	TrendingMoviesCache  = cache.New[trendingMoviesCacheKey, []MovieDTO](15 * time.Minute)
	ContentDetailsCache  = cache.New[ContentDetailsCacheKey, *ContentDetailsDTO](60 * time.Minute)
	SimilarMoviesCache   = cache.New[similarMoviesCacheKey, []MovieDTO](30 * time.Minute)
	MovieCreditsCache    = cache.New[movieCreditsCacheKey, *MovieCreditsDTO](60 * time.Minute)
	MovieImdbIDCache     = cache.New[movieImdbIDCacheKey, TmdbMovieImdbId](60 * time.Minute)
	GenreCategoriesCache = cache.New[genreCategoriesCacheKey, []CategoryDTO](24 * time.Hour)
	MoviesByGenreCache   = cache.New[MovieGenreCacheKey, []MovieDTO](30 * time.Minute)
)
