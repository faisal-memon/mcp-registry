// Package registry discovers labelled MCP containers and exposes their tools.
package registry

// Container is the small part of Docker container metadata the registry needs.
type Container struct {
	ID     string
	Name   string
	Labels map[string]string
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
