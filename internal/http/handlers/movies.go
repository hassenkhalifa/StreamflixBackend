package handlers

import (
	"StreamflixBackend/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	_ "io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	_ "github.com/gin-gonic/gin"
)

var generateMovies = []models.Movie{}
var generatedContentDetails = models.ContentDetails{}

func RandomMovieList() []models.Movie {
	// Si déjà généré, retourner la liste existante
	if len(generateMovies) > 0 {
		return generateMovies
	}

	// Sinon, générer
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

	// Générer le cast
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

func GetMoviesByGenre(genre string) []models.Movie {
	var movies []models.Movie
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

func GetMoviesByID(movieID int) models.Movie {
	return generateMovies[movieID]

}

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

	// Validation de la page
	if _, err := strconv.Atoi(page); err != nil {
		return nil, fmt.Errorf("invalid page")
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

	return out, nil
}

func GetContentDetails(tmdbBearerToken string, imageBase string, movieID int) (*models.ContentDetailsDTO, error) {
	if imageBase == "" {
		imageBase = "https://image.tmdb.org/t/p/w500"
	}

	if tmdbBearerToken == "" {
		return nil, fmt.Errorf("TMDB bearer token missing")
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

	// Extraction de l'année
	year := 0
	if tmdbRes.ReleaseDate != "" {
		if t, err := time.Parse("2006-01-02", tmdbRes.ReleaseDate); err == nil {
			year = t.Year()
		}
	}

	// Construction des URLs d'images
	image := ""
	if tmdbRes.PosterPath != "" {
		image = fmt.Sprintf("%s%s", imageBase, tmdbRes.PosterPath)
	}

	backdropImage := ""
	if tmdbRes.BackdropPath != "" {
		backdropImage = fmt.Sprintf("https://image.tmdb.org/t/p/original%s", tmdbRes.BackdropPath)
	}

	// Extraction des genres
	genres := make([]string, 0, len(tmdbRes.Genres))
	for _, g := range tmdbRes.Genres {
		genres = append(genres, g.Name)
	}

	// Durée formatée
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

	// Langues
	languages := ""
	if len(tmdbRes.SpokenLanguages) > 0 {
		langNames := make([]string, 0, len(tmdbRes.SpokenLanguages))
		for _, lang := range tmdbRes.SpokenLanguages {
			langNames = append(langNames, lang.Name)
		}
		languages = strings.Join(langNames, ", ")
	}

	// Producteur (première compagnie de production)
	producer := ""
	if len(tmdbRes.ProductionCompanies) > 0 {
		producer = tmdbRes.ProductionCompanies[0].Name
	}

	return &models.ContentDetailsDTO{
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
	}, nil
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

	return out, nil
}

func GetMovieCredits(tmdbToken string, imageBase string, movieID int) (*models.MovieCreditsDTO, error) {
	if imageBase == "" {
		imageBase = "https://image.tmdb.org/t/p/w500"
	}

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/credits", movieID)

	req, _ := http.NewRequest("GET", url, nil)
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
		Cast: make([]models.CastMemberDTO, 0),
	}

	// --- Director ---
	for _, c := range tmdb.Crew {
		if c.Job == "Director" {
			out.Director = c.Name
			break
		}
	}

	// --- Producer ---
	for _, c := range tmdb.Crew {
		if c.Job == "Producer" || c.Job == "Executive Producer" {
			out.Producer = c.Name
			break
		}
	}

	// --- Writer ---
	for _, c := range tmdb.Crew {
		if c.Job == "Screenplay" || c.Job == "Writer" || c.Job == "Novel" {
			out.Writer = c.Name
			break
		}
	}

	// --- ACTORS (limit 12) ---
	for i, actor := range tmdb.Cast {
		if i >= 12 {
			break
		}

		image := ""
		if actor.ProfilePath != "" {
			image = imageBase + actor.ProfilePath
		}

		out.Cast = append(out.Cast, models.CastMemberDTO{
			Name:  actor.Name,
			Role:  actor.Character,
			Image: image,
		})
	}

	return out, nil
}

func GetMovieImdbID(tmdbBearerToken string, movieID int) (models.TmdbMovieImdbId, error) {
	log.Printf("   → GetMovieImdbID: Début pour movieID=%d", movieID)

	if tmdbBearerToken == "" {
		log.Println("   ❌ Token TMDB manquant")
		return models.TmdbMovieImdbId{}, fmt.Errorf("TMDB bearer token missing")
	}

	u := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d", movieID)
	log.Printf("   → URL TMDB: %s", u)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		log.Printf("   ❌ Erreur création requête: %v", err)
		return models.TmdbMovieImdbId{}, err
	}
	req.Header.Set("Authorization", "Bearer "+tmdbBearerToken)
	req.Header.Set("Accept", "application/json")

	log.Println("   → Envoi de la requête à TMDB...")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("   ❌ Erreur requête HTTP: %v", err)
		return models.TmdbMovieImdbId{}, err
	}
	defer resp.Body.Close()

	log.Printf("   → Status Code TMDB: %d", resp.StatusCode)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("   ❌ Réponse TMDB: %s", string(body))
		return models.TmdbMovieImdbId{}, fmt.Errorf("tmdb error: status %d", resp.StatusCode)
	}

	var tmdbRes models.TmdbMovieImdbId
	if err := json.NewDecoder(resp.Body).Decode(&tmdbRes); err != nil {
		log.Printf("   ❌ Erreur décodage JSON: %v", err)
		return models.TmdbMovieImdbId{}, fmt.Errorf("failed to decode tmdb response")
	}

	log.Printf("   ✅ IMDB ID trouvé: %s", tmdbRes.ImdbId)
	return tmdbRes, nil
}

