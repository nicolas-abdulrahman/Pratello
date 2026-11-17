package main

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Server wraps http.Server with production timeouts and graceful shutdown support.
type Server struct {
	httpServer *http.Server
}

// NewServer constructs a new HTTP Server instance.
func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  1 * time.Minute,
		},
	}
}

// Start launches ListenAndServe, treating http.ErrServerClosed as a clean exit.
func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server with a 10-second drain timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	drainCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(drainCtx)
}
