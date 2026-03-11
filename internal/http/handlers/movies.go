package handlers

import (
	"StreamflixBackend/internal/cache"
	"StreamflixBackend/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

// ============================================================================
// CACHE KEYS
// ============================================================================

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

type contentDetailsCacheKey struct {
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

// ============================================================================
// CACHES
// ============================================================================

var (
	popularMoviesCache   = cache.New[popularMoviesCacheKey, []models.MovieDTO](30 * time.Minute)
	topRatedMoviesCache  = cache.New[topRatedMoviesCacheKey, []models.MovieDTO](30 * time.Minute)
	trendingMoviesCache  = cache.New[trendingMoviesCacheKey, []models.MovieDTO](15 * time.Minute)
	contentDetailsCache  = cache.New[contentDetailsCacheKey, *models.ContentDetailsDTO](60 * time.Minute)
	similarMoviesCache   = cache.New[similarMoviesCacheKey, []models.MovieDTO](30 * time.Minute)
	movieCreditsCache    = cache.New[movieCreditsCacheKey, *models.MovieCreditsDTO](60 * time.Minute)
	movieImdbIDCache     = cache.New[movieImdbIDCacheKey, models.TmdbMovieImdbId](60 * time.Minute)
	genreCategoriesCache = cache.New[genreCategoriesCacheKey, []models.CategoryDTO](24 * time.Hour)
)

// ============================================================================
// FAKE DATA (inchangé)
// ============================================================================

var generateMovies = []models.Movie{}
var generatedContentDetails = models.ContentDetails{}

func RandomMovieList() []models.Movie {
	if len(generateMovies) > 0 {
		return generateMovies
	}
	var movies []models.Movie
	for i := 0; i < gofakeit.Number(10, 20); i++ {
		movies = append(movies, models.Movie{
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

func GetContentDetailsRandomized() models.ContentDetails {
	if len(generatedContentDetails.Cast) > 0 {
		return generatedContentDetails
	}
	var details models.ContentDetails
	castCount := gofakeit.Number(3, 6)
	var cast []models.Cast
	for j := 0; j < castCount; j++ {
		cast = append(cast, models.Cast{
			Name:  gofakeit.Name(),
			Role:  gofakeit.JobTitle(),
			Image: "https://i.pravatar.cc/150?img=" + strconv.Itoa(gofakeit.Number(1, 70)),
		})
	}
	details = models.ContentDetails{
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

// ============================================================================
// HANDLERS
// ============================================================================

func GetPopularMovies(tmdbBearerToken string, imageBase string, genreMap map[int]string, page string) ([]models.MovieDTO, error) {
	if imageBase == "" {
		imageBase = "https://image.tmdb.org/t/p/w500"
	}
	if tmdbBearerToken == "" {
		return nil, fmt.Errorf("TMDB bearer token missing")
	}
	if page == "" {
		page = "1"
	}
	if _, err := strconv.Atoi(page); err != nil {
		return nil, fmt.Errorf("invalid page")
	}

	key := popularMoviesCacheKey{page: page, language: "fr-FR"}
	if cached, ok := popularMoviesCache.Get(key); ok {
		return cached, nil
	}

	u, _ := url.Parse("https://api.themoviedb.org/3/movie/popular")
	q := u.Query()
	q.Set("page", page)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tmdbBearerToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tmdb error: status %d", resp.StatusCode)
	}

	var tmdbRes models.TmdbPopularResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmdbRes); err != nil {
		return nil, fmt.Errorf("failed to decode tmdb response")
	}

	out := make([]models.MovieDTO, 0, len(tmdbRes.Results))
	for _, m := range tmdbRes.Results {
		year := 0
		if m.ReleaseDate != "" {
			if t, err := time.Parse("2006-01-02", m.ReleaseDate); err == nil {
				year = t.Year()
			}
		}
		image := ""
		if m.PosterPath != "" {
			image = fmt.Sprintf("%s%s", imageBase, m.PosterPath)
		}
		genres := make([]string, 0, len(m.GenreIDs))
		for _, gid := range m.GenreIDs {
			if name, ok := genreMap[gid]; ok && name != "" {
				genres = append(genres, name)
			}
		}
		out = append(out, models.MovieDTO{
			ID:     m.ID,
			Title:  m.Title,
			Image:  image,
			Year:   year,
			Genre:  genres,
			Rating: m.VoteAverage,
		})
	}

	popularMoviesCache.Set(key, out)
	return out, nil
}

func GetTopRatedMovies(bearerToken string, page int) ([]models.MovieDTO, error) {
	key := topRatedMoviesCacheKey{page: page}
	if cached, ok := topRatedMoviesCache.Get(key); ok {
		return cached, nil
	}

	u := fmt.Sprintf("https://api.themoviedb.org/3/movie/top_rated?language=fr-FR&page=%d", page)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", res.StatusCode)
	}

	var tmdbResp models.TMDBResponse
	if err := json.NewDecoder(res.Body).Decode(&tmdbResp); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	dtos := make([]models.MovieDTO, 0, len(tmdbResp.Results))
	for _, raw := range tmdbResp.Results {
		year := 0
		if len(raw.ReleaseDate) >= 4 {
			fmt.Sscanf(raw.ReleaseDate[:4], "%d", &year)
		}
		genres := make([]string, 0)
		for _, id := range raw.GenreIDs {
			if name, ok := models.MovieGenreMap[id]; ok {
				genres = append(genres, name)
			}
		}
		dtos = append(dtos, models.MovieDTO{
			ID:     raw.ID,
			Title:  raw.Title,
			Image:  "https://image.tmdb.org/t/p/w500" + raw.PosterPath,
			Year:   year,
			Genre:  genres,
			Rating: raw.VoteAverage,
		})
	}

	topRatedMoviesCache.Set(key, dtos)
	return dtos, nil
}

func GetTrendingMovies(bearerToken, timeWindow string, page int, language string) ([]models.MovieDTO, error) {
	if timeWindow == "" {
		timeWindow = "day"
	}
	if page <= 0 {
		page = 1
	}
	if language == "" {
		language = "fr-FR"
	}

	key := trendingMoviesCacheKey{timeWindow: timeWindow, page: page, language: language}
	if cached, ok := trendingMoviesCache.Get(key); ok {
		return cached, nil
	}

	u := fmt.Sprintf(
		"https://api.themoviedb.org/3/trending/movie/%s?language=%s&page=%d",
		timeWindow, language, page,
	)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status: %d body=%s", res.StatusCode, string(b))
	}

	var tmdbResp models.TMDBTrendingResponse
	if err := json.NewDecoder(res.Body).Decode(&tmdbResp); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	dtos := make([]models.MovieDTO, 0, len(tmdbResp.Results))
	for _, raw := range tmdbResp.Results {
		year := 0
		if len(raw.ReleaseDate) >= 4 {
			fmt.Sscanf(raw.ReleaseDate[:4], "%d", &year)
		}
		genres := make([]string, 0, len(raw.GenreIDs))
		for _, id := range raw.GenreIDs {
			if name, ok := models.MovieGenreMap[id]; ok {
				genres = append(genres, name)
			}
		}
		image := ""
		if raw.PosterPath != "" {
			image = "https://image.tmdb.org/t/p/w500" + raw.PosterPath
		}
		dtos = append(dtos, models.MovieDTO{
			ID:     raw.ID,
			Title:  raw.Title,
			Image:  image,
			Year:   year,
			Genre:  genres,
			Rating: raw.VoteAverage,
		})
	}

	trendingMoviesCache.Set(key, dtos)
	return dtos, nil
}

func GetContentDetails(tmdbBearerToken string, imageBase string, movieID int) (*models.ContentDetailsDTO, error) {
	if imageBase == "" {
		imageBase = "https://image.tmdb.org/t/p/w500"
	}
	if tmdbBearerToken == "" {
		return nil, fmt.Errorf("TMDB bearer token missing")
	}

	key := contentDetailsCacheKey{movieID: movieID}
	if cached, ok := contentDetailsCache.Get(key); ok {
		return cached, nil
	}

	u := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d", movieID)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tmdbBearerToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tmdb error: status %d", resp.StatusCode)
	}

	var tmdbRes models.TmdbMovieDetails
	if err := json.NewDecoder(resp.Body).Decode(&tmdbRes); err != nil {
		return nil, fmt.Errorf("failed to decode tmdb response")
	}

	year := 0
	if tmdbRes.ReleaseDate != "" {
		if t, err := time.Parse("2006-01-02", tmdbRes.ReleaseDate); err == nil {
			year = t.Year()
		}
	}

	image := ""
	if tmdbRes.PosterPath != "" {
		image = fmt.Sprintf("%s%s", imageBase, tmdbRes.PosterPath)
	}

	backdropImage := ""
	if tmdbRes.BackdropPath != "" {
		backdropImage = fmt.Sprintf("https://image.tmdb.org/t/p/original%s", tmdbRes.BackdropPath)
	}

	genres := make([]string, 0, len(tmdbRes.Genres))
	for _, g := range tmdbRes.Genres {
		genres = append(genres, g.Name)
	}

	duration := ""
	if tmdbRes.Runtime > 0 {
		hours := tmdbRes.Runtime / 60
		minutes := tmdbRes.Runtime % 60
		if hours > 0 {
			duration = fmt.Sprintf("%dh %dmin", hours, minutes)
		} else {
			duration = fmt.Sprintf("%dmin", minutes)
		}
	}

	languages := ""
	if len(tmdbRes.SpokenLanguages) > 0 {
		langNames := make([]string, 0, len(tmdbRes.SpokenLanguages))
		for _, lang := range tmdbRes.SpokenLanguages {
			langNames = append(langNames, lang.Name)
		}
		languages = strings.Join(langNames, ", ")
	}

	producer := ""
	if len(tmdbRes.ProductionCompanies) > 0 {
		producer = tmdbRes.ProductionCompanies[0].Name
	}

	result := &models.ContentDetailsDTO{
		ID:            tmdbRes.ID,
		Title:         tmdbRes.Title,
		Image:         image,
		Imdbid:        tmdbRes.ImdbId,
		BackdropImage: backdropImage,
		Year:          year,
		Genres:        genres,
		Rating:        tmdbRes.VoteAverage,
		Duration:      duration,
		Synopsis:      tmdbRes.Overview,
		Languages:     languages,
		Producer:      producer,
	}

	contentDetailsCache.Set(key, result)
	return result, nil
}

func GetSimilarMovies(tmdbBearerToken string, imageBase string, genreMap map[int]string, movieID int, page string) ([]models.MovieDTO, error) {
	if imageBase == "" {
		imageBase = "https://image.tmdb.org/t/p/w500"
	}
	if tmdbBearerToken == "" {
		return nil, fmt.Errorf("TMDB bearer token missing")
	}
	if page == "" {
		page = "1"
	}

	key := similarMoviesCacheKey{movieID: movieID, page: page}
	if cached, ok := similarMoviesCache.Get(key); ok {
		return cached, nil
	}

	u, _ := url.Parse(fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/similar", movieID))
	q := u.Query()
	q.Set("page", page)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tmdbBearerToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tmdb error: status %d", resp.StatusCode)
	}

	var tmdbRes models.TmdbPopularResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmdbRes); err != nil {
		return nil, fmt.Errorf("failed to decode tmdb response")
	}

	out := make([]models.MovieDTO, 0, len(tmdbRes.Results))
	for _, m := range tmdbRes.Results {
		year := 0
		if m.ReleaseDate != "" {
			if t, err := time.Parse("2006-01-02", m.ReleaseDate); err == nil {
				year = t.Year()
			}
		}
		image := ""
		if m.PosterPath != "" {
			image = fmt.Sprintf("%s%s", imageBase, m.BackdropPath)
		}
		genres := make([]string, 0, len(m.GenreIDs))
		for _, gid := range m.GenreIDs {
			if name, ok := genreMap[gid]; ok && name != "" {
				genres = append(genres, name)
			}
		}
		out = append(out, models.MovieDTO{
			ID:     m.ID,
			Title:  m.Title,
			Image:  image,
			Year:   year,
			Genre:  genres,
			Rating: m.VoteAverage,
		})
	}

	similarMoviesCache.Set(key, out)
	return out, nil
}

func GetMovieCredits(tmdbToken string, imageBase string, movieID int) (*models.MovieCreditsDTO, error) {
	if imageBase == "" {
		imageBase = "https://image.tmdb.org/t/p/w500"
	}

	key := movieCreditsCacheKey{movieID: movieID}
	if cached, ok := movieCreditsCache.Get(key); ok {
		return cached, nil
	}

	u := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/credits", movieID)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Authorization", "Bearer "+tmdbToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tmdb returned %d", resp.StatusCode)
	}

	var tmdb models.TmdbMovieCredits
	if err := json.NewDecoder(resp.Body).Decode(&tmdb); err != nil {
		return nil, err
	}

	out := &models.MovieCreditsDTO{
		Cast: make([]models.CastMemberMoviesDTO, 0),
	}

	for _, c := range tmdb.Crew {
		if c.Job == "Director" {
			out.Director = c.Name
			break
		}
	}
	for _, c := range tmdb.Crew {
		if c.Job == "Producer" || c.Job == "Executive Producer" {
			out.Producer = c.Name
			break
		}
	}
	for _, c := range tmdb.Crew {
		if c.Job == "Screenplay" || c.Job == "Writer" || c.Job == "Novel" {
			out.Writer = c.Name
			break
		}
	}

	for i, actor := range tmdb.Cast {
		if i >= 12 {
			break
		}
		image := ""
		if actor.ProfilePath != "" {
			image = imageBase + actor.ProfilePath
		}
		out.Cast = append(out.Cast, models.CastMemberMoviesDTO{
			Name:  actor.Name,
			Role:  actor.Character,
			Image: image,
		})
	}

	movieCreditsCache.Set(key, out)
	return out, nil
}

