package handlers

import (
	"StreamflixBackend/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func GetTrendingTV(bearerToken, timeWindow string, page int, language string) ([]models.TVDTO, error) {
	if timeWindow == "" {
		timeWindow = "day"
	}
	if page <= 0 {
		page = 1
	}
	if language == "" {
		language = "fr-FR"
	}

	url := fmt.Sprintf(
		"https://api.themoviedb.org/3/trending/tv/%s?language=%s&page=%d",
		timeWindow, language, page,
	)

	req, err := http.NewRequest("GET", url, nil)
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

	var tmdbResp models.TMDBTrendingTVResponse
	if err := json.NewDecoder(res.Body).Decode(&tmdbResp); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	dtos := make([]models.TVDTO, 0, len(tmdbResp.Results))
	for _, raw := range tmdbResp.Results {
		year := 0
		if len(raw.FirstAirDate) >= 4 {
			fmt.Sscanf(raw.FirstAirDate[:4], "%d", &year)
		}

		genres := make([]string, 0, len(raw.GenreIDs))
		for _, id := range raw.GenreIDs {
			if name, ok := models.TVGenreMap[id]; ok {
				genres = append(genres, name)
			}
		}

		image := ""
		if raw.PosterPath != "" {
			image = "https://image.tmdb.org/t/p/w500" + raw.PosterPath
		}

		dtos = append(dtos, models.TVDTO{
			ID:       raw.ID,
			Name:     raw.Name,
			Image:    image,
			Year:     year,
			Genres:   genres,
			Rating:   raw.VoteAverage,
			Language: raw.OriginalLanguage,
			Country:  raw.OriginCountry,
		})
	}

	return dtos, nil
}

func GetTVByGenre(bearerToken string, genreID int, page int, language string) ([]models.Tvdto, error) {
	url := fmt.Sprintf(
		"https://api.themoviedb.org/3/discover/tv"+
			"?with_genres=%d"+
			"&page=%d"+
			"&language=%s"+
			"&sort_by=vote_average.desc"+
			"&vote_count.gte=100",
		genreID, page, language,
	)

	req, err := http.NewRequest("GET", url, nil)
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

	var raw models.TMDBDiscoverTVResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	dtos := make([]models.Tvdto, 0, len(raw.Results))
	for _, tv := range raw.Results {
		year := 0
		if len(tv.FirstAirDate) >= 4 {
			if y, err := strconv.Atoi(tv.FirstAirDate[:4]); err == nil {
				year = y
			}
		}

		genres := make([]string, 0)
		for _, gid := range tv.GenreIDs {
			if name, ok := models.TvGenreMap[gid]; ok {
				genres = append(genres, name)
			}
		}

		image := ""
		if tv.PosterPath != "" {
			image = "https://image.tmdb.org/t/p/w500" + tv.PosterPath
		}

		dtos = append(dtos, models.Tvdto{
			ID:     tv.ID,
			Name:   tv.Name,
			Image:  image,
			Year:   year,
			Genres: genres,
			Rating: tv.VoteAverage,
		})
	}

	return dtos, nil
}

func GetPopularTVShows(bearerToken string, page int, language string) ([]models.Tvdto, error) {
	url := fmt.Sprintf(
		"https://api.themoviedb.org/3/tv/popular?language=%s&page=%d",
		language, page,
	)

	req, err := http.NewRequest("GET", url, nil)
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

	var tmdbResp models.TMDBTVShowResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmdbResp); err != nil {
		return nil, err
	}

	dtos := make([]models.Tvdto, 0, len(tmdbResp.Results))
	for _, show := range tmdbResp.Results {
		year := 0
		if len(show.FirstAirDate) >= 4 {
			if y, err := strconv.Atoi(show.FirstAirDate[:4]); err == nil {
				year = y
			}
		}

		genres := make([]string, 0)
		for _, gid := range show.GenreIDs {
			if name, ok := models.TvGenreMap[gid]; ok {
				genres = append(genres, name)
			}
		}

		image := ""
		if show.PosterPath != "" {
			image = "https://image.tmdb.org/t/p/w500" + show.PosterPath
		}

		dtos = append(dtos, models.Tvdto{
			ID:     show.ID,
			Name:   show.Name,
			Image:  image,
			Year:   year,
			Genres: genres,
			Rating: show.VoteAverage,
		})
	}

	return dtos, nil
}

