package models

// ============================================================================
// USER MODELS
// ============================================================================

type UserListItem struct {
	ID           int    `json:"id"`
	ContentID    int    `json:"contentId"`
	ContentType  string `json:"contentType"` // "movie" ou "tvshow"
	Title        string `json:"title"`
	Image        string `json:"image"`
	Description  string `json:"description"`
	Subtitle     string `json:"subtitle"`
	Duration     string `json:"duration"`
	AddedDate    string `json:"addedDate"`
	Percentage   int    `json:"percentage"`
	Progress     int    `json:"progress"`
	CurrentTime  string `json:"currentTime"`
	TotalTime    string `json:"totalTime"`
	PlayHref     string `json:"playHref"`
	FavoriteHref string `json:"favoriteHref"`
	Category     string `json:"category"` // "favorites", "history", "watchlist"
}
