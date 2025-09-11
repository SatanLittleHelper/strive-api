package main

import (
	"log"
	"net/http"

	"github.com/aleksandr/strive-api/internal/config"
	httphandler "github.com/aleksandr/strive-api/internal/http"
	"github.com/aleksandr/strive-api/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	logger.Info("Application starting", "config", cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", httphandler.HealthHandler)

	handler := httphandler.LoggingMiddleware(logger)(httphandler.RequestIDMiddleware()(mux))

	server := httphandler.NewServer(cfg, handler, logger)
	server.Start()
	server.WaitForShutdown()
}
