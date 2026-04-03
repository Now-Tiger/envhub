package config

import (
	"fmt"
	"time"

	"github.com/Now-Tiger/envhub/scripts"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Crypto   CryptoConfig
}

type ServerConfig struct {
	Port         string
	Env          string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type AuthConfig struct {
	SupabaseURL   string
	SupabaseKey   string
	JWTSecret     string
	JWTExpiry     time.Duration
	RefreshExpiry time.Duration
}

type CryptoConfig struct {
	MasterEncryptionKey string
}

func Load() (*Config, error) {
	cfg := &Config{

		Server: ServerConfig{
			Port:         scripts.GetEnv("PORT", "8080"),
			Env:          scripts.GetEnv("ENV", ""),
			ReadTimeout:  scripts.GetEnvAsDuration("SERVER_READ_TIMEOUT", "15s"),
			WriteTimeout: scripts.GetEnvAsDuration("SERVER_WRITE_TIMEOUT", "15s"),
			IdleTimeout:  scripts.GetEnvAsDuration("SERVER_IDLE_TIMEOUT", "60s"),
		},

		Database: DatabaseConfig{
			Host:            scripts.GetEnv("DB_HOST", ""),
			Port:            scripts.GetEnvAsInt("DB_PORT", 5432),
			User:            scripts.GetEnv("DB_USER", ""),
			Password:        scripts.GetEnv("DB_PASSWORD", ""),
			Name:            scripts.GetEnv("DB_NAME", ""),
			SSLMode:         scripts.GetEnv("DB_SSLMODE", ""),
			MaxConns:        int32(scripts.GetEnvAsInt("DB_MAX_CONNS", 25)),
			MinConns:        int32(scripts.GetEnvAsInt("DB_MIN_CONNS", 5)),
			MaxConnLifetime: scripts.GetEnvAsDuration("DB_MAX_CONN_LIFETIME", "1h"),
			MaxConnIdleTime: scripts.GetEnvAsDuration("DB_MAX_CONN_IDLE_TIME", "30m"),
		},

		Auth: AuthConfig{
			SupabaseURL:   scripts.GetEnv("SUPABASE_URL", ""),
			SupabaseKey:   scripts.GetEnv("SUPABASE_KEY", ""),
			JWTSecret:     scripts.GetEnv("JWT_SECRET", ""),
			JWTExpiry:     scripts.GetEnvAsDuration("JWT_EXPIRY", "24h"),
			RefreshExpiry: scripts.GetEnvAsDuration("JWT_REFRESH_EXPIRY", "168h"),
		},

		Crypto: CryptoConfig{
			MasterEncryptionKey: scripts.GetEnv("MASTER_ENCRYPTION_KEY", ""),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.Crypto.MasterEncryptionKey == "" {
		return fmt.Errorf("MASTER_ENCRYPTION_KEY is required")
	}
	return nil
}
