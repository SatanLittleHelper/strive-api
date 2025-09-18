package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server          ServerConfig
	Log             LogConfig
	DB              DatabaseConfig
	JWT             JWTConfig
	RateLimit       RateLimitConfig
	CORS            CORSConfig
	SecurityHeaders SecurityHeadersConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int32
	MinConns int32
}

type JWTConfig struct {
	Secret    string
	Issuer    string
	Audience  string
	ClockSkew time.Duration
}

type RateLimitConfig struct {
	AuthRequestsPerMinute    int
	GeneralRequestsPerMinute int
	BurstSize                int
	Enabled                  bool
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type SecurityHeadersConfig struct {
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	CSPDirective          string
	XFrameOptions         string
	XContentTypeOptions   string
	ReferrerPolicy        string
	XSSProtection         string
}

func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvInt("PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "INFO"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		DB: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "strive"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: int32(getEnvInt("DB_MAX_CONNS", 25)),
			MinConns: int32(getEnvInt("DB_MIN_CONNS", 5)),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", ""),
			Issuer:    getEnv("JWT_ISSUER", "strive-api"),
			Audience:  getEnv("JWT_AUDIENCE", "strive-app"),
			ClockSkew: getEnvDuration("JWT_CLOCK_SKEW", 2*time.Minute),
		},
		RateLimit: RateLimitConfig{
			AuthRequestsPerMinute:    getEnvInt("RATE_LIMIT_AUTH_PER_MINUTE", 5),
			GeneralRequestsPerMinute: getEnvInt("RATE_LIMIT_GENERAL_PER_MINUTE", 60),
			BurstSize:                getEnvInt("RATE_LIMIT_BURST_SIZE", 10),
			Enabled:                  getEnv("RATE_LIMIT_ENABLED", "true") == "true",
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:4200", "http://127.0.0.1:4200", "http://192.168.1.186:4200", "https://satanlittlehelper.github.io"}),
			AllowedMethods:   getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"}),
			ExposedHeaders:   getEnvSlice("CORS_EXPOSED_HEADERS", []string{"X-Request-ID"}),
			AllowCredentials: getEnv("CORS_ALLOW_CREDENTIALS", "true") == "true",
			MaxAge:           getEnvInt("CORS_MAX_AGE", 86400),
		},
		SecurityHeaders: SecurityHeadersConfig{
			HSTSMaxAge:            getEnvInt("SECURITY_HSTS_MAX_AGE", 31536000),
			HSTSIncludeSubdomains: getEnv("SECURITY_HSTS_INCLUDE_SUBDOMAINS", "true") == "true",
			CSPDirective:          getEnv("SECURITY_CSP_DIRECTIVE", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"),
			XFrameOptions:         getEnv("SECURITY_X_FRAME_OPTIONS", "DENY"),
			XContentTypeOptions:   getEnv("SECURITY_X_CONTENT_TYPE_OPTIONS", "nosniff"),
			ReferrerPolicy:        getEnv("SECURITY_REFERRER_POLICY", "strict-origin-when-cross-origin"),
			XSSProtection:         getEnv("SECURITY_XSS_PROTECTION", "1; mode=block"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Server.Port)
	}

	validLevels := map[string]bool{
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
	}
	if !validLevels[c.Log.Level] {
		return fmt.Errorf("invalid log level: %s", c.Log.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validFormats[c.Log.Format] {
		return fmt.Errorf("invalid log format: %s", c.Log.Format)
	}

	if c.DB.Port <= 0 || c.DB.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.DB.Port)
	}

	if c.DB.MaxConns <= 0 {
		return fmt.Errorf("invalid max connections: %d", c.DB.MaxConns)
	}

	if c.DB.MinConns < 0 || c.DB.MinConns > c.DB.MaxConns {
		return fmt.Errorf("invalid min connections: %d", c.DB.MinConns)
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	return nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DB.User, c.DB.Password, c.DB.Host, c.DB.Port, c.DB.DBName, c.DB.SSLMode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
