// Package models contient tous les modèles de données (DTOs) de l'application StreamFlix.
//
// Ce package définit les structures utilisées pour :
//   - Représenter les réponses des API externes (TMDB, Real-Debrid, Torrentio, TVMaze)
//   - Transférer les données entre les couches (handlers ↔ services)
//   - Sérialiser les réponses JSON de l'API StreamFlix
//   - Stocker les données en cache
//
// Les modèles sont organisés par domaine fonctionnel :
//   - content_models.go : films, séries TV, catégories, recherche
//   - rd_models.go : Real-Debrid (débridage, streaming, transcodage)
//   - zt_models.go : Zone Téléchargement (scraping, recherche)
//   - player_models.go : lecteur vidéo (qualités, pistes audio/sous-titres)
//   - user_models.go : données utilisateur (listes, favoris, historique)
//   - ffmpegs_models.go : FFmpeg/FFprobe (résolution, pistes audio)
//   - constants_models.go : constantes (genres, gradients, caches globaux)
package models

import (
	"StreamflixBackend/internal/cache"
	"time"
)

// Gradients contient une liste de classes CSS Tailwind représentant des dégradés de couleurs.
// Ces dégradés sont utilisés pour attribuer un style visuel aux éléments du frontend,
// par exemple pour colorer les cartes de catégories ou les bannières de contenu.
var Gradients = []string{
	"from-purple-500 to-pink-500",
	"from-blue-500 to-cyan-500",
	"from-green-500 to-teal-500",
	"from-yellow-500 to-orange-500",
	"from-red-500 to-pink-500",
	"from-indigo-500 to-purple-500",
	"from-gray-500 to-slate-500",
}

