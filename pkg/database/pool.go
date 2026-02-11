package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds database configuration
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// NewPool creates a new connection pool with optimal settings
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	var poolConfig *pgxpool.Config
	var err error

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		poolConfig, err = pgxpool.ParseConfig(dbURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
		}
	} else {
		// Fallback: Build connection string from individual config fields
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host,
			cfg.Port,
			cfg.User,
			cfg.Password,
			cfg.Database,
			cfg.SSLMode,
		)
		poolConfig, err = pgxpool.ParseConfig(dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pool config: %w", err)
		}
	}

	// Configure connection pool settings
	// These settings are optimized for a microservice with moderate load

	// MaxConns: Maximum number of connections in the pool
	// For a single API instance, 25 is a good starting point
	// Scale this based on: (core_count * 2) + effective_spindle_count
	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	} else {
		poolConfig.MaxConns = 25
	}

	// MinConns: Minimum number of connections to maintain
	// Keeps connections warm for fast response times
	if cfg.MinConns > 0 {
		poolConfig.MinConns = cfg.MinConns
	} else {
		poolConfig.MinConns = 5
	}

	// MaxConnLifetime: Maximum time a connection can be reused
	// Prevents stale connections and helps with load balancer rotation
	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	} else {
		poolConfig.MaxConnLifetime = 1 * time.Hour
	}

	// MaxConnIdleTime: Maximum time a connection can be idle
	// Closes idle connections to free up resources
	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	} else {
		poolConfig.MaxConnIdleTime = 30 * time.Minute
	}

	// HealthCheckPeriod: How often to check connection health
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// ConnectTimeout: Maximum time to wait for a connection
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second

	// Create the pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// FIX: Added retry logic for the initial connection ping.
	// In Docker environments, DNS resolution for service names (like 'postgres')
	// can sometimes lag slightly behind service startup, causing 'no such host' errors.
	const maxRetries = 10
	var pingErr error

	for i := 0; i < maxRetries; i++ {
		pingErr = pool.Ping(ctx)
		if pingErr == nil {
			// Connection successful
			return pool, nil
		}

		// Log the retry attempt (optional, using fmt here for simplicity)
		// fmt.Printf("Failed to ping database (attempt %d/%d): %v. Retrying...\n", i+1, maxRetries, pingErr)

		// Wait before retrying, increasing delay slightly
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("context canceled during connection retry: %w", ctx.Err())
		case <-time.After(time.Duration(i+1) * time.Second):
			// continue loop
		}
	}

	// If all retries fail, close pool and return error
	pool.Close()
	return nil, fmt.Errorf("failed to ping database after %d attempts: %w", maxRetries, pingErr)
}

// Stats returns current pool statistics
func Stats(pool *pgxpool.Pool) *pgxpool.Stat {
	return pool.Stat()
}

// Close gracefully closes the connection pool
func Close(pool *pgxpool.Pool) {
	pool.Close()
}
