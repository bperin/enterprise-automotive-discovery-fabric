package search_test

import (
	"context"
	"testing"

	"enterprise-search/internal/product"
	"enterprise-search/internal/search"
)

func TestSearchService(t *testing.T) {
	engine := search.NewMemorySearchEngine()
	svc := search.NewService(engine)
	ctx := context.Background()

	t.Run("Index Product and Execute Search", func(t *testing.T) {
		p := &product.Product{
			ID:            "prod-wheels-20",
			CanonicalName: "Heavy Duty Black 20-inch Wheel",
			Description:   "Alloy wheel compatible with Apex Ridge 1500",
			Category:      product.CategoryWheels,
			Brand:         "OEM HeavyDuty",
		}

		if err := svc.IndexProduct(ctx, p); err != nil {
			t.Fatalf("failed indexing product: %v", err)
		}

		resp, err := svc.ExecuteSearch(ctx, search.SearchQuery{
			Text:     "black wheel",
			PageSize: 10,
		})
		if err != nil {
			t.Fatalf("execute search failed: %v", err)
		}

		if resp.TotalCount == 0 {
			t.Errorf("expected at least 1 search result")
		}
		found := false
		for _, r := range resp.Results {
			if r.EntityID == "prod-wheels-20" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected entity ID 'prod-wheels-20' in search results")
		}
	})
}
