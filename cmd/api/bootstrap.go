package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Now-Tiger/envhub/config"
	"github.com/Now-Tiger/envhub/internal/handlers"
	"github.com/Now-Tiger/envhub/internal/repository"
	"github.com/Now-Tiger/envhub/internal/service"
	"github.com/Now-Tiger/envhub/pkg/crypto"
	"github.com/Now-Tiger/envhub/pkg/database"
)

// AppComponents holds all initialized application components.
type AppComponents struct {
	Pool               *pgxpool.Pool
	Querier            *repository.Queries
	Config             *config.Config
	PlanHandler        *handlers.PlanHandler
	ProjectHandler     *handlers.ProjectHandler
	SecretHandler      *handlers.SecretHandler
	EnvironmentHandler *handlers.EnvironmentHandler
	TeamHandler        *handlers.TeamHandler
	AuthHandler        *handlers.AuthHandler
	CLIHandler         *handlers.CLIHandler
}

// InitializeBootstrap sets up the database connection, services, and handlers.
// Returns the initialized AppComponents ready for routing.
func InitializeBootstrap(ctx context.Context) (*AppComponents, error) {
	// Load database configuration
	dbConfig, err := database.LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	// Create connection pool
	log.Println("Connecting to database...")
	pool, err := database.NewPool(ctx, dbConfig)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Initialize master key before creating services
	masterKey, err := crypto.MasterKeyFromBase64(cfg.Crypto.MasterEncryptionKey)
	if err != nil {
		return nil, err
	}

	// Create repository querier
	querier := repository.New(pool)

	// Initialize services with master key (no re-initialization needed)
	projectSvc := service.NewProjectService(querier, masterKey)
	secretSvc := service.NewSecretService(querier, masterKey)
	environmentSvc := service.NewEnvironmentService(querier)
	teamSvc := service.NewTeamService(querier)
	planSvc := service.NewPlanService(querier, pool)
	organizationSvc := service.NewOrganizationService(querier)

	// Initialize handlers
	planHandler := handlers.NewPlanHandler(planSvc)
	projectHandler := handlers.NewProjectHandler(projectSvc, planSvc, organizationSvc)
	secretHandler := handlers.NewSecretHandler(secretSvc)
	environmentHandler := handlers.NewEnvironmentHandler(environmentSvc)
	teamHandler := handlers.NewTeamHandler(teamSvc, querier)
	authHandler := handlers.NewAuthHandler(querier, cfg.Auth)
	cliHandler := handlers.NewCLIHandler(secretSvc)

	return &AppComponents{
		Pool:               pool,
		Querier:            querier,
		Config:             cfg,
		PlanHandler:        planHandler,
		ProjectHandler:     projectHandler,
		SecretHandler:      secretHandler,
		EnvironmentHandler: environmentHandler,
		TeamHandler:        teamHandler,
		AuthHandler:        authHandler,
		CLIHandler:         cliHandler,
	}, nil
}
