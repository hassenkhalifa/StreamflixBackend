package handlers

import (
	"StreamflixBackend/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GetTrendingTV récupère les séries TV tendance (day|week) et retourne une liste de DTO.
func GetTrendingTV(bearerToken, timeWindow string, page int, language string) ([]models.TVDTO, error) {
	if timeWindow == "" {
		timeWindow = "day" // "day" ou "week"
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
