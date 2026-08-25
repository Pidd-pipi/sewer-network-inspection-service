package main

import (
	"context"
	"sync"
	"time"
)

type OpsWorker struct {
	service    *OpsService
	interval   time.Duration
	staleAfter time.Duration
	actor      string
	stop       chan struct{}
	stopOnce   sync.Once
	runCancel  context.CancelFunc
	mu         sync.Mutex
	done       chan struct{}
	doneOnce   sync.Once
}

func newOpsWorker(service *OpsService, interval, staleAfter time.Duration, actor string) *OpsWorker {
	return &OpsWorker{
		service:    service,
		interval:   interval,
		staleAfter: staleAfter,
		actor:      actor,
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
	}
}

func (w *OpsWorker) Run(parent context.Context) {
	if parent == nil {
		parent = context.Background()
	}
	// Derive an internal context whose cancellation is driven by Close() as well as
	// the parent, so a batch currently in flight is aborted promptly on shutdown
	// instead of being left to finish (or hang) after the worker is told to stop.
	ctx, cancel := context.WithCancel(parent)
	w.mu.Lock()
	w.runCancel = cancel
	w.mu.Unlock()
	defer cancel()
	defer w.doneOnce.Do(func() { close(w.done) })

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-ticker.C:
			_, _ = w.service.BatchCloseStale(ctx, w.staleAfter, w.actor)
		}
	}
}

// Close signals the worker loop to stop and waits for it to finish. It cancels the
// worker's internal context first so any batch currently running aborts quickly,
// then unblocks the loop via the stop channel. Idempotent and always returns
// (Run guarantees closing w.done on every exit path).
func (w *OpsWorker) Close() {
	w.stopOnce.Do(func() { close(w.stop) })
	w.mu.Lock()
	if w.runCancel != nil {
		w.runCancel()
	}
	w.mu.Unlock()
	<-w.done
}

// Running reports whether the loop is still active. It reflects the real state of
// w.done rather than the context, so it stays accurate after Close() returns.
func (w *OpsWorker) Running() bool {
	select {
	case <-w.done:
		return false
	default:
		return true
	}
}
