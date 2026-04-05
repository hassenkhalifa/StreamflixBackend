package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standardized JSON response format.
type APIResponse struct {
	Data  interface{}  `json:"data"`
	Error *ErrorDetail `json:"error"`
}

// ErrorDetail represents a structured error in the response.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespondSuccess sends a standardized success response.
func RespondSuccess(c *gin.Context, code int, data interface{}) {
	c.JSON(code, APIResponse{
		Data:  data,
		Error: nil,
	})
}

// RespondError sends a standardized error response.
func RespondError(c *gin.Context, httpCode int, errorCode string, message string) {
	c.JSON(httpCode, APIResponse{
		Data: nil,
		Error: &ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
	})
}

// BadRequest sends a 400 error response.
func BadRequest(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// NotFound sends a 404 error response.
func NotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, "NOT_FOUND", message)
}

// InternalError sends a 500 error response without exposing internal details.
func InternalError(c *gin.Context, _ error) {
	RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue")
}

// RateLimited sends a 429 error response.
func RateLimited(c *gin.Context) {
	RespondError(c, http.StatusTooManyRequests, "RATE_LIMITED", "Trop de requetes, veuillez reessayer plus tard")
}

// APIError is kept for backward compatibility during migration.
func APIError(c *gin.Context, code int, message string) {
	RespondError(c, code, "ERROR", message)
}
