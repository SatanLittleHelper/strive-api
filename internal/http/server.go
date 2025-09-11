package http

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/logger"
)

type Server struct {
	httpServer *http.Server
	logger     *logger.Logger
}

func NewServer(cfg *config.Config, handler http.Handler, log *logger.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + strconv.Itoa(cfg.Server.Port),
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       cfg.Server.ReadTimeout,
			WriteTimeout:      cfg.Server.WriteTimeout,
			IdleTimeout:       cfg.Server.IdleTimeout,
		},
		logger: log,
	}
}

func (s *Server) Start() {
	s.logger.Info("Starting server", "addr", s.httpServer.Addr)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server error", "error", err)
			log.Fatalf("server error: %v", err)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) WaitForShutdown() {
	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-shutdownCtx.Done()
	s.logger.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Stop(ctx); err != nil {
		s.logger.Error("Graceful shutdown failed", "error", err)
		log.Printf("graceful shutdown failed: %v", err)
	} else {
		s.logger.Info("Server stopped gracefully")
	}
}
