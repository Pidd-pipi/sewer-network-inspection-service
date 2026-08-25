package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOpsContextParentCancel(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	derived, derivedCancel := opsContext(parent, 5*time.Second)
	defer derivedCancel()
	cancel()
	select {
	case <-derived.Done():
		if !errors.Is(derived.Err(), context.Canceled) {
			t.Fatalf("derived err = %v, want context.Canceled", derived.Err())
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("derived context did not observe parent cancellation")
	}
}

func TestOpsDelayCancelFast(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan error, 1)
	go func() { done <- opsDelay(ctx, 5*time.Second) }()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("opsDelay err = %v, want context.Canceled", err)
		}
	case <-time.After(700 * time.Millisecond):
		t.Fatal("opsDelay did not return promptly after cancellation")
	}
}

func TestOpsTransitionCancelledNotApplied(t *testing.T) {
	svc := newOpsService(seedOpsRecords())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.Transition(ctx, "ops-202", 0, OpsStatusActive, "zhang"); err == nil {
		t.Fatal("expected error on canceled transition")
	}
	rec, err := svc.Get(context.Background(), "ops-202")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status != OpsStatusQueued {
		t.Fatalf("status = %s, want queued (canceled transition was still applied)", rec.Status)
	}
}

func TestOpsCreateCancelledNotStored(t *testing.T) {
	svc := newOpsService(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rec := OpsRecord{ID: "ops-new", Subject: "新任务", Owner: "zhang", Priority: OpsPriorityNormal, Labels: map[string]string{"site": "north"}}
	if _, err := svc.Create(ctx, rec); err == nil {
		t.Fatal("expected error on canceled create")
	}
	if svc.Count() != 0 {
		t.Fatal("record was stored despite canceled context")
	}
}

func TestOpsSearchCancelled(t *testing.T) {
	svc := newOpsService(seedOpsRecords())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.Search(ctx, OpsQuery{}); err == nil {
		t.Fatal("expected error from canceled search")
	}
}
