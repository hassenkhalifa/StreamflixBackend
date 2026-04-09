// Package http contient le routeur HTTP Gin et la configuration des routes de l'API StreamFlix.
//
// Ce package est responsable de :
//   - La création et configuration du moteur Gin (mode, middleware)
//   - L'enregistrement de toutes les routes API v1 et legacy
//   - La liaison entre les routes et les handlers/services
//   - Le health check endpoint
//
// Architecture des routes :
//   - /health : vérification de santé du serveur
//   - /api/v1/movies/* : endpoints films (populaires, tendances, détails, recherche)
//   - /api/v1/tv/* : endpoints séries TV
//   - /api/v1/player/* : endpoints lecteur vidéo
//   - /api/v1/zt/* : endpoints Zone Téléchargement
//   - /api/v1/user/* : endpoints utilisateur
//   - Routes legacy : compatibilité ascendante avec les anciens endpoints
package http

import (
	"StreamflixBackend/internal/config"
	"StreamflixBackend/internal/http/handlers"
	"StreamflixBackend/internal/http/middleware"
	"StreamflixBackend/internal/models"
	"StreamflixBackend/internal/utils"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// NewRouter crée et configure le moteur Gin avec l'ensemble des routes et middlewares.
//
// La configuration inclut :
//   - Le mode Gin (debug/release) selon la configuration.
//   - La pile de middlewares : recovery, logger, headers de sécurité, CORS et rate-limiting.
//   - L'endpoint /health pour la vérification de santé du serveur.
//   - L'initialisation du parser Zone Téléchargement avec rate-limiting de 2 secondes.
//   - L'enregistrement des groupes de routes API v1 (movies, tv, player, zt, user).
//   - L'enregistrement des routes legacy pour la compatibilité ascendante.
//
// Paramètres :
//   - cfg : configuration de l'application contenant les tokens API, le mode Gin,
//     les origines CORS et les limites de rate-limiting.
//
// Retourne le moteur Gin configuré, prêt à être démarré.
func NewRouter(cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.GinMode)

	router := gin.New()

	// Middleware stack
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.Cors(cfg.CORSOrigins))

	// Rate limiter
	limiter := middleware.NewRateLimiter(cfg.RateLimitPerMin)
	router.Use(limiter.RateLimitWithBypass())

	// Health check
	startTime := time.Now()
	router.GET("/health", func(c *gin.Context) {
		utils.RespondSuccess(c, http.StatusOK, gin.H{
			"status":  "healthy",
			"version": "1.0.0",
			"uptime":  time.Since(startTime).String(),
		})
	})

	// ZT parser
	parser := handlers.NewZtParser(
		cfg.Environment == "development",
		2*time.Second,
		cfg.TMDBToken,
	)
	parser.UseBaseUrl("https://www.zone-telechargement.irish")

	// API routes
	api := router.Group("/api/v1")
	registerMovieRoutes(api, cfg)
	registerTVRoutes(api, cfg)
	registerPlayerRoutes(api, cfg)
	registerZTRoutes(api, parser)
	registerUserRoutes(api)

	// Legacy routes (backward compatibility)
	registerLegacyRoutes(router, cfg, parser)

	return router
}

// registerMovieRoutes enregistre les routes du groupe /api/v1/movies.
//
// Routes configurées :
//   - GET /popular : films populaires
//   - GET /top-rated : films les mieux notés
//   - GET /trending : films en tendance
//   - GET /search : recherche de films
//   - GET /by-genre : films par genre
//   - GET /genres : liste des genres disponibles
//   - GET /:id : détails d'un film
//   - GET /:id/credits : crédits d'un film
//   - GET /:id/similar : films similaires
//   - GET /:id/imdb : identifiant IMDB d'un film
func registerMovieRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	movies := rg.Group("/movies")
	{
		movies.GET("/popular", moviePopularHandler(cfg))
		movies.GET("/top-rated", movieTopRatedHandler(cfg))
		movies.GET("/trending", movieTrendingHandler(cfg))
		movies.GET("/search", movieSearchHandler(cfg))
		movies.GET("/by-genre", moviesByGenreHandler(cfg))
		movies.GET("/genres", movieGenresHandler(cfg))
		movies.GET("/:id", movieDetailsHandler(cfg))
		movies.GET("/:id/credits", movieCreditsHandler(cfg))
		movies.GET("/:id/similar", movieSimilarHandler(cfg))
		movies.GET("/:id/imdb", movieImdbIDHandler(cfg))
	}
}

