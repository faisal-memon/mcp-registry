package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerServesSnapshot(t *testing.T) {
	service := NewCache(nil, nil, testLabelConfig())
	service.snapshot = Snapshot{Servers: []Server{{
		Name:    "muni",
		URL:     "http://muni:8787/mcp",
		Enabled: true,
		Tools:   []Tool{{Name: "nearest_stops", InputSchema: map[string]any{"type": "object"}}},
	}}}

	request := httptest.NewRequest(http.MethodGet, "/registry.json", nil)
	response := httptest.NewRecorder()
	service.Handler().ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	var got Snapshot
	require.NoError(t, json.NewDecoder(response.Body).Decode(&got))
	require.Len(t, got.Servers, 1)
	assert.Equal(t, "muni", got.Servers[0].Name)
}

func TestHandlerHealthCheck(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	NewCache(nil, nil, testLabelConfig()).Handler().ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
}
