package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port            string
	GinMode         string
	Environment     string
	RealDebridToken string
	TMDBToken       string
	CORSOrigins     []string
	RateLimitPerMin int
	LogLevel        slog.Level
	UserAgent       string
	HTTPTimeout     time.Duration
	CacheTTL        time.Duration
}

// Load reads the .env.dev file and returns a populated Config.
func Load() *Config {
	_ = godotenv.Load(".env.dev")

	cfg := &Config{
		Port:            getEnv("PORT", "2000"),
		GinMode:         getEnv("GIN_MODE", "debug"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		RealDebridToken: getEnv("REALDEBRID_TOKEN", ""),
		TMDBToken:       getEnv("TMDB_TOKEN", ""),
		CORSOrigins:     parseCORSOrigins(getEnv("CORS_ORIGINS", "http://localhost:3000")),
		RateLimitPerMin: getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
		LogLevel:        parseLogLevel(getEnv("LOG_LEVEL", "info")),
		UserAgent:       getEnv("USER_AGENT", "StreamFlix/1.0"),
		HTTPTimeout:     time.Duration(getEnvAsInt("HTTP_TIMEOUT_SECONDS", 10)) * time.Second,
		CacheTTL:        time.Duration(getEnvAsInt("CACHE_TTL_MINUTES", 5)) * time.Minute,
	}

	return cfg
}

// IsProduction returns true if the environment is production.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// Validate logs warnings for missing mandatory configuration.
func (c *Config) Validate() {
	if c.RealDebridToken == "" || c.RealDebridToken == "change_me" {
		slog.Warn("REALDEBRID_TOKEN is not configured")
	}
	if c.TMDBToken == "" || c.TMDBToken == "change_me" {
		slog.Warn("TMDB_TOKEN is not configured")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func parseCORSOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
