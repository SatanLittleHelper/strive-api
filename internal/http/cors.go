package http

import (
	"net/http"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/rs/cors"
)

func NewCORSMiddleware(cfg *config.CORSConfig) func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		ExposedHeaders:   cfg.ExposedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})

	return c.Handler
}
