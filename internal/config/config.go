package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server ServerConfig
	Log    LogConfig
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

	return nil
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