func GetMovieStreamsFromImdb(imdbID, torrentioRealDebridKey, realDebridApiKey string) ([]StreamResult, error) {
	if imdbID == "" {
		return nil, fmt.Errorf("imdbID is required")
	}
	if realDebridApiKey == "" {
		return nil, fmt.Errorf("RealDebrid API key is required")
	}

	// 1. Clients
	torrentioClient := NewTorrentioClient(torrentioRealDebridKey)
	realDebridClient := NewRealDebridClient(realDebridApiKey)

	// 2. Service
	streamingService := NewStreamingService(torrentioClient, realDebridClient)

	// 3. Workflow complet
	streams, err := streamingService.GetStreamForMovie(imdbID)
	if err != nil {
		return nil, err
	}

	return streams, nil
}

func GetTorrentioStreams(imdbID string) (*models.TorrentioResponse, error) {
	log.Printf("   → GetTorrentioStreams: Début pour imdbID=%s", imdbID)

	url := fmt.Sprintf("https://torrentio.strem.fun/stream/movie/%s.json", imdbID)
	log.Printf("   → URL Torrentio: %s", url)

	log.Println("   → Envoi de la requête à Torrentio...")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 🔥 Obligatoire sinon Torrentio bloque (403)
	req.Header.Set("User-Agent", "curl/8.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("   ❌ Erreur requête HTTP: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("   → Status Code Torrentio: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("   ❌ Réponse Torrentio : %s", string(body))
		return nil, fmt.Errorf("Torrentio error: %d", resp.StatusCode)
	}

	var result models.TorrentioResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("   ❌ Erreur décodage JSON: %v", err)
		return nil, err
	}

	log.Printf("   ✅ %d streams récupérés", len(result.Streams))
	return &result, nil
}
func AddMagnetRealDebrid(apiKey, infoHash string) (*models.RdAddMagnetResponse, error) {
	log.Printf("   → AddMagnetRealDebrid: InfoHash=%s", infoHash)

	apiURL := "https://api.real-debrid.com/rest/1.0/torrents/addMagnet"
	magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s", infoHash)
	log.Printf("   → Magnet construit: %s", magnet)

	data := url.Values{}
	data.Set("magnet", magnet)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Printf("   ❌ Erreur création requête: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	log.Println("   → Envoi de la requête à Real-Debrid...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("   ❌ Erreur requête HTTP: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("   → Status Code Real-Debrid: %d", resp.StatusCode)

	var result models.RdAddMagnetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("   ❌ Erreur décodage JSON: %v", err)
		return nil, err
	}

	log.Printf("   ✅ Torrent ID: %s", result.Id)
	return &result, nil
}
func SelectFilesRealDebrid(apiKey, torrentId string) error {
	log.Printf("   → SelectFilesRealDebrid: TorrentID=%s", torrentId)

	apiURL := fmt.Sprintf("https://api.real-debrid.com/rest/1.0/torrents/selectFiles/%s", torrentId)
	log.Printf("   → URL: %s", apiURL)

	data := url.Values{}
	data.Set("files", "all")

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Printf("   ❌ Erreur création requête: %v", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	log.Println("   → Envoi de la requête...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("   ❌ Erreur requête HTTP: %v", err)
		return err
	}
	defer resp.Body.Close()

	log.Printf("   → Status Code: %d", resp.StatusCode)

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("   ❌ Erreur Real-Debrid: %s", string(body))
		return fmt.Errorf("select files error: %s", string(body))
	}

	log.Println("   ✅ Fichiers sélectionnés")
	return nil
}

func GetRealDebridTorrentInfo(apiKey, torrentId string) (*models.RdTorrentInfo, error) {
	log.Printf("   → GetRealDebridTorrentInfo: TorrentID=%s", torrentId)

	apiURL := fmt.Sprintf("https://api.real-debrid.com/rest/1.0/torrents/info/%s", torrentId)
	log.Printf("   → URL: %s", apiURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("   ❌ Erreur création requête: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	log.Println("   → Envoi de la requête...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("   ❌ Erreur requête HTTP: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("   → Status Code: %d", resp.StatusCode)

	var info models.RdTorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Printf("   ❌ Erreur décodage JSON: %v", err)
		return nil, err
	}

	log.Printf("   ✅ Infos récupérées: Status=%s, Progress=%.2f%%", info.Status, info.Progress)
	return &info, nil
}

func UnrestrictRealDebridLink(apiKey, rawLink string) (*models.RdUnrestrictResponse, error) {
	log.Printf("   → UnrestrictRealDebridLink: RawLink=%s", rawLink)

	cleanLink := strings.ReplaceAll(rawLink, `\/`, `/`)
	log.Printf("   → Lien nettoyé: %s", cleanLink)

	apiURL := "https://api.real-debrid.com/rest/1.0/unrestrict/link"

	data := url.Values{}
	data.Set("link", cleanLink)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Printf("   ❌ Erreur création requête: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	log.Println("   → Envoi de la requête...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("   ❌ Erreur requête HTTP: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("   → Status Code: %d", resp.StatusCode)

	var info models.RdUnrestrictResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Printf("   ❌ Erreur décodage JSON: %v", err)
		return nil, err
	}

	log.Printf("   ✅ Lien direct: %s", info.Download)
	return &info, nil
}