func GetTVInfo(bearerToken string, seriesID int, language string) (*models.TVInfoResponse, error) {
	detailsURL := fmt.Sprintf(
		"https://api.themoviedb.org/3/tv/%d?language=%s&append_to_response=credits,similar,content_ratings",
		seriesID, language,
	)

	raw, err := tmdbGet[models.TMDBTVDetailsRaw](bearerToken, detailsURL)
	if err != nil {
		return nil, fmt.Errorf("tv details: %w", err)
	}

	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		seasons  = make([]models.SeasonDTO, 0, len(raw.Seasons))
		fetchErr error
	)

	for _, s := range raw.Seasons {
		if s.SeasonNumber == 0 {
			continue
		}
		wg.Add(1)
		go func(seasonNum, episodeCount int, airDate string) {
			defer wg.Done()

			url := fmt.Sprintf(
				"https://api.themoviedb.org/3/tv/%d/season/%d?language=%s",
				seriesID, seasonNum, language,
			)
			seasonRaw, err := tmdbGet[models.TMDBSeasonDetailsRaw](bearerToken, url)
			if err != nil {
				mu.Lock()
				fetchErr = err
				mu.Unlock()
				return
			}

			episodes := make([]models.EpisodeDTO, 0, len(seasonRaw.Episodes))
			for _, ep := range seasonRaw.Episodes {
				still := ""
				if ep.StillPath != "" {
					still = "https://image.tmdb.org/t/p/w300" + ep.StillPath
				}
				episodes = append(episodes, models.EpisodeDTO{
					EpisodeNumber: ep.EpisodeNumber,
					Name:          ep.Name,
					Overview:      ep.Overview,
					AirDate:       ep.AirDate,
					Runtime:       ep.Runtime,
					Still:         still,
				})
			}

			year := 0
			if len(airDate) >= 4 {
				year, _ = strconv.Atoi(airDate[:4])
			}

			mu.Lock()
			seasons = append(seasons, models.SeasonDTO{
				SeasonNumber: seasonNum,
				EpisodeCount: episodeCount,
				Year:         year,
				Episodes:     episodes,
			})
			mu.Unlock()
		}(s.SeasonNumber, s.EpisodeCount, s.AirDate)
	}
	wg.Wait()

	if fetchErr != nil {
		return nil, fetchErr
	}

	sortSeasons(seasons)

	year := 0
	if len(raw.FirstAirDate) >= 4 {
		year, _ = strconv.Atoi(raw.FirstAirDate[:4])
	}

	image := ""
	if raw.PosterPath != "" {
		image = "https://image.tmdb.org/t/p/w500" + raw.PosterPath
	}
	backdrop := ""
	if raw.BackdropPath != "" {
		backdrop = "https://image.tmdb.org/t/p/original" + raw.BackdropPath
	}

	genres := make([]string, 0, len(raw.Genres))
	for _, g := range raw.Genres {
		genres = append(genres, g.Name)
	}

	createdBy := make([]string, 0)
	for _, c := range raw.CreatedBy {
		createdBy = append(createdBy, c.Name)
	}

	networks := make([]string, 0)
	for _, n := range raw.Networks {
		networks = append(networks, n.Name)
	}

	languages := make([]string, 0)
	for _, l := range raw.SpokenLanguages {
		languages = append(languages, l.EnglishName)
	}

	classification := ""
	for _, r := range raw.ContentRatings.Results {
		if r.ISO31661 == "FR" {
			classification = r.Rating
			break
		}
	}

	runtime := 0
	if len(raw.EpisodeRunTime) > 0 {
		runtime = raw.EpisodeRunTime[0]
	}

	cast := make([]models.CastMemberDTO, 0, 10)
	for i, c := range raw.Credits.Cast {
		if i >= 10 {
			break
		}
		image := ""
		if c.ProfilePath != "" {
			image = "https://image.tmdb.org/t/p/w185" + c.ProfilePath
		}
		cast = append(cast, models.CastMemberDTO{
			ID:    c.ID,
			Name:  c.Name,
			Role:  c.Character,
			Image: image,
		})
	}

	similar := make([]models.SimilarTVDTO, 0, len(raw.Similar.Results))
	for _, s := range raw.Similar.Results {
		img := ""
		if s.BackdropPath != "" {
			img = "https://image.tmdb.org/t/p/w500" + s.BackdropPath
		} else if s.PosterPath != "" {
			img = "https://image.tmdb.org/t/p/w500" + s.PosterPath
		}
		y := 0
		if len(s.FirstAirDate) >= 4 {
			y, _ = strconv.Atoi(s.FirstAirDate[:4])
		}
		sGenres := make([]string, 0)
		for _, gid := range s.GenreIDs {
			if name, ok := models.TvGenreMap[gid]; ok {
				sGenres = append(sGenres, name)
			}
		}
		similar = append(similar, models.SimilarTVDTO{
			ID:     s.ID,
			Title:  s.Name,
			Image:  img,
			Year:   y,
			Genres: sGenres,
			Rating: s.VoteAverage,
		})
	}

	return &models.TVInfoResponse{
		ContentData: models.TVDetailsDTO{
			ID:             raw.ID,
			Name:           raw.Name,
			Image:          image,
			BackdropImage:  backdrop,
			Year:           year,
			Rating:         raw.VoteAverage,
			EpisodeRuntime: runtime,
			Genres:         genres,
			Synopsis:       raw.Overview,
			CreatedBy:      strings.Join(createdBy, ", "),
			Networks:       strings.Join(networks, ", "),
			Languages:      strings.Join(languages, ", "),
			Classification: classification,
			Cast:           cast,
		},
		Seasons:      seasons,
		Credits:      cast,
		SimilarItems: similar,
	}, nil
}

