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

// NewRouter creates and configures the Gin engine with all routes and middleware.
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

func registerPlayerRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	player := rg.Group("/player")
	{
		player.GET("/movie/:id", videoPlayerMovieHandler(cfg))
		player.GET("/series/:id/:season/:episode", videoPlayerSeriesHandler(cfg))
	}
}

func registerZTRoutes(rg *gin.RouterGroup, parser *handlers.ZtParserService) {
	zt := rg.Group("/zt")
	{
		zt.GET("/search", ztSearchHandler(parser))
		zt.GET("/search-all", ztSearchAllHandler(parser))
		zt.GET("/basic/:category/:id", ztBasicHandler(parser))
		zt.GET("/:category/:id", ztDetailsHandler(parser))
	}
}

func registerUserRoutes(rg *gin.RouterGroup) {
	user := rg.Group("/user")
	{
		user.GET("/list", func(c *gin.Context) {
			utils.RespondSuccess(c, http.StatusOK, handlers.GetUserListItems())
		})
	}
}

// registerLegacyRoutes preserves backward compatibility with old endpoints.
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
