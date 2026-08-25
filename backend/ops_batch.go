package main

import (
	"context"
	"strconv"
	"time"
)

type OpsBatchResult struct {
	Scanned int
	Closed  int
	Failed  int
}

func (r OpsBatchResult) Empty() bool { return r.Scanned == 0 && r.Closed == 0 && r.Failed == 0 }

// auditBatch records the outcome of a batch run only when work was actually attempted.
// A batch that scans nothing, or aborts before scanning (e.g. store error), produces no
// audit event — so failures are not mistakenly recorded as successes.
func (s *OpsService) auditBatch(actor, action string, r OpsBatchResult, err error) {
	if r.Empty() && err != nil {
		// The run failed before doing any per-item work; surface the failure directly.
		cause := "ok"
		if err != nil {
			cause = err.Error()
		}
		s.audit.AddWithDetails("batch", action, actor, map[string]string{
			"scanned": "0", "closed": "0", "failed": "0", "result": "error", "error": cause,
		})
		return
	}
	if r.Empty() {
		return
	}
	result := "success"
	if r.Closed == 0 && r.Failed == 0 {
		// Work was attempted but no item matched the batch criteria — a genuine no-op,
		// not a success to be confused with "closed N".
		result = "noop"
	} else if r.Failed > 0 {
		if r.Closed == 0 {
			result = "failure"
		} else {
			result = "partial"
		}
	}
	s.audit.AddWithDetails("batch", action, actor, map[string]string{
		"scanned": strconv.Itoa(r.Scanned),
		"closed":  strconv.Itoa(r.Closed),
		"failed":  strconv.Itoa(r.Failed),
		"result":  result,
	})
}

func (s *OpsService) BatchCloseStale(ctx context.Context, maxAge time.Duration, actor string) (OpsBatchResult, error) {
	items, err := s.store.List(ctx)
	if err != nil {
		return OpsBatchResult{}, err
	}
	result := OpsBatchResult{Scanned: len(items)}
	// Wrap in a closure so the deferred call reads the final result, not the
	// zero/empty value captured at the defer statement.
	defer func() { s.auditBatch(actor, "batch_close", result, nil) }()
	for _, item := range items {
		if ctx.Err() != nil {
			break
		}
		if item.Status != OpsStatusActive {
			continue
		}
		if opsAge(s.clock.Now(), item.UpdatedAt) < maxAge {
			continue
		}
		if _, err := s.Transition(ctx, item.ID, item.Revision, OpsStatusClosed, actor); err != nil {
			result.Failed++
			continue
		}
		result.Closed++
	}
	return result, nil
}

func (s *OpsService) BatchResolve(ctx context.Context, ids []string, actor string) (OpsBatchResult, error) {
	result := OpsBatchResult{Scanned: len(ids)}
	defer func() { s.auditBatch(actor, "batch_resolve", result, nil) }()
	for _, id := range ids {
		if ctx.Err() != nil {
			break
		}
		record, err := s.store.Get(ctx, id)
		if err != nil {
			// Invalid / missing ticket: count it as a failure so the batch tally stays
			// consistent (Scanned == Closed + Failed) and the caller can see it.
			result.Failed++
			continue
		}
		if _, err := s.Transition(ctx, id, record.Revision, OpsStatusClosed, actor); err != nil {
			result.Failed++
			continue
		}
		result.Closed++
	}
	return result, nil
}
