package main

import (
	"testing"
	"time"

	"github.com/faisal-memon/mcp-registry/internal/registry"
	"github.com/stretchr/testify/assert"
)

func TestConfigFromEnvironment(t *testing.T) {
	t.Setenv("LISTEN_ADDR", "127.0.0.1:9090")
	t.Setenv("RECONCILE_INTERVAL", "45s")
	t.Setenv("MCP_LABEL_ENABLED_KEY", "example.com/mcp")
	t.Setenv("MCP_LABEL_ENABLED_VALUE", "enabled")
	t.Setenv("MCP_LABEL_NAME_KEY", "example.com/name")
	t.Setenv("MCP_LABEL_URL_KEY", "example.com/url")
	t.Setenv("MCP_LABEL_DESCRIPTION_KEY", "example.com/description")

	assert.Equal(t, Config{
		ListenAddress:     "127.0.0.1:9090",
		ReconcileInterval: 45 * time.Second,
		Labels: registry.LabelConfig{
			EnabledKey:     "example.com/mcp",
			EnabledValue:   "enabled",
			NameKey:        "example.com/name",
			URLKey:         "example.com/url",
			DescriptionKey: "example.com/description",
		},
	}, configFromEnvironment())
}

func TestConfigFromEnvironmentUsesDefaults(t *testing.T) {
	for _, name := range []string{
		"LISTEN_ADDR",
		"RECONCILE_INTERVAL",
		"MCP_LABEL_ENABLED_KEY",
		"MCP_LABEL_ENABLED_VALUE",
		"MCP_LABEL_NAME_KEY",
		"MCP_LABEL_URL_KEY",
		"MCP_LABEL_DESCRIPTION_KEY",
	} {
		t.Setenv(name, "")
	}

	assert.Equal(t, Config{
		ListenAddress:     ":8080",
		ReconcileInterval: 30 * time.Second,
		Labels: registry.LabelConfig{
			EnabledKey:     "mcp.enabled",
			EnabledValue:   "true",
			NameKey:        "mcp.name",
			URLKey:         "mcp.url",
			DescriptionKey: "mcp.description",
		},
	}, configFromEnvironment())
}
