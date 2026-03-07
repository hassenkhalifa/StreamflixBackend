package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"status": false,
		"error":  message,
	})
}

func InternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"status": false,
		"error":  err.Error(),
	})
}
