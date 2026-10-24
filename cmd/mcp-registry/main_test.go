package main

import (
	"testing"
	"time"

	"github.com/faisal-memon/mcp-registry/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigFromEnvironment(t *testing.T) {
	t.Setenv("LISTEN_ADDR", "127.0.0.1:9090")
	t.Setenv("RECONCILE_INTERVAL", "45s")
	t.Setenv("MCP_LABEL_ENABLED_KEY", "example.com/mcp")
	t.Setenv("MCP_LABEL_ENABLED_VALUE", "enabled")
	t.Setenv("MCP_LABEL_PORT_KEY", "example.com/port")
	t.Setenv("MCP_LABEL_PATH_KEY", "example.com/path")
	t.Setenv("MCP_DEFAULT_PATH", "/custom")

	config, err := configFromEnvironment()
	require.NoError(t, err)
	assert.Equal(t, Config{
		ListenAddress:     "127.0.0.1:9090",
		ReconcileInterval: 45 * time.Second,
		Labels: registry.LabelConfig{
			EnabledKey:   "example.com/mcp",
			EnabledValue: "enabled",
			PortKey:      "example.com/port",
			PathKey:      "example.com/path",
			DefaultPath:  "/custom",
		},
	}, config)
}

func TestConfigFromEnvironmentUsesDefaults(t *testing.T) {
	for _, name := range []string{
		"LISTEN_ADDR",
		"RECONCILE_INTERVAL",
		"MCP_LABEL_ENABLED_KEY",
		"MCP_LABEL_ENABLED_VALUE",
		"MCP_LABEL_PORT_KEY",
		"MCP_LABEL_PATH_KEY",
		"MCP_DEFAULT_PATH",
	} {
		t.Setenv(name, "")
	}

	config, err := configFromEnvironment()
	require.NoError(t, err)
	assert.Equal(t, Config{
		ListenAddress:     ":8080",
		ReconcileInterval: 30 * time.Second,
		Labels: registry.LabelConfig{
			EnabledKey:   "mcp.enabled",
			EnabledValue: "true",
			PortKey:      "mcp.port",
			PathKey:      "mcp.path",
			DefaultPath:  "/mcp",
		},
	}, config)
}

func TestConfigFromEnvironmentRejectsInvalidRefreshInterval(t *testing.T) {
	t.Setenv("RECONCILE_INTERVAL", "never")

	_, err := configFromEnvironment()

	assert.Error(t, err)
}
