package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var opsAuditSequence uint64

// opsAuditCap bounds the number of retained audit events. Without a cap the
// slice grows forever under steady write traffic, which is the observed
// memory leak. When the cap is reached the oldest events are dropped.
const opsAuditCap = 10000

func newOpsAuditID() string { return fmt.Sprintf("evt-%06d", atomic.AddUint64(&opsAuditSequence, 1)) }

type OpsAudit struct {
	mu     sync.RWMutex
	events []OpsEvent
}

func newOpsAudit() *OpsAudit { return &OpsAudit{events: []OpsEvent{}} }
func (a *OpsAudit) Add(recordID, typ, actor string) OpsEvent {
	event := OpsEvent{ID: newOpsAuditID(), RecordID: recordID, Type: typ, Actor: actor, At: time.Now().UTC().Format(time.RFC3339Nano)}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = append(a.events, event)
	// Drop the oldest event once we blow past the cap so memory stays bounded.
	if len(a.events) > opsAuditCap {
		// Shift left by the overflow, preserving ordering without re-allocating
		// each time. Copy into a fresh slice to avoid pinning the dropped tail.
		overflow := len(a.events) - opsAuditCap
		copy(a.events, a.events[overflow:])
		keep := a.events[:opsAuditCap]
		fresh := make([]OpsEvent, len(keep))
		copy(fresh, keep)
		a.events = fresh
	}
	return event
}
func (a *OpsAudit) For(recordID string) []OpsEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := []OpsEvent{}
	for _, event := range a.events {
		if event.RecordID == recordID {
			out = append(out, event)
		}
	}
	return out
}
func (a *OpsAudit) Since(start time.Time) []OpsEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := []OpsEvent{}
	for _, event := range a.events {
		parsed, err := time.Parse(time.RFC3339Nano, event.At)
		if err == nil && !parsed.Before(start) {
			out = append(out, event)
		}
	}
	return out
}
func (a *OpsAudit) Count() int { a.mu.RLock(); defer a.mu.RUnlock(); return len(a.events) }
func (a *OpsAudit) Latest() (OpsEvent, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.events) == 0 {
		return OpsEvent{}, false
	}
	return a.events[len(a.events)-1], true
}
func (a *OpsAudit) Clear() { a.mu.Lock(); defer a.mu.Unlock(); a.events = a.events[:0] }
