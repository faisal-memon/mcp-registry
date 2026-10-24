package registry

import (
	"fmt"
	"strings"
)

// LabelConfig describes the Docker labels that identify MCP containers and their endpoints.
type LabelConfig struct {
	EnabledKey   string
	EnabledValue string
	PortKey      string
	PathKey      string
	DefaultPath  string
}

func (c LabelConfig) serverFromContainer(container Container) (Server, bool) {
	if container.Name == "" || container.Labels[c.EnabledKey] != c.EnabledValue {
		return Server{}, false
	}
	port := container.Labels[c.PortKey]
	if port == "" {
		return Server{}, false
	}
	path := container.Labels[c.PathKey]
	if path == "" {
		path = c.DefaultPath
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return Server{
		Name: container.Name,
		URL:  fmt.Sprintf("http://%s:%s%s", container.Name, port, path),
	}, true
}
