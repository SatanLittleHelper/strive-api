package main

import (
	"log"
	"net/http"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/database"
	httphandler "github.com/aleksandr/strive-api/internal/http"
	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/migrate"
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

	mux := http.NewServeMux()
	mux.HandleFunc("/health", httphandler.HealthHandler)

	handler := httphandler.LoggingMiddleware(logger)(httphandler.RequestIDMiddleware()(mux))

	server := httphandler.NewServer(cfg, handler, logger)
	server.Start()
	server.WaitForShutdown()
}
