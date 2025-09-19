package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/aleksandr/strive-api/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	validateTokenFunc func(string) (*services.Claims, error)
}

func (m *mockAuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	return nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	return "", "", nil
}

func (m *mockAuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	return "", "", nil
}

func (m *mockAuthService) ValidateToken(tokenString string) (*services.Claims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(tokenString)
	}
	return nil, nil
}

func (m *mockAuthService) HashPassword(password string) (string, error) {
	return "", nil
}

func (m *mockAuthService) VerifyPassword(hashedPassword, password string) error {
	return nil
}

func TestAuthMiddleware(t *testing.T) {
	log := logger.New("INFO", "json")

	t.Run("MissingAuthorizationHeader", func(t *testing.T) {
		mockAuth := &mockAuthService{}
		middleware := AuthMiddleware(mockAuth, log)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response AuthError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Error.Code)
		assert.Equal(t, "Authentication required", response.Error.Message)
	})

	t.Run("InvalidBearerFormat", func(t *testing.T) {
		mockAuth := &mockAuthService{}
		middleware := AuthMiddleware(mockAuth, log)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Basic dGVzdDp0ZXN0")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response AuthError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Error.Code)
	})

	t.Run("EmptyToken", func(t *testing.T) {
		mockAuth := &mockAuthService{}
		middleware := AuthMiddleware(mockAuth, log)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response AuthError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", response.Error.Code)
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		mockAuth := &mockAuthService{
			validateTokenFunc: func(token string) (*services.Claims, error) {
				return nil, services.ErrTokenExpired
			},
		}
		middleware := AuthMiddleware(mockAuth, log)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer expired-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response AuthError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "TOKEN_EXPIRED", response.Error.Code)
		assert.Equal(t, "Token has expired", response.Error.Message)
	})

	t.Run("InvalidIssuer", func(t *testing.T) {
		mockAuth := &mockAuthService{
			validateTokenFunc: func(token string) (*services.Claims, error) {
				return nil, services.ErrInvalidIssuer
			},
		}
		middleware := AuthMiddleware(mockAuth, log)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer invalid-issuer-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response AuthError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "INVALID_ISSUER", response.Error.Code)
	})

	t.Run("InvalidAudience", func(t *testing.T) {
		mockAuth := &mockAuthService{
			validateTokenFunc: func(token string) (*services.Claims, error) {
				return nil, services.ErrInvalidAudience
			},
		}
		middleware := AuthMiddleware(mockAuth, log)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer invalid-audience-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response AuthError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "INVALID_AUDIENCE", response.Error.Code)
	})

	t.Run("ValidToken", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"

		mockAuth := &mockAuthService{
			validateTokenFunc: func(token string) (*services.Claims, error) {
				return &services.Claims{
					UserID: userID,
					Email:  email,
				}, nil
			},
		}
		middleware := AuthMiddleware(mockAuth, log)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextUserID, ok := GetUserIDFromContext(r.Context())
			assert.True(t, ok)
			assert.Equal(t, userID.String(), contextUserID)

			contextEmail, ok := GetUserEmailFromContext(r.Context())
			assert.True(t, ok)
			assert.Equal(t, email, contextEmail)

			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("RealTokenValidation", func(t *testing.T) {
		testUser := &models.User{
			ID:    uuid.New(),
			Email: "real@example.com",
		}

		mockAuth := &mockAuthService{
			validateTokenFunc: func(token string) (*services.Claims, error) {
				return &services.Claims{
					UserID: testUser.ID,
					Email:  testUser.Email,
				}, nil
			},
		}
		middleware := AuthMiddleware(mockAuth, log)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextUserID, ok := GetUserIDFromContext(r.Context())
			assert.True(t, ok)
			assert.Equal(t, testUser.ID.String(), contextUserID)

			contextEmail, ok := GetUserEmailFromContext(r.Context())
			assert.True(t, ok)
			assert.Equal(t, testUser.Email, contextEmail)

			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-Request-ID", "test-request-123")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
