package main

import (
	"log"
	"net/http"

	_ "github.com/aleksandr/strive-api/docs"
	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/database"
	httphandler "github.com/aleksandr/strive-api/internal/http"
	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/migrate"
	"github.com/aleksandr/strive-api/internal/repositories"
	"github.com/aleksandr/strive-api/internal/services"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Strive API
// @version 1.0
// @description API for workout diary with user authentication
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg := loadConfig()
	logger := setupLogger(cfg)
	db := setupDatabase(cfg, logger)
	defer db.Close()

	runMigrations(cfg, logger)

	// Initialize services and handlers
	authService := setupServices(db, cfg)
	handlers := setupHandlers(authService, logger, db)

	// Setup routes and middleware
	handler := setupRoutes(handlers, logger, authService)

	// Start server
	server := httphandler.NewServer(cfg, handler, logger)
	server.Start()
	server.WaitForShutdown()
}

func loadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return cfg
}

func setupLogger(cfg *config.Config) *logger.Logger {
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	logger.Info("Application starting", "config", cfg)
	return logger
}

func setupDatabase(cfg *config.Config, logger *logger.Logger) *database.Database {
	db, err := database.New(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

func runMigrations(cfg *config.Config, logger *logger.Logger) {
	if err := migrate.Run(cfg, logger); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}

func setupServices(db *database.Database, cfg *config.Config) services.AuthService {
	userRepo := repositories.NewUserRepository(db.Pool())
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret)
	return authService
}

type Handlers struct {
	Auth   *httphandler.AuthHandlers
	Health *httphandler.DetailedHealthHandler
}

func setupHandlers(authService services.AuthService, logger *logger.Logger, db *database.Database) *Handlers {
	return &Handlers{
		Auth:   httphandler.NewAuthHandlers(authService, logger),
		Health: httphandler.NewDetailedHealthHandler(logger, db.Pool()),
	}
}

func setupRoutes(handlers *Handlers, logger *logger.Logger, authService services.AuthService) http.Handler {
	mux := http.NewServeMux()

	// Setup public routes
	setupPublicRoutes(mux, handlers)

	// Setup protected routes
	setupProtectedRoutes(mux, authService, logger)

	// Apply middleware
	return applyMiddleware(mux, logger)
}

func setupPublicRoutes(mux *http.ServeMux, handlers *Handlers) {
	// Health endpoints
	mux.HandleFunc("/health", handlers.Health.Health)
	mux.HandleFunc("/health/db", handlers.Health.DatabaseHealth)
	mux.HandleFunc("/health/detailed", handlers.Health.DetailedHealth)

	// Auth endpoints
	mux.HandleFunc("/api/v1/auth/register", handlers.Auth.Register)
	mux.HandleFunc("/api/v1/auth/login", handlers.Auth.Login)
	mux.HandleFunc("/api/v1/auth/refresh", handlers.Auth.Refresh)

	// Documentation
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
}

func setupProtectedRoutes(mux *http.ServeMux, authService services.AuthService, logger *logger.Logger) {
	// Create protected sub-router
	protectedMux := http.NewServeMux()

	// Protected endpoints
	protectedMux.HandleFunc("/api/v1/user/profile", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"This is a protected endpoint"}`))
	})

	// Apply auth middleware to protected routes
	protectedHandler := httphandler.LoggingMiddleware(logger)(
		httphandler.RequestIDMiddleware()(
			httphandler.AuthMiddleware(authService)(protectedMux),
		),
	)

	// Mount protected routes
	mux.Handle("/api/v1/user/", http.StripPrefix("/api/v1/user", protectedHandler))
}

func applyMiddleware(mux *http.ServeMux, logger *logger.Logger) http.Handler {
	corsMiddleware := httphandler.NewCORSMiddleware()

	return corsMiddleware(
		httphandler.LoggingMiddleware(logger)(
			httphandler.RequestIDMiddleware()(mux),
		),
	)
}
