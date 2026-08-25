package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsRetentionBounded(t *testing.T) {
	metrics := newOpsMetrics()
	for i := 0; i < opsMetricsCap+2000; i++ {
		metrics.Record("GET", "/api/x", http.StatusOK, 3)
	}
	if got := len(metrics.Snapshot()); got > opsMetricsCap {
		t.Fatalf("metrics samples grew to %d, cap %d", got, opsMetricsCap)
	}
}

func TestMetricsSnapshotIsolated(t *testing.T) {
	metrics := newOpsMetrics()
	metrics.Record("GET", "/a", http.StatusOK, 5)
	snap := metrics.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("snapshot len = %d", len(snap))
	}
	snap[0].Path = "tampered"
	again := metrics.Snapshot()
	if again[0].Path != "/a" {
		t.Fatalf("snapshot mutation leaked into metrics: %q", again[0].Path)
	}
}

func TestMetricsSummaryCountsErrors(t *testing.T) {
	metrics := newOpsMetrics()
	metrics.Record("GET", "/a", http.StatusOK, 5)
	metrics.Record("GET", "/b", http.StatusInternalServerError, 9)
	metrics.Record("GET", "/c", http.StatusBadGateway, 7)
	summary := metrics.Summary()
	if summary.Errors != 2 {
		t.Fatalf("summary errors = %d, want 2", summary.Errors)
	}
	if summary.Requests != 3 {
		t.Fatalf("summary requests = %d, want 3", summary.Requests)
	}
}

func TestHealthMethodNotAllowed(t *testing.T) {
	router := http.NewServeMux()
	router.Handle("/healthz", healthHandler("sewer-network-inspection-service"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/healthz", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /healthz: got %d, want 405", rec.Code)
	}
}
