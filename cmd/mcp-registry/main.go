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

const (
	defaultEnabledLabelKey   = "mcp.enabled"
	defaultEnabledLabelValue = "true"
	defaultPortLabelKey      = "mcp.port"
	defaultPathLabelKey      = "mcp.path"
	defaultMCPPath           = "/mcp"
)

type Config struct {
	ListenAddress     string
	ReconcileInterval time.Duration
	Labels            registry.LabelConfig
}

func main() {
	config := configFromEnvironment()
	source, err := docker.NewSource()
	if err != nil {
		log.Printf("create Docker client: %v", err)
		return
	}
	service := registry.New(source, registry.MCPDiscoverer{}, config.Labels)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reconcile(ctx, service)
	go reconcilePeriodically(ctx, service, config.ReconcileInterval)

	server := &http.Server{
		Addr:              config.ListenAddress,
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

func configFromEnvironment() Config {
	config := Config{
		ListenAddress:     environmentValue("LISTEN_ADDR", ":8080"),
		ReconcileInterval: 30 * time.Second,
		Labels: registry.LabelConfig{
			EnabledKey:   defaultEnabledLabelKey,
			EnabledValue: defaultEnabledLabelValue,
			PortKey:      defaultPortLabelKey,
			PathKey:      defaultPathLabelKey,
			DefaultPath:  defaultMCPPath,
		},
	}

	if value := os.Getenv("RECONCILE_INTERVAL"); value != "" {
		interval, err := time.ParseDuration(value)
		if err == nil && interval > 0 {
			config.ReconcileInterval = interval
		} else {
			log.Printf("invalid RECONCILE_INTERVAL %q; using %s", value, config.ReconcileInterval)
		}
	}
	config.Labels.EnabledKey = environmentValue("MCP_LABEL_ENABLED_KEY", config.Labels.EnabledKey)
	config.Labels.EnabledValue = environmentValue("MCP_LABEL_ENABLED_VALUE", config.Labels.EnabledValue)
	config.Labels.PortKey = environmentValue("MCP_LABEL_PORT_KEY", config.Labels.PortKey)
	config.Labels.PathKey = environmentValue("MCP_LABEL_PATH_KEY", config.Labels.PathKey)
	config.Labels.DefaultPath = environmentValue("MCP_DEFAULT_PATH", config.Labels.DefaultPath)
	return config
}

func environmentValue(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
