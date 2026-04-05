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

type visitor struct {
	tokens    float64
	lastSeen  time.Time
	maxTokens float64
	rate      float64 // tokens per second
}

// RateLimiter implements an in-memory per-IP token bucket rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64 // tokens per second
	burst    int     // max tokens
}

// NewRateLimiter creates a rate limiter allowing `requestsPerMinute` requests per minute.
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     float64(requestsPerMinute) / 60.0,
		burst:    requestsPerMinute,
	}

	// Cleanup expired visitors every 3 minutes
	go rl.cleanup()

	return rl
}

// Middleware returns a Gin middleware that enforces the rate limit.
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

		// Refill tokens based on elapsed time
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

// HealthCheckBypass allows health check and specific paths to bypass rate limiting.
func HealthCheckBypass() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Set("skip_rate_limit", true)
		}
		c.Next()
	}
}

// RateLimitWithBypass checks if the rate limit should be skipped.
func (rl *RateLimiter) RateLimitWithBypass() gin.HandlerFunc {
	mw := rl.Middleware()
	return func(c *gin.Context) {
		if skip, exists := c.Get("skip_rate_limit"); exists && skip.(bool) {
			c.Next()
			return
		}
		// Skip rate limit for health endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Skip rate limit for OPTIONS preflight
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		mw(c)
	}
}
