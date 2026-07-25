package analytics

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidEvent = errors.New("invalid event")

// Event is the application-owned analytics record.
type Event struct {
	ID        string
	Name      string
	Timestamp time.Time
	Payload   string
}

type QueryFilter struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
}

// EventStore is the consumer-owned analytics read/write boundary.
type EventStore interface {
	Write(ctx context.Context, event *Event) error
	Query(ctx context.Context, filter QueryFilter) ([]*Event, error)
}
