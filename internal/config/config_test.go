package config

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear env vars to test defaults
	envVars := []string{"PORT", "GIN_MODE", "ENVIRONMENT", "REALDEBRID_TOKEN", "TMDB_TOKEN",
		"CORS_ORIGINS", "RATE_LIMIT_PER_MINUTE", "LOG_LEVEL", "USER_AGENT",
		"HTTP_TIMEOUT_SECONDS", "CACHE_TTL_MINUTES"}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	cfg := Load()

	assert.Equal(t, "2000", cfg.Port)
	assert.Equal(t, "debug", cfg.GinMode)
	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, "", cfg.RealDebridToken)
	assert.Equal(t, "", cfg.TMDBToken)
	assert.Equal(t, []string{"http://localhost:3000"}, cfg.CORSOrigins)
	assert.Equal(t, 60, cfg.RateLimitPerMin)
	assert.Equal(t, slog.LevelInfo, cfg.LogLevel)
	assert.Equal(t, "StreamFlix/1.0", cfg.UserAgent)
	assert.Equal(t, 10*time.Second, cfg.HTTPTimeout)
	assert.Equal(t, 5*time.Minute, cfg.CacheTTL)
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("PORT", "8080")
	os.Setenv("GIN_MODE", "release")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("REALDEBRID_TOKEN", "test_token")
	os.Setenv("TMDB_TOKEN", "tmdb_test")
	os.Setenv("CORS_ORIGINS", "https://example.com,https://other.com")
	os.Setenv("RATE_LIMIT_PER_MINUTE", "120")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("USER_AGENT", "TestAgent/1.0")
	os.Setenv("HTTP_TIMEOUT_SECONDS", "20")
	os.Setenv("CACHE_TTL_MINUTES", "10")
	defer func() {
		envVars := []string{"PORT", "GIN_MODE", "ENVIRONMENT", "REALDEBRID_TOKEN", "TMDB_TOKEN",
			"CORS_ORIGINS", "RATE_LIMIT_PER_MINUTE", "LOG_LEVEL", "USER_AGENT",
			"HTTP_TIMEOUT_SECONDS", "CACHE_TTL_MINUTES"}
		for _, v := range envVars {
			os.Unsetenv(v)
		}
	}()

	cfg := Load()

	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "release", cfg.GinMode)
	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, "test_token", cfg.RealDebridToken)
	assert.Equal(t, "tmdb_test", cfg.TMDBToken)
	assert.Equal(t, []string{"https://example.com", "https://other.com"}, cfg.CORSOrigins)
	assert.Equal(t, 120, cfg.RateLimitPerMin)
	assert.Equal(t, slog.LevelError, cfg.LogLevel)
	assert.Equal(t, "TestAgent/1.0", cfg.UserAgent)
	assert.Equal(t, 20*time.Second, cfg.HTTPTimeout)
	assert.Equal(t, 10*time.Minute, cfg.CacheTTL)
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected bool
	}{
		{"production", "production", true},
		{"development", "development", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.env}
			assert.Equal(t, tt.expected, cfg.IsProduction())
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseLogLevel(tt.input))
		})
	}
}

func TestParseCORSOrigins(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single origin", "http://localhost:3000", []string{"http://localhost:3000"}},
		{"multiple origins", "http://a.com,http://b.com", []string{"http://a.com", "http://b.com"}},
		{"with spaces", " http://a.com , http://b.com ", []string{"http://a.com", "http://b.com"}},
		{"empty string", "", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCORSOrigins(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
