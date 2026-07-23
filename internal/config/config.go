package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration loaded from environment variables
type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	BaseURL       string
}

// Load reads configuration from environment variables with sensible defaults
// for local development. In production, always set these via actual env vars.
func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
	}
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt returns an environment variable as an integer or a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}