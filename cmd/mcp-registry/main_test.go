package main

import (
	"testing"

	"github.com/faisal-memon/mcp-registry/internal/registry"
	"github.com/stretchr/testify/assert"
)

func TestLabelConfigFromEnvironment(t *testing.T) {
	t.Setenv("MCP_LABEL_ENABLED_KEY", "example.com/mcp")
	t.Setenv("MCP_LABEL_ENABLED_VALUE", "enabled")
	t.Setenv("MCP_LABEL_NAME_KEY", "example.com/name")
	t.Setenv("MCP_LABEL_URL_KEY", "example.com/url")
	t.Setenv("MCP_LABEL_DESCRIPTION_KEY", "example.com/description")

	assert.Equal(t, registry.LabelConfig{
		EnabledKey:     "example.com/mcp",
		EnabledValue:   "enabled",
		NameKey:        "example.com/name",
		URLKey:         "example.com/url",
		DescriptionKey: "example.com/description",
	}, labelConfigFromEnvironment())
}

func TestLabelConfigFromEnvironmentUsesDefaults(t *testing.T) {
	for _, name := range []string{
		"MCP_LABEL_ENABLED_KEY",
		"MCP_LABEL_ENABLED_VALUE",
		"MCP_LABEL_NAME_KEY",
		"MCP_LABEL_URL_KEY",
		"MCP_LABEL_DESCRIPTION_KEY",
	} {
		t.Setenv(name, "")
	}
	assert.Equal(t, registry.DefaultLabelConfig(), labelConfigFromEnvironment())
}
