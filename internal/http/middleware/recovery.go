package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"StreamflixBackend/internal/utils"

	"github.com/gin-gonic/gin"
)

// Recovery retourne un middleware Gin de récupération de panics.
//
// Ce middleware intercepte toute panic survenant dans la chaîne de handlers Gin
// via un appel defer/recover. Lorsqu'une panic est capturée, le middleware :
//
//  1. Enregistre un log d'erreur structuré via slog contenant :
//     - l'erreur à l'origine de la panic
//     - le chemin de la requête (path)
//     - la méthode HTTP
//     - la stack trace complète (debug.Stack)
//
//  2. Retourne au client une réponse JSON générique avec le statut 500
//     (Internal Server Error) utilisant le format standardisé utils.APIResponse.
//     Le message d'erreur est volontairement générique ("Une erreur interne est survenue")
//     afin de ne pas exposer les détails internes de l'application au client,
//     évitant ainsi toute fuite d'information sensible (chemins de fichiers,
//     noms de fonctions internes, données en mémoire, etc.).
//
//  3. Interrompt la chaîne de middleware via c.Abort() pour empêcher tout
//     traitement ultérieur de la requête.
//
// Exemple d'utilisation :
//
//	router.Use(middleware.Recovery())
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				slog.Error("panic recovered",
					slog.Any("error", r),
					slog.String("path", c.Request.URL.Path),
					slog.String("method", c.Request.Method),
					slog.String("stack", stack),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, utils.APIResponse{
					Data: nil,
					Error: &utils.ErrorDetail{
						Code:    "INTERNAL_ERROR",
						Message: "Une erreur interne est survenue",
					},
				})
			}
		}()
		c.Next()
	}
}
