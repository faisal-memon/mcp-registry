// Package docker adapts Docker's API to the registry's small container interface.
package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/moby/client"

	"github.com/faisal-memon/mcp-registry/internal/registry"
)

// Source lists running containers from the local Docker daemon.
type Source struct {
	client *client.Client
}

// NewSource connects to Docker using its standard environment configuration.
func NewSource() (*Source, error) {
	dockerClient, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("create Docker client: %w", err)
	}
	return &Source{client: dockerClient}, nil
}

// List returns only running containers. Label filtering stays in the registry so it is testable.
func (s *Source) List(ctx context.Context) ([]registry.Container, error) {
	containers, err := s.client.ContainerList(ctx, client.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	result := make([]registry.Container, 0, len(containers.Items))
	for _, item := range containers.Items {
		if len(item.Names) == 0 {
			continue
		}
		result = append(result, registry.Container{
			ID:     strings.TrimPrefix(item.ID, "/"),
			Name:   strings.TrimPrefix(item.Names[0], "/"),
			Labels: item.Labels,
		})
	}
	return result, nil
}
