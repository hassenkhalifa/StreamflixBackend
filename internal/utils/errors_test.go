// Package utils fournit les tests unitaires pour les fonctions utilitaires
// de gestion des réponses HTTP et des erreurs API.
//
// Les tests vérifient le format standardisé des réponses JSON (APIResponse),
// les différents helpers d'erreur (BadRequest, NotFound, InternalError, RateLimited),
// ainsi que la rétrocompatibilité avec l'ancien helper APIError.
// Un accent particulier est mis sur la sécurité : les détails internes des erreurs
// ne doivent jamais être exposés dans les réponses.
package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// init configure Gin en mode test pour supprimer les logs superflus
// durant l'exécution des tests.
func init() {
	gin.SetMode(gin.TestMode)
}

// performRequest est une fonction utilitaire qui exécute un handler Gin
// dans un contexte de test isolé et retourne le ResponseRecorder pour inspection.
// Elle crée un contexte Gin avec une requête GET factice sur /test.
func performRequest(handler gin.HandlerFunc) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	handler(c)
	return w
}

// TestRespondSuccess vérifie que RespondSuccess retourne le statut HTTP correct,
// une réponse JSON valide avec les données fournies et aucune erreur (Error: nil).
func TestRespondSuccess(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		RespondSuccess(c, http.StatusOK, gin.H{"message": "hello"})
	})

	assert.Equal(t, http.StatusOK, w.Code)

	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Data)
}

// TestRespondError vérifie que RespondError retourne le statut HTTP correct,
// une réponse JSON avec le code d'erreur et le message spécifiés, et aucune donnée (Data: nil).
func TestRespondError(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid input")
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Nil(t, resp.Data)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "BAD_REQUEST", resp.Error.Code)
	assert.Equal(t, "invalid input", resp.Error.Message)
}

// TestBadRequest vérifie que le helper BadRequest retourne un statut 400
// avec le code d'erreur BAD_REQUEST dans la réponse JSON.
func TestBadRequest(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		BadRequest(c, "field is required")
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", resp.Error.Code)
}

// TestNotFound vérifie que le helper NotFound retourne un statut 404
// avec le code d'erreur NOT_FOUND dans la réponse JSON.
func TestNotFound(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		NotFound(c, "movie not found")
	})

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", resp.Error.Code)
}

// TestInternalError vérifie que le helper InternalError retourne un statut 500
// avec le code INTERNAL_ERROR et, par mesure de sécurité, ne divulgue pas
// les détails de l'erreur interne (ex: "database") dans la réponse.
func TestInternalError(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		InternalError(c, errors.New("database connection failed"))
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", resp.Error.Code)
	// Should NOT expose internal error details
	assert.NotContains(t, resp.Error.Message, "database")
}

// TestRateLimited vérifie que le helper RateLimited retourne un statut 429
// avec le code d'erreur RATE_LIMITED dans la réponse JSON.
func TestRateLimited(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		RateLimited(c)
	})

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "RATE_LIMITED", resp.Error.Code)
}

// TestAPIError_BackwardCompatible vérifie la rétrocompatibilité de l'ancien
// helper APIError. Ce test confirme que la fonction retourne le bon statut HTTP,
// assurant que le code existant utilisant APIError continue de fonctionner.
func TestAPIError_BackwardCompatible(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		APIError(c, http.StatusForbidden, "forbidden")
	})

	assert.Equal(t, http.StatusForbidden, w.Code)
}