// registerTVRoutes enregistre les routes du groupe /api/v1/tv.
//
// Routes configurées :
//   - GET /trending : séries TV en tendance
//   - GET /by-genre : séries TV par genre
//   - GET /popular : séries TV populaires
//   - GET /info : détails complets d'une série (avec saisons et épisodes)
//   - GET /search : recherche de séries TV
func registerTVRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	tv := rg.Group("/tv")
	{
		tv.GET("/trending", tvTrendingHandler(cfg))
		tv.GET("/by-genre", tvByGenreHandler(cfg))
		tv.GET("/popular", tvPopularHandler(cfg))
		tv.GET("/info", tvInfoHandler(cfg))
		tv.GET("/search", tvSearchHandler(cfg))
	}
}

// registerPlayerRoutes enregistre les routes du groupe /api/v1/player.
//
// Routes configurées :
//   - GET /movie/:id : lecteur vidéo pour un film (pipeline TMDB -> Torrentio -> Real-Debrid)
//   - GET /series/:id/:season/:episode : lecteur vidéo pour un épisode de série
func registerPlayerRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	player := rg.Group("/player")
	{
		player.GET("/movie/:id", videoPlayerMovieHandler(cfg))
		player.GET("/series/:id/:season/:episode", videoPlayerSeriesHandler(cfg))
	}
}

// registerZTRoutes enregistre les routes du groupe /api/v1/zt pour Zone Téléchargement.
//
// Routes configurées :
//   - GET /search : recherche paginée de contenus
//   - GET /search-all : recherche exhaustive sur toutes les pages
//   - GET /basic/:category/:id : informations de base d'un contenu (titre, liens)
//   - GET /:category/:id : détails complets enrichis via TMDB/TVMaze
func registerZTRoutes(rg *gin.RouterGroup, parser *handlers.ZtParserService) {
	zt := rg.Group("/zt")
	{
		zt.GET("/search", ztSearchHandler(parser))
		zt.GET("/search-all", ztSearchAllHandler(parser))
		zt.GET("/basic/:category/:id", ztBasicHandler(parser))
		zt.GET("/:category/:id", ztDetailsHandler(parser))
	}
}

// registerUserRoutes enregistre les routes du groupe /api/v1/user.
//
// Routes configurées :
//   - GET /list : récupère la liste des éléments utilisateur.
func registerUserRoutes(rg *gin.RouterGroup) {
	user := rg.Group("/user")
	{
		user.GET("/list", func(c *gin.Context) {
			utils.RespondSuccess(c, http.StatusOK, handlers.GetUserListItems())
		})
	}
}

