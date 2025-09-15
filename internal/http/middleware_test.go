package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	logger := logger.New("INFO", "json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	middleware := LoggingMiddleware(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "test response", rr.Body.String())
}

func TestRequestIDMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(requestIDKey)
		assert.NotNil(t, requestID)
		assert.NotEmpty(t, requestID)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	middleware := RequestIDMiddleware()
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "test response", rr.Body.String())
}

func TestRequestIDMiddleware_ExistingID(t *testing.T) {
	existingID := "existing-request-id"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(requestIDKey)
		assert.Equal(t, existingID, requestID)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	middleware := RequestIDMiddleware()
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-Request-ID", existingID)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "test response", rr.Body.String())
}

func TestMiddlewareChain(t *testing.T) {
	log := logger.New("INFO", "json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(requestIDKey)
		assert.NotNil(t, requestID)
		assert.NotEmpty(t, requestID)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	chain := LoggingMiddleware(log)(RequestIDMiddleware()(handler))

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rr := httptest.NewRecorder()

	chain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "test response", rr.Body.String())
}

func TestMiddlewareContext(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		assert.NotNil(t, ctx)

		requestID := ctx.Value(requestIDKey)
		assert.NotNil(t, requestID)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	middleware := RequestIDMiddleware()
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
