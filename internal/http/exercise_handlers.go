package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/aleksandr/strive-api/internal/services"
	"github.com/aleksandr/strive-api/internal/validation"
	"github.com/google/uuid"
)

type ExerciseHandlers struct {
	exerciseService services.ExerciseService
	logger          *logger.Logger
	validator       *validation.Validator
}

func NewExerciseHandlers(exerciseService services.ExerciseService, logger *logger.Logger) *ExerciseHandlers {
	return &ExerciseHandlers{
		exerciseService: exerciseService,
		logger:          logger,
		validator:       &validation.Validator{},
	}
}

// GetExercises returns a list of exercises with optional filtering
// @Summary Get exercises list
// @Description Returns a paginated list of exercises with optional filtering by muscle group, equipment, and category
// @Tags exercises
// @Accept json
// @Produce json
// @Param muscle_group_id query string false "Filter by muscle group ID"
// @Param equipment_id query string false "Filter by equipment ID"
// @Param category query int false "Filter by category (9=strength, 10=bodyweight, 11=weighted)"
// @Param search query string false "Search by exercise name"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.ExerciseListResponse "List of exercises"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/exercises [get]
func (h *ExerciseHandlers) GetExercises(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filters, err := h.parseExerciseFilters(r)
	if err != nil {
		http.Error(w, `{"error":{"code":"INVALID_PARAMETER","message":"Invalid request parameters"}}`, http.StatusBadRequest)
		return
	}

	response, err := h.exerciseService.GetExercises(r.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to get exercises", "error", err)
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Failed to get exercises"}}`, http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, response)
}

func (h *ExerciseHandlers) parseExerciseFilters(r *http.Request) (*models.ExerciseFilters, error) {
	filters := &models.ExerciseFilters{
		Page:  1,
		Limit: 20,
	}

	if err := h.parseMuscleGroupFilter(r, filters); err != nil {
		return nil, err
	}

	if err := h.parseEquipmentFilter(r, filters); err != nil {
		return nil, err
	}

	if err := h.parseCategoryFilter(r, filters); err != nil {
		return nil, err
	}

	if err := h.parsePaginationFilters(r, filters); err != nil {
		return nil, err
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	return filters, nil
}

func (h *ExerciseHandlers) parseMuscleGroupFilter(r *http.Request, filters *models.ExerciseFilters) error {
	if muscleGroupIDStr := r.URL.Query().Get("muscle_group_id"); muscleGroupIDStr != "" {
		muscleGroupID, err := uuid.Parse(muscleGroupIDStr)
		if err != nil {
			h.logger.Error("Invalid muscle group ID", "error", err, "muscle_group_id", muscleGroupIDStr)
			return err
		}
		filters.MuscleGroupID = &muscleGroupID
	}
	return nil
}

func (h *ExerciseHandlers) parseEquipmentFilter(r *http.Request, filters *models.ExerciseFilters) error {
	if equipmentIDStr := r.URL.Query().Get("equipment_id"); equipmentIDStr != "" {
		equipmentID, err := uuid.Parse(equipmentIDStr)
		if err != nil {
			h.logger.Error("Invalid equipment ID", "error", err, "equipment_id", equipmentIDStr)
			return err
		}
		filters.EquipmentID = &equipmentID
	}
	return nil
}

func (h *ExerciseHandlers) parseCategoryFilter(r *http.Request, filters *models.ExerciseFilters) error {
	if categoryStr := r.URL.Query().Get("category"); categoryStr != "" {
		category, err := strconv.Atoi(categoryStr)
		if err != nil {
			h.logger.Error("Invalid category parameter", "error", err, "category", categoryStr)
			return err
		}
		filters.Category = &category
	}
	return nil
}

func (h *ExerciseHandlers) parsePaginationFilters(r *http.Request, filters *models.ExerciseFilters) error {
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			h.logger.Error("Invalid page parameter", "error", err, "page", pageStr)
			return err
		}
		filters.Page = page
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			h.logger.Error("Invalid limit parameter", "error", err, "limit", limitStr)
			return err
		}
		filters.Limit = limit
	}

	return nil
}

func (h *ExerciseHandlers) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// GetExerciseByID returns detailed information about a specific exercise
// @Summary Get exercise by ID
// @Description Returns detailed information about a specific exercise including muscle groups, equipment, and alternatives
// @Tags exercises
// @Accept json
// @Produce json
// @Param id path string true "Exercise ID"
// @Success 200 {object} models.Exercise "Exercise details"
// @Failure 400 {object} ErrorResponse "Invalid exercise ID"
// @Failure 404 {object} ErrorResponse "Exercise not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/exercises/{id} [get]
func (h *ExerciseHandlers) GetExerciseByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	exerciseIDStr := r.URL.Path[len("/api/v1/exercises/"):]
	exerciseID, err := uuid.Parse(exerciseIDStr)
	if err != nil {
		h.logger.Error("Invalid exercise ID", "error", err, "exercise_id", exerciseIDStr)
		http.Error(w, `{"error":{"code":"INVALID_PARAMETER","message":"Invalid exercise ID"}}`, http.StatusBadRequest)
		return
	}

	exercise, err := h.exerciseService.GetExerciseByID(r.Context(), exerciseID)
	if err != nil {
		h.logger.Error("Failed to get exercise", "error", err, "exercise_id", exerciseID)
		http.Error(w, `{"error":{"code":"EXERCISE_NOT_FOUND","message":"Exercise not found"}}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(exercise); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// GetMuscleGroups returns a list of all muscle groups
// @Summary Get muscle groups
// @Description Returns a list of all available muscle groups
// @Tags exercises
// @Accept json
// @Produce json
// @Success 200 {array} models.MuscleGroup "List of muscle groups"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/exercises/muscle-groups [get]
func (h *ExerciseHandlers) GetMuscleGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	muscleGroups, err := h.exerciseService.GetMuscleGroups(r.Context())
	if err != nil {
		h.logger.Error("Failed to get muscle groups", "error", err)
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Failed to get muscle groups"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(muscleGroups); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// GetEquipment returns a list of all equipment
// @Summary Get equipment
// @Description Returns a list of all available equipment
// @Tags exercises
// @Accept json
// @Produce json
// @Success 200 {array} models.Equipment "List of equipment"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/exercises/equipment [get]
func (h *ExerciseHandlers) GetEquipment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	equipment, err := h.exerciseService.GetEquipment(r.Context())
	if err != nil {
		h.logger.Error("Failed to get equipment", "error", err)
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Failed to get equipment"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(equipment); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// GetCacheStatus returns the current cache status
// @Summary Get cache status
// @Description Returns information about the exercise cache including last update time and validity
// @Tags exercises
// @Accept json
// @Produce json
// @Success 200 {object} models.CacheStatus "Cache status information"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/exercises/cache/status [get]
func (h *ExerciseHandlers) GetCacheStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status, err := h.exerciseService.GetCacheStatus(r.Context())
	if err != nil {
		h.logger.Error("Failed to get cache status", "error", err)
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Failed to get cache status"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// RefreshCache manually refreshes the exercise cache
// @Summary Refresh cache
// @Description Manually refreshes the exercise cache from ExerciseDB API
// @Tags exercises
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Success message"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/exercises/cache/refresh [post]
func (h *ExerciseHandlers) RefreshCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.exerciseService.RefreshCache(r.Context())
	if err != nil {
		h.logger.Error("Failed to refresh cache", "error", err)
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Failed to refresh cache"}}`, http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Cache refreshed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}
