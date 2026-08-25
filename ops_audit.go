package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var opsAuditSequence uint64

// opsAuditCap bounds the number of retained audit events so that a high
// request volume cannot grow memory without limit. Older events are dropped
// once the cap is reached.
const opsAuditCap = 5000

func newOpsAuditID() string { return fmt.Sprintf("evt-%06d", atomic.AddUint64(&opsAuditSequence, 1)) }

type OpsAudit struct {
	mu     sync.RWMutex
	events []OpsEvent
}

func newOpsAudit() *OpsAudit { return &OpsAudit{events: []OpsEvent{}} }
func (a *OpsAudit) Add(recordID, typ, actor string) OpsEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	event := OpsEvent{ID: newOpsAuditID(), RecordID: recordID, Type: typ, Actor: actor, At: time.Now().UTC().Format(time.RFC3339Nano)}
	a.events = append(a.events, event)
	if len(a.events) > opsAuditCap {
		// Drop the oldest events to keep the audit log bounded.
		a.events = a.events[len(a.events)-opsAuditCap:]
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
