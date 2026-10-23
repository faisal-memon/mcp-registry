package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const httpShutdownTimeout = 5 * time.Second

// HTTPServer serves registry snapshots and shuts down gracefully with its context.
type HTTPServer struct {
	server *http.Server
}

// NewHTTPServer creates the registry HTTP server at address.
func NewHTTPServer(address string, cache *Cache) *HTTPServer {
	return &HTTPServer{server: &http.Server{
		Addr:              address,
		Handler:           cache.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}}
}

// Run serves requests until ctx is cancelled or the server fails.
func (s *HTTPServer) Run(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
			defer cancel()
			if err := s.server.Shutdown(shutdownCtx); err != nil {
				slog.Error("shut down MCP registry HTTP server", "error", err)
			}
		case <-stopped:
		}
	}()

	slog.Info("serving MCP registry", "address", s.server.Addr)
	err := s.server.ListenAndServe()
	close(stopped)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("serve MCP registry: %w", err)
}

// Handler serves the latest registry snapshot and a basic health check.
func (c *Cache) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /registry.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(c.Snapshot()); err != nil {
			http.Error(w, "encode registry", http.StatusInternalServerError)
		}
	})
	return mux
}
