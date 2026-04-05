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

func TestHealthEndpoint_HasSecurityHeaders(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

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

func TestLegacyRoutes_MoviesListExists(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/movieslist", nil)
	router.ServeHTTP(w, req)

	// Should return 200 with random movie list (no external API needed)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLegacyRoutes_CategoriesExists(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLegacyRoutes_ContentDetailsExists(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/contentDetails", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLegacyRoutes_UserListExists(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/list", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

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

func TestTVInfo_RequiresSeriesID(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/getTVInfo", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTVSearch_RequiresQuery(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/searchTV", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
