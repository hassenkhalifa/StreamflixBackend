// Package middleware fournit les middleware HTTP Gin pour l'application StreamFlix.
//
// Ce package contient les middleware transversaux appliqués à toutes les requêtes :
//   - CORS : gestion des origines cross-origin autorisées
//   - Logger : logging structuré de chaque requête via slog
//   - Recovery : récupération des panics avec réponse d'erreur standardisée
//   - Security : ajout des headers de sécurité HTTP
//   - RateLimit : limitation du nombre de requêtes par IP (token bucket)
//
// Tous les middleware suivent le pattern gin.HandlerFunc et sont composables
// via router.Use() dans le fichier router.go.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Cors retourne un middleware Gin configurant les headers CORS (Cross-Origin Resource Sharing).
//
// Le middleware valide l'en-tête Origin de chaque requête entrante par rapport à la liste
// des origines autorisées fournie en paramètre. Seules les origines présentes dans
// allowedOrigins reçoivent l'en-tête Access-Control-Allow-Origin en réponse ;
// les origines non reconnues ne provoquent pas d'erreur mais ne reçoivent pas cet en-tête,
// ce qui amène le navigateur à bloquer la requête côté client.
//
// Les origines sont nettoyées (espaces supprimés, chaînes vides ignorées) puis stockées
// dans un map pour une recherche en O(1). Le wildcard "*" n'est pas supporté en
// production afin de garantir une politique CORS stricte.
//
// Headers CORS définis sur chaque réponse :
//   - Access-Control-Allow-Methods : GET, POST, PUT, DELETE, OPTIONS
//   - Access-Control-Allow-Headers : Origin, Content-Type, Accept, Authorization, X-Requested-With
//   - Access-Control-Expose-Headers : Content-Length
//   - Access-Control-Allow-Credentials : true
//   - Access-Control-Max-Age : 86400 (24 heures de mise en cache preflight)
//
// Les requêtes preflight (méthode OPTIONS) sont interceptées et répondues immédiatement
// avec un statut 204 No Content sans passer aux handlers suivants.
//
// Exemple d'utilisation :
//
//	router.Use(middleware.Cors([]string{"https://streamflix.example.com", "http://localhost:3000"}))
func Cors(allowedOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			originSet[trimmed] = true
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if originSet[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
