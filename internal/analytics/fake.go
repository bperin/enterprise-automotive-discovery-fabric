package analytics

import (
	"context"
	"sync"
)

// FakeEventStore is a thread-safe local EventStore implementation.
type FakeEventStore struct {
	mu     sync.RWMutex
	events []*Event
}

var _ EventStore = (*FakeEventStore)(nil)

func NewFakeEventStore() *FakeEventStore {
	return &FakeEventStore{events: make([]*Event, 0)}
}

func (f *FakeEventStore) Write(_ context.Context, event *Event) error {
	if event == nil || event.ID == "" || event.Name == "" {
		return ErrInvalidEvent
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	copied := *event
	f.events = append(f.events, &copied)
	return nil
}

func (f *FakeEventStore) Query(_ context.Context, filter QueryFilter) ([]*Event, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	results := make([]*Event, 0, len(f.events))
	for _, event := range f.events {
		if filter.Name != "" && event.Name != filter.Name {
			continue
		}
		if !filter.StartTime.IsZero() && event.Timestamp.Before(filter.StartTime) {
			continue
		}
		if !filter.EndTime.IsZero() && event.Timestamp.After(filter.EndTime) {
			continue
		}
		copied := *event
		results = append(results, &copied)
	}
	return results, nil
}
