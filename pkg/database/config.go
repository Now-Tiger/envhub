package database

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// LoadConfigFromEnv loads database configuration from environment variables
func LoadConfigFromEnv() (Config, error) {
	cfg := Config{
		Host:     getEnv("DB_HOST", ""),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", ""),
		Password: getEnv("DB_PASSWORD", ""),
		Database: getEnv("DB_NAME", ""),
		SSLMode:  getEnv("DB_SSLMODE", ""),

		// Connection pool settings
		MaxConns:        int32(getEnvAsInt("DB_MAX_CONNS", 25)),
		MinConns:        int32(getEnvAsInt("DB_MIN_CONNS", 5)),
		MaxConnLifetime: getEnvAsDuration("DB_MAX_CONN_LIFETIME", "1h"),
		MaxConnIdleTime: getEnvAsDuration("DB_MAX_CONN_IDLE_TIME", "30m"),
	}

	// Validate required fields
	if cfg.User == "" {
		return cfg, fmt.Errorf("DB_USER is required")
	}
	if cfg.Password == "" {
		return cfg, fmt.Errorf("DB_PASSWORD is required")
	}
	if cfg.Database == "" {
		return cfg, fmt.Errorf("DB_NAME is required")
	}

	return cfg, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		valueStr = defaultValue
	}
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}
