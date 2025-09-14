package main

import (
	"flag"
	"log"
	"os"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/migrate"
)

func main() {
	var direction string
	flag.StringVar(&direction, "direction", "up", "Migration direction: up or down")
	flag.Parse()

	if direction != "up" && direction != "down" {
		log.Fatalf("Invalid direction: %s. Use 'up' or 'down'", direction)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	logger.Info("Starting database migrations", "direction", direction)

	switch direction {
	case "up":
		if err := migrate.Run(cfg, logger); err != nil {
			log.Fatalf("Failed to run up migrations: %v", err)
		}
		logger.Info("Up migrations completed successfully")
	case "down":
		if err := migrate.Down(cfg, logger); err != nil {
			log.Fatalf("Failed to run down migrations: %v", err)
		}
		logger.Info("Down migrations completed successfully")
	}

	os.Exit(0)
}
