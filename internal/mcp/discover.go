// Package mcp adapts the MCP SDK for service discovery.
package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/faisal-memon/mcp-registry/internal/registry"
)

// Discoverer uses the MCP Streamable HTTP transport to list a server's tools.
type Discoverer struct{}

// NewDiscoverer creates an MCP discovery client for the registry cache.
func NewDiscoverer() Discoverer {
	return Discoverer{}
}

// Discover connects briefly to one MCP endpoint and returns its advertised tools.
func (Discoverer) Discover(ctx context.Context, endpoint string) ([]registry.Tool, error) {
	client := sdk.NewClient(&sdk.Implementation{Name: "mcp-registry", Version: "v0.1.0"}, nil)
	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{
		Endpoint:             endpoint,
		DisableStandaloneSSE: true,
		MaxRetries:           0,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", endpoint, err)
	}
	defer func() { _ = session.Close() }()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list tools from %s: %w", endpoint, err)
	}
	tools := make([]registry.Tool, 0, len(result.Tools))
	for _, tool := range result.Tools {
		tools = append(tools, registry.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		})
	}
	return tools, nil
}
