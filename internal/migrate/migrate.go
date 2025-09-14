package migrate

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run(cfg *config.Config, log *logger.Logger) error {
	m, cleanup, err := createMigrator(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	log.Info("Running database migrations up")

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run up migrations: %w", err)
	}

	return logMigrationStatus(m, log, "up")
}

func Down(cfg *config.Config, log *logger.Logger) error {
	m, cleanup, err := createMigrator(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	log.Info("Running database migrations down")

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run down migrations: %w", err)
	}

	return logMigrationStatus(m, log, "down")
}

func createMigrator(cfg *config.Config) (*migrate.Migrate, func(), error) {
	db, err := sql.Open("pgx", cfg.DatabaseURL())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	migrationsPath := filepath.Join(wd, "migrations")

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		db.Close()
		return nil, nil, fmt.Errorf("migrations directory not found: %s", migrationsPath)
	}

	sourceURL := fmt.Sprintf("file://%s", migrationsPath)

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	cleanup := func() {
		m.Close()
		db.Close()
	}

	return m, cleanup, nil
}

func logMigrationStatus(m *migrate.Migrate, log *logger.Logger, direction string) error {
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	log.Info("Database migrations completed successfully",
		"direction", direction,
		"version", version,
		"dirty", dirty,
	)

	return nil
}
