package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleksandr/strive-api/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	mockService := new(MockAuthService)
	userID := uuid.New()
	email := "test@example.com"
	token := "valid-token"

	claims := &services.Claims{
		UserID: userID,
		Email:  email,
	}
	mockService.On("ValidateToken", token).Return(claims, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUserID := ctx.Value(UserIDKey)
		ctxEmail := ctx.Value(UserEmailKey)

		assert.Equal(t, userID, ctxUserID)
		assert.Equal(t, email, ctxEmail)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	middleware := AuthMiddleware(mockService)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mockService := new(MockAuthService)
	token := "invalid-token"

	mockService.On("ValidateToken", token).Return(nil, assert.AnError)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called with invalid token")
	})

	middleware := AuthMiddleware(mockService)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	mockService := new(MockAuthService)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called without token")
	})

	middleware := AuthMiddleware(mockService)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", http.NoBody)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	mockService := new(MockAuthService)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called with invalid format")
	})

	middleware := AuthMiddleware(mockService)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", http.NoBody)
	req.Header.Set("Authorization", "InvalidFormat token")
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_EmptyToken(t *testing.T) {
	mockService := new(MockAuthService)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called with empty token")
	})

	middleware := AuthMiddleware(mockService)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer ")
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_ContextValues(t *testing.T) {
	mockService := new(MockAuthService)
	userID := uuid.New()
	email := "test@example.com"
	token := "valid-token"

	claims := &services.Claims{
		UserID: userID,
		Email:  email,
	}
	mockService.On("ValidateToken", token).Return(claims, nil)

	var capturedUserID uuid.UUID
	var capturedEmail string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		capturedUserID = ctx.Value(UserIDKey).(uuid.UUID)
		capturedEmail = ctx.Value(UserEmailKey).(string)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	middleware := AuthMiddleware(mockService)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, userID, capturedUserID)
	assert.Equal(t, email, capturedEmail)
	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}
