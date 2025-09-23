package http

import (
	"encoding/json"
	"net/http"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/aleksandr/strive-api/internal/services"
	"github.com/aleksandr/strive-api/internal/validation"
)

type CalorieHandlers struct {
	calorieService services.CalorieService
	logger         *logger.Logger
	validator      *validation.Validator
}

func NewCalorieHandlers(calorieService services.CalorieService, logger *logger.Logger) *CalorieHandlers {
	return &CalorieHandlers{
		calorieService: calorieService,
		logger:         logger,
		validator:      &validation.Validator{},
	}
}

func (h *CalorieHandlers) CalculateCalories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data models.CalorieCalculationData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validator.Validate(data); err != nil {
		h.logger.Error("Validation failed", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userUUID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.Error("User ID not found in context", "error", err)
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	results, err := h.calorieService.CalculateCalories(r.Context(), userUUID, &data)
	if err != nil {
		h.logger.Error("Failed to calculate calories", "error", err)
		http.Error(w, "Failed to calculate calories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *CalorieHandlers) GetLastCalculation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userUUID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.Error("User ID not found in context", "error", err)
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	response, err := h.calorieService.GetLastCalculation(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("Failed to get last calculation", "error", err)
		http.Error(w, "No calculation found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}
