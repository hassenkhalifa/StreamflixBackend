package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger retourne un middleware Gin de logging structuré utilisant le package standard slog.
//
// Le middleware capture le chemin et la query string avant l'exécution du handler,
// puis enregistre un log structuré après la complétion de la requête.
//
// Champs enregistrés pour chaque requête :
//   - status : code de statut HTTP de la réponse (ex. 200, 404, 500)
//   - method : méthode HTTP utilisée (GET, POST, PUT, DELETE, etc.)
//   - path : chemin de l'URL demandé (ex. /api/v1/movies)
//   - query : paramètres de la query string brute (ex. page=1&limit=10)
//   - ip : adresse IP du client telle que déterminée par Gin (respecte X-Forwarded-For)
//   - latency : durée totale de traitement de la requête
//   - body_size : taille du corps de la réponse en octets
//   - errors : erreurs privées Gin accumulées pendant le traitement (si présentes)
//
// Le niveau de log est déterminé dynamiquement selon le code de statut :
//   - >= 500 : slog.LevelError (erreur serveur)
//   - >= 400 : slog.LevelWarn (erreur client)
//   - < 400 : slog.LevelInfo (succès)
//
// Exemple d'utilisation :
//
//	router.Use(middleware.Logger())
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
			slog.Int("body_size", c.Writer.Size()),
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		slog.LogAttrs(c.Request.Context(), level, "HTTP request",
			attrs...,
		)
	}
}
