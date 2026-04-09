package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"StreamflixBackend/internal/utils"

	"github.com/gin-gonic/gin"
)

// visitor représente l'état du token bucket pour un visiteur unique (identifié par IP).
//
// Chaque visiteur possède un seau de jetons (tokens) qui se remplit progressivement
// au fil du temps selon le débit configuré (rate). Un jeton est consommé à chaque
// requête autorisée. Lorsque le seau est vide (tokens < 1), la requête est rejetée.
type visitor struct {
	tokens    float64
	lastSeen  time.Time
	maxTokens float64
	rate      float64 // jetons par seconde
}

// RateLimiter implémente un limiteur de débit en mémoire par adresse IP,
// basé sur l'algorithme du token bucket (seau à jetons).
//
// Principe de l'algorithme token bucket :
// Chaque adresse IP dispose d'un seau contenant un nombre maximal de jetons (burst).
// Les jetons sont consommés à raison d'un par requête. Le seau se remplit
// continuellement à un débit constant (rate jetons/seconde). Si le seau est vide
// au moment de la requête, celle-ci est rejetée avec un statut 429 Too Many Requests.
// Ce mécanisme permet d'absorber des pics de trafic courts (burst) tout en
// maintenant un débit moyen contrôlé sur la durée.
//
// L'accès concurrent à la map des visiteurs est protégé par un sync.Mutex.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64 // jetons par seconde
	burst    int     // nombre maximal de jetons
}

// NewRateLimiter crée un nouveau limiteur de débit autorisant requestsPerMinute
// requêtes par minute.
//
// Le débit (rate) est calculé en jetons par seconde : requestsPerMinute / 60.
// La capacité maximale du seau (burst) est égale à requestsPerMinute, ce qui
// permet à un nouveau visiteur d'effectuer un burst initial de requêtes équivalent
// à une minute complète de quota.
//
// Une goroutine de nettoyage est lancée automatiquement en arrière-plan
// pour supprimer les visiteurs inactifs depuis plus de 5 minutes. Ce nettoyage
// s'exécute toutes les 3 minutes afin d'éviter une croissance non bornée de la
// map en mémoire.
//
// Exemple d'utilisation :
//
//	limiter := middleware.NewRateLimiter(120) // 120 requêtes/minute
//	router.Use(limiter.RateLimitWithBypass())
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     float64(requestsPerMinute) / 60.0,
		burst:    requestsPerMinute,
	}

	// Lancement de la goroutine de nettoyage des visiteurs expirés (toutes les 3 minutes)
	go rl.cleanup()

	return rl
}

// Middleware retourne un gin.HandlerFunc appliquant la limitation de débit.
//
// Pour chaque requête, le middleware :
//  1. Identifie le visiteur par son adresse IP (c.ClientIP()).
//  2. Crée une entrée dans la map si c'est un nouveau visiteur, avec un seau plein.
//  3. Recalcule le nombre de jetons disponibles en fonction du temps écoulé
//     depuis la dernière requête (recharge proportionnelle au débit).
//  4. Si le seau contient au moins 1 jeton, consomme un jeton et laisse passer la requête.
//  5. Si le seau est vide, rejette la requête avec un statut 429 via utils.RateLimited.
//
// Headers HTTP ajoutés à chaque réponse :
//   - X-RateLimit-Limit : capacité maximale du seau (burst)
//   - X-RateLimit-Remaining : nombre de jetons restants après cette requête
//   - Retry-After : (uniquement en cas de rejet) nombre de secondes avant qu'un
//     nouveau jeton soit disponible
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists {
			v = &visitor{
				tokens:    float64(rl.burst),
				lastSeen:  time.Now(),
				maxTokens: float64(rl.burst),
				rate:      rl.rate,
			}
			rl.visitors[ip] = v
		}

		// Recharge des jetons en fonction du temps écoulé
		now := time.Now()
		elapsed := now.Sub(v.lastSeen).Seconds()
		v.tokens += elapsed * v.rate
		if v.tokens > v.maxTokens {
			v.tokens = v.maxTokens
		}
		v.lastSeen = now

		remaining := int(v.tokens)
		if v.tokens < 1 {
			rl.mu.Unlock()

			retryAfter := int((1 - v.tokens) / v.rate)
			if retryAfter < 1 {
				retryAfter = 1
			}

			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.burst))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))

			utils.RateLimited(c)
			c.Abort()
			return
		}

		v.tokens--
		remaining = int(v.tokens)
		rl.mu.Unlock()

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.burst))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		c.Next()
	}
}

// cleanup est une goroutine de nettoyage qui supprime périodiquement les visiteurs
// inactifs de la map en mémoire.
//
// Elle s'exécute toutes les 3 minutes via un time.Ticker et supprime tout visiteur
// dont la dernière requête remonte à plus de 5 minutes. Cette opération empêche
// la map visitors de croître indéfiniment en cas de trafic provenant de nombreuses
// adresses IP distinctes.
//
// L'accès à la map est protégé par le mutex du RateLimiter.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 5*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// HealthCheckBypass retourne un middleware Gin qui marque les requêtes vers /health
// pour qu'elles contournent la limitation de débit.
//
// Ce middleware positionne la clé "skip_rate_limit" à true dans le contexte Gin
// pour les requêtes dont le chemin est /health. Cette valeur est ensuite
// consultée par [RateLimiter.RateLimitWithBypass] pour décider si la limitation
// de débit doit être appliquée.
//
// Ce mécanisme permet aux sondes de santé (health checks) des load balancers
// et orchestrateurs de fonctionner sans être affectées par le rate limiting.
//
// Exemple d'utilisation :
//
//	router.Use(middleware.HealthCheckBypass())
func HealthCheckBypass() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Set("skip_rate_limit", true)
		}
		c.Next()
	}
}

// RateLimitWithBypass retourne un middleware Gin combinant la limitation de débit
// avec une logique de contournement pour certaines requêtes.
//
// Les requêtes suivantes ne sont pas soumises au rate limiting :
//   - Requêtes marquées avec la clé "skip_rate_limit" dans le contexte Gin
//     (positionnée par [HealthCheckBypass])
//   - Requêtes vers le endpoint /health (vérification directe du chemin)
//   - Requêtes preflight CORS (méthode OPTIONS), qui ne doivent pas consommer
//     de jetons car elles sont automatiquement émises par les navigateurs
//
// Pour toutes les autres requêtes, le middleware délègue au [RateLimiter.Middleware]
// standard qui applique l'algorithme token bucket.
//
// Exemple d'utilisation :
//
//	limiter := middleware.NewRateLimiter(120)
//	router.Use(limiter.RateLimitWithBypass())
func (rl *RateLimiter) RateLimitWithBypass() gin.HandlerFunc {
	mw := rl.Middleware()
	return func(c *gin.Context) {
		if skip, exists := c.Get("skip_rate_limit"); exists && skip.(bool) {
			c.Next()
			return
		}
		// Contournement du rate limit pour le endpoint de santé
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Contournement du rate limit pour les requêtes preflight OPTIONS
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		mw(c)
	}
}
