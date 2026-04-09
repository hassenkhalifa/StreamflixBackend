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

// popularMoviesCacheKey est la clé de cache pour les films populaires,
// discriminée par numéro de page et langue.
type popularMoviesCacheKey struct {
	page     string
	language string
}

// topRatedMoviesCacheKey est la clé de cache pour les films les mieux notés,
// discriminée par numéro de page.
type topRatedMoviesCacheKey struct {
	page int
}

// trendingMoviesCacheKey est la clé de cache pour les films tendances,
// discriminée par fenêtre temporelle (day/week), page et langue.
type trendingMoviesCacheKey struct {
	timeWindow string
	page       int
	language   string
}

// contentDetailsCacheKey est la clé de cache pour les détails d'un contenu,
// discriminée par identifiant TMDB du film.
type contentDetailsCacheKey struct {
	movieID int
}

// similarMoviesCacheKey est la clé de cache pour les films similaires,
// discriminée par identifiant TMDB du film et numéro de page.
type similarMoviesCacheKey struct {
	movieID int
	page    string
}

// movieCreditsCacheKey est la clé de cache pour le casting et l'équipe technique,
// discriminée par identifiant TMDB du film.
type movieCreditsCacheKey struct {
	movieID int
}

// movieImdbIDCacheKey est la clé de cache pour l'identifiant IMDb d'un film ou d'une série,
// discriminée par identifiant TMDB.
type movieImdbIDCacheKey struct {
	movieID int
}

// genreCategoriesCacheKey est la clé de cache pour la liste des catégories de genres,
// discriminée par langue.
type genreCategoriesCacheKey struct {
	language string
}

// ============================================================================
// CACHES
// ============================================================================

// Caches en mémoire pour les réponses de l'API TMDB.
// Chaque cache est typé génériquement avec sa clé et sa valeur,
// et possède un TTL (durée de vie) adapté à la fréquence de mise à jour des données.
var (
	// popularMoviesCache met en cache les films populaires (TTL : 30 min).
	popularMoviesCache = cache.New[popularMoviesCacheKey, []models.MovieDTO](30 * time.Minute)
	// topRatedMoviesCache met en cache les films les mieux notés (TTL : 30 min).
	topRatedMoviesCache = cache.New[topRatedMoviesCacheKey, []models.MovieDTO](30 * time.Minute)
	// trendingMoviesCache met en cache les films tendances (TTL : 15 min).
	trendingMoviesCache = cache.New[trendingMoviesCacheKey, []models.MovieDTO](15 * time.Minute)
	// contentDetailsCache met en cache les détails complets d'un film (TTL : 60 min).
	contentDetailsCache = cache.New[contentDetailsCacheKey, *models.ContentDetailsDTO](60 * time.Minute)
	// similarMoviesCache met en cache les films similaires (TTL : 30 min).
	similarMoviesCache = cache.New[similarMoviesCacheKey, []models.MovieDTO](30 * time.Minute)
	// movieCreditsCache met en cache le casting et l'équipe technique (TTL : 60 min).
	movieCreditsCache = cache.New[movieCreditsCacheKey, *models.MovieCreditsDTO](60 * time.Minute)
	// movieImdbIDCache met en cache la correspondance TMDB ID → IMDb ID (TTL : 60 min).
	movieImdbIDCache = cache.New[movieImdbIDCacheKey, models.TmdbMovieImdbId](60 * time.Minute)
	// genreCategoriesCache met en cache les catégories de genres de films (TTL : 24 h).
	genreCategoriesCache = cache.New[genreCategoriesCacheKey, []models.CategoryDTO](24 * time.Hour)
)

// ============================================================================
// FAKE DATA (inchangé)
// ============================================================================

// generateMovies stocke la liste de films générés aléatoirement (singleton).
var generateMovies = []models.Movie{}

// generatedContentDetails stocke les détails de contenu générés aléatoirement (singleton).
var generatedContentDetails = models.ContentDetails{}

