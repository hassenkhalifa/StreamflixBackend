package handlers

import (
	"StreamflixBackend/internal/models"

	"github.com/brianvoe/gofakeit/v7"
)

func RandomCategories() []models.Categories {
	var categoriesArray []models.Categories

	for i := 0; i < gofakeit.Number(10, 20); i++ {
		categoriesArray = append(categoriesArray, models.Categories{
			ID:           gofakeit.Number(1, 1000),
			CategoryName: gofakeit.MovieGenre(),
			Description:  gofakeit.ProductDescription(),
			Href:         "/search",
			Color:        models.Gradients[gofakeit.Number(0, (len(models.Gradients)-1))],
			Previews:     []string{"https://picsum.photos/600/400"},
		})
	}

	return categoriesArray
}
