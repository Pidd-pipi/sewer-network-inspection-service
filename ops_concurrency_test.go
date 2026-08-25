package main

import (
	"context"
	"sync"
	"testing"
)

func seedInspectionRecord(id string) OpsRecord {
	return OpsRecord{
		ID:        id,
		Subject:   "MH-" + id + " 复检",
		Owner:     "zhang",
		Status:    OpsStatusActive,
		Priority:  OpsPriorityHigh,
		Revision:  1,
		Labels:    map[string]string{"site": "north", "operator": "zhang"},
		CreatedAt: "2026-08-20T08:00:00Z",
		UpdatedAt: "2026-08-20T08:00:00Z",
	}
}

func TestOpsListGetConcurrentPutRace(t *testing.T) {
	store := newOpsStore([]OpsRecord{seedInspectionRecord("ops-r1")})
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				_, _ = store.List(context.Background())
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				_, _ = store.Get(context.Background(), "ops-r1")
			}
		}()
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				rec, err := store.Get(context.Background(), "ops-r1")
				if err != nil {
					continue
				}
				rec.Labels["operator"] = "li"
				_ = store.Update(context.Background(), rec, 0)
			}
		}()
	}
	close(start)
	wg.Wait()
}

func TestAuditConcurrentAddRace(t *testing.T) {
	audit := newOpsAudit()
	const workers = 8
	const perWorker = 100
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start
			for j := 0; j < perWorker; j++ {
				audit.Add("ops-r1", "status_changed", "zhang")
			}
		}(i)
	}
	close(start)
	wg.Wait()
	if got := audit.Count(); got != workers*perWorker {
		t.Fatalf("audit count = %d, want %d (lost updates)", got, workers*perWorker)
	}
}

func TestAuditForConcurrentAddRace(t *testing.T) {
	audit := newOpsAudit()
	for i := 0; i < 10; i++ {
		audit.Add("ops-r1", "created", "zhang")
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				_ = audit.For("ops-r1")
			}
		}()
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				audit.Add("ops-r1", "status_changed", "li")
			}
		}()
	}
	close(start)
	wg.Wait()
}

func TestAuditRetentionBounded(t *testing.T) {
	audit := newOpsAudit()
	for i := 0; i < opsAuditCap+2000; i++ {
		audit.Add("ops-r1", "created", "zhang")
	}
	if got := audit.Count(); got > opsAuditCap {
		t.Fatalf("audit grew to %d, retention cap is %d", got, opsAuditCap)
	}
}

func TestOpsReadsReturnCopies(t *testing.T) {
	store := newOpsStore([]OpsRecord{seedInspectionRecord("ops-r1")})
	items, err := store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("list length = %d, want 1", len(items))
	}
	items[0].Labels["tampered"] = "yes"
	got, err := store.Get(context.Background(), "ops-r1")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.Labels["tampered"]; ok {
		t.Fatal("mutation of list result leaked into store")
	}
	got.Labels["tampered2"] = "yes"
	got2, err := store.Get(context.Background(), "ops-r1")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got2.Labels["tampered2"]; ok {
		t.Fatal("mutation of get result leaked into store")
	}
}