func GetMovieImdbID(tmdbBearerToken string, movieID int) (models.TmdbMovieImdbId, error) {
	log.Printf("   → GetMovieImdbID: movieID=%d", movieID)

	if tmdbBearerToken == "" {
		return models.TmdbMovieImdbId{}, fmt.Errorf("TMDB bearer token missing")
	}

	key := movieImdbIDCacheKey{movieID: movieID}
	if cached, ok := movieImdbIDCache.Get(key); ok {
		log.Printf("   ✅ Cache hit IMDB ID: %s", cached.ImdbId)
		return cached, nil
	}

	u := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d", movieID)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return models.TmdbMovieImdbId{}, err
	}
	req.Header.Set("Authorization", "Bearer "+tmdbBearerToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.TmdbMovieImdbId{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return models.TmdbMovieImdbId{}, fmt.Errorf("tmdb error: status %d body=%s", resp.StatusCode, string(body))
	}

	var tmdbRes models.TmdbMovieImdbId
	if err := json.NewDecoder(resp.Body).Decode(&tmdbRes); err != nil {
		return models.TmdbMovieImdbId{}, fmt.Errorf("failed to decode tmdb response")
	}

	movieImdbIDCache.Set(key, tmdbRes)
	log.Printf("   ✅ IMDB ID trouvé: %s", tmdbRes.ImdbId)
	return tmdbRes, nil
}

