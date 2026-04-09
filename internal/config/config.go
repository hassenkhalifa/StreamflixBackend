// Package config gère le chargement et la validation de la configuration
// de l'application StreamFlix à partir des variables d'environnement.
//
// La configuration est chargée depuis un fichier .env.dev (via godotenv)
// puis complétée par les variables d'environnement système. Chaque variable
// possède une valeur par défaut sensée pour le développement local.
//
// Variables d'environnement supportées :
//   - PORT : port d'écoute du serveur (défaut : "2000")
//   - GIN_MODE : mode Gin debug/release/test (défaut : "debug")
//   - ENVIRONMENT : environnement d'exécution (défaut : "development")
//   - REALDEBRID_TOKEN : token API Real-Debrid pour le débridage
//   - TMDB_TOKEN : token Bearer API TMDB pour les métadonnées de films
//   - CORS_ORIGINS : origines CORS autorisées séparées par des virgules
//   - RATE_LIMIT_PER_MINUTE : nombre max de requêtes par minute par IP
//   - LOG_LEVEL : niveau de log (debug/info/warn/error)
//   - USER_AGENT : User-Agent pour les requêtes HTTP sortantes
//   - HTTP_TIMEOUT_SECONDS : timeout des requêtes HTTP sortantes en secondes
//   - CACHE_TTL_MINUTES : durée de vie du cache en minutes
package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config contient toute la configuration de l'application chargée depuis
// les variables d'environnement.
//
// Chaque champ correspond à une variable d'environnement et possède une
// valeur par défaut pour le développement local. En production, les tokens
// API doivent être configurés via les variables d'environnement.
type Config struct {
	Port            string        // Port d'écoute du serveur HTTP (défaut : "2000")
	GinMode         string        // Mode Gin : "debug", "release" ou "test"
	Environment     string        // Environnement : "development" ou "production"
	RealDebridToken string        // Token API Real-Debrid pour le débridage de liens
	TMDBToken       string        // Token Bearer API TMDB pour les métadonnées de films/séries
	CORSOrigins     []string      // Liste des origines CORS autorisées
	RateLimitPerMin int           // Nombre maximum de requêtes par minute par IP
	LogLevel        slog.Level    // Niveau de logging (debug, info, warn, error)
	UserAgent       string        // User-Agent utilisé pour les requêtes HTTP sortantes
	HTTPTimeout     time.Duration // Timeout global pour les requêtes HTTP sortantes
	CacheTTL        time.Duration // Durée de vie par défaut des entrées en cache
}

// Load charge la configuration depuis le fichier .env.dev et les variables d'environnement.
//
// L'ordre de priorité est :
//  1. Variables d'environnement système (prioritaire)
//  2. Fichier .env.dev (si présent)
//  3. Valeurs par défaut
//
// Retourne un pointeur vers Config entièrement initialisé.
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

// IsProduction retourne true si l'environnement est configuré en production.
// Utilisé pour activer/désactiver certains comportements (ex: stack traces, mode debug).
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// Validate vérifie que les configurations obligatoires sont présentes et
// émet des avertissements via slog pour les tokens manquants ou non configurés.
//
// Les tokens vérifiés sont :
//   - REALDEBRID_TOKEN : nécessaire pour le débridage de liens
//   - TMDB_TOKEN : nécessaire pour accéder à l'API TMDB
func (c *Config) Validate() {
	if c.RealDebridToken == "" || c.RealDebridToken == "change_me" {
		slog.Warn("REALDEBRID_TOKEN is not configured")
	}
	if c.TMDBToken == "" || c.TMDBToken == "change_me" {
		slog.Warn("TMDB_TOKEN is not configured")
	}
}

// getEnv retourne la valeur de la variable d'environnement key,
// ou defaultValue si elle n'est pas définie ou vide.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retourne la valeur entière de la variable d'environnement key,
// ou defaultValue si elle n'est pas définie ou si la conversion échoue.
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// parseCORSOrigins découpe une chaîne d'origines CORS séparées par des virgules
// en une slice de strings nettoyées (espaces supprimés, chaînes vides ignorées).
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

// parseLogLevel convertit une chaîne de niveau de log en slog.Level.
// Les valeurs acceptées sont : "debug", "info", "warn"/"warning", "error".
// Toute autre valeur retourne slog.LevelInfo par défaut.
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
