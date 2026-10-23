package registry

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
