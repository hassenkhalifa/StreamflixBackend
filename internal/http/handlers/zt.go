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

// ZtParserService encapsule models.ZtParser pour y ajouter des méthodes métier
// dans le package handlers.
//
// Ce service gère le scraping du site Zone Téléchargement, la recherche de contenus,
// l'extraction des liens de téléchargement et l'enrichissement des données via
// les API externes TMDB et TVMaze. Il implémente un mécanisme de rate-limiting
// interne pour respecter les limites du site cible.
type ZtParserService struct {
	*models.ZtParser
}

// NewZtParser crée et initialise une nouvelle instance de ZtParserService.
//
// Le parser est configuré avec les catégories par défaut ("films", "series"),
// un mécanisme de rate-limiting basé sur requestTimeInBetween, et un jeton
// TMDB pour l'enrichissement des métadonnées.
//
// Paramètres :
//   - devMode : active le mode développement (logs détaillés sur stdout).
//   - requestTimeInBetween : délai minimum entre deux requêtes HTTP vers le site cible.
//   - moviesDbToken : clé API TMDB pour l'authentification aux services de métadonnées.
//
// Retourne un pointeur vers le ZtParserService initialisé.
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

// GetBaseUrl retourne l'URL de base du site Zone Téléchargement.
//
// L'accès est protégé par un mutex pour garantir la sécurité en contexte concurrent.
func (p *ZtParserService) GetBaseUrl() string {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	return p.BaseUrl
}

// GetAllCategories retourne la liste de toutes les catégories de contenu supportées.
//
// Les catégories par défaut sont "films" et "series".
func (p *ZtParserService) GetAllCategories() []string {
	return p.AllCategories
}

// SetRequestTimeInBetween définit le délai minimum entre deux requêtes HTTP successives.
//
// Ce mécanisme de rate-limiting protège contre un blocage par le site cible.
// La valeur doit être positive ; une valeur négative provoque une erreur.
// L'accès est protégé par un mutex pour garantir la sécurité en contexte concurrent.
func (p *ZtParserService) SetRequestTimeInBetween(value time.Duration) error {
	if value < 0 {
		return fmt.Errorf("value must be positive")
	}
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.RequestTimeInBetween = value
	return nil
}

// SetDevMode active ou désactive le mode développement.
//
// En mode développement, le service affiche des logs détaillés sur stdout
// (rate-limiting, chargement DOM, erreurs). L'accès est thread-safe.
func (p *ZtParserService) SetDevMode(value bool) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.DevMode = value
}

// SetMoviesDbToken définit la clé API TMDB utilisée pour l'enrichissement des métadonnées.
//
// L'accès est protégé par un mutex pour garantir la sécurité en contexte concurrent.
func (p *ZtParserService) SetMoviesDbToken(value string) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.MoviesDbKey = value
}

// UseBaseUrl définit l'URL de base du site Zone Téléchargement.
//
// Cette URL est utilisée comme préfixe pour toutes les requêtes de scraping.
// En mode développement, l'URL définie est affichée sur stdout.
// Retourne toujours true pour indiquer le succès de l'opération.
func (p *ZtParserService) UseBaseUrl(urlStr string) bool {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.BaseUrl = urlStr
	if p.DevMode {
		fmt.Printf("🔗 Base URL définie: %s\n", urlStr)
	}
	return true
}

// GetPayloadUrlFromQuery construit l'URL de recherche pour le site Zone Téléchargement.
//
// L'URL générée suit le format : {baseUrl}/?p={category}&search={query}&page={page}.
// La catégorie est normalisée en minuscules et validée contre la liste des catégories autorisées.
//
// Paramètres :
//   - category : catégorie de contenu ("films" ou "series").
//   - query : terme de recherche (sera encodé URL).
//   - page : numéro de page (doit être >= 1).
//
// Retourne l'URL construite ou une erreur si la catégorie est invalide ou la page < 1.
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

// GetDomElementFromUrl récupère le contenu HTML d'une URL et le parse en document DOM.
//
// Cette méthode implémente un mécanisme de rate-limiting : si le délai minimum
// entre deux requêtes (RequestTimeInBetween) n'est pas écoulé, elle attend
// automatiquement avant d'effectuer la requête. Le timeout HTTP est de 15 secondes.
//
// Le document retourné est un objet goquery.Document permettant la navigation
// et l'extraction de données du DOM avec des sélecteurs CSS.
//
// Paramètres :
//   - urlStr : URL complète de la page à charger.
//
// Retourne le document DOM parsé ou une erreur en cas d'échec HTTP ou de parsing.
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

// ParseMoviesFromSearchQuery effectue le scraping des résultats de recherche
// depuis le site Zone Téléchargement.
//
// Elle construit l'URL de recherche, charge le DOM, puis extrait pour chaque
// élément de résultat : le titre, l'URL, l'identifiant, l'image, la qualité,
// la langue et la date de publication. Les sélecteurs CSS utilisés ciblent
// la structure HTML du site (#dle-content .cover_global).
//
// Paramètres :
//   - category : catégorie de contenu ("films" ou "series").
//   - query : terme de recherche.
//   - page : numéro de page de résultats (>= 1).
//
// Retourne une slice de SearchResult ou une slice vide si aucun résultat n'est trouvé.
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

