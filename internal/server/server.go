package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/makcim392/maintenance-api/internal/health"
	"github.com/makcim392/maintenance-api/internal/logger"
)

// Server wraps the HTTP server with graceful shutdown capabilities
type Server struct {
	httpServer *http.Server
	logger     *logger.Logger
	health     *health.HealthChecker
}

// New creates a new server instance
func New(addr string, handler http.Handler, logger *logger.Logger, healthChecker *health.HealthChecker) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger: logger,
		health: healthChecker,
	}
}

// Start starts the server with graceful shutdown
func (s *Server) Start() error {
	// Create a context that listens for interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the server in a goroutine
	go func() {
		s.logger.LogInfo("Server starting on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.LogError(err, "Server failed to start")
		}
	}()

	// Start background health checks
	go func() {
		s.health.StartBackgroundChecks(ctx, 30*time.Second)
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	s.logger.LogInfo("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.LogError(err, "Server forced to shutdown")
		return err
	}

	s.logger.LogInfo("Server exited")
	return nil
}

// Stop immediately stops the server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}