package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandlers_Register(t *testing.T) {
	logger := logger.New("INFO", "json")

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful registration",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "Password123!",
			},
			mockSetup: func(m *MockAuthService) {
				user := &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
				m.On("Register", mock.Anything, mock.AnythingOfType("*models.CreateUserRequest")).
					Return(user, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "invalid request body",
			requestBody: map[string]string{
				"email": "invalid-email",
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "service error",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "Password123!",
			},
			mockSetup: func(m *MockAuthService) {
				m.On("Register", mock.Anything, mock.AnythingOfType("*models.CreateUserRequest")).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tt.mockSetup(mockService)

			cfg := &config.Config{}
			handlers := NewAuthHandlers(mockService, logger, cfg)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handlers.Register(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "message")
				assert.Contains(t, response, "user_id")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandlers_Login(t *testing.T) {
	logger := logger.New("INFO", "json")

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful login",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "Password123!",
			},
			mockSetup: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "test@example.com", "Password123!").
					Return("access_token", "refresh_token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "WrongPassword123!",
			},
			mockSetup: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "test@example.com", "WrongPassword123!").
					Return("", "", assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "invalid request body",
			requestBody: map[string]string{
				"email": "invalid-email",
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tt.mockSetup(mockService)

			cfg := &config.Config{}
			handlers := NewAuthHandlers(mockService, logger, cfg)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handlers.Login(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				// Check that access token is in JSON response and refresh token is in cookie
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "access_token")
				assert.Equal(t, "access_token", response["access_token"])

				cookies := rr.Result().Cookies()
				var refreshTokenCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == "refresh-token" {
						refreshTokenCookie = cookie
					}
				}

				assert.NotNil(t, refreshTokenCookie, "refresh-token cookie should be set")
				assert.Equal(t, "refresh_token", refreshTokenCookie.Value)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandlers_Me(t *testing.T) {
	logger := logger.New("INFO", "json")

	tests := []struct {
		name           string
		mockSetup      func(m *MockAuthService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful me request",
			mockSetup: func(m *MockAuthService) {
				// No mock setup needed as Me doesn't call auth service
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{}
			tt.mockSetup(mockService)

			cfg := &config.Config{}
			handlers := NewAuthHandlers(mockService, logger, cfg)

			req := httptest.NewRequest("GET", "/api/v1/auth/me", http.NoBody)
			rr := httptest.NewRecorder()

			// Add user context to request (simulating auth middleware)
			ctx := context.WithValue(req.Context(), UserIDKey, "test-user-id")
			ctx = context.WithValue(ctx, UserEmailKey, "test@example.com")
			req = req.WithContext(ctx)

			handlers.Me(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "user_id")
				assert.Contains(t, response, "email")
				assert.Contains(t, response, "message")
				assert.Equal(t, "test-user-id", response["user_id"])
				assert.Equal(t, "test@example.com", response["email"])
			}

			mockService.AssertExpectations(t)
		})
	}
}
