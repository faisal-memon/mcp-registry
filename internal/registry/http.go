package registry

import (
	"encoding/json"
	"net/http"
)

// Handler serves the latest registry snapshot and a basic health check.
func (r *Registry) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /registry.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(r.Snapshot()); err != nil {
			http.Error(w, "encode registry", http.StatusInternalServerError)
		}
	})
	return mux
}
