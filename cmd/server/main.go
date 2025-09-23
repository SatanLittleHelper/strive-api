package main

import (
	"log"
	"net/http"
	"os"

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

// @host strive-api-zjtl.onrender.com
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
	services := setupServices(db, cfg)
	handlers := setupHandlers(services, logger, db, cfg)

	// Setup routes and middleware
	handler := setupRoutes(handlers, logger, services.Auth, cfg)

	// Start server
	server := httphandler.NewServer(cfg, handler, logger)

	// Check if SSL certificates exist for HTTPS (only for localhost development)
	if fileExists("certs/localhost-cert.pem") && fileExists("certs/localhost-key.pem") {
		server.StartTLS("certs/localhost-cert.pem", "certs/localhost-key.pem")
		logger.Info("Starting with HTTPS (localhost development certificates found)")
	} else {
		server.Start()
		logger.Info("Starting with HTTP (production mode - SSL handled by platform)")
	}

	server.WaitForShutdown()
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
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

type Services struct {
	Auth    services.AuthService
	User    services.UserService
	Calorie services.CalorieService
}

func setupServices(db *database.Database, cfg *config.Config) *Services {
	userRepo := repositories.NewUserRepository(db.Pool())
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db.Pool())
	calorieRepo := repositories.NewCalorieRepository(db.Pool())

	authService := services.NewAuthService(userRepo, refreshTokenRepo, &cfg.JWT)
	userService := services.NewUserService(userRepo)
	calorieService := services.NewCalorieService(calorieRepo)

	return &Services{
		Auth:    authService,
		User:    userService,
		Calorie: calorieService,
	}
}

type Handlers struct {
	Auth    *httphandler.AuthHandlers
	User    *httphandler.UserHandlers
	Calorie *httphandler.CalorieHandlers
	Health  *httphandler.DetailedHealthHandler
}

func setupHandlers(services *Services, logger *logger.Logger, db *database.Database, cfg *config.Config) *Handlers {
	return &Handlers{
		Auth:    httphandler.NewAuthHandlers(services.Auth, logger, cfg),
		User:    httphandler.NewUserHandlers(services.User, logger),
		Calorie: httphandler.NewCalorieHandlers(services.Calorie, logger),
		Health:  httphandler.NewDetailedHealthHandler(logger, db.Pool()),
	}
}

func setupRoutes(handlers *Handlers, logger *logger.Logger, authService services.AuthService, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// Setup public routes
	setupPublicRoutes(mux, handlers)

	// Setup protected routes
	setupProtectedRoutes(mux, authService, logger, handlers)

	// Apply middleware
	return applyMiddleware(mux, logger, cfg)
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
	mux.HandleFunc("/api/v1/auth/logout", handlers.Auth.Logout)

	// Documentation
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
}

func setupProtectedRoutes(mux *http.ServeMux, authService services.AuthService, logger *logger.Logger, handlers *Handlers) {
	// User protected routes
	userProtectedMux := http.NewServeMux()
	userProtectedMux.HandleFunc("/me", handlers.User.Me)
	userProtectedMux.HandleFunc("/theme", handlers.User.UpdateTheme)
	userProtectedHandler := httphandler.AuthMiddleware(authService, logger)(userProtectedMux)
	mux.Handle("/api/v1/user/", http.StripPrefix("/api/v1/user", userProtectedHandler))

	// Calorie calculator protected routes
	calorieProtectedMux := http.NewServeMux()
	calorieProtectedMux.HandleFunc("/calculate", handlers.Calorie.CalculateCalories)
	calorieProtectedMux.HandleFunc("/last", handlers.Calorie.GetLastCalculation)
	calorieProtectedHandler := httphandler.AuthMiddleware(authService, logger)(calorieProtectedMux)
	mux.Handle("/api/v1/calorie/", http.StripPrefix("/api/v1/calorie", calorieProtectedHandler))
}

func applyMiddleware(mux *http.ServeMux, logger *logger.Logger, cfg *config.Config) http.Handler {
	corsMiddleware := httphandler.NewCORSMiddleware(&cfg.CORS)
	rateLimiter := httphandler.NewRateLimiter(&cfg.RateLimit, logger)
	securityHeadersMiddleware := httphandler.NewSecurityHeadersMiddleware(&cfg.SecurityHeaders)

	return corsMiddleware(
		rateLimiter.RateLimitMiddleware()(
			securityHeadersMiddleware(
				httphandler.LoggingMiddleware(logger)(
					httphandler.RequestIDMiddleware()(mux),
				),
			),
		),
	)
}
