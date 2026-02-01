package handlers

import (
	"StreamflixBackend/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ZtParserService wraps models.ZtParser to add methods in this package
type ZtParserService struct {
	*models.ZtParser
}

// NewZtParser crée une nouvelle instance du parser
func NewZtParser(devMode bool, requestTimeInBetween time.Duration, moviesDbToken string) *ZtParserService {
	parser := &models.ZtParser{
		BaseUrl:              "",
		AllCategories:        []string{"films", "series"},
		LastRequestTimestamp: time.Time{},
		RequestTimeInBetween: requestTimeInBetween,
		DevMode:              devMode,
		MoviesDbKey:          moviesDbToken,
	}

	if devMode {
		fmt.Println("ztParser: Dev mode enabled.")
	}

	return &ZtParserService{ZtParser: parser}
}

// GetBaseUrl retourne l'URL de base
func (p *ZtParserService) GetBaseUrl() string {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	return p.BaseUrl
}

// GetAllCategories retourne toutes les catégories
func (p *ZtParserService) GetAllCategories() []string {
	return p.AllCategories
}

// SetRequestTimeInBetween définit le délai entre les requêtes
func (p *ZtParserService) SetRequestTimeInBetween(value time.Duration) error {
	if value < 0 {
		return fmt.Errorf("value must be positive")
	}
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.RequestTimeInBetween = value
	return nil
}

// SetDevMode active/désactive le mode dev
func (p *ZtParserService) SetDevMode(value bool) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.DevMode = value
}

// SetMoviesDbToken définit le token TMDB
func (p *ZtParserService) SetMoviesDbToken(value string) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.MoviesDbKey = value
}

// UseBaseUrl définit l'URL de base
func (p *ZtParserService) UseBaseUrl(urlStr string) bool {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.BaseUrl = urlStr
	if p.DevMode {
		fmt.Printf("🔗 Base URL définie: %s\n", urlStr)
	}
	return true
}

// GetPayloadUrlFromQuery construit l'URL de recherche
func (p *ZtParserService) GetPayloadUrlFromQuery(category, query string, page int) (string, error) {
	if page < 1 {
		return "", fmt.Errorf("page must be >= 1")
	}

	category = strings.TrimSpace(strings.ToLower(category))

	validCategory := false
	for _, cat := range p.AllCategories {
		if cat == category {
			validCategory = true
			break
		}
	}

	if !validCategory {
		return "", fmt.Errorf("category must be one of: %s", strings.Join(p.AllCategories, ", "))
	}

	baseUrl := p.GetBaseUrl()
	return fmt.Sprintf("%s/?p=%s&search=%s&page=%d",
		baseUrl,
		category,
		url.QueryEscape(query),
		page,
	), nil
}

// GetDomElementFromUrl récupère et parse le DOM d'une URL
func (p *ZtParserService) GetDomElementFromUrl(urlStr string) (*goquery.Document, error) {
	p.Mu.Lock()

	// Calcul du temps écoulé depuis la dernière requête
	elapsed := time.Since(p.LastRequestTimestamp)

	// Si pas assez de temps écoulé, on attend
	if elapsed < p.RequestTimeInBetween {
		sleepDuration := p.RequestTimeInBetween - elapsed
		if p.DevMode {
			fmt.Printf("⏳ Rate limit: attente de %v\n", sleepDuration)
		}
		p.Mu.Unlock()
		time.Sleep(sleepDuration)
		p.Mu.Lock()
	}

	// Mise à jour du timestamp
	p.LastRequestTimestamp = time.Now()
	p.Mu.Unlock()

	// Requête HTTP avec timeout
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("erreur requête HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("code HTTP: %d", resp.StatusCode)
	}

	// Parse HTML avec goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur parse HTML: %w", err)
	}

	if p.DevMode {
		fmt.Printf("✓ DOM chargé depuis: %s\n", urlStr)
	}

	return doc, nil
}

