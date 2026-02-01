package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port            string
	RealDebridToken string
	Environment     string
}

func Load() *Config {
	port := getEnv("PORT", "2000")
	rdToken := getEnv("REALDEBRID_TOKEN", "")
	env := getEnv("ENVIRONMENT", "development")

	if rdToken == "" {
		log.Println("WARNING: REALDEBRID_TOKEN not set")
	}

	return &Config{
		Port:            port,
		RealDebridToken: rdToken,
		Environment:     env,
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
