package main

import (
	"context"
	"time"
)

type OpsWorker struct {
	service    *OpsService
	interval   time.Duration
	staleAfter time.Duration
	actor      string
	stop       chan struct{}
	done       chan struct{}
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

func (w *OpsWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = w.service.BatchCloseStale(ctx, w.staleAfter, w.actor)
		}
	}
}

func (w *OpsWorker) Close() {
	close(w.stop)
	<-w.done
}

func (w *OpsWorker) Running() bool {
	select {
	case <-w.done:
		return false
	default:
		return true
	}
}
