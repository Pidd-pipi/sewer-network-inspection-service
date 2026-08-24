package main

import (
	"errors"
	"sync"
)

var errInspectionNotFound = errors.New("inspection not found")
var errInspectionTransition = errors.New("inspection status transition is not allowed")

type InspectionStore struct {
	mu    sync.RWMutex
	items map[string]Inspection
}

func newInspectionStore() *InspectionStore {
	return &InspectionStore{items: map[string]Inspection{
		"ins-101": {ID: "ins-101", Manhole: "MH-101", Condition: "infiltration", Priority: "high", Status: "open", LastInspected: "2026-08-17"},
		"ins-102": {ID: "ins-102", Manhole: "MH-102", Condition: "stable", Priority: "routine", Status: "scheduled", LastInspected: "2026-08-18"},
	}}
}

func (s *InspectionStore) list() []Inspection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Inspection, 0, len(s.items))
	for _, item := range s.items {
		out = append(out, item)
	}
	return out
}

func (s *InspectionStore) changeStatus(id, status string) (Inspection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return Inspection{}, errInspectionNotFound
	}
	item.Status = status
	s.items[id] = item
	return item, nil
}
