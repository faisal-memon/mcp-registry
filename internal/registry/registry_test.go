package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSource struct {
	containers []Container
	err        error
}

func (s fakeSource) List(context.Context) ([]Container, error) {
	return s.containers, s.err
}

type fakeDiscoverer struct {
	toolsByURL map[string][]Tool
	errByURL   map[string]error
	calls      []string
}

func (d *fakeDiscoverer) Discover(_ context.Context, endpoint string) ([]Tool, error) {
	d.calls = append(d.calls, endpoint)
	if err := d.errByURL[endpoint]; err != nil {
		return nil, err
	}
	return d.toolsByURL[endpoint], nil
}

func TestReconcileBuildsSortedSnapshotFromEnabledContainers(t *testing.T) {
	config := testLabelConfig()
	discoverer := &fakeDiscoverer{toolsByURL: map[string][]Tool{
		"http://muni-mcp:8787/mcp": {{Name: "nearest_stops", InputSchema: map[string]any{"type": "object"}}},
		"http://news-mcp:8080/mcp": {{Name: "latest_headlines", InputSchema: map[string]any{"type": "object"}}},
	}}
	service := New(fakeSource{containers: []Container{
		{ID: "ignored", Name: "ignored", Labels: map[string]string{}},
		{ID: "muni", Name: "muni-mcp", Labels: labels(config, "8787", "")},
		{ID: "news", Name: "news-mcp", Labels: labels(config, "8080", "")},
		{ID: "incomplete", Name: "incomplete-mcp", Labels: map[string]string{config.EnabledKey: config.EnabledValue}},
	}}, discoverer, config)

	require.NoError(t, service.Reconcile(context.Background()))

	want := Snapshot{Servers: []Server{
		{Name: "muni-mcp", URL: "http://muni-mcp:8787/mcp", Enabled: true, Tools: discoverer.toolsByURL["http://muni-mcp:8787/mcp"]},
		{Name: "news-mcp", URL: "http://news-mcp:8080/mcp", Enabled: true, Tools: discoverer.toolsByURL["http://news-mcp:8080/mcp"]},
	}}
	assert.Equal(t, want, service.Snapshot())
	assert.Equal(t, []string{"http://muni-mcp:8787/mcp", "http://news-mcp:8080/mcp"}, discoverer.calls)
}

func TestReconcileUsesMCPPathLabelWhenPresent(t *testing.T) {
	config := testLabelConfig()
	discoverer := &fakeDiscoverer{toolsByURL: map[string][]Tool{"http://weather-mcp:9000/custom": nil}}
	service := New(fakeSource{containers: []Container{{
		Name:   "weather-mcp",
		Labels: labels(config, "9000", "/custom"),
	}}}, discoverer, config)

	require.NoError(t, service.Reconcile(context.Background()))
	assert.Equal(t, []string{"http://weather-mcp:9000/custom"}, discoverer.calls)
}

func TestReconcileMarksUnavailableServiceWithoutDiscardingHealthyServices(t *testing.T) {
	config := testLabelConfig()
	discoverer := &fakeDiscoverer{
		toolsByURL: map[string][]Tool{"http://healthy-mcp:8080/mcp": {{Name: "available", InputSchema: map[string]any{"type": "object"}}}},
		errByURL:   map[string]error{"http://unavailable-mcp:8080/mcp": errors.New("connection refused")},
	}
	service := New(fakeSource{containers: []Container{
		{Name: "healthy-mcp", Labels: labels(config, "8080", "")},
		{Name: "unavailable-mcp", Labels: labels(config, "8080", "")},
	}}, discoverer, config)

	require.NoError(t, service.Reconcile(context.Background()))

	got := service.Snapshot()
	require.Len(t, got.Servers, 2)
	assert.True(t, got.Servers[0].Enabled)
	assert.False(t, got.Servers[1].Enabled)
}

func TestReconcileKeepsLastSnapshotWhenContainerListingFails(t *testing.T) {
	config := testLabelConfig()
	discoverer := &fakeDiscoverer{toolsByURL: map[string][]Tool{"http://muni-mcp:8787/mcp": nil}}
	service := New(fakeSource{containers: []Container{{Name: "muni-mcp", Labels: labels(config, "8787", "")}}}, discoverer, config)
	require.NoError(t, service.Reconcile(context.Background()))
	before := service.Snapshot()
	service.source = fakeSource{err: errors.New("Docker unavailable")}

	require.Error(t, service.Reconcile(context.Background()))
	assert.Equal(t, before, service.Snapshot())
}

func testLabelConfig() LabelConfig {
	return LabelConfig{
		EnabledKey:   "mcp.enabled",
		EnabledValue: "true",
		PortKey:      "mcp.port",
		PathKey:      "mcp.path",
		DefaultPath:  "/mcp",
	}
}

func labels(config LabelConfig, port, path string) map[string]string {
	return map[string]string{
		config.EnabledKey: config.EnabledValue,
		config.PortKey:    port,
		config.PathKey:    path,
	}
}
