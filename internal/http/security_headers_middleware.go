package http

import (
	"net/http"
	"strconv"

	"github.com/aleksandr/strive-api/internal/config"
)

func NewSecurityHeadersMiddleware(cfg *config.SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setSecurityHeaders(w, cfg)
			next.ServeHTTP(w, r)
		})
	}
}

func setSecurityHeaders(w http.ResponseWriter, cfg *config.SecurityHeadersConfig) {
	if cfg.HSTSMaxAge > 0 {
		hstsValue := "max-age=" + strconv.Itoa(cfg.HSTSMaxAge)
		if cfg.HSTSIncludeSubdomains {
			hstsValue += "; includeSubDomains"
		}
		w.Header().Set("Strict-Transport-Security", hstsValue)
	}

	if cfg.CSPDirective != "" {
		w.Header().Set("Content-Security-Policy", cfg.CSPDirective)
	}

	if cfg.XFrameOptions != "" {
		w.Header().Set("X-Frame-Options", cfg.XFrameOptions)
	}

	if cfg.XContentTypeOptions != "" {
		w.Header().Set("X-Content-Type-Options", cfg.XContentTypeOptions)
	}

	if cfg.ReferrerPolicy != "" {
		w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
	}

	if cfg.XSSProtection != "" {
		w.Header().Set("X-XSS-Protection", cfg.XSSProtection)
	}

	w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
	w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
}
