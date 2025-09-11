package main

import (
	"log"
	"net/http"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/database"
	httphandler "github.com/aleksandr/strive-api/internal/http"
	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/migrate"
	"github.com/aleksandr/strive-api/internal/repositories"
	"github.com/aleksandr/strive-api/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	logger.Info("Application starting", "config", cfg)

	if err := migrate.Run(cfg, logger); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	db, err := database.New(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories and services
	userRepo := repositories.NewUserRepository(db.Pool())
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret)

	// Initialize handlers
	authHandlers := httphandler.NewAuthHandlers(authService, logger)

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/health", httphandler.HealthHandler)
	mux.HandleFunc("/api/v1/auth/register", authHandlers.Register)
	mux.HandleFunc("/api/v1/auth/login", authHandlers.Login)

	// Protected routes (example)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/api/v1/user/profile", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"This is a protected endpoint"}`))
	})

	// Apply middleware to protected routes
	protectedHandler := httphandler.LoggingMiddleware(logger)(
		httphandler.RequestIDMiddleware()(
			httphandler.AuthMiddleware(authService)(protectedMux),
		),
	)

	// Combine public and protected routes
	mux.Handle("/api/v1/user/", http.StripPrefix("/api/v1/user", protectedHandler))

	// Apply middleware to main mux
	handler := httphandler.LoggingMiddleware(logger)(httphandler.RequestIDMiddleware()(mux))

	server := httphandler.NewServer(cfg, handler, logger)
	server.Start()
	server.WaitForShutdown()
}