// RandomMovieList génère et retourne une liste aléatoire de films factices.
// La liste est créée une seule fois (entre 10 et 20 films) puis mise en cache
// dans la variable generateMovies. Les appels suivants retournent la même liste.
// Chaque film possède un identifiant, un titre, une année, une note, des genres
// et une image placeholder.
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

// GetContentDetailsRandomized génère et retourne des détails de contenu factices.
// Les détails sont créés une seule fois puis mis en cache dans la variable
// generatedContentDetails. Les appels suivants retournent le même objet.
// L'objet contient un casting aléatoire (3 à 6 acteurs), un titre, une image,
// des genres, une durée, un synopsis et d'autres métadonnées fictives.
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

// GetPopularMovies récupère la liste des films populaires depuis l'API TMDB.
//
// Paramètres :
//   - tmdbBearerToken : jeton d'authentification Bearer pour l'API TMDB (obligatoire).
//   - imageBase : URL de base pour les affiches (par défaut "https://image.tmdb.org/t/p/w500").
//   - genreMap : correspondance entre identifiants de genre TMDB et noms lisibles.
//   - page : numéro de page de résultats (par défaut "1").
//
// Retourne une tranche de [models.MovieDTO] et une erreur éventuelle.
//
// Mise en cache : les résultats sont mis en cache par page et langue (TTL : 30 min)
// via popularMoviesCache.
//
// Endpoint TMDB appelé : GET /3/movie/popular
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

// GetTopRatedMovies récupère la liste des films les mieux notés depuis l'API TMDB.
//
// Paramètres :
//   - bearerToken : jeton d'authentification Bearer pour l'API TMDB.
//   - page : numéro de page de résultats.
//
// Retourne une tranche de [models.MovieDTO] et une erreur éventuelle.
// Les identifiants de genre sont résolus via [models.MovieGenreMap].
// La langue est fixée à "fr-FR".
//
// Mise en cache : les résultats sont mis en cache par page (TTL : 30 min)
// via topRatedMoviesCache.
//
// Endpoint TMDB appelé : GET /3/movie/top_rated
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

// GetTrendingMovies récupère la liste des films tendances depuis l'API TMDB.
//
// Paramètres :
//   - bearerToken : jeton d'authentification Bearer pour l'API TMDB.
//   - timeWindow : fenêtre temporelle, "day" ou "week" (par défaut "day").
//   - page : numéro de page de résultats (par défaut 1 si <= 0).
//   - language : code de langue BCP 47 (par défaut "fr-FR").
//
// Retourne une tranche de [models.MovieDTO] et une erreur éventuelle.
// Les identifiants de genre sont résolus via [models.MovieGenreMap].
//
// Mise en cache : les résultats sont mis en cache par fenêtre temporelle, page et langue
// (TTL : 15 min) via trendingMoviesCache.
//
// Endpoint TMDB appelé : GET /3/trending/movie/{timeWindow}
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

// GetContentDetails récupère les informations détaillées d'un film depuis l'API TMDB.
//
// Paramètres :
//   - tmdbBearerToken : jeton d'authentification Bearer pour l'API TMDB (obligatoire).
//   - imageBase : URL de base pour les affiches (par défaut "https://image.tmdb.org/t/p/w500").
//   - movieID : identifiant TMDB du film.
//
// Retourne un pointeur vers [models.ContentDetailsDTO] contenant le titre, l'image,
// l'identifiant IMDb, le backdrop, l'année, les genres, la note, la durée formatée
// (ex. "2h 15min"), le synopsis, les langues parlées et le premier producteur.
// Retourne une erreur en cas d'échec.
//
// Mise en cache : les résultats sont mis en cache par movieID (TTL : 60 min)
// via contentDetailsCache.
//
// Endpoint TMDB appelé : GET /3/movie/{movieID}
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

