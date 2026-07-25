// Package sources provides standardized, internally owned source adapters in DDD.
//
// TRANSFORMATION FROM THIRD-PARTY CONNECTORS TO IN-HOUSE GCP-NATIVE ADAPTERS:
//
// BEFORE:
//   Source Systems -> Third-Party Connector Vendor -> Third-Party Integration Platform -> Search Platform -> Applications
//   - High cost, opaque data mappings, duplicated authentication, vendor lock-in.
//
// AFTER (GCP-Native Design):
//   Source Systems -> Small In-House Source Adapters -> Pub/Sub + Canonical Model -> Search Indexes & Operational Stores
//   - Smaller, standardized, reusable, observable, independently replaceable Go adapters.
//   - Direct database access uses Workload Identity + Cloud SQL IAM authentication + read-only DB roles.
//   - External SaaS providers use OAuth 2.0 client credentials or delegated OAuth.
package sources

import (
	"context"
	"errors"
	"time"

	"enterprise-search/internal/product"
)

// IngestionEventType defines standard change propagation event types.
type IngestionEventType string

const (
	EventObjectDiscovered IngestionEventType = "source.object.discovered"
	EventObjectCreated    IngestionEventType = "source.object.created"
	EventObjectUpdated    IngestionEventType = "source.object.updated"
	EventObjectDeleted    IngestionEventType = "source.object.deleted"
	EventRecordProposed   IngestionEventType = "canonical.record.proposed"
	EventRecordPublished  IngestionEventType = "canonical.record.published"
)

// SourceObject represents raw source payload prior to canonical normalization.
type SourceObject struct {
	SourceID       string                  `json:"source_id"`
	SourceObjectID string                  `json:"source_object_id"`
	Version        string                  `json:"version"`
	ContentType    string                  `json:"content_type"` // e.g. "application/json", "application/pdf"
	Payload        []byte                  `json:"payload"`
	URI            string                  `json:"uri"`
	ObservedAt     time.Time               `json:"observed_at"`
	ReceivedAt     time.Time               `json:"received_at"`
	Metadata       map[string]string       `json:"metadata,omitempty"`
	SourceRef      product.SourceReference `json:"source_ref"`
}

// IngestionEvent is the canonical Pub/Sub event emitted by all source adapters in DDD.
type IngestionEvent struct {
	EventID        string                  `json:"event_id"`
	EventType      IngestionEventType      `json:"event_type"`
	SourceID       string                  `json:"source_id"`
	SourceObjectID string                  `json:"source_object_id"`
	SourceVersion  string                  `json:"source_version"`
	ContentType    string                  `json:"content_type"`
	URI            string                  `json:"uri"`
	Payload        []byte                  `json:"payload,omitempty"`
	ObservedAt     time.Time               `json:"observed_at"`
	ReceivedAt     time.Time               `json:"received_at"`
	TraceID        string                  `json:"trace_id"`
	SourceRef      product.SourceReference `json:"source_ref"`
}

var (
	ErrAdapterUnavailable = errors.New("source adapter unavailable")
	ErrInvalidSourceObject = errors.New("invalid source object payload")
)

// SourceAdapter is the DDD consumer port for all internal and external source adapters.
type SourceAdapter interface {
	ID() string
	SourceType() string
	FetchObject(ctx context.Context, objectID string) (*SourceObject, error)
	ProduceCanonicalEvents(ctx context.Context) ([]*IngestionEvent, error)
}
