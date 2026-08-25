package main

import (
	"context"
	"sort"
	"sync"
)

type OpsStore struct {
	mu    sync.RWMutex
	items map[string]OpsRecord
}

func newOpsStore(seed []OpsRecord) *OpsStore {
	s := &OpsStore{items: map[string]OpsRecord{}}
	for _, item := range seed {
		item = normalizeOpsRecord(item)
		s.items[item.ID] = item
	}
	return s
}
func (s *OpsStore) Get(ctx context.Context, id string) (OpsRecord, error) {
	select {
	case <-ctx.Done():
		return OpsRecord{}, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return OpsRecord{}, ErrOpsNotFound
	}
	// Return a clone so callers can never mutate the stored record or its
	// Labels map, which would otherwise let concurrent readers/writers
	// cross-contaminate each other's tags.
	return item.Clone(), nil
}
func (s *OpsStore) List(ctx context.Context) ([]OpsRecord, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]OpsRecord, 0, len(s.items))
	for _, item := range s.items {
		out = append(out, item.Clone())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (s *OpsStore) Put(ctx context.Context, item OpsRecord) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[item.ID]; ok {
		return ErrOpsConflict
	}
	s.items[item.ID] = normalizeOpsRecord(item)
	return nil
}
func (s *OpsStore) Update(ctx context.Context, item OpsRecord, expected int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.items[item.ID]
	if !ok {
		return ErrOpsNotFound
	}
	if expected > 0 && current.Revision != expected {
		return ErrOpsConflict
	}
	item.Revision = current.Revision + 1
	item.UpdatedAt = timeNowOps()
	s.items[item.ID] = item.Clone()
	return nil
}
func (s *OpsStore) Delete(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return ErrOpsNotFound
	}
	delete(s.items, id)
	return nil
}
func (s *OpsStore) Count() int { s.mu.RLock(); defer s.mu.RUnlock(); return len(s.items) }

func seedOpsRecords() []OpsRecord {
	return []OpsRecord{
		{ID: "ops-201", Subject: "MH-201 井盖渗漏复检", Owner: "zhang", Status: OpsStatusActive, Priority: OpsPriorityHigh, Revision: 1, Labels: map[string]string{"site": "north", "operator": "zhang", "evidence": "photo-201"}, CreatedAt: "2026-08-20T08:00:00Z", UpdatedAt: "2026-08-21T08:00:00Z"},
		{ID: "ops-202", Subject: "MH-202 管道沉积清理", Owner: "li", Status: OpsStatusQueued, Priority: OpsPriorityNormal, Revision: 1, Labels: map[string]string{"site": "south", "operator": "li", "evidence": "photo-202"}, CreatedAt: "2026-08-21T09:00:00Z", UpdatedAt: "2026-08-21T09:00:00Z"},
		{ID: "ops-203", Subject: "MH-203 井壁开裂观察", Owner: "wang", Status: OpsStatusPaused, Priority: OpsPriorityCritical, Revision: 2, Labels: map[string]string{"site": "east", "operator": "wang", "evidence": "photo-203"}, CreatedAt: "2026-08-19T10:00:00Z", UpdatedAt: "2026-08-22T10:00:00Z"},
	}
}
