package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInspectionHTTP(t *testing.T) {
	h := newRouter(newInspectionStore())
	t.Run("collection", func(t *testing.T) {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/api/inspections", nil))
		if r.Code != http.StatusOK {
			t.Fatalf("got %d", r.Code)
		}
	})
	t.Run("status change", func(t *testing.T) {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest(http.MethodPost, "/api/inspections/ins-101/status", bytes.NewBufferString(`{"status":"resolved"}`)))
		if r.Code != http.StatusOK {
			t.Fatalf("got %d", r.Code)
		}
	})
	t.Run("invalid status", func(t *testing.T) {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest(http.MethodPost, "/api/inspections/ins-101/status", bytes.NewBufferString(`{"status":"ignored"}`)))
		if r.Code != http.StatusBadRequest {
			t.Fatalf("got %d", r.Code)
		}
	})
	t.Run("missing item", func(t *testing.T) {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest(http.MethodPost, "/api/inspections/nope/status", bytes.NewBufferString(`{"status":"open"}`)))
		if r.Code != http.StatusNotFound {
			t.Fatalf("got %d", r.Code)
		}
	})
}
