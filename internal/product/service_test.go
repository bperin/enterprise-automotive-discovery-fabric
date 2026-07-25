package product_test

import (
	"context"
	"testing"

	"enterprise-search/internal/product"
)

func TestProductService(t *testing.T) {
	repo := product.NewMemoryRepository()
	svc := product.NewService(repo)
	ctx := context.Background()

	t.Run("Create and Get Product", func(t *testing.T) {
		p := &product.Product{
			CanonicalName:  "20-Inch Black Wheel",
			Brand:          "OEM Performance",
			Category:       product.CategoryWheels,
			ManufacturerID: "MFR-99",
			Identifiers: []product.ProductIdentifier{
				{
					Type:       product.IDOEMPartNumber,
					Value:      "8415-4233",
					Normalized: "84154233",
				},
			},
		}

		created, err := svc.CreateOrUpdateProduct(ctx, p)
		if err != nil {
			t.Fatalf("unexpected error creating product: %v", err)
		}
		if created.ID == "" {
			t.Errorf("expected product ID to be generated")
		}

		retrieved, err := svc.GetProduct(ctx, created.ID)
		if err != nil {
			t.Fatalf("failed to retrieve created product: %v", err)
		}
		if retrieved.CanonicalName != p.CanonicalName {
			t.Errorf("expected %s, got %s", p.CanonicalName, retrieved.CanonicalName)
		}
	})

	t.Run("Lookup by Normalized Identifier", func(t *testing.T) {
		p := &product.Product{
			CanonicalName: "Brake Pad Kit",
			Identifiers: []product.ProductIdentifier{
				{
					Type:  product.IDOEMPartNumber,
					Value: "BP-100-22",
				},
			},
		}
		_, err := svc.CreateOrUpdateProduct(ctx, p)
		if err != nil {
			t.Fatalf("failed creating product: %v", err)
		}

		found, err := svc.LookupIdentifier(ctx, "bp10022")
		if err != nil {
			t.Fatalf("expected identifier lookup to succeed, got %v", err)
		}
		if found.CanonicalName != "Brake Pad Kit" {
			t.Errorf("expected 'Brake Pad Kit', got %s", found.CanonicalName)
		}
	})
}
