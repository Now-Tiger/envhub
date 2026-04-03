package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthResponse represents the health check response structure.
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// DBHealthResponse represents the database health check response structure.
type DBHealthResponse struct {
	Status   string    `json:"status"`
	Database string    `json:"database"`
	Pool     PoolStats `json:"pool"`
	Message  string    `json:"message,omitempty"`
	Success  bool      `json:"success"`
}

// PoolStats represents database pool statistics.
type PoolStats struct {
	Total    int32 `json:"total"`
	Idle     int32 `json:"idle"`
	Acquired int32 `json:"acquired"`
}

// HealthCheck returns a simple health check handler.
func HealthCheck(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "ok",
			Service:   "envhub-api",
			Timestamp: time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode health check response: %v", err)
		}
	}
}

// DBHealthCheck checks database connectivity.
func DBHealthCheck(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if err := pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			response := DBHealthResponse{
				Status:   "error",
				Database: "disconnected",
				Pool:     PoolStats{},
				Message:  err.Error(),
				Success:  false,
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Printf("Failed to encode error response: %v", err)
			}
			return
		}

		stats := pool.Stat()
		w.WriteHeader(http.StatusOK)

		response := DBHealthResponse{
			Status:   "ok",
			Database: "connected",
			Pool: PoolStats{
				Total:    stats.TotalConns(),
				Idle:     stats.IdleConns(),
				Acquired: stats.AcquiredConns(),
			},
			Success: true,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode DB health response: %v", err)
		}
	}
}
