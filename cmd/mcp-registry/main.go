// Command mcp-registry automatically discovers labelled MCP containers.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/faisal-memon/mcp-registry/internal/docker"
	"github.com/faisal-memon/mcp-registry/internal/mcp"
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
	if err := run(); err != nil {
		slog.Error("run MCP registry", "error", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := configFromEnvironment()
	if err != nil {
		return err
	}

	dockerSource, err := docker.NewSource()
	if err != nil {
		return fmt.Errorf("create Docker source: %w", err)
	}

	cache := registry.NewCache(dockerSource, mcp.NewDiscoverer(), config.Labels)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cache.Start(ctx, config.ReconcileInterval); err != nil {
		return fmt.Errorf("start registry cache: %w", err)
	}

	server := registry.NewHTTPServer(config.ListenAddress, cache)
	return server.Run(ctx)
}

func configFromEnvironment() (Config, error) {
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
		if err != nil || interval <= 0 {
			return Config{}, fmt.Errorf("RECONCILE_INTERVAL must be a positive duration: %q", value)
		}
		config.ReconcileInterval = interval
	}
	config.Labels.EnabledKey = environmentValue("MCP_LABEL_ENABLED_KEY", config.Labels.EnabledKey)
	config.Labels.EnabledValue = environmentValue("MCP_LABEL_ENABLED_VALUE", config.Labels.EnabledValue)
	config.Labels.PortKey = environmentValue("MCP_LABEL_PORT_KEY", config.Labels.PortKey)
	config.Labels.PathKey = environmentValue("MCP_LABEL_PATH_KEY", config.Labels.PathKey)
	config.Labels.DefaultPath = environmentValue("MCP_DEFAULT_PATH", config.Labels.DefaultPath)
	return config, nil
}

func environmentValue(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
