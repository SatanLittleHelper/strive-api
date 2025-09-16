package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/logger"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	config   *config.RateLimitConfig
	logger   *logger.Logger
}

type RateLimitError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewRateLimiter(cfg *config.RateLimitConfig, log *logger.Logger) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		config:   cfg,
		logger:   log,
	}
}

func (rl *RateLimiter) isAllowed(clientID string, limit int) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-time.Minute)

	if rl.requests[clientID] == nil {
		rl.requests[clientID] = make([]time.Time, 0)
	}

	requests := rl.requests[clientID]
	var validRequests []time.Time

	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	if len(validRequests) >= limit {
		return false
	}

	validRequests = append(validRequests, now)
	rl.requests[clientID] = validRequests

	return true
}

func (rl *RateLimiter) cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-time.Minute)

	for clientID, requests := range rl.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, clientID)
		} else {
			rl.requests[clientID] = validRequests
		}
	}
}

func (rl *RateLimiter) writeRateLimitError(w http.ResponseWriter, r *http.Request, limit int) {
	rateLimitErr := RateLimitError{}
	rateLimitErr.Error.Code = "RATE_LIMIT_EXCEEDED"
	rateLimitErr.Error.Message = fmt.Sprintf("Rate limit exceeded. Maximum %d requests per minute allowed.", limit)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "60")
	w.WriteHeader(http.StatusTooManyRequests)

	if err := json.NewEncoder(w).Encode(rateLimitErr); err != nil {
		rl.logger.Error("Failed to encode rate limit error", "error", err)
		http.Error(w, `{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Rate limit exceeded"}}`, http.StatusTooManyRequests)
		return
	}

	rl.logger.Warn("Rate limit exceeded",
		"client_ip", getClientIP(r),
		"path", r.URL.Path,
		"method", r.Method,
		"limit", limit,
		"user_agent", r.Header.Get("User-Agent"),
	)
}

func (rl *RateLimiter) RateLimitMiddleware() func(http.Handler) http.Handler {
	if !rl.config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := getClientIP(r)
			limit := rl.config.GeneralRequestsPerMinute

			if isAuthEndpoint(r.URL.Path) {
				limit = rl.config.AuthRequestsPerMinute
			}

			if !rl.isAllowed(clientID, limit) {
				rl.writeRateLimitError(w, r, limit)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAuthEndpoint(path string) bool {
	authPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
	}

	for _, authPath := range authPaths {
		if path == authPath {
			return true
		}
	}
	return false
}
