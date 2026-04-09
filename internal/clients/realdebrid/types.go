package realdebrid

import "fmt"

// UnrestrictResponse représente la réponse de l'API Real-Debrid lors de la
// dé-restriction d'un lien hébergeur via l'endpoint /unrestrict/link.
//
// Les champs principaux sont :
//   - Download : URL de téléchargement direct du fichier dé-restreint.
//   - Filename : nom du fichier original.
//   - Filesize : taille du fichier en octets.
//   - Streamable : indique si le fichier est streamable (1) ou non (0).
type UnrestrictResponse struct {
	ID         string `json:"id"`         // Identifiant unique de la dé-restriction.
	Filename   string `json:"filename"`   // Nom du fichier original.
	MimeType   string `json:"mimeType"`   // Type MIME du fichier (ex. "video/mp4").
	Filesize   int64  `json:"filesize"`   // Taille du fichier en octets.
	Link       string `json:"link"`       // Lien hébergeur original.
	Host       string `json:"host"`       // Nom de l'hébergeur (ex. "uptobox").
	Download   string `json:"download"`   // URL de téléchargement direct.
	Streamable int    `json:"streamable"` // 1 si le fichier est streamable, 0 sinon.
}

// ErrorResponse représente une réponse d'erreur retournée par l'API Real-Debrid.
//
// Elle implémente l'interface error pour pouvoir être utilisée directement
// comme valeur d'erreur Go. Le champ ErrorCode est optionnel (pointeur)
// car toutes les erreurs Real-Debrid ne fournissent pas un code numérique.
type ErrorResponse struct {
	Message   string `json:"error"`      // Message d'erreur lisible.
	ErrorCode *int   `json:"error_code"` // Code d'erreur numérique Real-Debrid (optionnel).
}

// Error implémente l'interface error pour ErrorResponse.
//
// Si un code d'erreur est présent, le format est "message (code=N)".
// Sinon, seul le message est retourné.
func (e ErrorResponse) Error() string {
	if e.ErrorCode != nil {
		return fmt.Sprintf("%s (code=%d)", e.Message, *e.ErrorCode)
	}
	return e.Message
}
