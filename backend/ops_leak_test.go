package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDLogBounded(t *testing.T) {
	opsRequestLog = nil
	handler := requestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for i := 0; i < opsRequestLogCap+500; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	}
	if got := len(opsRequestLog); got > opsRequestLogCap {
		t.Fatalf("request id log grew to %d, cap %d", got, opsRequestLogCap)
	}
}

func TestRequestPathLogBounded(t *testing.T) {
	opsRequestPaths = nil
	metrics := newOpsMetrics()
	handler := opsMetricsMiddleware(metrics, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for i := 0; i < opsRequestLogCap+500; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	}
	if got := len(opsRequestPaths); got > opsRequestLogCap {
		t.Fatalf("request path log grew to %d, cap %d", got, opsRequestLogCap)
	}
}

func TestEnterpriseLatencyBounded(t *testing.T) {
	opsMiddlewareLatencySamples = nil
	handler := opsEnterpriseMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for i := 0; i < opsLatencySamplesCap+500; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	}
	if got := len(opsMiddlewareLatencySamples); got > opsLatencySamplesCap {
		t.Fatalf("enterprise latency samples grew to %d, cap %d", got, opsLatencySamplesCap)
	}
}

func TestMetricsLatencyBounded(t *testing.T) {
	opsMetricsLatencySamples = nil
	metrics := newOpsMetrics()
	handler := opsMetricsMiddleware(metrics, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for i := 0; i < opsLatencySamplesCap+500; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	}
	if got := len(opsMetricsLatencySamples); got > opsLatencySamplesCap {
		t.Fatalf("metrics latency samples grew to %d, cap %d", got, opsLatencySamplesCap)
	}
}
