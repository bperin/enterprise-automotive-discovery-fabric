package sources_test

import (
	"context"
	"testing"

	"enterprise-search/internal/sources"
)

func TestSourceAdapters(t *testing.T) {
	ctx := context.Background()

	t.Run("Cloud SQL Adapter CDC Ingestion Event", func(t *testing.T) {
		adapter := sources.NewCloudSQLAdapter("dealer-db-1", "trader-prod-v2", "automotive_discovery")
		if adapter.ID() != "dealer-db-1" {
			t.Errorf("expected ID 'dealer-db-1', got %s", adapter.ID())
		}

		events, err := adapter.ProduceCanonicalEvents(ctx)
		if err != nil {
			t.Fatalf("failed producing canonical events: %v", err)
		}

		if len(events) == 0 {
			t.Fatalf("expected at least 1 canonical event")
		}
		if events[0].EventType != sources.EventObjectUpdated {
			t.Errorf("expected EventObjectUpdated, got %s", events[0].EventType)
		}
	})

	t.Run("SaaS OAuth Adapter Ingestion Event", func(t *testing.T) {
		adapter := sources.NewSaaSAdapter("supplier-acdelco", "ACDelco")
		events, err := adapter.ProduceCanonicalEvents(ctx)
		if err != nil {
			t.Fatalf("failed producing SaaS events: %v", err)
		}

		if len(events) == 0 {
			t.Fatalf("expected at least 1 SaaS canonical event")
		}
		if events[0].SourceID != "supplier-acdelco" {
			t.Errorf("expected source ID 'supplier-acdelco', got %s", events[0].SourceID)
		}
	})
}
