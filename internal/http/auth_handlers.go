package http

import (
	"encoding/json"
	"net/http"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/aleksandr/strive-api/internal/services"
	"github.com/aleksandr/strive-api/internal/validation"
)

type AuthHandlers struct {
	authService    services.AuthService
	logger         *logger.Logger
	securityLogger *SecurityLogger
}

func NewAuthHandlers(authService services.AuthService, logger *logger.Logger) *AuthHandlers {
	return &AuthHandlers{
		authService:    authService,
		logger:         logger,
		securityLogger: NewSecurityLogger(logger),
	}
}

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"password123"`
}

// LoginRequest represents user login credentials
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int    `json:"expires_in" example:"900"`
	TokenType    string `json:"token_type" example:"Bearer"`
}

// ErrorResponse represents API error response
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code" example:"VALIDATION_ERROR"`
		Message string `json:"message" example:"Invalid input data"`
	} `json:"error"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration data"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode register request", "error", err)
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"Invalid JSON"}}`, http.StatusBadRequest)
		return
	}

	// Validate input
	var validationErrors validation.ValidationErrors
	if err := validation.ValidateEmail(req.Email); err != nil {
		validationErrors = append(validationErrors, validation.ValidationError{
			Field:   "email",
			Message: err.Error(),
		})
	}
	if err := validation.ValidatePassword(req.Password); err != nil {
		validationErrors = append(validationErrors, validation.ValidationError{
			Field:   "password",
			Message: err.Error(),
		})
	}

	if len(validationErrors) > 0 {
		h.logger.Warn("Validation failed for register request", "errors", validationErrors)
		var errorMessages []string
		for _, err := range validationErrors {
			errorMessages = append(errorMessages, err.Message)
		}
		h.securityLogger.LogInvalidInput(r, errorMessages)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors.ToJSON())
		return
	}

	createReq := &models.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	user, err := h.authService.Register(r.Context(), createReq)
	if err != nil {
		h.logger.Error("Failed to register user", "error", err, "email", req.Email)
		http.Error(w, `{"error":{"code":"REGISTRATION_FAILED","message":"Failed to register user"}}`, http.StatusBadRequest)
		return
	}

	h.logger.Info("User registered successfully", "user_id", user.ID, "email", user.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user_id": user.ID,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User login credentials"
// @Success 200 {object} AuthResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Router /api/v1/auth/login [post]
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode login request", "error", err)
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"Invalid JSON"}}`, http.StatusBadRequest)
		return
	}

	// Validate input
	var validationErrors validation.ValidationErrors
	if err := validation.ValidateEmail(req.Email); err != nil {
		validationErrors = append(validationErrors, validation.ValidationError{
			Field:   "email",
			Message: err.Error(),
		})
	}
	if req.Password == "" {
		validationErrors = append(validationErrors, validation.ValidationError{
			Field:   "password",
			Message: "password is required",
		})
	}

	if len(validationErrors) > 0 {
		h.logger.Warn("Validation failed for login request", "errors", validationErrors)
		var errorMessages []string
		for _, err := range validationErrors {
			errorMessages = append(errorMessages, err.Message)
		}
		h.securityLogger.LogInvalidInput(r, errorMessages)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors.ToJSON())
		return
	}

	accessToken, refreshToken, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to login user", "error", err, "email", req.Email)
		h.securityLogger.LogFailedAuth(r, "invalid_credentials")
		http.Error(w, `{"error":{"code":"INVALID_CREDENTIALS","message":"Invalid email or password"}}`, http.StatusUnauthorized)
		return
	}

	h.logger.Info("User logged in successfully", "email", req.Email)

	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
		TokenType:    "Bearer",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
