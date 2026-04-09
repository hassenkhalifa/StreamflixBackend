// Package middleware fournit les tests unitaires pour les middlewares HTTP de StreamFlix.
//
// Les tests couvrent les middlewares suivants :
//   - CORS : vérification des origines autorisées et des requêtes preflight
//   - Security Headers : validation des en-têtes de sécurité HTTP
//   - Recovery : gestion des panics avec réponse JSON structurée
//   - Rate Limiter : limitation du débit par adresse IP, indépendance entre IPs
//   - Logger : middleware de journalisation des requêtes
//   - RateLimitWithBypass : contournement du rate limiting pour les endpoints critiques
package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"StreamflixBackend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// init configure Gin en mode test pour supprimer les logs superflus
// et éviter les effets de bord pendant l'exécution des tests.
func init() {
	gin.SetMode(gin.TestMode)
}

// setupRouter crée un routeur Gin minimal avec les middlewares spécifiés.
// Cette fonction utilitaire est utilisée par tous les tests pour isoler
// chaque middleware dans un environnement propre.
func setupRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	for _, mw := range middlewares {
		router.Use(mw)
	}
	return router
}

// ===========================================================================
// Tests CORS - Vérification du partage de ressources entre origines
// ===========================================================================

// TestCors_AllowedOrigin vérifie le comportement du middleware CORS avec
// différentes origines. Utilise un pattern de tests pilotés par table pour
// tester les origines autorisées (localhost, example.com) et refusées (evil.com).
// L'en-tête Access-Control-Allow-Origin ne doit être présent que pour les origines autorisées.
func TestCors_AllowedOrigin(t *testing.T) {
	router := setupRouter(Cors([]string{"http://localhost:3000", "https://example.com"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
	}{
		{"allowed origin", "http://localhost:3000", "http://localhost:3000"},
		{"another allowed origin", "https://example.com", "https://example.com"},
		{"disallowed origin", "https://evil.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

// TestCors_Preflight vérifie que les requêtes preflight (OPTIONS) sont
// correctement gérées avec un statut 204 No Content et un en-tête
// Access-Control-Max-Age de 86400 secondes (24 heures).
func TestCors_Preflight(t *testing.T) {
	router := setupRouter(Cors([]string{"http://localhost:3000"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

// ===========================================================================
// Tests des en-têtes de sécurité HTTP
// ===========================================================================

// TestSecurityHeaders vérifie que le middleware ajoute tous les en-têtes de
// sécurité requis : X-Content-Type-Options (nosniff), X-Frame-Options (DENY),
// X-XSS-Protection (1; mode=block) et Referrer-Policy (strict-origin-when-cross-origin).
func TestSecurityHeaders(t *testing.T) {
	router := setupRouter(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

// ===========================================================================
// Tests de récupération après panic
// ===========================================================================

// TestRecovery_PanicHandled vérifie que le middleware Recovery intercepte les panics,
// retourne un statut 500 avec une réponse JSON structurée (code INTERNAL_ERROR)
// et ne divulgue pas le message de panic dans la réponse (sécurité).
func TestRecovery_PanicHandled(t *testing.T) {
	router := setupRouter(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("something went wrong")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "INTERNAL_ERROR", resp.Error.Code)
	// Should not expose panic message
	assert.NotContains(t, resp.Error.Message, "something went wrong")
}

// TestRecovery_NoPanic vérifie que le middleware Recovery n'interfère pas
// avec le traitement normal des requêtes lorsqu'aucune panic ne se produit.
func TestRecovery_NoPanic(t *testing.T) {
	router := setupRouter(Recovery())
	router.GET("/ok", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ===========================================================================
// Tests du limiteur de débit (Rate Limiter)
// ===========================================================================

// TestRateLimiter_AllowsRequests vérifie qu'une requête en dessous de la limite
// passe correctement et que les en-têtes de rate limiting (X-RateLimit-Limit,
// X-RateLimit-Remaining) sont présents dans la réponse.
func TestRateLimiter_AllowsRequests(t *testing.T) {
	limiter := NewRateLimiter(60) // 60 req/min
	router := setupRouter(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// First request should pass
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
}

// TestRateLimiter_BlocksExcessRequests vérifie que les requêtes au-delà de la
// limite configurée (5 req/min) sont bloquées avec un statut 429 Too Many Requests
// et un en-tête Retry-After indiquant quand réessayer.
func TestRateLimiter_BlocksExcessRequests(t *testing.T) {
	limiter := NewRateLimiter(5) // 5 req/min - very low limit for testing
	router := setupRouter(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Exhaust the limit
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Next request should be rate limited
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.NotEmpty(t, w.Header().Get("Retry-After"))
}

// TestRateLimiter_DifferentIPsIndependent vérifie que le rate limiting est
// appliqué indépendamment par adresse IP. L'épuisement du quota d'une IP
// ne doit pas affecter les requêtes provenant d'une autre IP.
func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	limiter := NewRateLimiter(2)
	router := setupRouter(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Exhaust limit for IP1
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		router.ServeHTTP(w, req)
	}

	// IP2 should still work
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ===========================================================================
// Tests du middleware de journalisation (Logger)
// ===========================================================================

// TestLogger_NoError vérifie que le middleware Logger n'interfère pas avec
// le traitement normal des requêtes et que la réponse reste intacte.
func TestLogger_NoError(t *testing.T) {
	router := setupRouter(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ===========================================================================
// Tests du Rate Limiting avec contournement (Bypass)
// ===========================================================================

// TestRateLimitBypass_HealthEndpoint vérifie que l'endpoint /health est exempté
// du rate limiting. Même avec une limite très basse (1 req/min), les requêtes
// vers /health doivent toujours passer (5 requêtes successives testées).
func TestRateLimitBypass_HealthEndpoint(t *testing.T) {
	limiter := NewRateLimiter(1) // Very low limit
	router := setupRouter(limiter.RateLimitWithBypass())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Should always pass for /health
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}