// registerLegacyRoutes enregistre les anciennes routes pour assurer la compatibilité ascendante.
//
// Ces routes réutilisent les mêmes handlers factory que les routes v1 mais sont
// montées directement sur le routeur racine (sans préfixe /api/v1).
// Elles correspondent aux endpoints de la version initiale de l'API StreamFlix.
func registerLegacyRoutes(router *gin.Engine, cfg *config.Config, parser *handlers.ZtParserService) {
	router.GET("/movieslist", func(c *gin.Context) {
		c.JSON(200, handlers.RandomMovieList())
	})
	router.GET("/popularMovies", moviePopularHandler(cfg))
	router.GET("/getTopRatedMovies", movieTopRatedHandler(cfg))
	router.GET("/getTrendingMovies", movieTrendingHandler(cfg))
	router.GET("/categories", func(c *gin.Context) {
		c.JSON(200, handlers.RandomCategories())
	})
	router.GET("/movie/:id", movieDetailsHandler(cfg))
	router.GET("/movie/:id/credits", movieCreditsHandler(cfg))
	router.GET("/movie/:id/similar", movieSimilarHandler(cfg))
	router.GET("/moviesbygenre", moviesByGenreHandler(cfg))
	router.GET("/getMovieGenreList", movieGenresHandler(cfg))
	router.GET("/searchMovies", movieSearchHandler(cfg))
	router.GET("/user/list", func(c *gin.Context) {
		c.JSON(200, handlers.GetUserListItems())
	})
	router.GET("/contentDetails", func(c *gin.Context) {
		c.JSON(200, handlers.GetContentDetailsRandomized())
	})
	router.GET("/videoPlayer/:id", videoPlayerMovieHandler(cfg))
	router.GET("/videoPlayer/:id/:season/:episode", videoPlayerSeriesHandler(cfg))
	router.GET("/zt/search", ztSearchHandler(parser))
	router.GET("/zt/search-all", ztSearchAllHandler(parser))
	router.GET("/zt/basic/:category/:id", ztBasicHandler(parser))
	router.GET("/zt/:category/:id", ztDetailsHandler(parser))
	router.GET("/movieimdbid/:tmdbid", func(c *gin.Context) {
		tmdbid := c.Param("tmdbid")
		tmdbidconv, err := strconv.Atoi(tmdbid)
		if err != nil {
			utils.BadRequest(c, "tmdbid invalide")
			return
		}
		movieImdbId, err := handlers.GetMovieImdbID(cfg.TMDBToken, tmdbidconv)
		if err != nil {
			utils.InternalError(c, err)
			return
		}
		c.JSON(200, movieImdbId)
	})
	router.GET("/getTrendingTV", tvTrendingHandler(cfg))
	router.GET("/getTVShowsByGenre", tvByGenreHandler(cfg))
	router.GET("/getPopularTVShows", tvPopularHandler(cfg))
	router.GET("/getTVInfo", tvInfoHandler(cfg))
	router.GET("/searchTV", tvSearchHandler(cfg))
}

// ===========================================================================
// Movie Handlers
// ===========================================================================

// moviePopularHandler retourne un handler Gin qui récupère les films populaires.
//
// Flux de la requête : appel TMDB /movie/popular -> conversion en DTO -> réponse JSON.
// Aucun paramètre de query n'est requis.
func moviePopularHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		movies, err := handlers.GetPopularMovies(cfg.TMDBToken, "", models.MovieGenreMap, "")
		if err != nil {
			slog.Error("failed to get popular movies", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, movies)
	}
}

// movieTopRatedHandler retourne un handler Gin qui récupère les films les mieux notés.
//
// Flux de la requête : appel TMDB /movie/top_rated (page 1) -> conversion en DTO -> réponse JSON.
func movieTopRatedHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		movies, err := handlers.GetTopRatedMovies(cfg.TMDBToken, 1)
		if err != nil {
			slog.Error("failed to get top rated movies", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, movies)
	}
}

// movieTrendingHandler retourne un handler Gin qui récupère les films en tendance.
//
// Paramètres de query optionnels :
//   - time_window : "day" (défaut) ou "week".
//   - language : code langue BCP 47 (défaut "fr-FR").
//
// Flux de la requête : extraction des paramètres -> appel TMDB /trending/movie -> réponse JSON.
func movieTrendingHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		timeWindow := c.DefaultQuery("time_window", "day")
		language := c.DefaultQuery("language", "fr-FR")

		movies, err := handlers.GetTrendingMovies(cfg.TMDBToken, timeWindow, 1, language)
		if err != nil {
			slog.Error("failed to get trending movies", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, movies)
	}
}

// movieDetailsHandler retourne un handler Gin qui récupère les détails d'un film.
//
// Paramètre de chemin requis :
//   - :id : identifiant TMDB du film (entier).
//
// Flux de la requête : validation de l'ID -> appel TMDB /movie/{id} -> réponse JSON.
func movieDetailsHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		movieID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			utils.BadRequest(c, "id invalide")
			return
		}
		details, err := handlers.GetContentDetails(cfg.TMDBToken, "", movieID)
		if err != nil {
			slog.Error("failed to get movie details", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, details)
	}
}

// movieCreditsHandler retourne un handler Gin qui récupère les crédits d'un film.
//
// Paramètre de chemin requis :
//   - :id : identifiant TMDB du film (entier positif).
//
// Flux de la requête : validation de l'ID -> appel TMDB /movie/{id}/credits -> réponse JSON.
func movieCreditsHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		movieID, err := strconv.Atoi(c.Param("id"))
		if err != nil || movieID <= 0 {
			utils.BadRequest(c, "id invalide")
			return
		}
		out, err := handlers.GetMovieCredits(cfg.TMDBToken, "", movieID)
		if err != nil {
			slog.Error("failed to get movie credits", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, out)
	}
}

