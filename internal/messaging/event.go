package messaging

import (
	"context"
	"time"
)

// Event represents a versioned message envelope.
type Event struct {
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
}

// ContentUploadedEventV1 is the payload for version "v1" of the content uploaded event.
type ContentUploadedEventV1 struct {
	Bucket     string `json:"bucket"`
	ObjectName string `json:"object_name"`
}

// Publisher defines the port for publishing events.
type Publisher interface {
	Publish(ctx context.Context, topicID string, event *Event) error
}

// Handler defines the port for consuming events.
type Handler interface {
	Handle(ctx context.Context, event *Event) error
}

// HandlerFunc is an adapter to allow the use of ordinary functions as event handlers.
type HandlerFunc func(ctx context.Context, event *Event) error

// Handle calls f(ctx, event).
func (f HandlerFunc) Handle(ctx context.Context, event *Event) error {
	return f(ctx, event)
}

// Subscriber defines the port for subscribing to events.
type Subscriber interface {
	Subscribe(ctx context.Context, subscriptionID string, handler Handler) error
}
