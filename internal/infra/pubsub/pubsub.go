package pubsub

import (
	"context"
	"fmt"
	"time"

	pubsubSDK "cloud.google.com/go/pubsub"
	"enterprise-search/internal/messaging"
)

// PubSubAdapter implements messaging.Publisher and messaging.Subscriber using the official GCP Pub/Sub SDK.
// It relies on Application Default Credentials (ADC) for authentication.
type PubSubAdapter struct {
	client *pubsubSDK.Client
}

var _ messaging.Publisher = (*PubSubAdapter)(nil)
var _ messaging.Subscriber = (*PubSubAdapter)(nil)

// NewPubSubAdapter constructs a new PubSubAdapter.
func NewPubSubAdapter(ctx context.Context, projectID string) (*PubSubAdapter, error) {
	client, err := pubsubSDK.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("create pubsub client: %w", err)
	}
	return &PubSubAdapter{client: client}, nil
}

// Close closes the underlying Pub/Sub client.
func (a *PubSubAdapter) Close() error {
	return a.client.Close()
}

// Publish publishes a versioned event to the specified topic.
func (a *PubSubAdapter) Publish(ctx context.Context, topicID string, event *messaging.Event) error {
	topic := a.client.Topic(topicID)
	res := topic.Publish(ctx, &pubsubSDK.Message{
		Data: event.Data,
		Attributes: map[string]string{
			"id":        event.ID,
			"version":   event.Version,
			"type":      event.Type,
			"timestamp": event.Timestamp.Format(time.RFC3339),
		},
	})
	_, err := res.Get(ctx)
	if err != nil {
		return fmt.Errorf("publish message to topic %s: %w", topicID, err)
	}
	return nil
}

// Subscribe subscribes to the specified subscription and processes messages using the provided handler.
func (a *PubSubAdapter) Subscribe(ctx context.Context, subscriptionID string, handler messaging.Handler) error {
	sub := a.client.Subscription(subscriptionID)
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsubSDK.Message) {
		var timestamp time.Time
		if tsStr, ok := msg.Attributes["timestamp"]; ok {
			if t, err := time.Parse(time.RFC3339, tsStr); err == nil {
				timestamp = t
			}
		}
		if timestamp.IsZero() {
			timestamp = msg.PublishTime
		}

		event := &messaging.Event{
			ID:        msg.Attributes["id"],
			Version:   msg.Attributes["version"],
			Type:      msg.Attributes["type"],
			Timestamp: timestamp,
			Data:      msg.Data,
		}

		if err := handler.Handle(ctx, event); err != nil {
			msg.Nack()
		} else {
			msg.Ack()
		}
	})
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", subscriptionID, err)
	}
	return nil
}
