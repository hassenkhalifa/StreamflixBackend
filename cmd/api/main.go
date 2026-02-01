package main

import (
	"StreamflixBackend/internal/http/handlers"
	"StreamflixBackend/internal/http/middleware"
	"StreamflixBackend/internal/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func loadEnvDev() (string, string) {
	if err := godotenv.Load(".env.dev"); err != nil {
		log.Fatalf("cannot load .env.dev: %v", err)
	}
	token_rdt := os.Getenv("REALDEBRID_TOKEN")
	token_tmdb := os.Getenv("TMDB_TOKEN")
	if token_rdt == "" {
		log.Fatal("REALDEBRID_TOKEN is missing")
	}
	if token_tmdb == "" {
		log.Fatal("TMDB_TOKEN is missing")
	}
	return token_rdt, token_tmdb
}

func main() {
	router := gin.Default()
	token_rdt, token_tmdb := loadEnvDev()
	router.Use(middleware.Cors())
	parser := handlers.NewZtParser(
		true,          // devMode
		2*time.Second, // requestTimeInBetween
		token_tmdb,    // moviesDbToken
	)
	// Configuration de l'URL de base
	parser.UseBaseUrl("https://www.zone-telechargement.irish")

	// Création du handler

	// Ajouter CORS

	router.GET("/movieslist", func(c *gin.Context) {
		c.JSON(200, handlers.RandomMovieList())
	})

	router.GET("/popularMovies", func(c *gin.Context) {

		movies, err := handlers.GetPopularMovies(token_tmdb, "", models.MovieGenreMap, "")
		if err != nil {
			return
		}
		c.JSON(200, movies)

	})

	router.GET("/categories", func(c *gin.Context) {
		c.JSON(200, handlers.RandomCategories())
	})

	router.GET("/movie/:id", func(c *gin.Context) {
		movieID, _ := strconv.Atoi(c.Param("id"))

		details, err := handlers.GetContentDetails(token_tmdb, "", movieID)
		if err != nil {
			return
		}
		c.JSON(200, details)
	})

	router.GET("/movie/:id/credits", func(c *gin.Context) {
		idStr := c.Param("id")
		movieID, err := strconv.Atoi(idStr)
		if err != nil || movieID <= 0 {
			c.JSON(400, gin.H{"error": "invalid id"})
			return
		}

		out, err := handlers.GetMovieCredits(token_tmdb, "", movieID)
		if err != nil {
			c.JSON(502, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, out)
	})
	router.GET("/movie/:id/similar", func(c *gin.Context) {
		idStr := c.Param("id")
		movieID, err := strconv.Atoi(idStr)
		if err != nil || movieID <= 0 {
			c.JSON(400, gin.H{"error": "invalid id"})
			return
		}

		similar, _ := handlers.GetSimilarMovies(token_tmdb, "", models.MovieGenreMap, movieID, "")

		c.JSON(200, similar)
	})
	router.GET("/moviesbygenre", func(c *gin.Context) {
		genre := c.Query("genre")

		c.JSON(200, gin.H{"movie": handlers.GetMoviesByGenre(genre)})
	})
	router.GET("/user/list", func(c *gin.Context) {

		c.JSON(200, handlers.GetUserListItems())
	})
	router.GET("/contentDetails", func(c *gin.Context) {

		c.JSON(200, handlers.GetContentDetailsRandomized())
	})

	router.GET("/videoPlayer/:id", func(c *gin.Context) {
		log.Println("========== DÉBUT /videoPlayer/:id ==========")

		// Étape 1: Récupération de l'ID depuis l'URL
		idParam := c.Param("id")
		log.Printf("📥 [1/8] ID reçu depuis l'URL: %s", idParam)

		imdbid, err := strconv.Atoi(idParam)
		if err != nil {
			log.Printf("❌ Erreur conversion ID en int: %v", err)
			c.JSON(400, gin.H{"error": "ID invalide"})
			return
		}
		log.Printf("✅ ID converti en int: %d", imdbid)

		// Étape 2: Récupération de l'IMDB ID depuis TMDB
		log.Printf("🔍 [2/8] Appel TMDB pour récupérer l'IMDB ID du film %d...", imdbid)
		movieImdbId, err := handlers.GetMovieImdbID(token_tmdb, imdbid)
		if err != nil {
			log.Printf("❌ Erreur GetMovieImdbID: %v", err)
			c.JSON(500, gin.H{"error": "Impossible de récupérer l'IMDB ID"})
			return
		}
		log.Printf("✅ IMDB ID récupéré: %s", movieImdbId.ImdbId)

		// Étape 3: Récupération des streams Torrentio
		log.Printf("🌐 [3/8] Appel Torrentio pour IMDB ID: %s...", movieImdbId.ImdbId)
		streams, err := handlers.GetTorrentioStreams(movieImdbId.ImdbId)
		if err != nil {
			log.Printf("❌ Erreur GetTorrentioStreams: %v", err)
			c.JSON(500, gin.H{"error": "Impossible de récupérer les streams"})
			return
		}
		log.Printf("✅ Nombre de streams trouvés: %d", len(streams.Streams))
		if len(streams.Streams) > 0 {
			log.Printf("   Premier stream: %s (InfoHash: %s)", streams.Streams[0].Name, streams.Streams[0].InfoHash)
		} else {
			log.Println("❌ Aucun stream disponible")
			c.JSON(404, gin.H{"error": "Aucun stream trouvé"})
			return
		}

		// Étape 4: Ajout du magnet à Real-Debrid
		log.Printf("🧲 [4/8] Ajout du magnet à Real-Debrid (InfoHash: %s)...", streams.Streams[0].InfoHash)
		debrid, err := handlers.AddMagnetRealDebrid(token_rdt, streams.Streams[0].InfoHash)
		if err != nil {
			log.Printf("❌ Erreur AddMagnetRealDebrid: %v", err)
			c.JSON(500, gin.H{"error": "Impossible d'ajouter le magnet"})
			return
		}
		log.Printf("✅ Magnet ajouté avec succès. Torrent ID: %s", debrid.Id)

		// Étape 5: Sélection des fichiers
		log.Printf("📂 [5/8] Sélection des fichiers pour le torrent %s...", debrid.Id)
		err = handlers.SelectFilesRealDebrid(token_rdt, debrid.Id)
		if err != nil {
			log.Printf("❌ Erreur SelectFilesRealDebrid: %v", err)
			c.JSON(500, gin.H{"error": "Impossible de sélectionner les fichiers"})
			return
		}
		log.Println("✅ Fichiers sélectionnés avec succès")

		// Étape 6: Récupération des infos du torrent
		log.Printf("ℹ️  [6/8] Récupération des infos du torrent %s...", debrid.Id)
		info, err := handlers.GetRealDebridTorrentInfo(token_rdt, debrid.Id)
		if err != nil {
			log.Printf("❌ Erreur GetRealDebridTorrentInfo: %v", err)
			c.JSON(500, gin.H{"error": "Impossible de récupérer les infos du torrent"})
			return
		}
		log.Printf("✅ Infos torrent récupérées:")
		log.Printf("   - Filename: %s", info.Filename)
		log.Printf("   - Status: %s", info.Status)
		log.Printf("   - Progress: %.2f%%", info.Progress)
		log.Printf("   - Nombre de liens: %d", len(info.Links))
		if len(info.Links) > 0 {
			log.Printf("   - Premier lien: %s", info.Links[0])
		} else {
			log.Println("❌ Aucun lien disponible")
			c.JSON(500, gin.H{"error": "Aucun lien disponible"})
			return
		}

		// Étape 7: Unrestrict et récupération du MPD
		log.Printf("🔓 [7/8] Unrestrict du lien: %s...", info.Links[0])
		fileid, err := handlers.UnrestrictAndGetMPD(token_rdt, info.Links[0])
		if err != nil {
			log.Printf("❌ Erreur UnrestrictAndGetMPD: %v", err)
			c.JSON(500, gin.H{"error": "Impossible d'unrestrict le lien"})
			return
		}
		log.Printf("✅ Lien unrestricted. File ID: %s", fileid)

		// Étape 8: Récupération du lecteur vidéo
		log.Printf("🎬 [8/8] Récupération du lecteur vidéo pour File ID: %s...", fileid)
		videoPlayer, err := handlers.GetVideoPlayer(token_rdt, fileid)
		if err != nil {
			log.Printf("❌ Erreur GetVideoPlayer: %v", err)
			c.JSON(500, gin.H{"error": "Impossible de récupérer le lecteur vidéo"})
			return
		}
		log.Println("✅ Lecteur vidéo récupéré avec succès")

		log.Println("========== FIN /videoPlayer/:id (SUCCESS) ==========")
		c.JSON(200, videoPlayer)
	})
	router.GET("/zt/search", func(c *gin.Context) {
		category := c.Query("category")
		query := c.Query("query")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

		if category == "" || query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "category and query are required"})
			return
		}

		results, err := parser.Search(category, query, page)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": true, "data": results})
	})

	router.GET("/zt/search-all", func(c *gin.Context) {
		category := c.Query("category")
		query := c.Query("query")

		if category == "" || query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "category and query are required"})
			return
		}

		results, err := parser.SearchAll(category, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": true, "data": results})
	})

	router.GET("/zt/basic/:category/:id", func(c *gin.Context) {
		category := c.Param("category")
		id := c.Param("id")

		basicInfo, err := parser.GetMovieNameFromId(category, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": true, "data": basicInfo})
	})

	router.GET("/zt/:category/:id", func(c *gin.Context) {
		category := c.Param("category")
		id := c.Param("id")

		movieData, err := parser.GetQueryDatas(category, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": true, "data": movieData})
	})

	router.GET("/movieimdbid/:tmdbid", func(c *gin.Context) {
		tmdbid := c.Param("tmdbid")
		tmdbidconv, _ := strconv.Atoi(tmdbid)
		movieImdbId, _ := handlers.GetMovieImdbID(token_tmdb, tmdbidconv)

		c.JSON(200, movieImdbId)
	})

	router.Run(":2000")
}
