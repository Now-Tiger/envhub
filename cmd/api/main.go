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
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Now-Tiger/envhub/config"
	"github.com/Now-Tiger/envhub/internal/handlers"
	appmiddleware "github.com/Now-Tiger/envhub/internal/middleware"
	"github.com/Now-Tiger/envhub/internal/repository"
	"github.com/Now-Tiger/envhub/internal/service"
	"github.com/Now-Tiger/envhub/internal/utils"
	"github.com/Now-Tiger/envhub/pkg/crypto"
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

	// Load app config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
		return
	}

	// Create repository querier
	querier := repository.New(pool)

	// Initialize master key
	masterKey, err := crypto.MasterKeyFromBase64(cfg.Crypto.MasterEncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize master key: %v", err)
		return
	}

	// Initialize services
	projectSvc := service.NewProjectService(querier, masterKey)
	secretSvc := service.NewSecretService(querier, masterKey)
	environmentSvc := service.NewEnvironmentService(querier)
	teamSvc := service.NewTeamService(querier)
	planSvc := service.NewPlanService(querier, pool)
	organizationSvc := service.NewOrganizationService(querier)

	// Initialize handlers
	projectHandler := handlers.NewProjectHandler(projectSvc, planSvc, organizationSvc)
	secretHandler := handlers.NewSecretHandler(secretSvc)
	environmentHandler := handlers.NewEnvironmentHandler(environmentSvc)
	teamHandler := handlers.NewTeamHandler(teamSvc, querier)
	planHandler := handlers.NewPlanHandler(planSvc)
	authHandler := handlers.NewAuthHandler(querier, cfg.Auth)
	cliHandler := handlers.NewCLIHandler(secretSvc)

	// Initialize new router
	r := chi.NewRouter()

	// Chi middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS - strict policy
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting - 100 requests per minute per IP
	r.Use(middleware.Throttle(100))

	// Health routes
	r.Get("/health", healthCheckHandler(pool))
	r.Get("/health/db", dbHealthCheckHandler(pool))

	// Setup API routes
	setupRoutes(r, cfg, querier, planHandler, projectHandler, secretHandler, environmentHandler, teamHandler, authHandler, cliHandler)

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

// setupRoutes configures all API routes
func setupRoutes(
	r *chi.Mux,
	cfg *config.Config,
	querier *repository.Queries,
	planHandler *handlers.PlanHandler,
	projectHandler *handlers.ProjectHandler,
	secretHandler *handlers.SecretHandler,
	environmentHandler *handlers.EnvironmentHandler,
	teamHandler *handlers.TeamHandler,
	authHandler *handlers.AuthHandler,
	cliHandler *handlers.CLIHandler,
) {
	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		// Auth endpoints
		r.Post("/api/v1/auth/register", authHandler.Register)
		r.Post("/api/v1/auth/login", authHandler.Login)
		r.Post("/api/v1/auth/cli-login", authHandler.CLILogin)

		// Public plans endpoint - anyone can view available plans
		r.Get("/api/v1/plans", planHandler.ListPlans)

		// Stripe webhook endpoint (public - verified by Stripe signature in production)
		r.Post("/api/v1/webhooks/stripe", planHandler.HandleStripeWebhook)
	})

	// Protected routes (JWT or API token auth)
	r.Group(func(r chi.Router) {
		// Auth middleware
		r.Use(appmiddleware.JWTAuthMiddleware(cfg.Auth))
		r.Use(appmiddleware.APITokenMiddleware(querier))
		r.Use(appmiddleware.UserRateLimitMiddleware())

		// Auth endpoints
		r.Get("/api/v1/auth/me", authHandler.Me)

		// API token management
		r.Route("/api/v1/auth/tokens", func(r chi.Router) {
			r.Get("/", authHandler.ListTokens)
			r.Delete("/{id}", authHandler.RevokeToken)
		})

		// Project routes
		r.Route("/api/v1/projects", func(r chi.Router) {
			r.Post("/", projectHandler.Create)
			r.Get("/", projectHandler.List)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", projectHandler.Get)
				r.Patch("/", projectHandler.Update)
				r.Delete("/", projectHandler.Delete)
				r.Post("/rotate-dek", projectHandler.RotateDEK)

				// Secrets routes
				r.Get("/secrets", secretHandler.List)
				r.Post("/secrets", secretHandler.Create)
				r.Get("/secrets/{envName}", secretHandler.GetByEnvironment)

				// Environment routes
				r.Route("/environments", func(r chi.Router) {
					r.Post("/", environmentHandler.Create)
					r.Get("/", environmentHandler.List)

					r.Route("/{envId}", func(r chi.Router) {
						r.Get("/", environmentHandler.Get)
						r.Patch("/", environmentHandler.Update)
						r.Delete("/", environmentHandler.Delete)
					})
				})

				// Team routes
				r.Route("/members", func(r chi.Router) {
					r.Post("/", teamHandler.AddMember)
					r.Get("/", teamHandler.ListMembers)

					r.Route("/{userId}", func(r chi.Router) {
						r.Delete("/", teamHandler.RemoveMember)
						r.Patch("/", teamHandler.UpdateMemberRole)
					})
				})
			})
		})

		// Organization subscription endpoint
		r.Get("/api/v1/organizations/{id}/subscription", planHandler.GetOrgSubscription)

		// CLI routes - optimized endpoints for CLI
		r.Route("/api/v1/cli", func(r chi.Router) {
			r.Get("/secrets/{project}/{env}", cliHandler.GetSecrets)
		})
	})

	// Admin routes (protected - owner only)
	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.JWTAuthMiddleware(cfg.Auth))

		// Admin subscription management endpoint
		r.Post("/api/v1/admin/subscription", planHandler.AdminUpdateSubscription)
	})
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