// GetSimilarMovies récupère la liste des films similaires à un film donné depuis l'API TMDB.
//
// Paramètres :
//   - tmdbBearerToken : jeton d'authentification Bearer pour l'API TMDB (obligatoire).
//   - imageBase : URL de base pour les images (par défaut "https://image.tmdb.org/t/p/w500").
//   - genreMap : correspondance entre identifiants de genre TMDB et noms lisibles.
//   - movieID : identifiant TMDB du film de référence.
//   - page : numéro de page de résultats (par défaut "1").
//
// Retourne une tranche de [models.MovieDTO] et une erreur éventuelle.
//
// Mise en cache : les résultats sont mis en cache par movieID et page (TTL : 30 min)
// via similarMoviesCache.
//
// Endpoint TMDB appelé : GET /3/movie/{movieID}/similar
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

// GetMovieCredits récupère le casting et l'équipe technique d'un film depuis l'API TMDB.
//
// Paramètres :
//   - tmdbToken : jeton d'authentification Bearer pour l'API TMDB.
//   - imageBase : URL de base pour les photos de profil (par défaut "https://image.tmdb.org/t/p/w500").
//   - movieID : identifiant TMDB du film.
//
// Retourne un pointeur vers [models.MovieCreditsDTO] contenant le réalisateur, le producteur,
// le scénariste et jusqu'à 12 acteurs principaux avec leur nom, rôle et photo.
// Retourne une erreur en cas d'échec.
//
// Mise en cache : les résultats sont mis en cache par movieID (TTL : 60 min)
// via movieCreditsCache.
//
// Endpoint TMDB appelé : GET /3/movie/{movieID}/credits
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

// GetMovieImdbID récupère l'identifiant IMDb d'un film à partir de son identifiant TMDB.
//
// Paramètres :
//   - tmdbBearerToken : jeton d'authentification Bearer pour l'API TMDB (obligatoire).
//   - movieID : identifiant TMDB du film.
//
// Retourne un [models.TmdbMovieImdbId] contenant l'identifiant IMDb et une erreur éventuelle.
//
// Mise en cache : le résultat est mis en cache par movieID (TTL : 60 min)
// via movieImdbIDCache.
//
// Endpoint TMDB appelé : GET /3/movie/{movieID}
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

// GetSeriesImdbID récupère l'identifiant IMDb d'une série à partir de son identifiant TMDB.
//
// Paramètres :
//   - tmdbBearerToken : jeton d'authentification Bearer pour l'API TMDB (obligatoire).
//   - movieID : identifiant TMDB de la série.
//
// Retourne un [models.TmdbMovieImdbId] contenant l'identifiant IMDb et une erreur éventuelle.
//
// Mise en cache : le résultat est mis en cache par movieID (TTL : 60 min)
// via movieImdbIDCache (partagé avec GetMovieImdbID).
//
// Endpoint TMDB appelé : GET /3/tv/{movieID}/external_ids
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

// GetMoviesByGenre récupère les films d'un genre donné depuis l'API TMDB,
// triés par note moyenne décroissante avec un minimum de 100 votes.
//
// Paramètres :
//   - bearerToken : jeton d'authentification Bearer pour l'API TMDB.
//   - genreID : identifiant numérique du genre TMDB.
//   - page : numéro de page de résultats.
//   - language : code de langue BCP 47 (ex. "fr-FR").
//
// Retourne une tranche de [models.MovieDTO] et une erreur éventuelle.
//
// Mise en cache : les résultats sont mis en cache par genreID, page et langue
// via [models.MoviesByGenreCache].
//
// Endpoint TMDB appelé : GET /3/discover/movie?with_genres={genreID}&sort_by=vote_average.desc&vote_count.gte=100
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

// GetMovieGenreCategories récupère la liste des genres de films depuis l'API TMDB
// et les transforme en catégories avec description, lien et couleur.
//
// Paramètres :
//   - bearerToken : jeton d'authentification Bearer pour l'API TMDB.
//   - language : code de langue BCP 47 (par défaut "fr-FR" si vide).
//
// Retourne une tranche de [models.CategoryDTO] contenant l'identifiant, le nom,
// la description, le lien href et la couleur de chaque genre.
// Les couleurs sont déterminées par [models.GenreCategoryColor].
//
// Mise en cache : les résultats sont mis en cache par langue (TTL : 24 h)
// via genreCategoriesCache.
//
// Endpoint TMDB appelé : GET /3/genre/movie/list
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

