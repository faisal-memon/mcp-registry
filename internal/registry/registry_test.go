package registry

import (
	"context"
	"errors"
	"reflect"
	"testing"
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
	config := DefaultLabelConfig()
	discoverer := &fakeDiscoverer{toolsByURL: map[string][]Tool{
		"http://muni:8787/mcp": {{Name: "nearest_stops", InputSchema: map[string]any{"type": "object"}}},
		"http://news:8080/mcp": {{Name: "latest_headlines", InputSchema: map[string]any{"type": "object"}}},
	}}
	service := New(fakeSource{containers: []Container{
		{ID: "ignored", Labels: map[string]string{}},
		{ID: "muni", Labels: labels(config, "muni", "http://muni:8787/mcp", "Nearby Muni stops")},
		{ID: "news", Labels: labels(config, "news", "http://news:8080/mcp", "Local news")},
		{ID: "incomplete", Labels: map[string]string{config.EnabledKey: config.EnabledValue, config.NameKey: "missing-url"}},
	}}, discoverer, config)

	if err := service.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	got := service.Snapshot()
	want := Snapshot{Servers: []Server{
		{Name: "muni", URL: "http://muni:8787/mcp", Enabled: true, Description: "Nearby Muni stops", Tools: discoverer.toolsByURL["http://muni:8787/mcp"]},
		{Name: "news", URL: "http://news:8080/mcp", Enabled: true, Description: "Local news", Tools: discoverer.toolsByURL["http://news:8080/mcp"]},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Snapshot() = %#v, want %#v", got, want)
	}
	if wantCalls := []string{"http://muni:8787/mcp", "http://news:8080/mcp"}; !reflect.DeepEqual(discoverer.calls, wantCalls) {
		t.Fatalf("Discover calls = %v, want %v", discoverer.calls, wantCalls)
	}
}

func TestReconcileMarksUnavailableServiceWithoutDiscardingHealthyServices(t *testing.T) {
	config := DefaultLabelConfig()
	discoverer := &fakeDiscoverer{
		toolsByURL: map[string][]Tool{"http://healthy/mcp": {{Name: "available", InputSchema: map[string]any{"type": "object"}}}},
		errByURL:   map[string]error{"http://unavailable/mcp": errors.New("connection refused")},
	}
	service := New(fakeSource{containers: []Container{
		{Labels: labels(config, "healthy", "http://healthy/mcp", "")},
		{Labels: labels(config, "unavailable", "http://unavailable/mcp", "")},
	}}, discoverer, config)

	if err := service.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	got := service.Snapshot()
	if len(got.Servers) != 2 {
		t.Fatalf("Snapshot() = %#v, want healthy and unavailable services", got)
	}
	if !got.Servers[0].Enabled || got.Servers[1].Enabled {
		t.Fatalf("enabled states = %t, %t; want true, false", got.Servers[0].Enabled, got.Servers[1].Enabled)
	}
}

func TestReconcileKeepsLastSnapshotWhenContainerListingFails(t *testing.T) {
	config := DefaultLabelConfig()
	discoverer := &fakeDiscoverer{toolsByURL: map[string][]Tool{"http://muni/mcp": nil}}
	service := New(fakeSource{containers: []Container{{Labels: labels(config, "muni", "http://muni/mcp", "")}}}, discoverer, config)
	if err := service.Reconcile(context.Background()); err != nil {
		t.Fatalf("initial Reconcile() error = %v", err)
	}
	before := service.Snapshot()
	service.source = fakeSource{err: errors.New("Docker unavailable")}

	if err := service.Reconcile(context.Background()); err == nil {
		t.Fatal("Reconcile() error = nil, want error")
	}
	if got := service.Snapshot(); !reflect.DeepEqual(got, before) {
		t.Fatalf("Snapshot() after failed list = %#v, want %#v", got, before)
	}
}

func labels(config LabelConfig, name, url, description string) map[string]string {
	return map[string]string{
		config.EnabledKey:     config.EnabledValue,
		config.NameKey:        name,
		config.URLKey:         url,
		config.DescriptionKey: description,
	}
}

func TestLabelConfigFromEnvironmentUsesOverrides(t *testing.T) {
	config := labelConfigFrom(func(name string) string {
		return map[string]string{
			"MCP_LABEL_ENABLED_KEY":     "example.com/mcp",
			"MCP_LABEL_ENABLED_VALUE":   "enabled",
			"MCP_LABEL_NAME_KEY":        "example.com/name",
			"MCP_LABEL_URL_KEY":         "example.com/url",
			"MCP_LABEL_DESCRIPTION_KEY": "example.com/description",
		}[name]
	})
	labels := labels(config, "weather", "http://weather/mcp", "Weather")
	server, ok := config.serverFromLabels(labels)
	if !ok {
		t.Fatal("serverFromLabels() = false, want true")
	}
	if server.Name != "weather" || server.URL != "http://weather/mcp" || server.Description != "Weather" {
		t.Fatalf("serverFromLabels() = %#v", server)
	}
}
