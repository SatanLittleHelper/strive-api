package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DetailedHealthHandler struct {
	logger *logger.Logger
	db     *pgxpool.Pool
}

func NewDetailedHealthHandler(logger *logger.Logger, db *pgxpool.Pool) *DetailedHealthHandler {
	return &DetailedHealthHandler{
		logger: logger,
		db:     db,
	}
}

type DetailedHealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Services  map[string]ServiceInfo `json:"services,omitempty"`
}

type ServiceInfo struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Health godoc
// @Summary Health check
// @Description Check if the API is running
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} DetailedHealthResponse "API is healthy"
// @Router /health [get]
func (h *DetailedHealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := DetailedHealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// DatabaseHealth godoc
// @Summary Database health check
// @Description Check if the database is accessible
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} DetailedHealthResponse "Database is healthy"
// @Failure 503 {object} DetailedHealthResponse "Database is unhealthy"
// @Router /health/db [get]
func (h *DetailedHealthHandler) DatabaseHealth(w http.ResponseWriter, r *http.Request) {
	response := DetailedHealthResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  make(map[string]ServiceInfo),
	}

	// Check database connection
	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := h.db.Ping(ctx); err != nil {
			h.logger.Error("Database health check failed", "error", err)
			response.Status = "unhealthy"
			response.Services["database"] = ServiceInfo{
				Status:  "down",
				Message: "Database connection failed",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(response)
			return
		}

		response.Services["database"] = ServiceInfo{
			Status: "up",
		}
	} else {
		response.Status = "unhealthy"
		response.Services["database"] = ServiceInfo{
			Status:  "down",
			Message: "Database not configured",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	response.Status = "ok"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// DetailedHealth godoc
// @Summary Detailed health check
// @Description Check all system components
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} DetailedHealthResponse "All systems healthy"
// @Failure 503 {object} DetailedHealthResponse "Some systems unhealthy"
// @Router /health/detailed [get]
func (h *DetailedHealthHandler) DetailedHealth(w http.ResponseWriter, r *http.Request) {
	response := DetailedHealthResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  make(map[string]ServiceInfo),
	}

	allHealthy := true

	// Check database
	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := h.db.Ping(ctx); err != nil {
			h.logger.Error("Database health check failed", "error", err)
			response.Services["database"] = ServiceInfo{
				Status:  "down",
				Message: "Database connection failed",
			}
			allHealthy = false
		} else {
			response.Services["database"] = ServiceInfo{
				Status: "up",
			}
		}
	} else {
		response.Services["database"] = ServiceInfo{
			Status:  "down",
			Message: "Database not configured",
		}
		allHealthy = false
	}

	// Add other service checks here as needed
	response.Services["api"] = ServiceInfo{
		Status: "up",
	}

	if allHealthy {
		response.Status = "ok"
		w.WriteHeader(http.StatusOK)
	} else {
		response.Status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
