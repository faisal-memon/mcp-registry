// Package registry discovers labelled MCP containers and exposes their tools.
package registry

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// Container is the small part of Docker container metadata the registry needs.
type Container struct {
	ID     string
	Labels map[string]string
}

// ContainerSource lists currently running containers.
type ContainerSource interface {
	List(context.Context) ([]Container, error)
}

// Discoverer reads the tools exposed by one MCP endpoint.
type Discoverer interface {
	Discover(context.Context, string) ([]Tool, error)
}

// Tool is the JSON-compatible MCP tool shape exposed by the registry.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	InputSchema any    `json:"inputSchema"`
}

// Server is one MCP service returned to an agent. It remains listed but unavailable when discovery fails.
type Server struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
	Tools       []Tool `json:"tools"`
}

// Snapshot matches the existing static registry document while including discovered tools.
type Snapshot struct {
	Servers []Server `json:"servers"`
}

// Registry builds an in-memory registry from live containers.
type Registry struct {
	source     ContainerSource
	discoverer Discoverer
	labels     LabelConfig

	mu       sync.RWMutex
	snapshot Snapshot
}

// New creates an empty registry backed by the supplied container source and MCP discoverer.
func New(source ContainerSource, discoverer Discoverer, labels LabelConfig) *Registry {
	return &Registry{source: source, discoverer: discoverer, labels: labels}
}

// Snapshot returns a copy of the latest successfully discovered services.
func (r *Registry) Snapshot() Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	servers := append([]Server(nil), r.snapshot.Servers...)
	return Snapshot{Servers: servers}
}

// Reconcile replaces the snapshot with MCP services discovered from running labelled containers.
func (r *Registry) Reconcile(ctx context.Context) error {
	containers, err := r.source.List(ctx)
	if err != nil {
		return fmt.Errorf("list MCP containers: %w", err)
	}

	servers := make([]Server, 0, len(containers))
	for _, container := range containers {
		server, ok := r.labels.serverFromLabels(container.Labels)
		if !ok {
			continue
		}
		tools, err := r.discoverer.Discover(ctx, server.URL)
		if err == nil {
			server.Enabled = true
			server.Tools = tools
		}
		servers = append(servers, server)
	}
	sort.Slice(servers, func(i, j int) bool { return servers[i].Name < servers[j].Name })

	r.mu.Lock()
	r.snapshot = Snapshot{Servers: servers}
	r.mu.Unlock()
	return nil
}
