package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/aleksandr/strive-api/internal/logger"
)

type SecurityLogger struct {
	logger *logger.Logger
}

func NewSecurityLogger(logger *logger.Logger) *SecurityLogger {
	return &SecurityLogger{
		logger: logger,
	}
}

func (sl *SecurityLogger) LogSecurityEvent(event string, r *http.Request, details map[string]interface{}) {
	fields := map[string]interface{}{
		"event":       event,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"method":      r.Method,
		"path":        r.URL.Path,
		"remote_addr": getClientIP(r),
		"user_agent":  r.UserAgent(),
	}

	for k, v := range details {
		fields[k] = v
	}

	sl.logger.Warn("Security event", "fields", fields)
}

func (sl *SecurityLogger) LogFailedAuth(r *http.Request, reason string) {
	sl.LogSecurityEvent("failed_authentication", r, map[string]interface{}{
		"reason": reason,
	})
}

func (sl *SecurityLogger) LogSuspiciousActivity(r *http.Request, reason string) {
	sl.LogSecurityEvent("suspicious_activity", r, map[string]interface{}{
		"reason": reason,
	})
}

func (sl *SecurityLogger) LogRateLimitExceeded(r *http.Request, limit int) {
	sl.LogSecurityEvent("rate_limit_exceeded", r, map[string]interface{}{
		"limit": limit,
	})
}

func (sl *SecurityLogger) LogInvalidInput(r *http.Request, validationErrors []string) {
	sl.LogSecurityEvent("invalid_input", r, map[string]interface{}{
		"validation_errors": validationErrors,
	})
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if idx := strings.Index(ip, ","); idx != -1 {
			return strings.TrimSpace(ip[:idx])
		}
		return strings.TrimSpace(ip)
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	return r.RemoteAddr
}

// SecurityMiddleware logs security-related events
func (sl *SecurityLogger) SecurityMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isSuspiciousRequest(r) {
				sl.LogSuspiciousActivity(r, "suspicious_request_pattern")
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isSuspiciousRequest(r *http.Request) bool {
	suspiciousPatterns := []string{
		"../",
		"<script",
		"javascript:",
		"eval(",
		"exec(",
		"union select",
		"drop table",
		"delete from",
	}

	path := strings.ToLower(r.URL.Path)
	query := strings.ToLower(r.URL.RawQuery)

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(path, pattern) || strings.Contains(query, pattern) {
			return true
		}
	}

	return len(r.URL.RawQuery) > 2048
}