// ParseMoviesFromSearchQuery parse les résultats de recherche
func (p *ZtParserService) ParseMoviesFromSearchQuery(category, query string, page int) ([]models.SearchResult, error) {
	payloadUrl, err := p.GetPayloadUrlFromQuery(category, query, page)
	if err != nil {
		return nil, err
	}

	doc, err := p.GetDomElementFromUrl(payloadUrl)
	if err != nil {
		return nil, err
	}

	var responseMovieList []models.SearchResult

	movieListElements := doc.Find("#dle-content .cover_global")
	if movieListElements.Length() == 0 {
		return responseMovieList, nil
	}

	movieListElements.Each(func(index int, element *goquery.Selection) {
		titleAnchor := element.Find(".cover_infos_title a")
		title := titleAnchor.Text()
		href, exists := titleAnchor.Attr("href")

		if !exists || title == "" {
			return
		}

		theUrl := p.GetBaseUrl() + href

		// Extraction de l'ID depuis l'URL
		idRegex := regexp.MustCompile(`[?&]id=([0-9]{1,5})\-`)
		idMatches := idRegex.FindStringSubmatch(theUrl)
		var id string
		if len(idMatches) > 1 {
			id = idMatches[1]
		}

		imgSrc, _ := element.Find("img").Attr("src")
		detailRelease := element.Find(".cover_infos_global .detail_release")
		quality := detailRelease.Find("span").Eq(0).Text()
		language := detailRelease.Find("span").Eq(1).Text()

		timeText := element.Find("time").Text()
		publishDate, _ := time.Parse("2006-01-02", timeText)

		movieData := models.SearchResult{
			Title:              title,
			Url:                theUrl,
			Id:                 id,
			Image:              p.GetBaseUrl() + imgSrc,
			Quality:            quality,
			Language:           language,
			PublishedOn:        publishDate,
			PublishedTimestamp: publishDate.Unix(),
		}

		responseMovieList = append(responseMovieList, movieData)
	})

	return responseMovieList, nil
}

// Search effectue une recherche sur une page spécifique
func (p *ZtParserService) Search(category, query string, page int) (interface{}, error) {
	results, err := p.ParseMoviesFromSearchQuery(category, query, page)
	if err != nil {
		if p.DevMode {
			fmt.Println(err)
		}
		return models.ErrorResponse{
			Status: false,
			Error:  err.Error(),
		}, err
	}
	return results, nil
}

// SearchAll effectue une recherche sur toutes les pages
func (p *ZtParserService) SearchAll(category, query string) (interface{}, error) {
	var responseMovieList []models.SearchResult
	searchPage := 0

	for {
		searchPage++
		tempMovieList, err := p.ParseMoviesFromSearchQuery(category, query, searchPage)
		if err != nil {
			if p.DevMode {
				fmt.Println(err)
			}
			return models.ErrorResponse{
				Status: false,
				Error:  err.Error(),
			}, err
		}

		if len(tempMovieList) == 0 {
			break
		}

		responseMovieList = append(responseMovieList, tempMovieList...)
		fmt.Printf("Added %d movies from page %d\n", len(tempMovieList), searchPage)
	}

	return responseMovieList, nil
}

