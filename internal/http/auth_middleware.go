package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/services"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
)

type AuthError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeAuthError(w http.ResponseWriter, log *logger.Logger, r *http.Request, code, message, reason string) {
	authErr := AuthError{}
	authErr.Error.Code = code
	authErr.Error.Message = message

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	if err := json.NewEncoder(w).Encode(authErr); err != nil {
		log.Error("Failed to encode auth error", "error", err)
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`, http.StatusInternalServerError)
		return
	}

	logAuthFailure(log, r, reason)
}

func logAuthFailure(log *logger.Logger, r *http.Request, reason string) {
	log.Warn("Authentication failed",
		"reason", reason,
		"path", r.URL.Path,
		"method", r.Method,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.Header.Get("User-Agent"),
		"request_id", r.Header.Get("X-Request-ID"),
	)
}

func AuthMiddleware(authService services.AuthService, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeAuthError(w, log, r, "UNAUTHORIZED", "Authorization header required", "missing_auth_header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeAuthError(w, log, r, "BEARER_REQUIRED", "Bearer token required", "invalid_auth_format")
				return
			}

			tokenString := parts[1]
			if tokenString == "" {
				writeAuthError(w, log, r, "TOKEN_EMPTY", "Token cannot be empty", "empty_token")
				return
			}

			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				var code, message, reason string

				switch {
				case errors.Is(err, services.ErrTokenExpired):
					code, message, reason = "TOKEN_EXPIRED", "Token has expired", "token_expired"
				case errors.Is(err, services.ErrTokenNotBefore):
					code, message, reason = "TOKEN_NOT_VALID_YET", "Token is not valid yet", "token_not_before"
				case errors.Is(err, services.ErrInvalidSignature):
					code, message, reason = "INVALID_TOKEN", "Invalid token signature", "invalid_signature"
				case errors.Is(err, services.ErrInvalidIssuer):
					code, message, reason = "INVALID_ISSUER", "Invalid token issuer", "invalid_issuer"
				case errors.Is(err, services.ErrInvalidAudience):
					code, message, reason = "INVALID_AUDIENCE", "Invalid token audience", "invalid_audience"
				default:
					code, message, reason = "INVALID_TOKEN", "Invalid or malformed token", "token_validation_failed"
				}

				writeAuthError(w, log, r, code, message, reason)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID.String())
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}
