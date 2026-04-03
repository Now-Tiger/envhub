package database

import (
	"fmt"

	env "github.com/Now-Tiger/envhub/scripts"
)

// LoadConfigFromEnv loads database configuration from environment variables
func LoadConfigFromEnv() (Config, error) {
	cfg := Config{
		Host:     env.GetEnv("DB_HOST", ""),
		Port:     env.GetEnvAsInt("DB_PORT", 5432),
		User:     env.GetEnv("DB_USER", ""),
		Password: env.GetEnv("DB_PASSWORD", ""),
		Database: env.GetEnv("DB_NAME", ""),
		SSLMode:  env.GetEnv("DB_SSLMODE", ""),

		// Connection pool settings
		MaxConns:        int32(env.GetEnvAsInt("DB_MAX_CONNS", 25)),
		MinConns:        int32(env.GetEnvAsInt("DB_MIN_CONNS", 5)),
		MaxConnLifetime: env.GetEnvAsDuration("DB_MAX_CONN_LIFETIME", "1h"),
		MaxConnIdleTime: env.GetEnvAsDuration("DB_MAX_CONN_IDLE_TIME", "30m"),
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
