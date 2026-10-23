package registry

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPDiscoverer uses the MCP Streamable HTTP transport to list a server's tools.
type MCPDiscoverer struct{}

// Discover connects briefly to one MCP endpoint and returns its advertised tools.
func (MCPDiscoverer) Discover(ctx context.Context, endpoint string) ([]Tool, error) {
	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-registry", Version: "v0.1.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
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
	tools := make([]Tool, 0, len(result.Tools))
	for _, tool := range result.Tools {
		tools = append(tools, Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		})
	}
	return tools, nil
}
