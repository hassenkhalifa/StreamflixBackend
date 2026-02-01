package models

// ============================================================================
// CONSTANTS
// ============================================================================

var Gradients = []string{
	"from-purple-500 to-pink-500",
	"from-blue-500 to-cyan-500",
	"from-green-500 to-teal-500",
	"from-yellow-500 to-orange-500",
	"from-red-500 to-pink-500",
	"from-indigo-500 to-purple-500",
	"from-gray-500 to-slate-500",
}

// MovieGenres - Constantes des genres de films
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

// TVGenres - Constantes des genres de séries TV
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

// MovieGenreMap - Map ID → Nom (français)
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

// TVGenreMap - Map ID → Nom (français)
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
