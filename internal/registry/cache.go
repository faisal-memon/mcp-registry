// Package registry discovers labelled MCP containers and exposes their tools.
package registry

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"
)

const reconcileTimeout = 10 * time.Second

// Cache holds the latest registry snapshot and refreshes it from live containers.
type Cache struct {
	source     ContainerSource
	discoverer Discoverer
	labels     LabelConfig

	mu       sync.RWMutex
	snapshot Snapshot
}

// NewCache creates an empty cache backed by the supplied container source and MCP discoverer.
func NewCache(source ContainerSource, discoverer Discoverer, labels LabelConfig) *Cache {
	return &Cache{source: source, discoverer: discoverer, labels: labels}
}

// Start populates the cache once, then refreshes it until ctx is cancelled.
// An initial discovery failure prevents the registry from starting with an empty snapshot.
func (c *Cache) Start(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("cache refresh interval must be positive")
	}
	if err := c.reconcile(ctx); err != nil {
		return err
	}
	go c.refreshPeriodically(ctx, interval)
	return nil
}

// Snapshot returns a copy of the latest successfully discovered services.
func (c *Cache) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	servers := append([]Server(nil), c.snapshot.Servers...)
	return Snapshot{Servers: servers}
}

func (c *Cache) refreshPeriodically(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.reconcile(ctx); err != nil {
				slog.Warn("refresh MCP registry cache", "error", err)
			}
		}
	}
}

// reconcile replaces the snapshot with MCP services discovered from running labelled containers.
func (c *Cache) reconcile(ctx context.Context) error {
	attemptCtx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	containers, err := c.source.List(attemptCtx)
	if err != nil {
		return fmt.Errorf("list MCP containers: %w", err)
	}

	servers := make([]Server, 0, len(containers))
	for _, container := range containers {
		server, ok := c.labels.serverFromContainer(container)
		if !ok {
			continue
		}
		tools, err := c.discoverer.Discover(attemptCtx, server.URL)
		if err != nil {
			slog.Warn("discover MCP server", "name", server.Name, "url", server.URL, "error", err)
		} else {
			server.Enabled = true
			server.Tools = tools
		}
		servers = append(servers, server)
	}
	sort.Slice(servers, func(i, j int) bool { return servers[i].Name < servers[j].Name })

	c.mu.Lock()
	c.snapshot = Snapshot{Servers: servers}
	c.mu.Unlock()
	return nil
}
