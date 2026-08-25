package main

import (
	"strings"
	"testing"
)

func TestStateHistoryAppendIsolation(t *testing.T) {
	m := newOpsStateMachine()
	if err := m.Move(OpsStatusQueued, OpsStatusActive, "first"); err != nil {
		t.Fatal(err)
	}
	h := m.History()
	h[0].Reason = "tampered"
	if got := m.History()[0].Reason; got != "first" {
		t.Fatalf("history mutated through returned slice: %q", got)
	}
}

func TestStateHistoryRejectsFailedMove(t *testing.T) {
	m := newOpsStateMachine()
	if err := m.Move(OpsStatusClosed, OpsStatusActive, "illegal"); err == nil {
		t.Fatal("expected transition error")
	}
	if got := len(m.History()); got != 0 {
		t.Fatalf("failed move recorded %d history entries, want 0", got)
	}
}

func TestStateHistoryNoOpNotRecorded(t *testing.T) {
	m := newOpsStateMachine()
	if err := m.Move(OpsStatusActive, OpsStatusActive, "noop"); err != nil {
		t.Fatal(err)
	}
	if got := len(m.History()); got != 0 {
		t.Fatalf("no-op move recorded %d history entries, want 0", got)
	}
}

func TestStateLastOnEmptyNoPanic(t *testing.T) {
	m := newOpsStateMachine()
	_, ok := m.Last()
	if ok {
		t.Fatal("empty history reported a last transition")
	}
}

func TestExportDoesNotMutateInput(t *testing.T) {
	items := []OpsRecord{
		{ID: "ops-a", Subject: "低优先级", Owner: "zhang", Status: OpsStatusQueued, Priority: OpsPriorityLow, UpdatedAt: "2026-08-21T09:00:00Z"},
		{ID: "ops-b", Subject: "高优先级", Owner: "li", Status: OpsStatusActive, Priority: OpsPriorityCritical, UpdatedAt: "2026-08-20T09:00:00Z"},
	}
	before := []string{items[0].ID, items[1].ID}
	_ = exportRecordsCSV(items)
	if items[0].ID != before[0] || items[1].ID != before[1] {
		t.Fatalf("export reordered caller slice: %s, %s", items[0].ID, items[1].ID)
	}
}

func TestExportLinesAligned(t *testing.T) {
	items := []OpsRecord{
		{ID: "ops-a", Subject: "任务一", Owner: "zhang", Status: OpsStatusQueued, Priority: OpsPriorityNormal, UpdatedAt: "2026-08-21T09:00:00Z"},
		{ID: "ops-b", Subject: "任务二", Owner: "li", Status: OpsStatusActive, Priority: OpsPriorityHigh, UpdatedAt: "2026-08-20T09:00:00Z"},
		{ID: "ops-c", Subject: "任务三", Owner: "wang", Status: OpsStatusPaused, Priority: OpsPriorityCritical, UpdatedAt: "2026-08-19T09:00:00Z"},
	}
	csv := exportRecordsCSV(items)
	rows := strings.Split(strings.TrimSuffix(csv, "\n"), "\n")
	if len(rows) != len(items)+1 {
		t.Fatalf("csv has %d lines, want %d (header + %d records)", len(rows), len(items)+1, len(items))
	}
}

func TestExportHasHeader(t *testing.T) {
	items := []OpsRecord{{ID: "ops-a", Subject: "任务一", Owner: "zhang", Status: OpsStatusQueued, Priority: OpsPriorityNormal, UpdatedAt: "2026-08-21T09:00:00Z"}}
	csv := exportRecordsCSV(items)
	if !strings.HasPrefix(csv, "id,subject,owner,status,priority,updated_at") {
		t.Fatalf("csv missing header line: %q", csv)
	}
}

func TestExportEmptyHasHeader(t *testing.T) {
	csv := exportRecordsCSV(nil)
	if !strings.HasPrefix(csv, "id,subject,owner,status,priority,updated_at") {
		t.Fatalf("empty export should still return header, got %q", csv)
	}
}
