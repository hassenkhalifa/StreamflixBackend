// Package http fournit les tests d'intégration pour le routeur HTTP de StreamFlix.
//
// Ces tests vérifient le bon fonctionnement du routeur complet, incluant :
//   - L'endpoint de santé (/health) avec ses données et en-têtes de sécurité
//   - La politique CORS (origines autorisées et refusées)
//   - Les routes legacy (rétrocompatibilité avec l'ancienne API)
//   - Les endpoints API v1 (films, séries TV)
//   - La validation des paramètres d'entrée (IDs invalides, paramètres manquants)
//
// Tous les tests utilisent un routeur configuré avec testConfig() et des requêtes
// HTTP simulées via httptest, sans appels réseau externes.
package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"StreamflixBackend/internal/config"
	"StreamflixBackend/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testConfig crée une configuration de test avec des valeurs prédéfinies.
// Les tokens sont factices (test_token, test_tmdb_token) car les tests
// ne font pas d'appels aux API externes. Le rate limit est élevé (1000)
// pour éviter les faux positifs dans les tests.
func testConfig() *config.Config {
	return &config.Config{
		Port:            "2000",
		GinMode:         "test",
		Environment:     "development",
		RealDebridToken: "test_token",
		TMDBToken:       "test_tmdb_token",
		CORSOrigins:     []string{"http://localhost:3000"},
		RateLimitPerMin: 1000,
		UserAgent:       "Test/1.0",
	}
}

// TestHealthEndpoint vérifie que l'endpoint /health retourne un statut 200
// avec une réponse JSON contenant le statut "healthy", la version "1.0.0"
// et un champ uptime non vide. Valide aussi la structure APIResponse standard.
func TestHealthEndpoint(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Data)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "healthy", data["status"])
	assert.Equal(t, "1.0.0", data["version"])
	assert.NotEmpty(t, data["uptime"])
}

// TestHealthEndpoint_HasSecurityHeaders vérifie que les en-têtes de sécurité
// (X-Content-Type-Options, X-Frame-Options) sont présents sur les réponses
// du routeur, confirmant que le middleware SecurityHeaders est actif.
func TestHealthEndpoint_HasSecurityHeaders(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

// TestCORS_RejectsUnauthorizedOrigin vérifie qu'une requête provenant d'une
// origine non autorisée (https://evil.com) ne reçoit pas l'en-tête
// Access-Control-Allow-Origin dans la réponse.
func TestCORS_RejectsUnauthorizedOrigin(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "https://evil.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should NOT have Access-Control-Allow-Origin for unauthorized origin
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

// TestCORS_AllowsAuthorizedOrigin vérifie qu'une requête provenant d'une
// origine autorisée (http://localhost:3000) reçoit correctement l'en-tête
// Access-Control-Allow-Origin correspondant.
func TestCORS_AllowsAuthorizedOrigin(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

// TestLegacyRoutes_MoviesListExists vérifie que la route legacy /movieslist
// existe et retourne un statut 200. Cette route retourne une liste de films
// aléatoire sans appel API externe.
func TestLegacyRoutes_MoviesListExists(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/movieslist", nil)
	router.ServeHTTP(w, req)

	// Should return 200 with random movie list (no external API needed)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestLegacyRoutes_CategoriesExists vérifie que la route legacy /categories
// existe et retourne un statut 200.
func TestLegacyRoutes_CategoriesExists(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestLegacyRoutes_ContentDetailsExists vérifie que la route legacy /contentDetails
// existe et retourne un statut 200.
func TestLegacyRoutes_ContentDetailsExists(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/contentDetails", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestLegacyRoutes_UserListExists vérifie que la route legacy /user/list
// existe et retourne un statut 200.
func TestLegacyRoutes_UserListExists(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/list", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAPIV1_MoviesEndpointsExist vérifie que les endpoints API v1 pour les films
// (/api/v1/movies/popular, /api/v1/movies/top-rated, /api/v1/movies/trending)
// sont enregistrés et ne retournent pas 404. Les réponses peuvent être 500
// en raison des tokens invalides dans l'environnement de test.
func TestAPIV1_MoviesEndpointsExist(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	// These will fail with external API calls (expected) but should not 404
	endpoints := []string{
		"/api/v1/movies/popular",
		"/api/v1/movies/top-rated",
		"/api/v1/movies/trending",
	}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, ep, nil)
			router.ServeHTTP(w, req)
			// Should not be 404 (might be 500 due to invalid tokens in test)
			assert.NotEqual(t, http.StatusNotFound, w.Code, "endpoint %s should exist", ep)
		})
	}
}

// TestAPIV1_TVEndpointsExist vérifie que les endpoints API v1 pour les séries TV
// (/api/v1/tv/trending, /api/v1/tv/popular) sont enregistrés et ne retournent pas 404.
func TestAPIV1_TVEndpointsExist(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	endpoints := []string{
		"/api/v1/tv/trending",
		"/api/v1/tv/popular",
	}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, ep, nil)
			router.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusNotFound, w.Code, "endpoint %s should exist", ep)
		})
	}
}

// TestBadRequest_InvalidMovieID vérifie qu'un ID de film invalide (non numérique)
// retourne un statut 400 Bad Request avec le code d'erreur BAD_REQUEST
// dans la réponse JSON structurée.
func TestBadRequest_InvalidMovieID(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/movie/invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "BAD_REQUEST", resp.Error.Code)
}

// TestZTSearch_RequiresParams vérifie que l'endpoint /zt/search retourne
// un statut 400 lorsque les paramètres de recherche obligatoires sont absents.
func TestZTSearch_RequiresParams(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/zt/search", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp.Error)
}

// TestTVInfo_RequiresSeriesID vérifie que l'endpoint /getTVInfo retourne
// un statut 400 lorsque le paramètre seriesId est absent.
func TestTVInfo_RequiresSeriesID(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/getTVInfo", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTVSearch_RequiresQuery vérifie que l'endpoint /searchTV retourne
// un statut 400 lorsque le paramètre de requête de recherche est absent.
func TestTVSearch_RequiresQuery(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/searchTV", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