// MovieGenres regroupe les identifiants numériques des genres cinématographiques
// tels que définis par l'API TMDB. Chaque champ correspond à un genre de film
// et contient l'identifiant TMDB associé (par exemple ACTION = 28).
// Cette structure anonyme sert de référence pour filtrer ou catégoriser les films.
var MovieGenres = struct {
	ACTION          int // Genre Action (TMDB ID : 28)
	ADVENTURE       int // Genre Aventure (TMDB ID : 12)
	ANIMATION       int // Genre Animation (TMDB ID : 16)
	COMEDY          int // Genre Comédie (TMDB ID : 35)
	CRIME           int // Genre Crime (TMDB ID : 80)
	DOCUMENTARY     int // Genre Documentaire (TMDB ID : 99)
	DRAMA           int // Genre Drame (TMDB ID : 18)
	FAMILY          int // Genre Familial (TMDB ID : 10751)
	FANTASY         int // Genre Fantastique (TMDB ID : 14)
	HISTORY         int // Genre Histoire (TMDB ID : 36)
	HORROR          int // Genre Horreur (TMDB ID : 27)
	MUSIC           int // Genre Musique (TMDB ID : 10402)
	MYSTERY         int // Genre Mystère (TMDB ID : 9648)
	ROMANCE         int // Genre Romance (TMDB ID : 10749)
	SCIENCE_FICTION int // Genre Science-Fiction (TMDB ID : 878)
	TV_MOVIE        int // Genre Téléfilm (TMDB ID : 10770)
	THRILLER        int // Genre Thriller (TMDB ID : 53)
	WAR             int // Genre Guerre (TMDB ID : 10752)
	WESTERN         int // Genre Western (TMDB ID : 37)
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

// TVGenres regroupe les identifiants numériques des genres de séries TV
// tels que définis par l'API TMDB. Chaque champ correspond à un genre télévisuel
// et contient l'identifiant TMDB associé (par exemple DRAMA = 18).
// Cette structure anonyme sert de référence pour filtrer ou catégoriser les séries.
var TVGenres = struct {
	ACTION_ADVENTURE int // Genre Action & Aventure (TMDB ID : 10759)
	ANIMATION        int // Genre Animation (TMDB ID : 16)
	COMEDY           int // Genre Comédie (TMDB ID : 35)
	CRIME            int // Genre Crime (TMDB ID : 80)
	DOCUMENTARY      int // Genre Documentaire (TMDB ID : 99)
	DRAMA            int // Genre Drame (TMDB ID : 18)
	FAMILY           int // Genre Familial (TMDB ID : 10751)
	KIDS             int // Genre Enfants (TMDB ID : 10762)
	MYSTERY          int // Genre Mystère (TMDB ID : 9648)
	NEWS             int // Genre Actualités (TMDB ID : 10763)
	REALITY          int // Genre Téléréalité (TMDB ID : 10764)
	SCI_FI_FANTASY   int // Genre Science-Fiction & Fantastique (TMDB ID : 10765)
	SOAP             int // Genre Feuilleton (TMDB ID : 10766)
	TALK             int // Genre Talk-show (TMDB ID : 10767)
	WAR_POLITICS     int // Genre Guerre & Politique (TMDB ID : 10768)
	WESTERN          int // Genre Western (TMDB ID : 37)
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

// MovieGenreMap associe chaque identifiant numérique de genre TMDB à son libellé
// en français pour les films. Cette map est utilisée pour convertir les IDs de genre
// reçus de l'API TMDB en noms lisibles affichés dans l'interface utilisateur.
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

// TVGenreMap associe chaque identifiant numérique de genre TMDB à son libellé
// en français pour les séries TV. Cette map est utilisée pour convertir les IDs
// de genre reçus de l'API TMDB en noms lisibles dans l'interface utilisateur.
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

// TvGenreMap est une variante de TVGenreMap qui utilise les noms de genres
// en anglais (ou mixte anglais/français) pour les séries TV.
// Cette map est utilisée dans les contextes nécessitant les libellés originaux
// de l'API TMDB plutôt que les traductions françaises.
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

// GenreCategoryColor associe chaque nom de genre (en français) à une classe CSS
// Tailwind représentant un dégradé de couleur. Ces couleurs sont utilisées pour
// styliser les cartes de catégorie dans le frontend, chaque genre ayant sa propre
// identité visuelle (par exemple "Horreur" utilise un dégradé rouge foncé vers noir).
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

// Caches globaux de l'application. Chaque variable est un cache typé avec une durée
// de vie (TTL) spécifique. Ils permettent de réduire les appels répétés aux API
// externes (TMDB) en stockant temporairement les résultats en mémoire.
var (
	// PopularMoviesCache met en cache les films populaires pendant 30 minutes.
	PopularMoviesCache = cache.New[popularMoviesCacheKey, []MovieDTO](30 * time.Minute)
	// TopRatedMoviesCache met en cache les films les mieux notés pendant 30 minutes.
	TopRatedMoviesCache = cache.New[topRatedMoviesCacheKey, []MovieDTO](30 * time.Minute)
	// TrendingMoviesCache met en cache les films tendance pendant 15 minutes (TTL plus court car les tendances changent vite).
	TrendingMoviesCache = cache.New[trendingMoviesCacheKey, []MovieDTO](15 * time.Minute)
	// ContentDetailsCache met en cache les détails d'un contenu (film ou série) pendant 60 minutes.
	ContentDetailsCache = cache.New[ContentDetailsCacheKey, *ContentDetailsDTO](60 * time.Minute)
	// SimilarMoviesCache met en cache les films similaires pendant 30 minutes.
	SimilarMoviesCache = cache.New[similarMoviesCacheKey, []MovieDTO](30 * time.Minute)
	// MovieCreditsCache met en cache les crédits (casting, réalisateur) d'un film pendant 60 minutes.
	MovieCreditsCache = cache.New[movieCreditsCacheKey, *MovieCreditsDTO](60 * time.Minute)
	// MovieImdbIDCache met en cache l'identifiant IMDB d'un film pendant 60 minutes.
	MovieImdbIDCache = cache.New[movieImdbIDCacheKey, TmdbMovieImdbId](60 * time.Minute)
	// GenreCategoriesCache met en cache la liste des catégories de genres pendant 24 heures (données rarement modifiées).
	GenreCategoriesCache = cache.New[genreCategoriesCacheKey, []CategoryDTO](24 * time.Hour)
	// MoviesByGenreCache met en cache les films filtrés par genre pendant 30 minutes.
	MoviesByGenreCache = cache.New[MovieGenreCacheKey, []MovieDTO](30 * time.Minute)
)
