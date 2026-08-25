package main

import (
	"context"
	"testing"
	"time"
)

func TestBatchCloseStaleErrorNoPhantomAudit(t *testing.T) {
	svc := newOpsService(seedOpsRecords())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.BatchCloseStale(ctx, time.Hour, "system"); err == nil {
		t.Fatal("expected error from canceled batch close")
	}
	if got := len(svc.Audit("batch")); got != 0 {
		t.Fatalf("batch_close audit recorded %d events despite failure, want 0", got)
	}
}

func TestBatchResolveFailureNotCounted(t *testing.T) {
	svc := newOpsService(seedOpsRecords())
	result, err := svc.BatchResolve(context.Background(), []string{"ops-201", "nope-404"}, "zhang")
	if err != nil {
		t.Fatal(err)
	}
	if result.Closed != 1 {
		t.Fatalf("Closed = %d, want 1", result.Closed)
	}
	if result.Failed != 1 {
		t.Fatalf("Failed = %d, want 1 (missing record must be counted as failure)", result.Failed)
	}
}

func TestWorkerCloseReturns(t *testing.T) {
	svc := newOpsService(seedOpsRecords())
	worker := newOpsWorker(svc, 10*time.Millisecond, time.Hour, "system")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Run(ctx)
	done := make(chan struct{})
	go func() {
		worker.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(700 * time.Millisecond):
		t.Fatal("worker.Close did not return")
	}
}

func TestWorkerCloseStopsRunning(t *testing.T) {
	svc := newOpsService(seedOpsRecords())
	worker := newOpsWorker(svc, 10*time.Millisecond, time.Hour, "system")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Run(ctx)
	done := make(chan struct{})
	go func() {
		worker.Close()
		close(done)
	}()
	deadline := time.Now().Add(700 * time.Millisecond)
	for time.Now().Before(deadline) {
		if !worker.Running() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("worker still running after Close")
}

func TestBatchResolveCancelled(t *testing.T) {
	svc := newOpsService(seedOpsRecords())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.BatchResolve(ctx, []string{"ops-201"}, "zhang"); err == nil {
		t.Fatal("expected cancellation error from canceled batch resolve")
	}
}

func TestWorkerCloseIdempotent(t *testing.T) {
	svc := newOpsService(seedOpsRecords())
	worker := newOpsWorker(svc, 10*time.Millisecond, time.Hour, "system")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Run(ctx)
	done := make(chan struct{})
	go func() {
		worker.Close()
		worker.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(700 * time.Millisecond):
		t.Fatal("double worker.Close did not return cleanly")
	}
}
