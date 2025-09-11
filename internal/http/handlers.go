package http

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

// HealthHandler godoc
// @Summary Health check
// @Description Check if the API is running
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "API is healthy"
// @Router /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}
