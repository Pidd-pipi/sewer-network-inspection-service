package main

import (
	"sync"
	"time"
)

type OpsMetric struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	LatencyMS  int    `json:"latencyMs"`
	RecordedAt string `json:"recordedAt"`
}

type OpsMetricSummary struct {
	Requests     int `json:"requests"`
	Errors       int `json:"errors"`
	AvgLatencyMS int `json:"avgLatencyMs"`
}

type OpsMetrics struct {
	mu       sync.RWMutex
	samples  []OpsMetric
	requests int
	errors   int
}

const opsMetricsCap = 5000

func newOpsMetrics() *OpsMetrics {
	return &OpsMetrics{samples: make([]OpsMetric, 0, 64)}
}

func (m *OpsMetrics) Record(method, path string, status, latencyMS int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.samples = append(m.samples, OpsMetric{
		Method:     method,
		Path:       path,
		Status:     status,
		LatencyMS:  latencyMS,
		RecordedAt: time.Now().UTC().Format(time.RFC3339Nano),
	})
	if len(m.samples) > opsMetricsCap {
		// Drop the oldest samples to keep the metrics buffer bounded.
		m.samples = m.samples[len(m.samples)-opsMetricsCap:]
	}
	m.requests++
	if status >= 500 {
		m.errors++
	}
}

func (m *OpsMetrics) Snapshot() []OpsMetric {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]OpsMetric, len(m.samples))
	copy(out, m.samples)
	return out
}

func (m *OpsMetrics) Summary() OpsMetricSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()
	total := len(m.samples)
	if total == 0 {
		return OpsMetricSummary{}
	}
	var latency int
	for _, sample := range m.samples {
		latency += sample.LatencyMS
	}
	return OpsMetricSummary{Requests: m.requests, Errors: m.errors, AvgLatencyMS: latency / total}
}

func (m *OpsMetrics) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.samples = m.samples[:0]
	m.requests = 0
	m.errors = 0
}
