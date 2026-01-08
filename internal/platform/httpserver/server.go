package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// Server wraps an HTTP server with graceful shutdown
type Server struct {
	httpServer *http.Server
	logger     zerolog.Logger
}

// NewServer creates a new HTTP server
func NewServer(addr string, handler http.Handler, logger zerolog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger: logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info().Str("addr", s.httpServer.Addr).Msg("Starting HTTP server")
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