// movieSimilarHandler retourne un handler Gin qui récupère les films similaires à un film donné.
//
// Paramètre de chemin requis :
//   - :id : identifiant TMDB du film (entier positif).
//
// Flux de la requête : validation de l'ID -> appel TMDB /movie/{id}/similar -> réponse JSON.
func movieSimilarHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		movieID, err := strconv.Atoi(c.Param("id"))
		if err != nil || movieID <= 0 {
			utils.BadRequest(c, "id invalide")
			return
		}
		similar, err := handlers.GetSimilarMovies(cfg.TMDBToken, "", models.MovieGenreMap, movieID, "")
		if err != nil {
			slog.Error("failed to get similar movies", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, similar)
	}
}

// moviesByGenreHandler retourne un handler Gin qui récupère les films filtrés par genre.
//
// Paramètres de query optionnels :
//   - genre_id : identifiant du genre TMDB (défaut "28" = Action).
//   - page : numéro de page (défaut "1").
//   - language : code langue BCP 47 (défaut "fr-FR").
//
// Flux de la requête : extraction et validation des paramètres -> appel TMDB /discover/movie -> réponse JSON.
func moviesByGenreHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		genreIDStr := c.DefaultQuery("genre_id", "28")
		genreID, err := strconv.Atoi(genreIDStr)
		if err != nil {
			utils.BadRequest(c, "genre_id invalide")
			return
		}
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		language := c.DefaultQuery("language", "fr-FR")

		movies, err := handlers.GetMoviesByGenre(cfg.TMDBToken, genreID, page, language)
		if err != nil {
			slog.Error("failed to get movies by genre", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, movies)
	}
}

// movieGenresHandler retourne un handler Gin qui récupère la liste des genres de films disponibles.
//
// Paramètre de query optionnel :
//   - language : code langue BCP 47 (défaut "fr-FR").
//
// Flux de la requête : appel TMDB /genre/movie/list -> réponse JSON avec la map des genres.
func movieGenresHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		language := c.DefaultQuery("language", "fr-FR")
		genreMap, err := handlers.GetMovieGenreCategories(cfg.TMDBToken, language)
		if err != nil {
			slog.Error("failed to get movie genres", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, genreMap)
	}
}

// movieSearchHandler retourne un handler Gin qui effectue une recherche avancée de films.
//
// Paramètres de query :
//   - query : terme de recherche (texte libre).
//   - genres : identifiants de genres séparés par des virgules.
//   - years : années de sortie séparées par des virgules.
//   - sort_by : critère de tri TMDB.
//   - page : numéro de page (défaut "1").
//   - language : code langue BCP 47 (défaut "fr-FR").
//   - rating : note minimale (défaut "0").
//
// Flux de la requête : extraction des paramètres -> appel TMDB /search/movie ou /discover/movie -> réponse JSON.
func movieSearchHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		ratingStr := c.DefaultQuery("rating", "0")
		page, _ := strconv.Atoi(pageStr)
		rating, _ := strconv.ParseFloat(ratingStr, 64)

		movies, err := handlers.SearchMovies(models.SearchMoviesParams{
			BearerToken: cfg.TMDBToken,
			Query:       c.Query("query"),
			GenresCSV:   c.Query("genres"),
			YearsCSV:    c.Query("years"),
			SortBy:      c.Query("sort_by"),
			Page:        page,
			Language:    c.DefaultQuery("language", "fr-FR"),
			Rating:      rating,
		})
		if err != nil {
			slog.Error("failed to search movies", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, movies)
	}
}

// movieImdbIDHandler retourne un handler Gin qui récupère l'identifiant IMDB d'un film
// à partir de son identifiant TMDB.
//
// Paramètre de chemin requis :
//   - :id : identifiant TMDB du film (entier).
//
// Flux de la requête : validation de l'ID -> appel TMDB /movie/{id}/external_ids -> réponse JSON.
func movieImdbIDHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tmdbid, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			utils.BadRequest(c, "id invalide")
			return
		}
		movieImdbId, err := handlers.GetMovieImdbID(cfg.TMDBToken, tmdbid)
		if err != nil {
			slog.Error("failed to get movie IMDB ID", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, movieImdbId)
	}
}

