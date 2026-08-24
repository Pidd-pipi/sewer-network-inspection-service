package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func newRouter(store *InspectionStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler("sewer-network-inspection-service"))
	mux.HandleFunc("/api/inspections", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/inspections" {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		writeJSON(w, http.StatusOK, store.list())
	})
	mux.HandleFunc("/api/inspections/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/inspections/")
		if !strings.HasSuffix(path, "/status") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "inspection status path not found"})
			return
		}
		id := strings.TrimSuffix(path, "/status")
		if id == "" || strings.Contains(id, "/") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "inspection not found"})
			return
		}
		var change StatusChange
		if err := json.NewDecoder(r.Body).Decode(&change); err != nil || validateInspectionStatus(change.Status) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid status is required"})
			return
		}
		item, err := store.changeStatus(id, change.Status)
		if errors.Is(err, errInspectionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, item)
	})
	return mux
}