func GetSeriesImdbID(tmdbBearerToken string, movieID int) (models.TmdbMovieImdbId, error) {
	log.Printf("   → GetMovieImdbID: movieID=%d", movieID)

	if tmdbBearerToken == "" {
		return models.TmdbMovieImdbId{}, fmt.Errorf("TMDB bearer token missing")
	}

	key := movieImdbIDCacheKey{movieID: movieID}
	if cached, ok := movieImdbIDCache.Get(key); ok {
		log.Printf("   ✅ Cache hit IMDB ID: %s", cached.ImdbId)
		return cached, nil
	}

	u := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/external_ids", movieID)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return models.TmdbMovieImdbId{}, err
	}
	req.Header.Set("Authorization", "Bearer "+tmdbBearerToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.TmdbMovieImdbId{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return models.TmdbMovieImdbId{}, fmt.Errorf("tmdb error: status %d body=%s", resp.StatusCode, string(body))
	}

	var tmdbRes models.TmdbMovieImdbId
	if err := json.NewDecoder(resp.Body).Decode(&tmdbRes); err != nil {
		return models.TmdbMovieImdbId{}, fmt.Errorf("failed to decode tmdb response")
	}

	movieImdbIDCache.Set(key, tmdbRes)
	log.Printf("   ✅ IMDB ID trouvé: %s", tmdbRes.ImdbId)
	return tmdbRes, nil
}