// ===========================================================================
// TV Handlers
// ===========================================================================

// tvTrendingHandler retourne un handler Gin qui récupère les séries TV en tendance.
//
// Paramètres de query optionnels :
//   - time_window : "day" (défaut) ou "week".
//   - language : code langue BCP 47 (défaut "fr-FR").
//   - page : numéro de page (défaut 1).
//
// Flux de la requête : extraction des paramètres -> appel TMDB /trending/tv -> réponse JSON.
func tvTrendingHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		timeWindow := c.DefaultQuery("time_window", "day")
		language := c.DefaultQuery("language", "fr-FR")
		page := 1
		if p := c.Query("page"); p != "" {
			if v, err := strconv.Atoi(p); err == nil {
				page = v
			}
		}
		tv, err := handlers.GetTrendingTV(cfg.TMDBToken, timeWindow, page, language)
		if err != nil {
			slog.Error("failed to get trending TV", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, tv)
	}
}

// tvByGenreHandler retourne un handler Gin qui récupère les séries TV filtrées par genre.
//
// Paramètres de query optionnels :
//   - genre_id : identifiant du genre TMDB (défaut "18" = Drame).
//   - page : numéro de page (défaut "1").
//   - language : code langue BCP 47 (défaut "fr-FR").
//
// Flux de la requête : extraction et validation des paramètres -> appel TMDB /discover/tv -> réponse JSON.
func tvByGenreHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		genreIDStr := c.DefaultQuery("genre_id", "18")
		pageStr := c.DefaultQuery("page", "1")
		language := c.DefaultQuery("language", "fr-FR")

		genreID, err := strconv.Atoi(genreIDStr)
		if err != nil {
			utils.BadRequest(c, "genre_id invalide")
			return
		}
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		tvs, err := handlers.GetTVByGenre(cfg.TMDBToken, genreID, page, language)
		if err != nil {
			slog.Error("failed to get TV by genre", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, tvs)
	}
}

// tvPopularHandler retourne un handler Gin qui récupère les séries TV populaires.
//
// Paramètres de query optionnels :
//   - language : code langue BCP 47 (défaut "fr-FR").
//   - page : numéro de page (défaut "1").
//
// Flux de la requête : extraction des paramètres -> appel TMDB /tv/popular -> réponse JSON.
func tvPopularHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		language := c.DefaultQuery("language", "fr-FR")
		pageStr := c.DefaultQuery("page", "1")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		shows, err := handlers.GetPopularTVShows(cfg.TMDBToken, page, language)
		if err != nil {
			slog.Error("failed to get popular TV shows", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, shows)
	}
}

// tvInfoHandler retourne un handler Gin qui récupère les informations détaillées d'une série TV.
//
// Paramètres de query :
//   - series_id (requis) : identifiant TMDB de la série (entier).
//   - language (optionnel) : code langue BCP 47 (défaut "fr-FR").
//
// Flux de la requête : validation de series_id -> appel TMDB avec détails, saisons,
// crédits et séries similaires en parallèle -> réponse JSON.
func tvInfoHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Query("series_id")
		if idStr == "" {
			utils.BadRequest(c, "series_id requis")
			return
		}
		seriesID, err := strconv.Atoi(idStr)
		if err != nil {
			utils.BadRequest(c, "series_id invalide")
			return
		}
		language := c.DefaultQuery("language", "fr-FR")
		info, err := handlers.GetTVInfo(cfg.TMDBToken, seriesID, language)
		if err != nil {
			slog.Error("failed to get TV info", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, info)
	}
}

// tvSearchHandler retourne un handler Gin qui effectue une recherche de séries TV par mot-clé.
//
// Paramètres de query :
//   - query (requis) : terme de recherche.
//   - language (optionnel) : code langue BCP 47 (défaut "fr-FR").
//   - page (optionnel) : numéro de page (défaut "1").
//
// Flux de la requête : validation de query -> appel TMDB /search/tv -> réponse JSON.
func tvSearchHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("query")
		if query == "" {
			utils.BadRequest(c, "query requis")
			return
		}
		language := c.DefaultQuery("language", "fr-FR")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		results, err := handlers.SearchTV(cfg.TMDBToken, query, language, page)
		if err != nil {
			slog.Error("failed to search TV", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		utils.RespondSuccess(c, http.StatusOK, results)
	}
}

