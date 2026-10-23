// Command mcp-registry automatically discovers labelled MCP containers.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/faisal-memon/mcp-registry/internal/docker"
	"github.com/faisal-memon/mcp-registry/internal/registry"
)

func main() {
	source, err := docker.NewSource()
	if err != nil {
		log.Printf("create Docker client: %v", err)
		return
	}
	service := registry.New(source, registry.MCPDiscoverer{})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reconcile(ctx, service)
	go reconcilePeriodically(ctx, service, reconcileInterval())

	server := &http.Server{
		Addr:              listenAddress(),
		Handler:           service.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("serving MCP registry at %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("MCP registry stopped unexpectedly: %v", err)
	}
}

func reconcile(ctx context.Context, service *registry.Registry) {
	attemptCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := service.Reconcile(attemptCtx); err != nil {
		log.Printf("reconcile MCP registry: %v", err)
	}
}

func reconcilePeriodically(ctx context.Context, service *registry.Registry, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reconcile(ctx, service)
		}
	}
}

func listenAddress() string {
	if address := os.Getenv("LISTEN_ADDR"); address != "" {
		return address
	}
	return ":8080"
}

func reconcileInterval() time.Duration {
	if value := os.Getenv("RECONCILE_INTERVAL"); value != "" {
		interval, err := time.ParseDuration(value)
		if err == nil && interval > 0 {
			return interval
		}
		log.Printf("invalid RECONCILE_INTERVAL %q; using 30s", value)
	}
	return 30 * time.Second
}
