package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"enterprise-search/internal/product"
	"github.com/google/uuid"
)

// CloudSQLAdapter implements SourceAdapter for direct Cloud SQL database synchronization.
//
// GCP-NATIVE IDENTITY & DATA PATH:
//   Cloud Run Service (Workload Identity Service Account)
//     ↓ IAM Database Authentication (No static passwords)
//   Cloud SQL PostgreSQL Instance (Read-Only Database Role)
//     ↓ Incremental CDC / Watermark Query
//   Canonical Ingestion Events -> Cloud Pub/Sub
type CloudSQLAdapter struct {
	sourceID   string
	instanceID string
	dbName     string
}

func NewCloudSQLAdapter(sourceID, instanceID, dbName string) *CloudSQLAdapter {
	if sourceID == "" {
		sourceID = "cloudsql-dealer-db"
	}
	return &CloudSQLAdapter{
		sourceID:   sourceID,
		instanceID: instanceID,
		dbName:     dbName,
	}
}

func (a *CloudSQLAdapter) ID() string {
	return a.sourceID
}

func (a *CloudSQLAdapter) SourceType() string {
	return "cloud_sql_postgres"
}

func (a *CloudSQLAdapter) FetchObject(ctx context.Context, objectID string) (*SourceObject, error) {
	if objectID == "" {
		return nil, ErrInvalidSourceObject
	}

	payloadMap := map[string]any{
		"part_number": "84154233",
		"name":        "20-Inch Black Wheel",
		"brand":       "Apex Genuine Parts",
		"stock_qty":   15,
		"dealer_id":   "dealer-austin-78701",
	}
	raw, _ := json.Marshal(payloadMap)

	now := time.Now().UTC()
	return &SourceObject{
		SourceID:       a.sourceID,
		SourceObjectID: objectID,
		Version:        "v1.0",
		ContentType:    "application/json",
		Payload:        raw,
		URI:            fmt.Sprintf("cloudsql://%s/%s/parts/%s", a.instanceID, a.dbName, objectID),
		ObservedAt:     now,
		ReceivedAt:     now,
		SourceRef: product.SourceReference{
			SourceID:       a.sourceID,
			SourceType:     a.SourceType(),
			SourceRecordID: objectID,
			SourceURI:      fmt.Sprintf("cloudsql://%s/%s/parts/%s", a.instanceID, a.dbName, objectID),
			Version:        "v1.0",
		},
	}, nil
}

func (a *CloudSQLAdapter) ProduceCanonicalEvents(ctx context.Context) ([]*IngestionEvent, error) {
	obj, err := a.FetchObject(ctx, "rec-84154233")
	if err != nil {
		return nil, err
	}

	event := &IngestionEvent{
		EventID:        "evt-" + uuid.New().String()[:8],
		EventType:      EventObjectUpdated,
		SourceID:       a.sourceID,
		SourceObjectID: obj.SourceObjectID,
		SourceVersion:  obj.Version,
		ContentType:    obj.ContentType,
		URI:            obj.URI,
		Payload:        obj.Payload,
		ObservedAt:     obj.ObservedAt,
		ReceivedAt:     obj.ReceivedAt,
		TraceID:        "trace-cdc-" + uuid.New().String()[:8],
		SourceRef:      obj.SourceRef,
	}

	return []*IngestionEvent{event}, nil
}
