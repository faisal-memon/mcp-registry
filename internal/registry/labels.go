package registry

// LabelConfig describes the Docker labels that identify and describe MCP containers.
type LabelConfig struct {
	EnabledKey     string
	EnabledValue   string
	NameKey        string
	URLKey         string
	DescriptionKey string
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
