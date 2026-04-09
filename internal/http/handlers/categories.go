// Package handlers contient les fonctions de logique métier pour l'API StreamFlix.
//
// Ce package implémente les appels aux API externes (TMDB, Real-Debrid, Torrentio,
// Zone Téléchargement, TVMaze) et la transformation des données en DTOs pour le frontend.
//
// Organisation par domaine fonctionnel :
//   - movies.go : films (populaires, tendances, détails, recherche, crédits)
//   - series.go : séries TV (tendances, par genre, détails, recherche)
//   - player.go : construction du lecteur vidéo avec qualités et pistes
//   - realdebrid.go : interactions avec l'API Real-Debrid (unrestrict, transcode, media infos)
//   - real_debrid.go : client Real-Debrid complet (magnet, torrent, streaming)
//   - streaming.go : service de streaming (workflow magnet → stream)
//   - torrentio.go : client Torrentio (recherche de streams torrent)
//   - zt.go : parser Zone Téléchargement (scraping HTML via goquery)
//   - trancodevideo.go : transcodage vidéo via FFmpeg
//   - categories.go : génération de catégories aléatoires (données mock)
//   - user.go : génération de listes utilisateur (données mock)
//
// Note : les caches sont définis dans movies.go pour les données TMDB
// et dans constants_models.go pour les caches exportés.
package handlers

import (
	"StreamflixBackend/internal/models"

	"github.com/brianvoe/gofakeit/v7"
)

// RandomCategories génère une liste aléatoire de catégories fictives pour le frontend.
//
// Cette fonction produit entre 10 et 20 catégories, chacune contenant :
//   - un identifiant aléatoire (1 à 1000),
//   - un nom de genre cinématographique généré via gofakeit,
//   - une description produit fictive,
//   - un lien de redirection fixe vers "/search",
//   - un dégradé de couleur choisi aléatoirement parmi models.Gradients,
//   - une image de prévisualisation provenant de picsum.photos.
//
// Aucun paramètre n'est requis.
//
// Retourne un slice de [models.Categories] contenant les catégories générées.
// Cette fonction ne retourne pas d'erreur ; les données sont entièrement fictives (mock).
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