func tmdbGet[T any](bearerToken, url string) (*T, error) {
	req, err := http.NewRequest("GET", url, nil)
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

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func sortSeasons(seasons []models.SeasonDTO) {
	for i := 1; i < len(seasons); i++ {
		for j := i; j > 0 && seasons[j].SeasonNumber < seasons[j-1].SeasonNumber; j-- {
			seasons[j], seasons[j-1] = seasons[j-1], seasons[j]
		}
	}
}

func SearchTV(bearerToken, query, language string, page int) ([]models.TVSearchDTO, error) {
	if language == "" {
		language = "fr-FR"
	}
	if page <= 0 {
		page = 1
	}

	url := fmt.Sprintf(
		"https://api.themoviedb.org/3/search/tv?query=%s&language=%s&page=%d",
		query, language, page,
	)

	req, err := http.NewRequest("GET", url, nil)
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

	var tmdbResp models.TVSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&tmdbResp); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	if tmdbResp.Results == nil {
		return []models.TVSearchDTO{}, nil
	}

	return mapTVSearchToDTO(tmdbResp.Results), nil
}

func mapTVSearchToDTO(results []models.TVSearchResult) []models.TVSearchDTO {
	dtos := make([]models.TVSearchDTO, 0, len(results))
	for _, r := range results {
		year := 0
		if len(r.FirstAirDate) >= 4 {
			year, _ = strconv.Atoi(r.FirstAirDate[:4])
		}
		dtos = append(dtos, models.TVSearchDTO{
			ID:     r.ID,
			Title:  r.Name,
			Image:  "https://image.tmdb.org/t/p/w500" + r.PosterPath,
			Year:   year,
			Genre:  MapGenreIDs(r.GenreIDs, models.TvGenreMap),
			Rating: r.VoteAverage,
		})
	}
	return dtos
}

func MapGenreIDs(genreIDs []int, genreMap map[int]string) []string {
	genres := make([]string, 0, len(genreIDs))
	for _, id := range genreIDs {
		if name, ok := genreMap[id]; ok {
			genres = append(genres, name)
		}
	}
	return genres
}
