package main

import (
	"context"
	"testing"
)

func TestNormalizeLabelsNonNil(t *testing.T) {
	store := newOpsStore(nil)
	rec := OpsRecord{ID: "ops-n1", Subject: "无标签任务", Owner: "zhang", Priority: OpsPriorityNormal}
	if err := store.Put(context.Background(), rec); err != nil {
		t.Fatal(err)
	}
	store.mu.RLock()
	stored := store.items["ops-n1"]
	store.mu.RUnlock()
	if stored.Labels == nil {
		t.Fatal("stored record still has nil labels map")
	}
	stored.Labels["site"] = "north"
	got, err := store.Get(context.Background(), "ops-n1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Labels["site"] != "north" {
		t.Fatal("label write to stored record not visible via Get")
	}
}

func TestCloneLabelsIsolated(t *testing.T) {
	original := OpsRecord{ID: "ops-n2", Subject: "克隆隔离", Owner: "li", Priority: OpsPriorityLow, Labels: map[string]string{"site": "north"}}
	clone := original.Clone()
	clone.Labels["site"] = "south"
	if original.Labels["site"] != "north" {
		t.Fatalf("clone mutation leaked into original: %v", original.Labels)
	}
}

func TestOpsSearchZeroPageNoPanic(t *testing.T) {
	svc := newOpsService(seedOpsRecords())
	page, err := svc.Search(context.Background(), OpsQuery{Page: 0, PageSize: 25})
	if err != nil {
		t.Fatal(err)
	}
	if page.Page != 1 {
		t.Fatalf("page = %d, want 1 (zero page should default)", page.Page)
	}
}

func TestOpsSearchHugePageNoPanic(t *testing.T) {
	svc := newOpsService(seedOpsRecords())
	page, err := svc.Search(context.Background(), OpsQuery{Page: 999, PageSize: 25})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 0 {
		t.Fatalf("huge page returned %d items, want empty", len(page.Items))
	}
}

