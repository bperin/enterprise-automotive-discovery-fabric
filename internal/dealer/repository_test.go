package dealer

import (
	"context"
	"testing"
)

func TestMemoryRepository_Dealers(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	d, err := repo.GetByID(ctx, "dealer-austin-01")
	if err != nil {
		t.Fatalf("unexpected error getting default dealer: %v", err)
	}

	if d.Name != "Northstar Motors Austin" {
		t.Errorf("expected Northstar Motors Austin, got %s", d.Name)
	}

	locs, err := repo.FindByPostalCode(ctx, "78701", 50)
	if err != nil {
		t.Fatalf("unexpected error finding by postal code: %v", err)
	}

	if len(locs) != 1 {
		t.Errorf("expected 1 location in 78701, got %d", len(locs))
	}
}
