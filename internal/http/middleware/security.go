package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders retourne un middleware Gin ajoutant des en-têtes HTTP de sécurité
// à chaque réponse.
//
// Ce middleware renforce la sécurité côté client en instruisant les navigateurs
// via les en-têtes de réponse suivants :
//
//   - X-Content-Type-Options: nosniff
//     Empêche le navigateur de deviner le type MIME du contenu (MIME sniffing),
//     forçant le respect du Content-Type déclaré par le serveur.
//
//   - X-Frame-Options: DENY
//     Interdit l'intégration de la page dans une iframe, protégeant contre
//     les attaques de type clickjacking.
//
//   - X-XSS-Protection: 1; mode=block
//     Active le filtre XSS intégré du navigateur et bloque le rendu de la page
//     si une attaque XSS réfléchie est détectée.
//
//   - Referrer-Policy: strict-origin-when-cross-origin
//     Limite les informations envoyées dans l'en-tête Referer : l'URL complète
//     est transmise pour les requêtes same-origin, mais seule l'origine est
//     envoyée pour les requêtes cross-origin (et rien en cas de downgrade HTTPS vers HTTP).
//
//   - Content-Security-Policy: default-src 'self'
//     Restreint le chargement de toutes les ressources (scripts, styles, images, etc.)
//     à la même origine que la page, bloquant l'injection de contenu tiers.
//
// Exemple d'utilisation :
//
//	router.Use(middleware.SecurityHeaders())
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
