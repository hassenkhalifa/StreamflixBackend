package main

import (
	"fmt"
	"strconv"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gin-gonic/gin"
)

type movie struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Year     int      `json:"year"`
	Rating   float32  `json:"rating"`
	Genre    []string `json:"genre"`
	ImageURL string   `json:"image"`
}
type categories struct {
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

var gradients = []string{
	"from-purple-500 to-pink-500",
	"from-blue-500 to-cyan-500",
	"from-green-500 to-teal-500",
	"from-yellow-500 to-orange-500",
	"from-red-500 to-pink-500",
	"from-indigo-500 to-purple-500",
	"from-gray-500 to-slate-500",
}

type UserListItem struct {
	ID           int    `json:"id"`
	ContentID    int    `json:"contentId"`
	ContentType  string `json:"contentType"` // "movie" ou "tvshow"
	Title        string `json:"title"`
	Image        string `json:"image"`
	Description  string `json:"description"`
	Subtitle     string `json:"subtitle"`
	Duration     string `json:"duration"`
	AddedDate    string `json:"addedDate"`
	Percentage   int    `json:"percentage"`
	Progress     int    `json:"progress"`
	CurrentTime  string `json:"currentTime"`
	TotalTime    string `json:"totalTime"`
	PlayHref     string `json:"playHref"`
	FavoriteHref string `json:"favoriteHref"`
	Category     string `json:"category"` // "favorites", "history", "watchlist"
}
type VideoPlayer struct {
	Src         string `json:"src"`
	Title       string `json:"title"`
	Poster      string `json:"poster"`
	Quality     string `json:"quality"`
	AudioFormat string `json:"audioFormat"`
	Autoplay    bool   `json:"autoplay"`
}

var generatedVideoPlayer = VideoPlayer{}
var generatedContentDetails = ContentDetails{}
var generatedUserListItem = []UserListItem{}
var generateMovies = []movie{}

func getVideoPlayer() VideoPlayer {
	if generatedVideoPlayer.Poster != "" {
		return generatedVideoPlayer
	}

	generatedVideoPlayer = VideoPlayer{
		Src:         "https://dash.akamaized.net/akamai/bbb_30fps/bbb_30fps.mpd",
		Title:       gofakeit.MovieName(),
		Poster:      "https://picsum.photos/1200/800",
		Quality:     []string{"HD", "FHD", "UHD"}[gofakeit.Number(0, 2)],
		AudioFormat: []string{"Stereo", "5.1", "7.1", "Dolby Atmos"}[gofakeit.Number(0, 3)],
		Autoplay:    gofakeit.Bool(),
	}

	return generatedVideoPlayer
}
func randomMovieList() []movie {
	// Si déjà généré, retourner la liste existante
	if len(generateMovies) > 0 {
		return generateMovies
	}

	// Sinon, générer
	var movies []movie
	for i := 0; i < gofakeit.Number(10, 20); i++ {
		movies = append(movies, movie{
			ID:       gofakeit.Number(1, 1000),
			Title:    gofakeit.MovieName(),
			Year:     gofakeit.Year(),
			Rating:   gofakeit.Float32Range(0.5, 5.0),
			Genre:    []string{gofakeit.MovieGenre(), gofakeit.MovieGenre()},
			ImageURL: "https://picsum.photos/400/600",
		})
	}

	generateMovies = movies
	return movies
}
func getContentDetails() ContentDetails {
	if len(generatedContentDetails.Cast) > 0 {
		return generatedContentDetails
	}
	var details ContentDetails

	// Générer le cast
	castCount := gofakeit.Number(3, 6)
	var cast []Cast

	for j := 0; j < castCount; j++ {
		cast = append(cast, Cast{
			Name:  gofakeit.Name(),
			Role:  gofakeit.JobTitle(),
			Image: "https://i.pravatar.cc/150?img=" + strconv.Itoa(gofakeit.Number(1, 70)),
		})
	}

	details = ContentDetails{
		ID:             gofakeit.Number(100, 999),
		Title:          gofakeit.MovieName(),
		Image:          "https://picsum.photos/1200/800",
		BackdropImage:  "https://picsum.photos/1200/800",
		Year:           gofakeit.Year(),
		Genres:         []string{gofakeit.MovieGenre(), gofakeit.MovieGenre()},
		Rating:         gofakeit.Float32Range(1, 5),
		Duration:       strconv.Itoa(gofakeit.Number(90, 180)/60) + "h " + strconv.Itoa(gofakeit.Number(0, 59)) + "min",
		Synopsis:       gofakeit.Paragraph(1, 3, 20, " "),
		Director:       gofakeit.Name(),
		Producer:       gofakeit.Company(),
		Languages:      "Français, Anglais",
		Classification: "Tout public",
		Cast:           cast,
	}
	generatedContentDetails = details
	return details
}
func randomCategories() []categories {
	var categoriesArray []categories

	for i := 0; i < gofakeit.Number(10, 20); i++ {
		categoriesArray = append(categoriesArray, categories{
			ID:           gofakeit.Number(1, 1000),
			CategoryName: gofakeit.MovieGenre(),
			Description:  gofakeit.ProductDescription(),
			Href:         "/search",
			Color:        gradients[gofakeit.Number(0, (len(gradients)-1))],
			Previews:     []string{"https://picsum.photos/600/400"},
		})
	}

	return categoriesArray
}
func getMoviesByID(movieID int) movie {
	return generateMovies[movieID]

}
func getMoviesByGenre(genre string) []movie {
	var movies []movie
	for _, movie := range generateMovies {
		// Vérifier si le genre est dans le slice Genre
		for _, g := range movie.Genre {
			if g == genre {
				movies = append(movies, movie)
				break // éviter les doublons si le genre apparaît 2 fois
			}
		}
	}

	return movies
}

func getUserListItems() []UserListItem {
	if len(generatedUserListItem) > 0 {
		return generatedUserListItem
	}
	var items []UserListItem
	categories := []string{"favorites", "history", "watchlist"}
	contentTypes := []string{"movie", "tvshow"}

	for i := 0; i < gofakeit.Number(5, 15); i++ {
		contentID := gofakeit.Number(100, 999)
		contentType := contentTypes[gofakeit.Number(0, 1)]
		percentage := gofakeit.Number(0, 100)

		// Générer durée et temps
		var duration, totalTime, currentTime, subtitle, playHref string

		if contentType == "movie" {
			hours := gofakeit.Number(1, 3)
			minutes := gofakeit.Number(0, 59)
			duration = fmt.Sprintf("%dh %02dmin", hours, minutes)
			totalTime = fmt.Sprintf("%d:%02d:00", hours, minutes)

			// Calculer currentTime basé sur percentage
			totalMinutes := hours*60 + minutes
			currentMinutes := (totalMinutes * percentage) / 100
			currentTime = fmt.Sprintf("%d:%02d:00", currentMinutes/60, currentMinutes%60)

			subtitle = fmt.Sprintf("Film • %d", gofakeit.Year())
			playHref = fmt.Sprintf("/play?id=%d&type=movie", contentID)
		} else {
			minutes := gofakeit.Number(40, 65)
			duration = fmt.Sprintf("%dmin", minutes)
			totalTime = fmt.Sprintf("%d:00", minutes)

			currentMinutes := (minutes * percentage) / 100
			currentTime = fmt.Sprintf("%d:%02d", currentMinutes, gofakeit.Number(0, 59))

			season := gofakeit.Number(1, 8)
			episode := gofakeit.Number(1, 12)
			subtitle = fmt.Sprintf("Série • S%d E%d", season, episode)
			playHref = fmt.Sprintf("/play?id=%d&type=tvshow&season=%d&episode=%d", contentID, season, episode)
		}

		items = append(items, UserListItem{
			ID:           i + 1,
			ContentID:    contentID,
			ContentType:  contentType,
			Title:        gofakeit.MovieName(),
			Image:        "https://picsum.photos/600/400",
			Description:  gofakeit.Sentence(8),
			Subtitle:     subtitle,
			Duration:     duration,
			AddedDate:    gofakeit.Date().Format("2006-01-02"),
			Percentage:   percentage,
			Progress:     percentage,
			CurrentTime:  currentTime,
			TotalTime:    totalTime,
			PlayHref:     playHref,
			FavoriteHref: fmt.Sprintf("/list/toggle/%d", contentID),
			Category:     categories[gofakeit.Number(0, len(categories)-1)],
		})
	}
	generatedUserListItem = items
	return items
}

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	router := gin.Default()

	// Ajouter CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	router.GET("/movieslist", func(c *gin.Context) {
		c.JSON(200, randomMovieList())
	})

	router.GET("/categories", func(c *gin.Context) {
		c.JSON(200, randomCategories())
	})

	router.GET("/movies/:id", func(c *gin.Context) {
		movieID, _ := strconv.Atoi(c.Query("id"))

		c.JSON(200, gin.H{"movie": getMoviesByID(movieID)})
	})
	router.GET("/moviesbygenre", func(c *gin.Context) {
		genre := c.Query("genre")

		c.JSON(200, gin.H{"movie": getMoviesByGenre(genre)})
	})
	router.GET("/user/list", func(c *gin.Context) {

		c.JSON(200, getUserListItems())
	})
	router.GET("/contentDetails", func(c *gin.Context) {

		c.JSON(200, getContentDetails())
	})

	router.GET("/videoPlayer", func(c *gin.Context) {

		c.JSON(200, getVideoPlayer())
	})

	router.Run(":2000")
}