// GetTorrentioMoviesStreams récupère les flux de streaming disponibles pour un film
// via l'API Torrentio.
//
// Paramètres :
//   - imdbID : identifiant IMDb du film (ex. "tt1234567").
//
// Retourne un pointeur vers [models.TorrentioResponse] contenant la liste des flux
// disponibles et une erreur éventuelle.
// Aucune mise en cache n'est effectuée (requête transactionnelle).
//
// API externe appelée : GET https://torrentio.strem.fun/stream/movie/{imdbID}.json
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
// GetTorrentioSeriesStreams récupère les flux de streaming disponibles pour une série
// via l'API Torrentio.
//
// Paramètres :
//   - imdbID : identifiant IMDb de la série (ex. "tt1234567").
//
// Retourne un pointeur vers [models.TorrentioResponse] contenant la liste des flux
// disponibles et une erreur éventuelle.
// Aucune mise en cache n'est effectuée (requête transactionnelle).
//
// API externe appelée : GET https://torrentio.strem.fun/stream/series/{imdbID}.json
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

// AddMagnetRealDebrid ajoute un lien magnet au service Real-Debrid pour téléchargement.
//
// Paramètres :
//   - apiKey : clé d'API Real-Debrid (utilisée comme Bearer token).
//   - infoHash : hash BitTorrent du contenu à ajouter.
//
// Retourne un pointeur vers [models.RdAddMagnetResponse] contenant l'identifiant
// du torrent créé et une erreur éventuelle.
// Le lien magnet est construit automatiquement à partir de l'infoHash.
//
// API externe appelée : POST https://api.real-debrid.com/rest/1.0/torrents/addMagnet
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

// SelectFilesRealDebrid sélectionne tous les fichiers d'un torrent Real-Debrid
// pour lancer le téléchargement.
//
// Paramètres :
//   - apiKey : clé d'API Real-Debrid (utilisée comme Bearer token).
//   - torrentId : identifiant du torrent retourné par AddMagnetRealDebrid.
//
// Retourne une erreur en cas d'échec. Le code de retour attendu est 204 ou 200.
//
// API externe appelée : POST https://api.real-debrid.com/rest/1.0/torrents/selectFiles/{torrentId}
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

// GetRealDebridTorrentInfo récupère les informations et l'état d'avancement d'un torrent
// sur le service Real-Debrid.
//
// Paramètres :
//   - apiKey : clé d'API Real-Debrid (utilisée comme Bearer token).
//   - torrentId : identifiant du torrent à interroger.
//
// Retourne un pointeur vers [models.RdTorrentInfo] contenant le statut et la progression
// du téléchargement, ainsi qu'une erreur éventuelle.
//
// API externe appelée : GET https://api.real-debrid.com/rest/1.0/torrents/info/{torrentId}
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

// UnrestrictRealDebridLink convertit un lien hébergeur restreint en lien de téléchargement
// direct via le service Real-Debrid.
//
// Paramètres :
//   - apiKey : clé d'API Real-Debrid (utilisée comme Bearer token).
//   - rawLink : lien brut à dérestreindre (les backslashes échappés sont nettoyés).
//
// Retourne un pointeur vers [models.RdUnrestrictResponse] contenant le lien
// de téléchargement direct et une erreur éventuelle.
//
// API externe appelée : POST https://api.real-debrid.com/rest/1.0/unrestrict/link
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

