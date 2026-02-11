package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Now-Tiger/envhub/internal/utils"
	"github.com/Now-Tiger/envhub/pkg/database"
)

func main() {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load database configuration
	dbConfig, err := database.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
		return
	}

	// Create connection pool
	log.Println("Connecting to database...")
	pool, err := database.NewPool(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to create database pool: %v", err)
		return
	}
	defer database.Close(pool)

	log.Println("✅ Database connection pool created successfully")

	// Log pool stats
	stats := database.Stats(pool)
	log.Printf("📊 Pool stats - Total: %d, Idle: %d, Acquired: %d",
		stats.TotalConns(),
		stats.IdleConns(),
		stats.AcquiredConns(),
	)

	// Initialize new router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// Routes
	r.Get("/health", healthCheckHandler(pool))
	r.Get("/health/db", dbHealthCheckHandler(pool))

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Gracefully shutdown the server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

// healthCheckHandler returns a simple health check
func healthCheckHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := fmt.Sprintf(
			`{"status":"ok","service":"envhub-api","timestamp":"%s"}`,
			time.Now().Format(time.RFC3339),
		)

		// Using blank identifier to explicitly ignore the error
		// from writing to the response body, satisfying errcheck.
		_, _ = fmt.Fprint(w, response)
	}
}

// dbHealthCheckHandler checks database connectivity
func dbHealthCheckHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(utils.ErrorResponse{
				Success:    false,
				StatusCode: 500,
				Message:    err.Error(),
			})
			return
		}

		stats := pool.Stat()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := fmt.Sprintf(
			`{"status":"ok","database":"connected","pool":{"total":%d,"idle":%d,"acquired":%d}}`,
			stats.TotalConns(),
			stats.IdleConns(),
			stats.AcquiredConns(),
		)

		_, _ = fmt.Fprint(w, response)
	}
}