// GetMovieNameFromId récupère les informations de base depuis ZT
func (p *ZtParserService) GetMovieNameFromId(category, id string) (*models.ZtBasicInfo, error) {
	// Conversion de la catégorie
	if category == "films" {
		category = "film"
	} else if category == "series" {
		category = "serie"
	}

	fmt.Printf("categories: %s, id: %s\n", category, id)

	urlStr := fmt.Sprintf("%s/?p=%s&id=%s", p.GetBaseUrl(), category, id)
	doc, err := p.GetDomElementFromUrl(urlStr)
	if err != nil {
		return nil, err
	}

	mainHtml := doc.Find("#dle-content")
	htmlString, _ := mainHtml.Html()

	name := mainHtml.Find("h1").Eq(0).Text()

	// Extraction avec regex
	originalNameRegex := regexp.MustCompile(`<strong><u>Titre original</u> :</strong>\s*([^<]+)<br>`)
	originalNameMatch := originalNameRegex.FindStringSubmatch(htmlString)
	originalName := ""
	if len(originalNameMatch) > 1 {
		originalName = strings.TrimSpace(originalNameMatch[1])
	}

	info := &models.ZtBasicInfo{
		Name:         name,
		OriginalName: originalName,
	}

	if category == "film" {
		// Extraction pour les films
		qualityRegex := regexp.MustCompile(`<strong><u>Qualité</u> :</strong>\s*([^<]+)<br>`)
		qualityMatch := qualityRegex.FindStringSubmatch(htmlString)
		if len(qualityMatch) > 1 {
			info.Quality = strings.TrimSpace(qualityMatch[1])
		}

		languageRegex := regexp.MustCompile(`<strong><u>Langue</u> :</strong>\s*([^<]+)<br>`)
		languageMatch := languageRegex.FindStringSubmatch(htmlString)
		if len(languageMatch) > 1 {
			info.Language = strings.TrimSpace(languageMatch[1])
		}

		sizeRegex := regexp.MustCompile(`<strong><u>Taille du fichier</u> :</strong>\s*([^<]+)<br>`)
		sizeMatch := sizeRegex.FindStringSubmatch(htmlString)
		if len(sizeMatch) > 1 {
			info.Size = strings.TrimSpace(sizeMatch[1])
		}

		// Extraction des liens de téléchargement
		downloadLinks := make(map[string]string)
		postinfo := doc.Find("#news-id-23557 .postinfo")
		postinfo.Find("b > div").Each(func(index int, element *goquery.Selection) {
			hostName := element.Text()
			if hostName != "" {
				downloadUrl, exists := element.Parent().Next().Find("a").Attr("href")
				if exists {
					downloadLinks[hostName] = downloadUrl
				}
			}
		})
		info.Links = downloadLinks

	} else if category == "serie" {
		// Extraction pour les séries
		sizeRegex := regexp.MustCompile(`<strong><u>Taille d'un episode</u> :</strong>\s*([^<]+)<br>`)
		sizeMatch := sizeRegex.FindStringSubmatch(htmlString)
		if len(sizeMatch) > 1 {
			info.Size = strings.TrimSpace(sizeMatch[1])
		}

		// Extraction des liens par épisode
		downloadLinks := make(map[string]map[string]string)
		postinfo := doc.Find("#news-id-23557 .postinfo")
		postinfo.Find("b > div").Each(func(index int, element *goquery.Selection) {
			hostName := element.Text()
			if hostName != "" {
				element.Parent().NextAll().FilterFunction(func(i int, s *goquery.Selection) bool {
					return s.Is("b")
				}).Each(func(idx int, episodeElem *goquery.Selection) {
					episode := episodeElem.Find("a").Text()
					if episode != "" {
						if downloadLinks[episode] == nil {
							downloadLinks[episode] = make(map[string]string)
						}
						downloadUrl, exists := episodeElem.Find("a").Attr("href")
						if exists {
							downloadLinks[episode][hostName] = downloadUrl
						}
					}
				})
			}
		})
		info.Links = downloadLinks
	}

	return info, nil
}

// TheMovieDbAuthentication vérifie l'authentification TMDB
func (p *ZtParserService) TheMovieDbAuthentication() bool {
	urlStr := fmt.Sprintf("https://api.themoviedb.org/3/authentication/token/new?api_key=%s", p.MoviesDbKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(urlStr)
	if err != nil {
		if p.DevMode {
			fmt.Println("Erreur:", err)
		}
		return false
	}
	defer resp.Body.Close()

	if p.DevMode {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	}

	return resp.StatusCode == http.StatusOK
}

// GetMovieDatas récupère les détails complets d'un film
func (p *ZtParserService) GetMovieDatas(category, id string) (*models.MovieDetails, error) {
	if !p.TheMovieDbAuthentication() {
		return nil, fmt.Errorf("error while trying to authenticate to TheMovieDB API")
	}

	ztData, err := p.GetMovieNameFromId(category, id)
	if err != nil {
		return nil, err
	}

	// Recherche du film sur TMDB
	searchUrl := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s&query=%s",
		p.MoviesDbKey,
		url.QueryEscape(ztData.Name),
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(searchUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchResponse models.TmdbMovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, err
	}

	if len(searchResponse.Results) == 0 {
		return nil, fmt.Errorf("no movie details found in the API response")
	}

	movieDetails := searchResponse.Results[0]

	// Récupération de la liste des genres
	genresUrl := fmt.Sprintf("https://api.themoviedb.org/3/genre/movie/list?api_key=%s", p.MoviesDbKey)
	genresResp, err := client.Get(genresUrl)
	if err != nil {
		return nil, err
	}
	defer genresResp.Body.Close()

	var genresResponse models.TmdbGenresResponse
	if err := json.NewDecoder(genresResp.Body).Decode(&genresResponse); err != nil {
		return nil, err
	}

	// Mapping des genres
	var genreNames []string
	for _, genreId := range movieDetails.GenreIds {
		for _, genre := range genresResponse.Genres {
			if genre.Id == genreId {
				genreNames = append(genreNames, genre.Name)
				break
			}
		}
	}

	// Récupération des crédits
	creditsUrl := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/credits?api_key=%s",
		movieDetails.Id,
		p.MoviesDbKey,
	)
	creditsResp, err := client.Get(creditsUrl)
	if err != nil {
		return nil, err
	}
	defer creditsResp.Body.Close()

	var creditsResponse models.TmdbCreditsResponse
	if err := json.NewDecoder(creditsResp.Body).Decode(&creditsResponse); err != nil {
		return nil, err
	}

	// Extraction des réalisateurs
	var directors []string
	for _, crew := range creditsResponse.Crew {
		if crew.Job == "Director" {
			directors = append(directors, crew.Name)
		}
	}

	// Extraction des acteurs
	var actors []string
	for _, cast := range creditsResponse.Cast {
		actors = append(actors, cast.Name)
	}

	// Conversion des liens
	downloadLinks, _ := ztData.Links.(map[string]string)

	data := &models.MovieDetails{
		Title:        movieDetails.OriginalTitle,
		OriginalName: ztData.OriginalName,
		Language:     ztData.Language,
		Quality:      ztData.Quality,
		Size:         ztData.Size,
		Description:  movieDetails.Overview,
		Poster:       "https://image.tmdb.org/t/p/w780" + movieDetails.PosterPath,
		ReleaseDate:  movieDetails.ReleaseDate,
		Genres:       genreNames,
		VoteAverage:  movieDetails.VoteAverage,
		Directors:    directors,
		Actors:       actors,
		DownloadLink: downloadLinks,
	}

	return data, nil
}

