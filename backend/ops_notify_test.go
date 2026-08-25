package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNotifierOutboxBounded(t *testing.T) {
	notifier := newOpsNotifier(2)
	for i := 0; i < opsNotifierCap+1000; i++ {
		notifier.Queue("ops-r1", "mail", "notice")
	}
	if got := notifier.Pending(); got > opsNotifierCap {
		t.Fatalf("outbox grew to %d, cap %d", got, opsNotifierCap)
	}
}

func TestNotifierDispatchCancel(t *testing.T) {
	notifier := newOpsNotifier(2)
	for i := 0; i < 5; i++ {
		notifier.Queue("ops-r1", "mail", "notice")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan error, 1)
	go func() { done <- notifier.Dispatch(ctx, func(n OpsNotice) error { return nil }) }()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("dispatch err = %v, want context.Canceled", err)
		}
	case <-time.After(700 * time.Millisecond):
		t.Fatal("dispatch did not honor cancellation")
	}
}

func TestNotifierDispatchErrorStops(t *testing.T) {
	notifier := newOpsNotifier(2)
	notifier.Queue("ops-r1", "mail", "notice")
	done := make(chan error, 1)
	go func() { done <- notifier.Dispatch(context.Background(), func(n OpsNotice) error { return errors.New("send failed") }) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected send failure to stop dispatch")
		}
	case <-time.After(700 * time.Millisecond):
		t.Fatal("dispatch did not stop after send failure")
	}
}

func TestPageHitLogBounded(t *testing.T) {
	pageHitLog = nil
	handler := staticHandler()
	for i := 0; i < pageHitLogCap+500; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if got := len(pageHitLog); got > pageHitLogCap {
		t.Fatalf("page hit log grew to %d, cap %d", got, pageHitLogCap)
	}
}
