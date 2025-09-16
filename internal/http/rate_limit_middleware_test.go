package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/logger"
)

const testClientIP = "192.168.1.1:12345"

func TestRateLimiter_GeneralRequests(t *testing.T) {
	cfg := &config.RateLimitConfig{
		AuthRequestsPerMinute:    5,
		GeneralRequestsPerMinute: 3,
		BurstSize:                5,
		Enabled:                  true,
	}

	log := logger.New("INFO", "text")
	rateLimiter := NewRateLimiter(cfg, log)

	handler := rateLimiter.RateLimitMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	// Test general endpoint
	req := httptest.NewRequest("GET", "/health", http.NoBody)
	req.RemoteAddr = testClientIP

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// 4th request should be rate limited
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

func TestRateLimiter_AuthRequests(t *testing.T) {
	cfg := &config.RateLimitConfig{
		AuthRequestsPerMinute:    2,
		GeneralRequestsPerMinute: 10,
		BurstSize:                5,
		Enabled:                  true,
	}

	log := logger.New("INFO", "text")
	rateLimiter := NewRateLimiter(cfg, log)

	handler := rateLimiter.RateLimitMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	// Test auth endpoint
	req := httptest.NewRequest("POST", "/api/v1/auth/login", http.NoBody)
	req.RemoteAddr = testClientIP

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// 3rd request should be rate limited
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

func TestRateLimiter_Disabled(t *testing.T) {
	cfg := &config.RateLimitConfig{
		AuthRequestsPerMinute:    1,
		GeneralRequestsPerMinute: 1,
		BurstSize:                1,
		Enabled:                  false,
	}

	log := logger.New("INFO", "text")
	rateLimiter := NewRateLimiter(cfg, log)

	handler := rateLimiter.RateLimitMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/health", http.NoBody)
	req.RemoteAddr = testClientIP

	// All requests should succeed when rate limiting is disabled
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimiter_DifferentClients(t *testing.T) {
	cfg := &config.RateLimitConfig{
		AuthRequestsPerMinute:    2,
		GeneralRequestsPerMinute: 2,
		BurstSize:                5,
		Enabled:                  true,
	}

	log := logger.New("INFO", "text")
	rateLimiter := NewRateLimiter(cfg, log)

	handler := rateLimiter.RateLimitMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	// Test with different client IPs
	client1Req := httptest.NewRequest("GET", "/health", http.NoBody)
	client1Req.RemoteAddr = testClientIP

	client2Req := httptest.NewRequest("GET", "/health", http.NoBody)
	client2Req.RemoteAddr = "192.168.1.2:12345"

	// Both clients should be able to make requests independently
	for i := 0; i < 2; i++ {
		// Client 1
		w1 := httptest.NewRecorder()
		handler.ServeHTTP(w1, client1Req)
		if w1.Code != http.StatusOK {
			t.Errorf("Client 1 request %d: expected status 200, got %d", i+1, w1.Code)
		}

		// Client 2
		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, client2Req)
		if w2.Code != http.StatusOK {
			t.Errorf("Client 2 request %d: expected status 200, got %d", i+1, w2.Code)
		}
	}
}

func TestIsAuthEndpoint(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/api/v1/auth/login", true},
		{"/api/v1/auth/register", true},
		{"/api/v1/auth/refresh", true},
		{"/health", false},
		{"/api/v1/user/profile", false},
		{"/swagger/", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := IsAuthEndpoint(tt.path)
			if result != tt.expected {
				t.Errorf("isAuthEndpoint(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}
