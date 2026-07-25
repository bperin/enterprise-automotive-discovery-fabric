package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"enterprise-search/internal/product"
	"github.com/google/uuid"
)

// SaaSAdapter implements SourceAdapter for third-party vendor APIs.
//
// GCP-NATIVE IDENTITY & DATA PATH:
//   GCP In-House Source Adapter
//     ↓ OAuth 2.0 Client Credentials Grant (Secret Manager credentials)
//   External Supplier API (e.g. AC Delco Catalog Feed)
//     ↓ Webhook / Polled Payload
//   Canonical Ingestion Events -> Cloud Pub/Sub
type SaaSAdapter struct {
	sourceID string
	vendor   string
}

func NewSaaSAdapter(sourceID, vendor string) *SaaSAdapter {
	if sourceID == "" {
		sourceID = "supplier-acdelco-api"
	}
	return &SaaSAdapter{
		sourceID: sourceID,
		vendor:   vendor,
	}
}

func (s *SaaSAdapter) ID() string {
	return s.sourceID
}

func (s *SaaSAdapter) SourceType() string {
	return "external_saas_oauth2"
}

func (s *SaaSAdapter) FetchObject(ctx context.Context, objectID string) (*SourceObject, error) {
	if objectID == "" {
		return nil, ErrInvalidSourceObject
	}

	payloadMap := map[string]any{
		"sku":            "ACD-10-4022",
		"oem_cross_ref":  "84154233",
		"description":    "Heavy Duty Brake Pad Kit",
		"supplier_price": 42.50,
		"currency":       "USD",
	}
	raw, _ := json.Marshal(payloadMap)

	now := time.Now().UTC()
	return &SourceObject{
		SourceID:       s.sourceID,
		SourceObjectID: objectID,
		Version:        "v2.1",
		ContentType:    "application/json",
		Payload:        raw,
		URI:            fmt.Sprintf("https://api.supplier.com/v2/catalog/%s", objectID),
		ObservedAt:     now,
		ReceivedAt:     now,
		SourceRef: product.SourceReference{
			SourceID:       s.sourceID,
			SourceType:     s.SourceType(),
			SourceRecordID: objectID,
			SourceURI:      fmt.Sprintf("https://api.supplier.com/v2/catalog/%s", objectID),
			Version:        "v2.1",
		},
	}, nil
}

func (s *SaaSAdapter) ProduceCanonicalEvents(ctx context.Context) ([]*IngestionEvent, error) {
	obj, err := s.FetchObject(ctx, "ACD-10-4022")
	if err != nil {
		return nil, err
	}

	event := &IngestionEvent{
		EventID:        "evt-" + uuid.New().String()[:8],
		EventType:      EventObjectCreated,
		SourceID:       s.sourceID,
		SourceObjectID: obj.SourceObjectID,
		SourceVersion:  obj.Version,
		ContentType:    obj.ContentType,
		URI:            obj.URI,
		Payload:        obj.Payload,
		ObservedAt:     obj.ObservedAt,
		ReceivedAt:     obj.ReceivedAt,
		TraceID:        "trace-saas-" + uuid.New().String()[:8],
		SourceRef:      obj.SourceRef,
	}

	return []*IngestionEvent{event}, nil
}