func GetMoviesByGenre(bearerToken string, genreID int, page int, language string) ([]models.MovieDTO, error) {
	const imageBaseURL = "https://image.tmdb.org/t/p/w500"

	key := models.MovieGenreCacheKey{GenreID: genreID, Page: page, Language: language}
	if cached, ok := models.MoviesByGenreCache.Get(key); ok {
		return cached, nil
	}

	u := fmt.Sprintf(
		"https://api.themoviedb.org/3/discover/movie?with_genres=%d&sort_by=vote_average.desc&vote_count.gte=100&page=%d&language=%s",
		genreID, page, language,
	)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Results []models.TMDBMovieRaw `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	dtos := make([]models.MovieDTO, 0, len(result.Results))
	for _, m := range result.Results {
		year, _ := strconv.Atoi(safeYear(m.ReleaseDate))
		dtos = append(dtos, models.MovieDTO{
			ID:     m.ID,
			Title:  m.Title,
			Image:  imageBaseURL + m.PosterPath,
			Year:   year,
			Genre:  mapGenres(m.GenreIDs),
			Rating: m.VoteAverage,
		})
	}

	models.MoviesByGenreCache.Set(key, dtos)
	return dtos, nil
}

func GetMovieGenreCategories(bearerToken, language string) ([]models.CategoryDTO, error) {
	if strings.TrimSpace(language) == "" {
		language = "fr-FR"
	}

	key := genreCategoriesCacheKey{language: language}
	if cached, ok := genreCategoriesCache.Get(key); ok {
		return cached, nil
	}

	u, _ := url.Parse("https://api.themoviedb.org/3/genre/movie/list")
	q := u.Query()
	q.Set("language", language)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("tmdb error: status=%d body=%s", res.StatusCode, string(b))
	}

	var decoded models.TMDBGenreMovieListResponse
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, err
	}

	out := make([]models.CategoryDTO, 0, len(decoded.Genres))
	for _, g := range decoded.Genres {
		color := models.GenreCategoryColor[g.Name]
		if color == "" {
			color = "from-slate-600 to-slate-800"
		}
		out = append(out, models.CategoryDTO{
			ID:           g.ID,
			CategoryName: g.Name,
			Description:  genreDescription(g.Name),
			Href:         genreHref(g.ID),
			Color:        color,
		})
	}

	genreCategoriesCache.Set(key, out)
	return out, nil
}

// ============================================================================
// REAL-DEBRID & TORRENTIO (pas de cache — requêtes transactionnelles)
// ============================================================================

func GetTorrentioMoviesStreams(imdbID string) (*models.TorrentioResponse, error) {
	log.Printf("   → GetTorrentioStreams: imdbID=%s", imdbID)

	u := fmt.Sprintf("https://torrentio.strem.fun/stream/movie/%s.json", imdbID)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "curl/8.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Torrentio error: %d body=%s", resp.StatusCode, string(body))
	}

	var result models.TorrentioResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	log.Printf("   ✅ %d streams récupérés", len(result.Streams))
	return &result, nil
}
func GetTorrentioSeriesStreams(imdbID string) (*models.TorrentioResponse, error) {
	log.Printf("   → GetTorrentioStreams: imdbID=%s", imdbID)

	u := fmt.Sprintf("https://torrentio.strem.fun/stream/series/%s.json", imdbID)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "curl/8.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Torrentio error: %d body=%s", resp.StatusCode, string(body))
	}

	var result models.TorrentioResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	log.Printf("   ✅ %d streams récupérés", len(result.Streams))
	return &result, nil
}

func AddMagnetRealDebrid(apiKey, infoHash string) (*models.RdAddMagnetResponse, error) {
	log.Printf("   → AddMagnetRealDebrid: InfoHash=%s", infoHash)

	apiURL := "https://api.real-debrid.com/rest/1.0/torrents/addMagnet"
	magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s", infoHash)

	data := url.Values{}
	data.Set("magnet", magnet)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.RdAddMagnetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	log.Printf("   ✅ Torrent ID: %s", result.Id)
	return &result, nil
}

func SelectFilesRealDebrid(apiKey, torrentId string) error {
	log.Printf("   → SelectFilesRealDebrid: TorrentID=%s", torrentId)

	apiURL := fmt.Sprintf("https://api.real-debrid.com/rest/1.0/torrents/selectFiles/%s", torrentId)
	data := url.Values{}
	data.Set("files", "all")

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("select files error: %s", string(body))
	}

	log.Println("   ✅ Fichiers sélectionnés")
	return nil
}

func GetRealDebridTorrentInfo(apiKey, torrentId string) (*models.RdTorrentInfo, error) {
	log.Printf("   → GetRealDebridTorrentInfo: TorrentID=%s", torrentId)

	apiURL := fmt.Sprintf("https://api.real-debrid.com/rest/1.0/torrents/info/%s", torrentId)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info models.RdTorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	log.Printf("   ✅ Status=%s Progress=%.2f%%", info.Status, info.Progress)
	return &info, nil
}

func UnrestrictRealDebridLink(apiKey, rawLink string) (*models.RdUnrestrictResponse, error) {
	log.Printf("   → UnrestrictRealDebridLink: RawLink=%s", rawLink)

	cleanLink := strings.ReplaceAll(rawLink, `\/`, `/`)
	apiURL := "https://api.real-debrid.com/rest/1.0/unrestrict/link"

	data := url.Values{}
	data.Set("link", cleanLink)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info models.RdUnrestrictResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	log.Printf("   ✅ Lien direct: %s", info.Download)
	return &info, nil
}

// ============================================================================
// SEARCH (conserve son propre cache map existant)
// ============================================================================

func SearchMovies(p models.SearchMoviesParams) ([]models.MovieDTO, error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if strings.TrimSpace(p.Language) == "" {
		p.Language = "fr-FR"
	}
	if strings.TrimSpace(p.SortBy) == "" {
		p.SortBy = "popularity.desc"
	}

	hasQuery := strings.TrimSpace(p.Query) != ""
	hasGenres := strings.TrimSpace(p.GenresCSV) != ""
	years := parseIntsCSV(p.YearsCSV)
	hasYear := len(years) > 0

	if !hasQuery && !hasGenres && !hasYear {
		return nil, fmt.Errorf("au moins un paramètre requis : query, genres ou year")
	}

	if !hasYear {
		years = []int{0}
	}

	wantedGenres := parseGenresCSV(p.GenresCSV)
	seen := make(map[int]bool)
	out := make([]models.MovieDTO, 0)

	for _, year := range years {
		cacheKey := buildCacheKey(p, year)

		models.CacheMutex.RLock()
		cached, found := models.SearchCache[cacheKey]
		models.CacheMutex.RUnlock()

		var results []models.TMDBMovieRaw

		if found && time.Now().Before(cached.ExpiresAt) {
			results = cached.Results
		} else {
			var err error
			results, err = fetchMoviesPage(models.FetchParams{
				BearerToken: p.BearerToken,
				Query:       p.Query,
				Year:        year,
				SortBy:      p.SortBy,
				Page:        p.Page,
				Language:    p.Language,
				HasQuery:    hasQuery,
			})
			if err != nil {
				return nil, err
			}
			models.CacheMutex.Lock()
			models.SearchCache[cacheKey] = models.CachedSearch{
				Results:   results,
				ExpiresAt: time.Now().Add(models.CacheTTL),
			}
			models.CacheMutex.Unlock()
		}

		for _, m := range results {
			if len(wantedGenres) > 0 && !hasAllGenres(m.GenreIDs, wantedGenres) {
				continue
			}
			dto := toMovieDTO(m)
			if p.Rating > 0 && dto.Rating < p.Rating {
				continue
			}
			if !seen[m.ID] {
				seen[m.ID] = true
				out = append(out, dto)
			}
		}
	}

	return out, nil
}

// ============================================================================
// HELPERS
// ============================================================================

func safeYear(date string) string {
	if len(date) >= 4 {
		return date[:4]
	}
	return "0"
}

func mapGenres(ids []int) []string {
	genres := make([]string, 0, len(ids))
	for _, id := range ids {
		if name, ok := models.MovieGenreMap[id]; ok {
			genres = append(genres, name)
		}
	}
	return genres
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

func parseIntsCSV(s string) []int {
	result := []int{}
	for _, part := range splitCSV(s) {
		if n, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			result = append(result, n)
		}
	}
	return result
}

func parseGenresCSV(s string) map[int]bool {
	result := map[int]bool{}
	for _, part := range splitCSV(s) {
		if id, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			result[id] = true
		}
	}
	return result
}

func toMovieDTO(m models.TMDBMovieRaw) models.MovieDTO {
	year := 0
	if len(m.ReleaseDate) >= 4 {
		if y, err := strconv.Atoi(m.ReleaseDate[:4]); err == nil {
			year = y
		}
	}
	image := ""
	if m.PosterPath != "" {
		image = "https://image.tmdb.org/t/p/w500" + m.PosterPath
	}
	genres := make([]string, 0, len(m.GenreIDs))
	for _, id := range m.GenreIDs {
		if name, ok := models.MovieGenreMap[id]; ok {
			genres = append(genres, name)
		}
	}
	return models.MovieDTO{
		ID:     m.ID,
		Title:  m.Title,
		Image:  image,
		Year:   year,
		Genre:  genres,
		Rating: m.VoteAverage,
	}
}

func buildCacheKey(p models.SearchMoviesParams, year int) string {
	return fmt.Sprintf("q=%s|y=%d|page=%d|lang=%s",
		strings.TrimSpace(p.Query), year, p.Page, p.Language)
}

func hasAllGenres(movieGenres []int, wanted map[int]bool) bool {
	movieSet := make(map[int]bool, len(movieGenres))
	for _, id := range movieGenres {
		movieSet[id] = true
	}
	for id := range wanted {
		if !movieSet[id] {
			return false
		}
	}
	return true
}

func fetchMoviesPage(p models.FetchParams) ([]models.TMDBMovieRaw, error) {
	var endpoint *url.URL

	if p.HasQuery {
		endpoint, _ = url.Parse("https://api.themoviedb.org/3/search/movie")
		q := endpoint.Query()
		q.Set("query", p.Query)
		q.Set("page", strconv.Itoa(p.Page))
		q.Set("language", p.Language)
		if p.Year > 0 {
			q.Set("primary_release_year", strconv.Itoa(p.Year))
		}
		endpoint.RawQuery = q.Encode()
	} else {
		endpoint, _ = url.Parse("https://api.themoviedb.org/3/discover/movie")
		q := endpoint.Query()
		if p.Genre != "" {
			q.Set("with_genres", p.Genre)
		}
		q.Set("sort_by", p.SortBy)
		q.Set("page", strconv.Itoa(p.Page))
		q.Set("language", p.Language)
		if p.Year > 0 {
			q.Set("primary_release_year", strconv.Itoa(p.Year))
		}
		endpoint.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.BearerToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tmdb error: status=%d", resp.StatusCode)
	}

	var decoded models.TMDBDiscoverResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}

	return decoded.Results, nil
}

func genreDescription(name string) string {
	return "Découvrez les meilleurs films du genre " + name + "."
}

func genreHref(id int) string {
	return "/genres/" + strconv.Itoa(id)
}
