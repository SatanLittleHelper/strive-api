package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
				"password": "Password123",
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
				"password": "Password123",
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

			handlers := NewAuthHandlers(mockService, logger)

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
				"password": "password123",
			},
			mockSetup: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "test@example.com", "password123").
					Return("access_token", "refresh_token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			mockSetup: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "test@example.com", "wrongpassword").
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

			handlers := NewAuthHandlers(mockService, logger)

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
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "access_token")
				assert.Contains(t, response, "refresh_token")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}
