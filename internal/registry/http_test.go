package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerServesSnapshot(t *testing.T) {
	service := New(nil, nil)
	service.snapshot = Snapshot{Servers: []Server{{
		Name:    "muni",
		URL:     "http://muni:8787/mcp",
		Enabled: true,
		Tools:   []Tool{{Name: "nearest_stops", InputSchema: map[string]any{"type": "object"}}},
	}}}

	request := httptest.NewRequest(http.MethodGet, "/registry.json", nil)
	response := httptest.NewRecorder()
	service.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got Snapshot
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Servers) != 1 || got.Servers[0].Name != "muni" {
		t.Fatalf("response = %#v, want Muni server", got)
	}
}

func TestHandlerHealthCheck(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	New(nil, nil).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}
