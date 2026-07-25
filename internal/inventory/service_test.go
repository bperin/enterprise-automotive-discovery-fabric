package inventory_test

import (
	"context"
	"testing"
	"time"

	"enterprise-search/internal/inventory"
)

func TestInventoryService(t *testing.T) {
	repo := inventory.NewMemoryRepository()
	svc := inventory.NewService(repo)
	ctx := context.Background()

	loc, err := svc.SaveLocation(ctx, &inventory.Location{
		Name:       "Austin Central Warehouse",
		PostalCode: "78701",
		Latitude:   30.2672,
		Longitude:  -97.7431,
	})
	if err != nil {
		t.Fatalf("failed creating location: %v", err)
	}

	t.Run("Record and Query Freshest Observation", func(t *testing.T) {
		qty1 := 5
		t1 := time.Now().Add(-2 * time.Hour)
		_, err := svc.RecordObservation(ctx, &inventory.InventoryObservation{
			ProductID:  "wheel-102",
			LocationID: loc.ID,
			Quantity:   &qty1,
			ObservedAt: t1,
		})
		if err != nil {
			t.Fatalf("failed recording observation 1: %v", err)
		}

		qty2 := 12
		t2 := time.Now()
		_, err = svc.RecordObservation(ctx, &inventory.InventoryObservation{
			ProductID:  "wheel-102",
			LocationID: loc.ID,
			Quantity:   &qty2,
			ObservedAt: t2,
		})
		if err != nil {
			t.Fatalf("failed recording observation 2: %v", err)
		}

		freshest, err := svc.GetFreshestObservation(ctx, "wheel-102", loc.ID)
		if err != nil {
			t.Fatalf("get freshest observation failed: %v", err)
		}
		if freshest.Quantity == nil || *freshest.Quantity != 12 {
			t.Errorf("expected freshest quantity 12, got %v", freshest.Quantity)
		}
	})
}
