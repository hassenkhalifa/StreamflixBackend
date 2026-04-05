package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"StreamflixBackend/internal/utils"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from panics and logs the stack trace.
// It returns a generic 500 error to the client without exposing internal details.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				slog.Error("panic recovered",
					slog.Any("error", r),
					slog.String("path", c.Request.URL.Path),
					slog.String("method", c.Request.Method),
					slog.String("stack", stack),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, utils.APIResponse{
					Data: nil,
					Error: &utils.ErrorDetail{
						Code:    "INTERNAL_ERROR",
						Message: "Une erreur interne est survenue",
					},
				})
			}
		}()
		c.Next()
	}
}
