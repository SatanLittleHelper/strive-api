package http

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/aleksandr/strive-api/internal/services"
	"github.com/aleksandr/strive-api/internal/validation"
)

const (
	productionEnv = "production"
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

func getCookieSettings() bool {
	return os.Getenv("ENVIRONMENT") == productionEnv
}

func setSecureCookie(w http.ResponseWriter, name, value string, maxAge int) {
	secure := getCookieSettings()

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	}

	http.SetCookie(w, cookie)
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"password123"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int    `json:"expires_in" example:"900"`
	TokenType    string `json:"token_type" example:"Bearer"`
	Message      string `json:"message,omitempty" example:"Login successful"`
}

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
		_ = json.NewEncoder(w).Encode(validationErrors.ToJSON())
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
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
		_ = json.NewEncoder(w).Encode(validationErrors.ToJSON())
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

	setSecureCookie(w, "access-token", accessToken, 900)
	setSecureCookie(w, "refresh-token", refreshToken, 604800)

	response := AuthResponse{
		ExpiresIn: 900,
		TokenType: "Bearer",
		Message:   "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Refresh godoc
// @Summary Refresh access token
// @Description Refresh access token using refresh token from cookie
// @Tags authentication
// @Accept json
// @Produce json
// @Success 200 {object} AuthResponse "Token refreshed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Invalid refresh token"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshTokenCookie, err := r.Cookie("refresh-token")
	if err != nil {
		h.logger.Warn("Refresh token cookie not found")
		h.securityLogger.LogFailedAuth(r, "missing_refresh_token_cookie")
		http.Error(w, `{"error":{"code":"MISSING_REFRESH_TOKEN","message":"Refresh token cookie not found"}}`, http.StatusUnauthorized)
		return
	}

	if refreshTokenCookie.Value == "" {
		h.logger.Warn("Empty refresh token in cookie")
		h.securityLogger.LogInvalidInput(r, []string{"refresh_token is empty"})
		http.Error(w, `{"error":{"code":"INVALID_REFRESH_TOKEN","message":"Refresh token is empty"}}`, http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.authService.RefreshToken(r.Context(), refreshTokenCookie.Value)
	if err != nil {
		h.logger.Error("Failed to refresh token", "error", err)
		h.securityLogger.LogFailedAuth(r, "invalid_refresh_token")
		http.Error(w, `{"error":{"code":"INVALID_REFRESH_TOKEN","message":"Invalid or expired refresh token"}}`, http.StatusUnauthorized)
		return
	}

	h.logger.Info("Token refreshed successfully")

	setSecureCookie(w, "access-token", accessToken, 900)
	setSecureCookie(w, "refresh-token", refreshToken, 604800)

	response := AuthResponse{
		ExpiresIn: 900,
		TokenType: "Bearer",
		Message:   "Token refreshed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Logout godoc
// @Summary Logout user
// @Description Logout user and clear authentication cookies
// @Tags authentication
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	setSecureCookie(w, "access-token", "", -1)
	setSecureCookie(w, "refresh-token", "", -1)

	h.logger.Info("User logged out successfully")

	response := map[string]interface{}{
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Me returns information about the current authenticated user
// @Summary Get current user information
// @Description Returns information about the currently authenticated user
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User information"
// @Failure 401 {object} AuthError "Unauthorized"
// @Failure 500 {object} AuthError "Internal server error"
// @Router /api/v1/auth/me [get]
func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"User ID not found in context"}}`, http.StatusInternalServerError)
		return
	}

	userEmail, _ := GetUserEmailFromContext(r.Context())

	response := map[string]interface{}{
		"user_id": userID,
		"email":   userEmail,
		"message": "User authenticated successfully",
	}

	h.logger.Info("User profile requested", "user_id", userID, "email", userEmail)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
