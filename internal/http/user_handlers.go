package http

import (
	"encoding/json"
	"net/http"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/services"
	"github.com/aleksandr/strive-api/internal/validation"
)

type UserHandlers struct {
	userService    services.UserService
	logger         *logger.Logger
	securityLogger *SecurityLogger
}

func NewUserHandlers(userService services.UserService, logger *logger.Logger) *UserHandlers {
	return &UserHandlers{
		userService:    userService,
		logger:         logger,
		securityLogger: NewSecurityLogger(logger),
	}
}

type UpdateUserThemeRequest struct {
	Theme string `json:"theme" validate:"required,oneof=light dark" example:"dark"`
}

// Me returns information about the current authenticated user
// @Summary Get current user profile
// @Description Returns detailed information about the currently authenticated user including theme
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User "User profile information"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/user/me [get]
func (h *UserHandlers) Me(w http.ResponseWriter, r *http.Request) {
	userUUID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.Error("User ID not found in context", "error", err)
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"User ID not found in context"}}`, http.StatusInternalServerError)
		return
	}

	user, err := h.userService.GetUserProfile(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("Failed to get user profile", "error", err, "user_id", userUUID)
		http.Error(w, `{"error":{"code":"USER_NOT_FOUND","message":"User not found"}}`, http.StatusNotFound)
		return
	}

	h.logger.Info("User profile requested", "user_id", userUUID, "email", user.Email, "theme", user.Theme)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

// UpdateTheme godoc
// @Summary Update user theme
// @Description Update the theme preference for the current user
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateUserThemeRequest true "Theme update data"
// @Success 200 {object} map[string]interface{} "Theme updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/user/theme [put]
func (h *UserHandlers) UpdateTheme(w http.ResponseWriter, r *http.Request) {
	userUUID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.Error("User ID not found in context", "error", err)
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"User ID not found in context"}}`, http.StatusInternalServerError)
		return
	}

	var req UpdateUserThemeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode theme update request", "error", err)
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"Invalid JSON"}}`, http.StatusBadRequest)
		return
	}

	var validationErrors validation.ValidationErrors
	if req.Theme != "light" && req.Theme != "dark" {
		validationErrors = append(validationErrors, validation.ValidationError{
			Field:   "theme",
			Message: "theme must be either 'light' or 'dark'",
		})
	}

	if len(validationErrors) > 0 {
		h.logger.Warn("Validation failed for theme update request", "errors", validationErrors)
		var errorMessages []string
		for _, err := range validationErrors {
			errorMessages = append(errorMessages, err.Message)
		}
		h.securityLogger.LogInvalidInput(r, errorMessages)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(validationErrors.ToJSON())
		return
	}

	if err := h.userService.UpdateUserTheme(r.Context(), userUUID, req.Theme); err != nil {
		h.logger.Error("Failed to update user theme", "error", err, "user_id", userUUID, "theme", req.Theme)
		if err == services.ErrInvalidTheme {
			http.Error(w, `{"error":{"code":"INVALID_THEME","message":"Invalid theme value"}}`, http.StatusBadRequest)
		} else {
			http.Error(w, `{"error":{"code":"THEME_UPDATE_FAILED","message":"Failed to update theme"}}`, http.StatusInternalServerError)
		}
		return
	}

	h.logger.Info("User theme updated successfully", "user_id", userUUID, "theme", req.Theme)

	response := map[string]interface{}{
		"message": "Theme updated successfully",
		"theme":   req.Theme,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
