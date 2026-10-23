package registry

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestCacheReconcilesSortedSnapshotFromEnabledContainers(t *testing.T) {
	config := testLabelConfig()
	discoverer := &fakeDiscoverer{toolsByURL: map[string][]Tool{
		"http://muni-mcp:8787/mcp": {{Name: "nearest_stops", InputSchema: map[string]any{"type": "object"}}},
		"http://news-mcp:8080/mcp": {{Name: "latest_headlines", InputSchema: map[string]any{"type": "object"}}},
	}}
	service := NewCache(fakeSource{containers: []Container{
		{ID: "ignored", Name: "ignored", Labels: map[string]string{}},
		{ID: "muni", Name: "muni-mcp", Labels: labels(config, "8787", "")},
		{ID: "news", Name: "news-mcp", Labels: labels(config, "8080", "")},
		{ID: "incomplete", Name: "incomplete-mcp", Labels: map[string]string{config.EnabledKey: config.EnabledValue}},
	}}, discoverer, config)

	require.NoError(t, service.reconcile(context.Background()))

	want := Snapshot{Servers: []Server{
		{Name: "muni-mcp", URL: "http://muni-mcp:8787/mcp", Enabled: true, Tools: discoverer.toolsByURL["http://muni-mcp:8787/mcp"]},
		{Name: "news-mcp", URL: "http://news-mcp:8080/mcp", Enabled: true, Tools: discoverer.toolsByURL["http://news-mcp:8080/mcp"]},
	}}
	assert.Equal(t, want, service.Snapshot())
	assert.Equal(t, []string{"http://muni-mcp:8787/mcp", "http://news-mcp:8080/mcp"}, discoverer.calls)
}

func TestCacheUsesMCPPathLabelWhenPresent(t *testing.T) {
	config := testLabelConfig()
	discoverer := &fakeDiscoverer{toolsByURL: map[string][]Tool{"http://weather-mcp:9000/custom": nil}}
	service := NewCache(fakeSource{containers: []Container{{
		Name:   "weather-mcp",
		Labels: labels(config, "9000", "/custom"),
	}}}, discoverer, config)

	require.NoError(t, service.reconcile(context.Background()))
	assert.Equal(t, []string{"http://weather-mcp:9000/custom"}, discoverer.calls)
}

func TestCacheMarksUnavailableServiceWithoutDiscardingHealthyServices(t *testing.T) {
	config := testLabelConfig()
	discoverer := &fakeDiscoverer{
		toolsByURL: map[string][]Tool{"http://healthy-mcp:8080/mcp": {{Name: "available", InputSchema: map[string]any{"type": "object"}}}},
		errByURL:   map[string]error{"http://unavailable-mcp:8080/mcp": errors.New("connection refused")},
	}
	service := NewCache(fakeSource{containers: []Container{
		{Name: "healthy-mcp", Labels: labels(config, "8080", "")},
		{Name: "unavailable-mcp", Labels: labels(config, "8080", "")},
	}}, discoverer, config)

	require.NoError(t, service.reconcile(context.Background()))

	got := service.Snapshot()
	require.Len(t, got.Servers, 2)
	assert.True(t, got.Servers[0].Enabled)
	assert.False(t, got.Servers[1].Enabled)
}

func TestCacheKeepsLastSnapshotWhenContainerListingFails(t *testing.T) {
	config := testLabelConfig()
	discoverer := &fakeDiscoverer{toolsByURL: map[string][]Tool{"http://muni-mcp:8787/mcp": nil}}
	service := NewCache(fakeSource{containers: []Container{{Name: "muni-mcp", Labels: labels(config, "8787", "")}}}, discoverer, config)
	require.NoError(t, service.reconcile(context.Background()))
	before := service.Snapshot()
	service.source = fakeSource{err: errors.New("Docker unavailable")}

	require.Error(t, service.reconcile(context.Background()))
	assert.Equal(t, before, service.Snapshot())
}

func TestCacheStartPopulatesSnapshot(t *testing.T) {
	config := testLabelConfig()
	cache := NewCache(
		fakeSource{containers: []Container{{Name: "muni-mcp", Labels: labels(config, "8787", "")}}},
		&fakeDiscoverer{toolsByURL: map[string][]Tool{"http://muni-mcp:8787/mcp": nil}},
		config,
	)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	require.NoError(t, cache.Start(ctx, time.Hour))

	assert.Equal(t, Snapshot{Servers: []Server{{
		Name: "muni-mcp", URL: "http://muni-mcp:8787/mcp", Enabled: true,
	}}}, cache.Snapshot())
}

func TestCacheStartRejectsNonPositiveRefreshInterval(t *testing.T) {
	cache := NewCache(nil, nil, testLabelConfig())

	assert.Error(t, cache.Start(context.Background(), 0))
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
