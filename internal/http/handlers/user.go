package handlers

import (
	"StreamflixBackend/internal/models"
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
)

var generatedUserListItem = []models.UserListItem{}

func GetUserListItems() []models.UserListItem {
	if len(generatedUserListItem) > 0 {
		return generatedUserListItem
	}
	var items []models.UserListItem
	categories := []string{"favorites", "history", "watchlist"}
	contentTypes := []string{"movie", "tvshow"}

	for i := 0; i < gofakeit.Number(5, 15); i++ {
		contentID := gofakeit.Number(100, 999)
		contentType := contentTypes[gofakeit.Number(0, 1)]
		percentage := gofakeit.Number(0, 100)

		// Générer durée et temps
		var duration, totalTime, currentTime, subtitle, playHref string

		if contentType == "movie" {
			hours := gofakeit.Number(1, 3)
			minutes := gofakeit.Number(0, 59)
			duration = fmt.Sprintf("%dh %02dmin", hours, minutes)
			totalTime = fmt.Sprintf("%d:%02d:00", hours, minutes)

			// Calculer currentTime basé sur percentage
			totalMinutes := hours*60 + minutes
			currentMinutes := (totalMinutes * percentage) / 100
			currentTime = fmt.Sprintf("%d:%02d:00", currentMinutes/60, currentMinutes%60)

			subtitle = fmt.Sprintf("Film • %d", gofakeit.Year())
			playHref = fmt.Sprintf("/play?id=%d&type=movie", contentID)
		} else {
			minutes := gofakeit.Number(40, 65)
			duration = fmt.Sprintf("%dmin", minutes)
			totalTime = fmt.Sprintf("%d:00", minutes)

			currentMinutes := (minutes * percentage) / 100
			currentTime = fmt.Sprintf("%d:%02d", currentMinutes, gofakeit.Number(0, 59))

			season := gofakeit.Number(1, 8)
			episode := gofakeit.Number(1, 12)
			subtitle = fmt.Sprintf("Série • S%d E%d", season, episode)
			playHref = fmt.Sprintf("/play?id=%d&type=tvshow&season=%d&episode=%d", contentID, season, episode)
		}

		items = append(items, models.UserListItem{
			ID:           i + 1,
			ContentID:    contentID,
			ContentType:  contentType,
			Title:        gofakeit.MovieName(),
			Image:        "https://picsum.photos/600/400",
			Description:  gofakeit.Sentence(8),
			Subtitle:     subtitle,
			Duration:     duration,
			AddedDate:    gofakeit.Date().Format("2006-01-02"),
			Percentage:   percentage,
			Progress:     percentage,
			CurrentTime:  currentTime,
			TotalTime:    totalTime,
			PlayHref:     playHref,
			FavoriteHref: fmt.Sprintf("/list/toggle/%d", contentID),
			Category:     categories[gofakeit.Number(0, len(categories)-1)],
		})
	}
	generatedUserListItem = items
	return items
}
