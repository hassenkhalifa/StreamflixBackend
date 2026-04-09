package handlers

import (
	"StreamflixBackend/internal/models"
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
)

// generatedUserListItem est un cache en mémoire pour les éléments de liste utilisateur.
// Une fois générés lors du premier appel à GetUserListItems, les éléments sont stockés
// ici afin d'être retournés directement lors des appels suivants (pattern singleton).
var generatedUserListItem = []models.UserListItem{}

// GetUserListItems génère et retourne une liste fictive d'éléments de la bibliothèque utilisateur.
//
// Cette fonction utilise un cache interne (generatedUserListItem) : si les éléments ont déjà
// été générés lors d'un appel précédent, le cache est retourné immédiatement sans régénération.
//
// Lors de la première invocation, entre 5 et 15 éléments sont créés aléatoirement via gofakeit.
// Chaque élément représente un film ou une série TV et contient :
//   - un identifiant séquentiel (à partir de 1),
//   - un identifiant de contenu aléatoire (100 à 999),
//   - un type de contenu ("movie" ou "tvshow"),
//   - un titre de film fictif, une image placeholder (picsum.photos), une description,
//   - un sous-titre formaté (ex. "Film - 2023" ou "Série - S3 E7"),
//   - une durée formatée (ex. "2h 15min" pour un film, "52min" pour une série),
//   - un pourcentage de progression aléatoire (0 à 100) avec le temps courant et total calculés,
//   - des liens de lecture (playHref) et de favori (favoriteHref),
//   - une catégorie parmi "favorites", "history" ou "watchlist".
//
// Pour les films, la durée est entre 1h et 3h. Pour les séries, entre 40 et 65 minutes.
// Le temps courant (currentTime) est calculé proportionnellement au pourcentage de progression.
//
// Aucun paramètre n'est requis.
//
// Retourne un slice de [models.UserListItem]. Cette fonction ne retourne pas d'erreur ;
// toutes les données sont fictives (mock).
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
