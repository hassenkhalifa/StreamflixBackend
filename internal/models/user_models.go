package models

// ============================================================================
// USER MODELS
// ============================================================================

// UserListItem représente un élément dans une liste utilisateur (favoris, historique
// de visionnage ou liste de suivi). Cette structure unifie les différents types de contenu
// (films et séries TV) sous un même format, permettant au frontend d'afficher une liste
// homogène quel que soit le type de contenu.
//
// Les champs de progression (Percentage, Progress, CurrentTime, TotalTime) sont utilisés
// principalement pour l'historique de visionnage, tandis que les champs de navigation
// (PlayHref, FavoriteHref) permettent au frontend de construire les actions utilisateur.
type UserListItem struct {
	ID           int    `json:"id"`           // Identifiant unique de l'entrée dans la liste
	ContentID    int    `json:"contentId"`    // Identifiant TMDB du contenu (film ou série)
	ContentType  string `json:"contentType"`  // Type de contenu : "movie" pour un film, "tvshow" pour une série
	Title        string `json:"title"`        // Titre du contenu
	Image        string `json:"image"`        // URL de l'affiche du contenu
	Description  string `json:"description"`  // Description courte ou synopsis
	Subtitle     string `json:"subtitle"`     // Sous-titre affiché (par exemple le genre ou l'année)
	Duration     string `json:"duration"`     // Durée formatée du contenu (par exemple "2h15" ou "45min")
	AddedDate    string `json:"addedDate"`    // Date d'ajout à la liste au format lisible
	Percentage   int    `json:"percentage"`   // Pourcentage de progression de visionnage (0-100)
	Progress     int    `json:"progress"`     // Progression absolue en secondes
	CurrentTime  string `json:"currentTime"`  // Temps actuel formaté (par exemple "01:23:45")
	TotalTime    string `json:"totalTime"`    // Durée totale formatée (par exemple "02:15:00")
	PlayHref     string `json:"playHref"`     // URL pour reprendre ou lancer la lecture
	FavoriteHref string `json:"favoriteHref"` // URL pour ajouter/retirer des favoris
	Category     string `json:"category"`     // Catégorie de la liste : "favorites", "history" ou "watchlist"
}
