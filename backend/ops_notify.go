package main

import (
	"context"
	"sync"
	"time"
)

type OpsNotice struct {
	RecordID string `json:"recordId"`
	Channel  string `json:"channel"`
	Message  string `json:"message"`
	QueuedAt string `json:"queuedAt"`
}

type OpsNotifier struct {
	mu      sync.Mutex
	outbox  []OpsNotice
	workers chan struct{}
}

const opsNotifierCap = 2000

func newOpsNotifier(maxWorkers int) *OpsNotifier {
	if maxWorkers < 1 {
		maxWorkers = 1
	}
	return &OpsNotifier{
		outbox:  make([]OpsNotice, 0, 32),
		workers: make(chan struct{}, maxWorkers),
	}
}

func (n *OpsNotifier) Queue(recordID, channel, message string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.outbox = append(n.outbox, OpsNotice{
		RecordID: recordID,
		Channel:  channel,
		Message:  message,
		QueuedAt: time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func (n *OpsNotifier) Pending() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.outbox)
}

func (n *OpsNotifier) Peek() []OpsNotice {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]OpsNotice, len(n.outbox))
	copy(out, n.outbox)
	return out
}

func (n *OpsNotifier) Dispatch(ctx context.Context, send func(OpsNotice) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		n.mu.Lock()
		if len(n.outbox) == 0 {
			n.mu.Unlock()
			return nil
		}
		notice := n.outbox[0]
		n.mu.Unlock()
		if err := send(notice); err != nil {
			return err
		}
		n.mu.Lock()
		n.outbox = n.outbox[1:]
		n.mu.Unlock()
	}
}
