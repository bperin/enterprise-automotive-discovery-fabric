// Package main provides a standalone database and search index seeder for GCP Automotive Discovery Fabric.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/evidence"
	"enterprise-search/internal/fitment"
	"enterprise-search/internal/infra/db"
	"enterprise-search/internal/infra/postgres"
	"enterprise-search/internal/inventory"
	"enterprise-search/internal/product"
	"enterprise-search/internal/search"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("SQLITE_DSN")
	}

	database, err := db.Open("", dbURL)
	if err != nil {
		log.Fatalf("Failed opening database for seeding: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	// Initialize domain repositories and services
	productRepo := product.NewMemoryRepository()
	fitmentRepo := fitment.NewMemoryRepository()
	inventoryRepo := inventory.NewMemoryRepository()
	evidenceRepo := evidence.NewMemoryRepository()
	searchEngine := search.NewMemorySearchEngine()

	productSvc := product.NewService(productRepo)
	fitmentSvc := fitment.NewService(fitmentRepo)
	inventorySvc := inventory.NewService(inventoryRepo)
	evidenceSvc := evidence.NewService(evidenceRepo)
	searchSvc := search.NewService(searchEngine)

	fmt.Println("Seeding synthetic automotive discovery dataset...")

	// 1. Seed Products
	p1, err := productSvc.CreateOrUpdateProduct(ctx, &product.Product{
		ID:             "prod-wheel-20-black",
		ManufacturerID: "MFR-APEX-84154233",
		CanonicalName:  "20-Inch Black Heavy Duty Alloy Wheel",
		Description:    "High-strength 20-inch alloy wheel in gloss black finish for late-model Apex Ridge trucks.",
		Category:       product.CategoryWheels,
		Brand:          "Apex Genuine Parts",
		OEM:            true,
		Publication:    product.PublicationPublished,
		AccessPolicy: auth.AccessPolicy{
			Visibility: auth.VisibilityPublic,
		},
		Identifiers: []product.ProductIdentifier{
			{Type: product.IDOEMPartNumber, Value: "8415-4233", Normalized: "84154233"},
			{Type: product.IDUPC, Value: "019283746501", Normalized: "019283746501"},
		},
	})
	if err != nil {
		log.Fatalf("Failed seeding product 1: %v", err)
	}
	_ = searchSvc.IndexProduct(ctx, p1)

	p2, err := productSvc.CreateOrUpdateProduct(ctx, &product.Product{
		ID:             "prod-brake-pad-acdelco",
		ManufacturerID: "MFR-ALT-104022",
		CanonicalName:  "ProPerformance Ceramic Brake Pad Kit",
		Description:    "Noise-dampening ceramic brake pads for front axle applications.",
		Category:       product.CategoryBrakes,
		Brand:          "Apex Aftermarket",
		OEM:            false,
		Publication:    product.PublicationPublished,
		AccessPolicy: auth.AccessPolicy{
			Visibility: auth.VisibilityPublic,
		},
		Identifiers: []product.ProductIdentifier{
			{Type: product.IDOEMPartNumber, Value: "8432-1901", Normalized: "84321901"},
			{Type: product.IDSupplierSKU, Value: "APX-10-4022", Normalized: "apx104022"},
		},
	})
	if err != nil {
		log.Fatalf("Failed seeding product 2: %v", err)
	}
	_ = searchSvc.IndexProduct(ctx, p2)

	// 2. Seed Vehicle Configurations
	v1, err := fitmentSvc.SaveVehicleConfig(ctx, &fitment.VehicleConfiguration{
		ID:         "veh-2026-apex-ridge-1500",
		Year:       2026,
		Make:       "ApexMotors",
		Model:      "Ridge 1500",
		Trim:       "Trail Z71",
		Engine:     "5.3L V8",
		Drivetrain: "4WD",
	})
	if err != nil {
		log.Fatalf("Failed seeding vehicle 1: %v", err)
	}

	// 3. Seed Fitment Assertions
	_, err = fitmentSvc.CreateAssertion(ctx, &fitment.FitmentAssertion{
		ID:                     "fit-wheel-ridge-2026",
		ProductID:              p1.ID,
		VehicleConfigurationID: v1.ID,
		Compatibility:          fitment.DirectFit,
		Authority:              fitment.AuthAuthoritative,
		Confidence:             1.0,
		VerificationStatus:     fitment.VerifiedBySourceContract,
		SourceRef: product.SourceReference{
			SourceID:       "dealer-db-1",
			SourceType:     "cloud_sql_postgres",
			SourceRecordID: "rec-84154233",
		},
	})
	if err != nil {
		log.Fatalf("Failed seeding fitment: %v", err)
	}

	// 4. Seed Locations & Inventory Observations
	loc, err := inventorySvc.SaveLocation(ctx, &inventory.Location{
		ID:         "loc-austin-78701",
		Name:       "Austin Central Distribution Hub",
		PostalCode: "78701",
		Latitude:   30.2672,
		Longitude:  -97.7431,
	})
	if err != nil {
		log.Fatalf("Failed seeding location: %v", err)
	}

	qty := 18
	_, err = inventorySvc.RecordObservation(ctx, &inventory.InventoryObservation{
		ID:           "inv-obs-wheel-austin",
		ProductID:    p1.ID,
		LocationID:   loc.ID,
		Quantity:     &qty,
		Availability: inventory.InStock,
		Price: &inventory.Money{
			Amount:   349.99,
			Currency: "USD",
		},
		ObservedAt: time.Now().UTC(),
	})
	if err != nil {
		log.Fatalf("Failed seeding inventory observation: %v", err)
	}

	// 5. Seed Asset & Evidence Claims
	asset, err := evidenceSvc.RegisterAsset(ctx, &evidence.Asset{
		ID:        "asset-catalog-ridge-2026",
		URI:       "gs://automotive-assets/catalogs/ridge_2026_wheels.pdf",
		MediaType: "application/pdf",
		SHA256:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	})
	if err != nil {
		log.Fatalf("Failed seeding asset: %v", err)
	}

	page := 14
	ev, err := evidenceSvc.RecordEvidence(ctx, &evidence.Evidence{
		ID:      "ev-page-14-region-2",
		AssetID: asset.ID,
		Page:    &page,
		Text:    "Part 84154233 20-inch Gloss Black Wheel rated for 2026 Apex Ridge 1500 4WD Trail trim",
		Extractor: evidence.ExtractorIdentity{
			Name:    "product-identity-extractor",
			Version: "v1.0",
		},
	})
	if err != nil {
		log.Fatalf("Failed seeding evidence: %v", err)
	}

	_, err = evidenceSvc.ProposeClaim(ctx, &evidence.Claim{
		ID:          "claim-wheel-oem-number",
		SubjectType: "product",
		SubjectID:   p1.ID,
		FieldPath:   "identifiers.oem_part_number",
		Value: product.TypedValue{
			Type:      "string",
			StringVal: "84154233",
		},
		Confidence:  0.96,
		Authority:   fitment.AuthDerived,
		EvidenceIDs: []string{ev.ID},
	})
	if err != nil {
		log.Fatalf("Failed seeding claim: %v", err)
	}

	fmt.Println("Successfully seeded dataset!")
	fmt.Printf("  - Products: 2\n  - Vehicles: 1\n  - Locations: 1\n  - Cloud SQL Endpoint: %s\n", postgres.DefaultDatabaseURL)
}
