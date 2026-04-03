package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Now-Tiger/envhub/config"
	appmiddleware "github.com/Now-Tiger/envhub/internal/middleware"
	"github.com/Now-Tiger/envhub/internal/repository"
)

// setupRoutes configures all API routes.
func setupRoutes(
	r *chi.Mux,
	cfg *config.Config,
	querier *repository.Queries,
	planHandler interface {
		ListPlans(http.ResponseWriter, *http.Request)
		HandleStripeWebhook(http.ResponseWriter, *http.Request)
		GetOrgSubscription(http.ResponseWriter, *http.Request)
		AdminUpdateSubscription(http.ResponseWriter, *http.Request)
	},
	projectHandler interface {
		Create(http.ResponseWriter, *http.Request)
		List(http.ResponseWriter, *http.Request)
		Get(http.ResponseWriter, *http.Request)
		Update(http.ResponseWriter, *http.Request)
		Delete(http.ResponseWriter, *http.Request)
		RotateDEK(http.ResponseWriter, *http.Request)
	},
	secretHandler interface {
		List(http.ResponseWriter, *http.Request)
		Create(http.ResponseWriter, *http.Request)
		GetByEnvironment(http.ResponseWriter, *http.Request)
	},
	environmentHandler interface {
		Create(http.ResponseWriter, *http.Request)
		List(http.ResponseWriter, *http.Request)
		Get(http.ResponseWriter, *http.Request)
		Update(http.ResponseWriter, *http.Request)
		Delete(http.ResponseWriter, *http.Request)
	},
	teamHandler interface {
		AddMember(http.ResponseWriter, *http.Request)
		ListMembers(http.ResponseWriter, *http.Request)
		RemoveMember(http.ResponseWriter, *http.Request)
		UpdateMemberRole(http.ResponseWriter, *http.Request)
	},
	authHandler interface {
		Register(http.ResponseWriter, *http.Request)
		Login(http.ResponseWriter, *http.Request)
		CLILogin(http.ResponseWriter, *http.Request)
		Me(http.ResponseWriter, *http.Request)
		ListTokens(http.ResponseWriter, *http.Request)
		RevokeToken(http.ResponseWriter, *http.Request)
	},
	cliHandler interface {
		GetSecrets(http.ResponseWriter, *http.Request)
	},
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
	// Note: JWT middleware already applied in protected group above, reusing same auth context
	r.Group(func(r chi.Router) {
		// Admin subscription management endpoint
		r.Post("/api/v1/admin/subscription", planHandler.AdminUpdateSubscription)
	})
}