// SearchMovies effectue une recherche multi-critères de films via l'API TMDB.
//
// Algorithme de recherche :
//  1. Validation : au moins un paramètre parmi Query, GenresCSV ou YearsCSV doit être fourni.
//  2. Valeurs par défaut : Page=1, Language="fr-FR", SortBy="popularity.desc".
//  3. Itération multi-années : si plusieurs années sont spécifiées (séparées par des virgules),
//     la recherche est exécutée indépendamment pour chaque année puis les résultats sont fusionnés.
//     Si aucune année n'est spécifiée, une seule requête sans filtre d'année est effectuée.
//  4. Mise en cache : pour chaque combinaison (query, année, page, langue), les résultats bruts
//     TMDB sont mis en cache dans [models.SearchCache] avec un TTL défini par [models.CacheTTL].
//     Le cache est protégé par [models.CacheMutex] (RWMutex).
//  5. Filtrage par genre : si GenresCSV est fourni, seuls les films possédant TOUS les genres
//     demandés sont conservés (intersection, via hasAllGenres).
//  6. Filtrage par note : si Rating > 0, les films ayant une note inférieure sont exclus.
//  7. Déduplication : un film déjà présent dans les résultats (identifié par son ID TMDB)
//     n'est pas ajouté une seconde fois, ce qui évite les doublons entre années.
//
// Paramètres : [models.SearchMoviesParams] contenant BearerToken, Query, GenresCSV,
// YearsCSV, Rating, Page, Language et SortBy.
//
// Retourne une tranche de [models.MovieDTO] et une erreur éventuelle.
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

// safeYear extrait les 4 premiers caractères d'une date (l'année) de manière sûre.
// Retourne "0" si la chaîne contient moins de 4 caractères.
func safeYear(date string) string {
	if len(date) >= 4 {
		return date[:4]
	}
	return "0"
}

// mapGenres convertit une tranche d'identifiants de genre TMDB en noms lisibles
// à l'aide de [models.MovieGenreMap].
func mapGenres(ids []int) []string {
	genres := make([]string, 0, len(ids))
	for _, id := range ids {
		if name, ok := models.MovieGenreMap[id]; ok {
			genres = append(genres, name)
		}
	}
	return genres
}

// splitCSV découpe une chaîne de valeurs séparées par des virgules en tranche de chaînes.
// Retourne une tranche vide si la chaîne est vide ou ne contient que des espaces.
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

// parseIntsCSV analyse une chaîne de nombres entiers séparés par des virgules
// et retourne une tranche d'entiers. Les valeurs non numériques sont ignorées.
func parseIntsCSV(s string) []int {
	result := []int{}
	for _, part := range splitCSV(s) {
		if n, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			result = append(result, n)
		}
	}
	return result
}

// parseGenresCSV analyse une chaîne d'identifiants de genre séparés par des virgules
// et retourne un ensemble (map[int]bool) pour un filtrage par intersection rapide.
func parseGenresCSV(s string) map[int]bool {
	result := map[int]bool{}
	for _, part := range splitCSV(s) {
		if id, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			result[id] = true
		}
	}
	return result
}

// toMovieDTO convertit un résultat brut TMDB ([models.TMDBMovieRaw]) en [models.MovieDTO]
// en extrayant l'année depuis la date de sortie, en construisant l'URL de l'affiche
// et en résolvant les identifiants de genre via [models.MovieGenreMap].
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

// buildCacheKey construit une clé de cache textuelle pour la recherche de films
// à partir des paramètres de recherche et de l'année. Le format est
// "q={query}|y={year}|page={page}|lang={language}".
func buildCacheKey(p models.SearchMoviesParams, year int) string {
	return fmt.Sprintf("q=%s|y=%d|page=%d|lang=%s",
		strings.TrimSpace(p.Query), year, p.Page, p.Language)
}

// hasAllGenres vérifie qu'un film possède tous les genres demandés.
// Retourne true si chaque identifiant de genre présent dans wanted est également
// présent dans movieGenres (test d'inclusion / intersection complète).
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

// fetchMoviesPage effectue un appel à l'API TMDB pour récupérer une page de résultats bruts.
// Si HasQuery est vrai, l'endpoint /3/search/movie est utilisé avec le paramètre query.
// Sinon, l'endpoint /3/discover/movie est utilisé avec les filtres genre, tri et année.
// Retourne la tranche de [models.TMDBMovieRaw] contenue dans la réponse et une erreur éventuelle.
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

// genreDescription retourne une description par défaut en français pour un genre donné.
func genreDescription(name string) string {
	return "Découvrez les meilleurs films du genre " + name + "."
}

// genreHref construit le chemin URL relatif vers la page d'un genre à partir de son identifiant.
func genreHref(id int) string {
	return "/genres/" + strconv.Itoa(id)
}
