// Package utils fournit des utilitaires partagés pour l'application StreamFlix.
//
// Ce package contient principalement les fonctions de réponse HTTP standardisées
// utilisées par tous les handlers Gin de l'application. Toutes les réponses JSON
// suivent le format unifié { "data": ..., "error": ... } pour faciliter le
// traitement côté client.
//
// Exemple d'utilisation dans un handler :
//
//	func MyHandler(c *gin.Context) {
//	    data, err := service.GetData()
//	    if err != nil {
//	        utils.InternalError(c, err)
//	        return
//	    }
//	    utils.RespondSuccess(c, http.StatusOK, data)
//	}
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse est le format de réponse JSON standardisé de l'API StreamFlix.
//
// Toutes les réponses de l'API suivent ce format :
//   - En cas de succès : Data contient les données, Error est nil
//   - En cas d'erreur : Data est nil, Error contient les détails de l'erreur
//
// Exemple de réponse succès :
//
//	{"data": {"id": 1, "title": "Film"}, "error": null}
//
// Exemple de réponse erreur :
//
//	{"data": null, "error": {"code": "NOT_FOUND", "message": "Film non trouvé"}}
type APIResponse struct {
	Data  interface{}  `json:"data"`  // Données de la réponse (nil en cas d'erreur)
	Error *ErrorDetail `json:"error"` // Détails de l'erreur (nil en cas de succès)
}

// ErrorDetail représente les détails structurés d'une erreur dans la réponse API.
//
// Le champ Code contient un identifiant machine-readable (ex: "BAD_REQUEST", "NOT_FOUND")
// et Message contient une description lisible pour l'utilisateur.
type ErrorDetail struct {
	Code    string `json:"code"`    // Code d'erreur machine-readable (ex: "BAD_REQUEST")
	Message string `json:"message"` // Message d'erreur lisible par l'utilisateur
}

// RespondSuccess envoie une réponse JSON de succès standardisée.
//
// Paramètres :
//   - c : contexte Gin de la requête
//   - code : code de statut HTTP (ex: 200, 201)
//   - data : données à inclure dans le champ "data" de la réponse
func RespondSuccess(c *gin.Context, code int, data interface{}) {
	c.JSON(code, APIResponse{
		Data:  data,
		Error: nil,
	})
}

// RespondError envoie une réponse JSON d'erreur standardisée.
//
// Paramètres :
//   - c : contexte Gin de la requête
//   - httpCode : code de statut HTTP (ex: 400, 404, 500)
//   - errorCode : code d'erreur machine-readable (ex: "BAD_REQUEST")
//   - message : message d'erreur lisible pour l'utilisateur
func RespondError(c *gin.Context, httpCode int, errorCode string, message string) {
	c.JSON(httpCode, APIResponse{
		Data: nil,
		Error: &ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
	})
}

// BadRequest envoie une réponse HTTP 400 (Bad Request) standardisée.
// Utilisé quand les paramètres de la requête sont invalides ou manquants.
func BadRequest(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// NotFound envoie une réponse HTTP 404 (Not Found) standardisée.
// Utilisé quand une ressource demandée n'existe pas (film, série, stream).
func NotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, "NOT_FOUND", message)
}

// InternalError envoie une réponse HTTP 500 (Internal Server Error) standardisée.
//
// Le paramètre error est ignoré intentionnellement pour ne pas exposer les détails
// internes de l'erreur au client. L'erreur doit être loggée séparément via slog
// avant d'appeler cette fonction.
func InternalError(c *gin.Context, _ error) {
	RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue")
}

// RateLimited envoie une réponse HTTP 429 (Too Many Requests) standardisée.
// Utilisé par le middleware de rate limiting quand un client dépasse sa limite.
func RateLimited(c *gin.Context) {
	RespondError(c, http.StatusTooManyRequests, "RATE_LIMITED", "Trop de requetes, veuillez reessayer plus tard")
}

// APIError est conservé pour la compatibilité ascendante pendant la migration
// vers le nouveau format de réponse standardisé. Les nouveaux handlers doivent
// utiliser RespondError à la place.
//
// Deprecated: utiliser RespondError pour les nouveaux développements.
func APIError(c *gin.Context, code int, message string) {
	RespondError(c, code, "ERROR", message)
}
