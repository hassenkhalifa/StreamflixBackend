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

func init() {
	gin.SetMode(gin.TestMode)
}

func performRequest(handler gin.HandlerFunc) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	handler(c)
	return w
}

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

func TestAPIError_BackwardCompatible(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		APIError(c, http.StatusForbidden, "forbidden")
	})

	assert.Equal(t, http.StatusForbidden, w.Code)
}
