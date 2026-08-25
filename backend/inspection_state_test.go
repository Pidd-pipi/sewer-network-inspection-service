package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInspectionResolvedTerminal(t *testing.T) {
	store := newInspectionStore()
	if _, err := store.changeStatus("ins-101", "resolved"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.changeStatus("ins-101", "open"); err == nil {
		t.Fatal("resolved inspection was reopened")
	}
}

func TestInspectionTransitionErrorChain(t *testing.T) {
	store := newInspectionStore()
	if _, err := store.changeStatus("ins-101", "resolved"); err != nil {
		t.Fatal(err)
	}
	_, err := store.changeStatus("ins-101", "open")
	if !errors.Is(err, errInspectionTransition) {
		t.Fatalf("err = %v, want errInspectionTransition chain", err)
	}
}

func TestInspectionHTTPInvalidTransition409(t *testing.T) {
	router := newRouter(newInspectionStore())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/inspections/ins-101/status", strings.NewReader(`{"status":"resolved"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("setup resolve failed: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/inspections/ins-101/status", strings.NewReader(`{"status":"open"}`)))
	if rec.Code != http.StatusConflict {
		t.Fatalf("invalid transition: got %d, want 409", rec.Code)
	}
}

func TestInspectionInvalidStatus400(t *testing.T) {
	router := newRouter(newInspectionStore())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/inspections/ins-101/status", strings.NewReader(`{"status":"closed"}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown status: got %d, want 400", rec.Code)
	}
}

func TestInspectionValidTransitionsOK(t *testing.T) {
	store := newInspectionStore()
	for _, to := range []string{"scheduled", "resolved"} {
		if _, err := store.changeStatus("ins-102", to); err != nil {
			t.Fatalf("valid transition to %s rejected: %v", to, err)
		}
	}
}
