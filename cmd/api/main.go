package main

import (
	"StreamflixBackend/internal/http/handlers"
	"StreamflixBackend/internal/http/middleware"
	"StreamflixBackend/internal/models"
	"StreamflixBackend/internal/utils"
	"fmt"
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
			utils.InternalError(c, err)
		}
		c.JSON(200, movies)

	})
	router.GET("/getTopRatedMovies", func(c *gin.Context) {

		movies, err := handlers.GetTopRatedMovies(token_tmdb, 1)
		if err != nil {
			utils.InternalError(c, err)
		}
		c.JSON(200, movies)

	})

	router.GET("/getTrendingMovies", func(c *gin.Context) {
		// query params optionnels: ?time_window=day|week&language=fr-FR&page=1
		timeWindow := c.DefaultQuery("time_window", "day")
		language := c.DefaultQuery("language", "fr-FR")

		movies, err := handlers.GetTrendingMovies(token_tmdb, timeWindow, 1, language)
		if err != nil {
			utils.InternalError(c, err)
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
			utils.InternalError(c, err)
		}
		c.JSON(200, details)
	})

	router.GET("/movie/:id/credits", func(c *gin.Context) {
		idStr := c.Param("id")
		movieID, err := strconv.Atoi(idStr)
		if err != nil || movieID <= 0 {
			utils.InternalError(c, err)
			return
		}

		out, err := handlers.GetMovieCredits(token_tmdb, "", movieID)
		if err != nil {
			utils.InternalError(c, err)
			return
		}

		c.JSON(200, out)
	})
	router.GET("/movie/:id/similar", func(c *gin.Context) {
		idStr := c.Param("id")
		movieID, err := strconv.Atoi(idStr)
		if err != nil || movieID <= 0 {
			utils.InternalError(c, err)
			return
		}

		similar, _ := handlers.GetSimilarMovies(token_tmdb, "", models.MovieGenreMap, movieID, "")

		c.JSON(200, similar)
	})
	router.GET("/moviesbygenre", func(c *gin.Context) {
		genreIDStr := c.DefaultQuery("genre_id", "28")
		genreID, _ := strconv.Atoi(genreIDStr)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		language := c.DefaultQuery("language", "fr-FR")

		movies, err := handlers.GetMoviesByGenre(token_tmdb, genreID, page, language)
		if err != nil {
			utils.InternalError(c, err)
			return
		}
		c.JSON(http.StatusOK, movies)
	})
	router.GET("/getMovieGenreList", func(c *gin.Context) {
		language := c.DefaultQuery("language", "fr-FR")

		genreMap, err := handlers.GetMovieGenreCategories(token_tmdb, language)
		if err != nil {
			utils.InternalError(c, err)
			return
		}

		c.JSON(http.StatusOK, genreMap)
	})
	router.GET("/searchMovies", func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		ratingStr := c.DefaultQuery("rating", "0")
		page, _ := strconv.Atoi(pageStr)
		rating, _ := strconv.ParseFloat(ratingStr, 64)

		movies, err := handlers.SearchMovies(models.SearchMoviesParams{
			BearerToken: token_tmdb,
			Query:       c.Query("query"),
			GenresCSV:   c.Query("genres"),
			YearsCSV:    c.Query("years"),
			SortBy:      c.Query("sort_by"),
			Page:        page,
			Language:    c.DefaultQuery("language", "fr-FR"),
			Rating:      rating,
		})
		if err != nil {
			utils.InternalError(c, err)
			return
		}
		c.JSON(http.StatusOK, movies)
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
			utils.InternalError(c, err)
			return
		}
		log.Printf("✅ ID converti en int: %d", imdbid)

		// Étape 2: Récupération de l'IMDB ID depuis TMDB
		log.Printf("🔍 [2/8] Appel TMDB pour récupérer l'IMDB ID du film %d...", imdbid)
		movieImdbId, err := handlers.GetMovieImdbID(token_tmdb, imdbid)
		if err != nil {
			log.Printf("❌ Erreur GetMovieImdbID: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Printf("✅ IMDB ID récupéré: %s", movieImdbId.ImdbId)

		// Étape 3: Récupération des streams Torrentio
		log.Printf("🌐 [3/8] Appel Torrentio pour IMDB ID: %s...", movieImdbId.ImdbId)
		streams, err := handlers.GetTorrentioStreams(movieImdbId.ImdbId)
		if err != nil {
			log.Printf("❌ Erreur GetTorrentioStreams: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Printf("✅ Nombre de streams trouvés: %d", len(streams.Streams))
		if len(streams.Streams) > 0 {
			log.Printf("   Premier stream: %s (InfoHash: %s)", streams.Streams[0].Name, streams.Streams[0].InfoHash)
		} else {
			log.Println("❌ Aucun stream disponible")
			utils.APIError(c, http.StatusNotFound, "Aucun stream trouvé")
			return
		}

		// Étape 4: Ajout du magnet à Real-Debrid
		log.Printf("🧲 [4/8] Ajout du magnet à Real-Debrid (InfoHash: %s)...", streams.Streams[0].InfoHash)
		debrid, err := handlers.AddMagnetRealDebrid(token_rdt, streams.Streams[0].InfoHash)
		if err != nil {
			log.Printf("❌ Erreur AddMagnetRealDebrid: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Printf("✅ Magnet ajouté avec succès. Torrent ID: %s", debrid.Id)

		// Étape 5: Sélection des fichiers
		log.Printf("📂 [5/8] Sélection des fichiers pour le torrent %s...", debrid.Id)
		err = handlers.SelectFilesRealDebrid(token_rdt, debrid.Id)
		if err != nil {
			log.Printf("❌ Erreur SelectFilesRealDebrid: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Println("✅ Fichiers sélectionnés avec succès")

		// Étape 6: Récupération des infos du torrent
		log.Printf("ℹ️  [6/8] Récupération des infos du torrent %s...", debrid.Id)
		info, err := handlers.GetRealDebridTorrentInfo(token_rdt, debrid.Id)
		if err != nil {
			log.Printf("❌ Erreur GetRealDebridTorrentInfo: %v", err)
			utils.InternalError(c, err)
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
			utils.APIError(c, http.StatusNotFound, "Aucun lien disponible")
			return
		}

		// Étape 7: Unrestrict et récupération du MPD
		log.Printf("🔓 [7/8] Unrestrict du lien: %s...", info.Links[0])
		fileid, err := handlers.UnrestrictAndGetMPD(token_rdt, info.Links[0])
		if err != nil {
			log.Printf("❌ Erreur UnrestrictAndGetMPD: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Printf("✅ Lien unrestricted. File ID: %s", fileid)

		// Étape 8: Récupération du lecteur vidéo
		log.Printf("🎬 [8/8] Récupération du lecteur vidéo pour File ID: %s...", fileid)
		videoPlayer, err := handlers.GetVideoPlayer(token_rdt, fileid)
		if err != nil {
			log.Printf("❌ Erreur GetVideoPlayer: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Println("✅ Lecteur vidéo récupéré avec succès")

		log.Println("========== FIN /videoPlayer/:id (SUCCESS) ==========")
		c.JSON(200, videoPlayer)
	})
	router.GET("/videoPlayer/:id/:season/:episode", func(c *gin.Context) {
		log.Println("========== DÉBUT /videoPlayer/:id/:season/:episode ==========")

		// Étape 1: Récupération de l'ID depuis l'URL
		idParam := c.Param("id")
		season := c.Param("season")
		episode := c.Param("episode")
		log.Printf("📥 [1/8] ID reçu depuis l'URL: %s, saison: %s, épisode: %s", idParam, season, episode)

		imdbid, err := strconv.Atoi(idParam)
		if err != nil {
			log.Printf("❌ Erreur conversion ID en int: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Printf("✅ ID converti en int: %d", imdbid)

		// Étape 2: Récupération de l'IMDB ID depuis TMDB
		log.Printf("🔍 [2/8] Appel TMDB pour récupérer l'IMDB ID du film %d...", imdbid)
		movieImdbId, err := handlers.GetMovieImdbID(token_tmdb, imdbid)
		if err != nil {
			log.Printf("❌ Erreur GetMovieImdbID: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Printf("✅ IMDB ID récupéré: %s", movieImdbId.ImdbId)
		idParamFinal := fmt.Sprintf("%s:%s:%s", movieImdbId.ImdbId, c.Param("season"), c.Param("episode"))

		// Étape 3: Récupération des streams Torrentio
		log.Printf("🌐 [3/8] Appel Torrentio pour IMDB ID: %s...", idParamFinal)
		streams, err := handlers.GetTorrentioStreams(idParamFinal)
		if err != nil {
			log.Printf("❌ Erreur GetTorrentioStreams: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Printf("✅ Nombre de streams trouvés: %d", len(streams.Streams))
		if len(streams.Streams) > 0 {
			log.Printf("   Premier stream: %s (InfoHash: %s)", streams.Streams[0].Name, streams.Streams[0].InfoHash)
		} else {
			log.Println("❌ Aucun stream disponible")
			utils.APIError(c, http.StatusNotFound, "Aucun stream trouvé")
			return
		}

		// Étape 4: Ajout du magnet à Real-Debrid
		log.Printf("🧲 [4/8] Ajout du magnet à Real-Debrid (InfoHash: %s)...", streams.Streams[0].InfoHash)
		debrid, err := handlers.AddMagnetRealDebrid(token_rdt, streams.Streams[0].InfoHash)
		if err != nil {
			log.Printf("❌ Erreur AddMagnetRealDebrid: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Printf("✅ Magnet ajouté avec succès. Torrent ID: %s", debrid.Id)

		// Étape 5: Sélection des fichiers
		log.Printf("📂 [5/8] Sélection des fichiers pour le torrent %s...", debrid.Id)
		err = handlers.SelectFilesRealDebrid(token_rdt, debrid.Id)
		if err != nil {
			log.Printf("❌ Erreur SelectFilesRealDebrid: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Println("✅ Fichiers sélectionnés avec succès")

		// Étape 6: Récupération des infos du torrent
		log.Printf("ℹ️  [6/8] Récupération des infos du torrent %s...", debrid.Id)
		info, err := handlers.GetRealDebridTorrentInfo(token_rdt, debrid.Id)
		if err != nil {
			log.Printf("❌ Erreur GetRealDebridTorrentInfo: %v", err)
			utils.InternalError(c, err)
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
			utils.APIError(c, http.StatusNotFound, "Aucun lien disponible")
			return
		}

		// Étape 7: Unrestrict et récupération du MPD
		log.Printf("🔓 [7/8] Unrestrict du lien: %s...", info.Links[0])
		fileid, err := handlers.UnrestrictAndGetMPD(token_rdt, info.Links[0])
		if err != nil {
			log.Printf("❌ Erreur UnrestrictAndGetMPD: %v", err)
			utils.InternalError(c, err)
			return
		}
		log.Printf("✅ Lien unrestricted. File ID: %s", fileid)

		// Étape 8: Récupération du lecteur vidéo
		log.Printf("🎬 [8/8] Récupération du lecteur vidéo pour File ID: %s...", fileid)
		videoPlayer, err := handlers.GetVideoPlayer(token_rdt, fileid)
		if err != nil {
			log.Printf("❌ Erreur GetVideoPlayer: %v", err)
			utils.InternalError(c, err)
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
			utils.APIError(c, http.StatusBadRequest, "category and query are required")
			return
		}

		results, err := parser.Search(category, query, page)
		if err != nil {
			utils.InternalError(c, err)
			return
		}

		c.JSON(200, gin.H{"status": true, "data": results})
	})

	router.GET("/zt/search-all", func(c *gin.Context) {
		category := c.Query("category")
		query := c.Query("query")

		if category == "" || query == "" {
			utils.APIError(c, http.StatusBadRequest, "category and query are required")
			return
		}

		results, err := parser.SearchAll(category, query)
		if err != nil {
			utils.InternalError(c, err)
			return
		}

		c.JSON(200, gin.H{"status": true, "data": results})
	})

	router.GET("/zt/basic/:category/:id", func(c *gin.Context) {
		category := c.Param("category")
		id := c.Param("id")

		basicInfo, err := parser.GetMovieNameFromId(category, id)
		if err != nil {
			utils.InternalError(c, err)
			return
		}

		c.JSON(200, gin.H{"status": true, "data": basicInfo})
	})

	router.GET("/zt/:category/:id", func(c *gin.Context) {
		category := c.Param("category")
		id := c.Param("id")

		movieData, err := parser.GetQueryDatas(category, id)
		if err != nil {
			utils.InternalError(c, err)
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

	router.GET("/getTrendingTV", func(c *gin.Context) {
		timeWindow := c.DefaultQuery("time_window", "day")
		language := c.DefaultQuery("language", "fr-FR")

		// optionnel: page
		page := 1
		if p := c.Query("page"); p != "" {
			fmt.Sscanf(p, "%d", &page)
		}

		tv, err := handlers.GetTrendingTV(token_tmdb, timeWindow, page, language)
		if err != nil {
			utils.InternalError(c, err)
			return
		}

		c.JSON(200, tv)
	})

	router.GET("/getTVShowsByGenre", func(c *gin.Context) {
		genreIDStr := c.DefaultQuery("genre_id", "18")
		pageStr := c.DefaultQuery("page", "1")
		language := c.DefaultQuery("language", "fr-FR")

		genreID, err := strconv.Atoi(genreIDStr)
		if err != nil {
			utils.InternalError(c, err)
			return
		}

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		tvs, err := handlers.GetTVByGenre(token_tmdb, genreID, page, language)
		if err != nil {
			utils.InternalError(c, err)
			return
		}

		c.JSON(http.StatusOK, tvs)
	})

	router.GET("/getPopularTVShows", func(c *gin.Context) {
		language := c.DefaultQuery("language", "fr-FR")
		pageStr := c.DefaultQuery("page", "1")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		shows, err := handlers.GetPopularTVShows(token_tmdb, page, language)
		if err != nil {
			utils.InternalError(c, err)
			return
		}

		c.JSON(http.StatusOK, shows)
	})
	router.GET("/getTVInfo", func(c *gin.Context) {
		idStr := c.Query("series_id")
		if idStr == "" {
			utils.APIError(c, http.StatusBadRequest, "series_id requis")
			return
		}
		seriesID, err := strconv.Atoi(idStr)
		if err != nil {
			utils.APIError(c, http.StatusBadRequest, "series_id invalide")
			return
		}

		language := c.DefaultQuery("language", "fr-FR")

		info, err := handlers.GetTVInfo(token_tmdb, seriesID, language)
		if err != nil {
			utils.InternalError(c, err)
			return
		}

		c.JSON(http.StatusOK, info)
	})
	router.GET("/searchTV", func(c *gin.Context) {
		query := c.Query("query")
		if query == "" {
			utils.APIError(c, http.StatusBadRequest, "query requis")
			return
		}

		language := c.DefaultQuery("language", "fr-FR")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

		results, err := handlers.SearchTV(token_tmdb, query, language, page)
		if err != nil {
			utils.InternalError(c, err)
			return
		}

		c.JSON(http.StatusOK, results)
	})

	err := router.Run(":2000")
	if err != nil {
		return
	}
}