// Search effectue une recherche de contenu sur une page spécifique.
//
// C'est un wrapper autour de ParseMoviesFromSearchQuery qui gère les erreurs
// en retournant un ErrorResponse en cas d'échec. En mode développement,
// les erreurs sont également affichées sur stdout.
//
// Paramètres :
//   - category : catégorie de contenu ("films" ou "series").
//   - query : terme de recherche.
//   - page : numéro de page de résultats (>= 1).
//
// Retourne les résultats de recherche ([]SearchResult) ou un ErrorResponse en cas d'erreur.
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

// SearchAll effectue une recherche exhaustive sur toutes les pages disponibles.
//
// Elle itère page par page en appelant ParseMoviesFromSearchQuery jusqu'à ce
// qu'une page retourne zéro résultat, puis agrège tous les résultats dans
// une seule slice. Chaque page chargée respecte le rate-limiting configuré.
//
// Attention : cette méthode peut être lente si le nombre de pages est élevé,
// car chaque page nécessite une requête HTTP distincte avec un délai de rate-limiting.
//
// Paramètres :
//   - category : catégorie de contenu ("films" ou "series").
//   - query : terme de recherche.
//
// Retourne la liste complète des résultats ou un ErrorResponse en cas d'erreur.
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

// GetMovieNameFromId récupère les informations de base d'un film ou d'une série
// depuis le site Zone Téléchargement à partir de son identifiant.
//
// Elle charge la page de détail du contenu, puis extrait via des expressions
// régulières : le titre, le titre original, la qualité, la langue, la taille
// du fichier et les liens de téléchargement par hébergeur.
//
// Le comportement diffère selon la catégorie :
//   - "films" (converti en "film") : extrait qualité, langue, taille et liens directs.
//   - "series" (converti en "serie") : extrait la taille par épisode et les liens
//     organisés par épisode et par hébergeur.
//
// Paramètres :
//   - category : catégorie de contenu ("films" ou "series").
//   - id : identifiant numérique du contenu sur Zone Téléchargement.
//
// Retourne un ZtBasicInfo contenant les métadonnées et liens, ou une erreur.
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

// TheMovieDbAuthentication vérifie que la clé API TMDB configurée est valide.
//
// Elle effectue une requête vers l'endpoint /authentication/token/new de TMDB.
// En mode développement, le corps de la réponse est affiché sur stdout.
//
// Retourne true si l'authentification réussit (HTTP 200), false sinon.
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

// GetMovieDatas récupère les détails complets d'un film en combinant les données
// du site Zone Téléchargement et de l'API TMDB.
//
// Le processus est le suivant :
//  1. Vérification de l'authentification TMDB.
//  2. Récupération des informations de base depuis ZT (titre, liens).
//  3. Recherche du film sur TMDB par titre pour obtenir les métadonnées enrichies.
//  4. Récupération de la liste des genres et des crédits (réalisateurs, acteurs).
//  5. Assemblage de toutes les données dans un MovieDetails.
//
// Paramètres :
//   - category : catégorie de contenu ("films").
//   - id : identifiant numérique du contenu sur Zone Téléchargement.
//
// Retourne un MovieDetails complet ou une erreur si l'authentification, le scraping
// ou un appel API échoue.
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

// GetSeriesData récupère les détails complets d'une série en combinant les données
// du site Zone Téléchargement et de l'API TVMaze.
//
// Le processus est le suivant :
//  1. Récupération des informations de base depuis ZT (titre original, liens).
//  2. Recherche de la série sur TVMaze par titre original.
//  3. Récupération des détails de la série, du casting et de l'équipe technique.
//  4. Extraction des producteurs exécutifs comme "réalisateurs".
//  5. Nettoyage du résumé HTML (suppression des balises <p>).
//  6. Assemblage de toutes les données dans un SeriesDetails.
//
// Paramètres :
//   - category : catégorie de contenu ("series").
//   - id : identifiant numérique du contenu sur Zone Téléchargement.
//
// Retourne un SeriesDetails complet ou une erreur si le scraping ou un appel API échoue.
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

// GetQueryDatas est le point d'entrée unifié pour récupérer les détails d'un contenu.
//
// Elle délègue vers GetSeriesData si la catégorie est "series", ou vers
// GetMovieDatas pour toute autre catégorie (typiquement "films").
//
// Paramètres :
//   - category : catégorie de contenu ("films" ou "series").
//   - id : identifiant numérique du contenu sur Zone Téléchargement.
//
// Retourne un MovieDetails ou SeriesDetails selon la catégorie, ou une erreur.
func (p *ZtParserService) GetQueryDatas(category, id string) (interface{}, error) {
	if category == "series" {
		return p.GetSeriesData(category, id)
	}
	return p.GetMovieDatas(category, id)
}
