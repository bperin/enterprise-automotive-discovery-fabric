package discovery_test

import (
	"context"
	"testing"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/discovery"
	"enterprise-search/internal/product"
	"enterprise-search/internal/search"
)

func TestDiscoveryGatewayService(t *testing.T) {
	engine := search.NewMemorySearchEngine()
	searchSvc := search.NewService(engine)
	discoverySvc := discovery.NewService(searchSvc)
	ctx := context.Background()

	// Index a sample product
	p := &product.Product{
		ID:            "apex-hauler-2026",
		CanonicalName: "2026 Apex Hauler EV Truck",
		Description:   "Electric truck with 7,000 lbs max towing capacity and 300 miles estimated range.",
		Category:      product.CategoryGeneric,
		Brand:         "ApexMotors",
		Publication:   product.PublicationPublished,
		AccessPolicy: auth.AccessPolicy{
			Visibility: auth.VisibilityPublic,
		},
	}
	_ = searchSvc.IndexProduct(ctx, p)

	user := &auth.Principal{
		SubjectID: "user-123",
		Roles:     []string{"customer"},
	}

	t.Run("Unified Brand Discovery Query", func(t *testing.T) {
		req := discovery.DiscoveryRequest{
			Query:   "2026 Apex Hauler EV towing capacity",
			Brand:   "ApexMotors",
			Channel: "public_web",
			User:    user,
			Location: &discovery.GeoContext{
				PostalCode: "78701",
			},
		}

		resp, err := discoverySvc.ProcessQuery(ctx, req)
		if err != nil {
			t.Fatalf("ProcessQuery failed: %v", err)
		}

		if !resp.Grounded {
			t.Errorf("expected response to be grounded")
		}
		if len(resp.Results) == 0 {
			t.Errorf("expected at least 1 discovery result")
		}
		if len(resp.Citations) == 0 {
			t.Errorf("expected evidence citations in response")
		}
		if len(resp.Actions) == 0 {
			t.Errorf("expected suggested call-to-action actions")
		}
	})
}
