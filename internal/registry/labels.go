package registry

import "os"

const (
	defaultEnabledLabelKey     = "mcp.enabled"
	defaultEnabledLabelValue   = "true"
	defaultNameLabelKey        = "mcp.name"
	defaultURLLabelKey         = "mcp.url"
	defaultDescriptionLabelKey = "mcp.description"
)

// LabelConfig describes the Docker labels that identify and describe MCP containers.
type LabelConfig struct {
	EnabledKey     string
	EnabledValue   string
	NameKey        string
	URLKey         string
	DescriptionKey string
}

// DefaultLabelConfig returns neutral label names suitable for most Docker deployments.
func DefaultLabelConfig() LabelConfig {
	return LabelConfig{
		EnabledKey:     defaultEnabledLabelKey,
		EnabledValue:   defaultEnabledLabelValue,
		NameKey:        defaultNameLabelKey,
		URLKey:         defaultURLLabelKey,
		DescriptionKey: defaultDescriptionLabelKey,
	}
}

// LabelConfigFromEnvironment loads label names from environment variables, using neutral defaults.
func LabelConfigFromEnvironment() LabelConfig {
	return labelConfigFrom(os.Getenv)
}

func labelConfigFrom(lookup func(string) string) LabelConfig {
	defaults := DefaultLabelConfig()
	return LabelConfig{
		EnabledKey:     valueOrDefault(lookup("MCP_LABEL_ENABLED_KEY"), defaults.EnabledKey),
		EnabledValue:   valueOrDefault(lookup("MCP_LABEL_ENABLED_VALUE"), defaults.EnabledValue),
		NameKey:        valueOrDefault(lookup("MCP_LABEL_NAME_KEY"), defaults.NameKey),
		URLKey:         valueOrDefault(lookup("MCP_LABEL_URL_KEY"), defaults.URLKey),
		DescriptionKey: valueOrDefault(lookup("MCP_LABEL_DESCRIPTION_KEY"), defaults.DescriptionKey),
	}
}

func valueOrDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func (c LabelConfig) serverFromLabels(labels map[string]string) (Server, bool) {
	if labels[c.EnabledKey] != c.EnabledValue {
		return Server{}, false
	}
	name, url := labels[c.NameKey], labels[c.URLKey]
	if name == "" || url == "" {
		return Server{}, false
	}
	return Server{
		Name:        name,
		URL:         url,
		Description: labels[c.DescriptionKey],
	}, true
}
