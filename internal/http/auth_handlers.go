package http

import (
	"encoding/json"
	"net/http"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/aleksandr/strive-api/internal/services"
)

type AuthHandlers struct {
	authService services.AuthService
	logger      *logger.Logger
}

func NewAuthHandlers(authService services.AuthService, logger *logger.Logger) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		logger:      logger,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode register request", "error", err)
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"Invalid JSON"}}`, http.StatusBadRequest)
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

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode login request", "error", err)
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"Invalid JSON"}}`, http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to login user", "error", err, "email", req.Email)
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