// GetSeriesData récupère les détails complets d'une série
func (p *ZtParserService) GetSeriesData(category, id string) (*models.SeriesDetails, error) {
	ztData, err := p.GetMovieNameFromId(category, id)
	if err != nil {
		return nil, err
	}

	// Recherche sur TVMaze
	searchUrl := fmt.Sprintf("https://api.tvmaze.com/search/shows?q=%s",
		url.QueryEscape(ztData.OriginalName),
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(searchUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchResponse []models.TvmazeSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, err
	}

	if len(searchResponse) == 0 {
		return nil, fmt.Errorf("no results found for the given query")
	}

	show := searchResponse[0].Show

	// Récupération des détails de la série
	showUrl := fmt.Sprintf("https://api.tvmaze.com/shows/%d", show.Id)
	showResp, err := client.Get(showUrl)
	if err != nil {
		return nil, err
	}
	defer showResp.Body.Close()

	var showDetails models.TvmazeShow
	if err := json.NewDecoder(showResp.Body).Decode(&showDetails); err != nil {
		return nil, err
	}

	// Récupération du cast
	castUrl := fmt.Sprintf("https://api.tvmaze.com/shows/%d/cast", show.Id)
	castResp, err := client.Get(castUrl)
	if err != nil {
		return nil, err
	}
	defer castResp.Body.Close()

	var cast []models.TvmazeCastMember
	if err := json.NewDecoder(castResp.Body).Decode(&cast); err != nil {
		return nil, err
	}

	var actors []string
	for _, member := range cast {
		actors = append(actors, member.Person.Name)
	}

	// Récupération de l'équipe
	crewUrl := fmt.Sprintf("https://api.tvmaze.com/shows/%d/crew", show.Id)
	crewResp, err := client.Get(crewUrl)
	if err != nil {
		return nil, err
	}
	defer crewResp.Body.Close()

	var crew []models.TvmazeCrewMember
	if err := json.NewDecoder(crewResp.Body).Decode(&crew); err != nil {
		return nil, err
	}

	var directors []string
	for _, member := range crew {
		if member.Type == "Executive Producer" || member.Type == "Co-Executive Producer" {
			directors = append(directors, member.Person.Name)
		}
	}

	// Nettoyage de la description
	description := strings.ReplaceAll(showDetails.Summary, "<​p>", "")
	description = strings.ReplaceAll(description, "<​/p>", "")

	// Conversion des liens
	downloadLinks, _ := ztData.Links.(map[string]map[string]string)

	data := &models.SeriesDetails{
		Title:        showDetails.Name,
		OriginalName: ztData.OriginalName,
		Language:     ztData.Language,
		Quality:      ztData.Quality,
		Size:         ztData.Size,
		Description:  description,
		Poster:       showDetails.Image.Original,
		ReleaseDate:  showDetails.Premiered,
		Genres:       showDetails.Genres,
		VoteAverage:  showDetails.Rating.Average,
		Directors:    directors,
		Actors:       actors,
		DownloadLink: downloadLinks,
	}

	return data, nil
}

// GetQueryDatas récupère les détails selon le type (film ou série)
func (p *ZtParserService) GetQueryDatas(category, id string) (interface{}, error) {
	if category == "series" {
		return p.GetSeriesData(category, id)
	}
	return p.GetMovieDatas(category, id)
}