// ===========================================================================
// Player Handlers
// ===========================================================================

// videoPlayerMovieHandler retourne un handler Gin qui résout un flux vidéo pour un film.
//
// Paramètre de chemin requis :
//   - :id : identifiant TMDB du film (entier).
//
// Flux de la requête (pipeline complet) :
//  1. Conversion de l'ID TMDB en ID IMDB via l'API TMDB.
//  2. Récupération des streams Torrentio pour l'ID IMDB.
//  3. Ajout du magnet du premier stream sur Real-Debrid.
//  4. Sélection des fichiers du torrent sur Real-Debrid.
//  5. Récupération des liens de téléchargement.
//  6. Dé-restriction du lien et obtention de l'URL MPD.
//  7. Construction de la réponse du lecteur vidéo.
func videoPlayerMovieHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		imdbid, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			utils.BadRequest(c, "id invalide")
			return
		}

		slog.Info("video player request", slog.Int("tmdb_id", imdbid))

		movieImdbId, err := handlers.GetMovieImdbID(cfg.TMDBToken, imdbid)
		if err != nil {
			slog.Error("failed to get IMDB ID", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		streams, err := handlers.GetTorrentioMoviesStreams(movieImdbId.ImdbId)
		if err != nil {
			slog.Error("failed to get torrentio streams", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		if len(streams.Streams) == 0 {
			utils.NotFound(c, "Aucun stream trouve")
			return
		}

		debrid, err := handlers.AddMagnetRealDebrid(cfg.RealDebridToken, streams.Streams[0].InfoHash)
		if err != nil {
			slog.Error("failed to add magnet", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		if err := handlers.SelectFilesRealDebrid(cfg.RealDebridToken, debrid.Id); err != nil {
			slog.Error("failed to select files", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		info, err := handlers.GetRealDebridTorrentInfo(cfg.RealDebridToken, debrid.Id)
		if err != nil {
			slog.Error("failed to get torrent info", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		if len(info.Links) == 0 {
			utils.NotFound(c, "Aucun lien disponible")
			return
		}

		fileid, err := handlers.UnrestrictAndGetMPD(cfg.RealDebridToken, info.Links[0])
		if err != nil {
			slog.Error("failed to unrestrict link", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		videoPlayer, err := handlers.GetVideoPlayer(cfg.RealDebridToken, fileid)
		if err != nil {
			slog.Error("failed to get video player", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		utils.RespondSuccess(c, http.StatusOK, videoPlayer)
	}
}

// videoPlayerSeriesHandler retourne un handler Gin qui résout un flux vidéo pour un épisode de série.
//
// Paramètres de chemin requis :
//   - :id : identifiant TMDB de la série (entier).
//   - :season : numéro de la saison.
//   - :episode : numéro de l'épisode.
//
// Flux de la requête (pipeline complet) :
//  1. Conversion de l'ID TMDB en ID IMDB via l'API TMDB.
//  2. Construction de l'identifiant Torrentio au format "imdb_id:saison:épisode".
//  3. Récupération des streams Torrentio pour l'épisode.
//  4. Ajout du magnet du premier stream sur Real-Debrid.
//  5. Sélection des fichiers, dé-restriction et construction de la réponse.
func videoPlayerSeriesHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		imdbid, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			utils.BadRequest(c, "id invalide")
			return
		}
		season := c.Param("season")
		episode := c.Param("episode")

		slog.Info("video player series request",
			slog.Int("tmdb_id", imdbid),
			slog.String("season", season),
			slog.String("episode", episode),
		)

		movieImdbId, err := handlers.GetSeriesImdbID(cfg.TMDBToken, imdbid)
		if err != nil {
			slog.Error("failed to get series IMDB ID", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		idParamFinal := fmt.Sprintf("%s:%s:%s", movieImdbId.ImdbId, season, episode)

		streams, err := handlers.GetTorrentioSeriesStreams(idParamFinal)
		if err != nil {
			slog.Error("failed to get torrentio streams", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		if len(streams.Streams) == 0 {
			utils.NotFound(c, "Aucun stream trouve")
			return
		}

		debrid, err := handlers.AddMagnetRealDebrid(cfg.RealDebridToken, streams.Streams[0].InfoHash)
		if err != nil {
			slog.Error("failed to add magnet", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		if err := handlers.SelectFilesRealDebrid(cfg.RealDebridToken, debrid.Id); err != nil {
			slog.Error("failed to select files", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		info, err := handlers.GetRealDebridTorrentInfo(cfg.RealDebridToken, debrid.Id)
		if err != nil {
			slog.Error("failed to get torrent info", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}
		if len(info.Links) == 0 {
			utils.NotFound(c, "Aucun lien disponible")
			return
		}

		fileid, err := handlers.UnrestrictAndGetMPD(cfg.RealDebridToken, info.Links[0])
		if err != nil {
			slog.Error("failed to unrestrict link", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		videoPlayer, err := handlers.GetVideoPlayer(cfg.RealDebridToken, fileid)
		if err != nil {
			slog.Error("failed to get video player", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		utils.RespondSuccess(c, http.StatusOK, videoPlayer)
	}
}

// ===========================================================================
// ZT Handlers
// ===========================================================================

// ztSearchHandler retourne un handler Gin qui effectue une recherche paginée sur Zone Téléchargement.
//
// Paramètres de query requis :
//   - category : catégorie de contenu ("films" ou "series").
//   - query : terme de recherche.
//
// Paramètre de query optionnel :
//   - page : numéro de page (défaut "1").
//
// Flux de la requête : validation des paramètres -> scraping ZT -> réponse JSON.
func ztSearchHandler(parser *handlers.ZtParserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Query("category")
		query := c.Query("query")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

		if category == "" || query == "" {
			utils.BadRequest(c, "category and query are required")
			return
		}

		results, err := parser.Search(category, query, page)
		if err != nil {
			slog.Error("failed to search ZT", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		utils.RespondSuccess(c, http.StatusOK, results)
	}
}

// ztSearchAllHandler retourne un handler Gin qui effectue une recherche exhaustive
// sur toutes les pages disponibles de Zone Téléchargement.
//
// Paramètres de query requis :
//   - category : catégorie de contenu ("films" ou "series").
//   - query : terme de recherche.
//
// Attention : cette requête peut être lente car elle parcourt toutes les pages
// de résultats avec le rate-limiting configuré.
func ztSearchAllHandler(parser *handlers.ZtParserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Query("category")
		query := c.Query("query")

		if category == "" || query == "" {
			utils.BadRequest(c, "category and query are required")
			return
		}

		results, err := parser.SearchAll(category, query)
		if err != nil {
			slog.Error("failed to search all ZT", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		utils.RespondSuccess(c, http.StatusOK, results)
	}
}

// ztBasicHandler retourne un handler Gin qui récupère les informations de base
// d'un contenu Zone Téléchargement (titre, titre original, liens de téléchargement).
//
// Paramètres de chemin requis :
//   - :category : catégorie de contenu ("films" ou "series").
//   - :id : identifiant numérique du contenu sur ZT.
//
// Flux de la requête : scraping de la page ZT -> extraction des métadonnées -> réponse JSON.
func ztBasicHandler(parser *handlers.ZtParserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Param("category")
		id := c.Param("id")

		basicInfo, err := parser.GetMovieNameFromId(category, id)
		if err != nil {
			slog.Error("failed to get ZT basic info", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		utils.RespondSuccess(c, http.StatusOK, basicInfo)
	}
}

// ztDetailsHandler retourne un handler Gin qui récupère les détails complets
// d'un contenu Zone Téléchargement enrichi via TMDB (films) ou TVMaze (séries).
//
// Paramètres de chemin requis :
//   - :category : catégorie de contenu ("films" ou "series").
//   - :id : identifiant numérique du contenu sur ZT.
//
// Flux de la requête : scraping ZT -> enrichissement TMDB/TVMaze -> réponse JSON.
func ztDetailsHandler(parser *handlers.ZtParserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Param("category")
		id := c.Param("id")

		movieData, err := parser.GetQueryDatas(category, id)
		if err != nil {
			slog.Error("failed to get ZT details", slog.String("error", err.Error()))
			utils.InternalError(c, err)
			return
		}

		utils.RespondSuccess(c, http.StatusOK, movieData)
	}
}
