package main

import (
	"context"
	"time"
)

type OpsBatchResult struct {
	Scanned int
	Closed  int
	Failed  int
}

func (s *OpsService) BatchCloseStale(ctx context.Context, maxAge time.Duration, actor string) (OpsBatchResult, error) {
	items, err := s.store.List(ctx)
	if err != nil {
		return OpsBatchResult{}, err
	}
	result := OpsBatchResult{Scanned: len(items)}
	for _, item := range items {
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
	for _, id := range ids {
		record, err := s.store.Get(ctx, id)
		if err != nil {
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
