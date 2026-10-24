package registry

import (
	"context"
	"encoding/json"
	"net"

	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestHTTPServerStopsWhenContextIsCancelled(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	require.NoError(t, listener.Close())

	cache := NewCache(nil, nil, testLabelConfig())
	server := NewHTTPServer(address, cache)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() { done <- server.Run(ctx) }()

	require.Eventually(t, func() bool {
		connection, err := net.DialTimeout("tcp", address, 10*time.Millisecond)
		if err != nil {
			return false
		}
		return connection.Close() == nil
	}, time.Second, 10*time.Millisecond)

	cancel()
	require.NoError(t, <-done)
}
