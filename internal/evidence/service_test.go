package evidence_test

import (
	"context"
	"testing"

	"enterprise-search/internal/evidence"
	"enterprise-search/internal/product"
)

func TestEvidenceService(t *testing.T) {
	repo := evidence.NewMemoryRepository()
	svc := evidence.NewService(repo)
	ctx := context.Background()

	asset, err := svc.RegisterAsset(ctx, &evidence.Asset{
		URI:       "gs://automotive-assets/catalogs/wheels_2022.pdf",
		MediaType: "application/pdf",
		SHA256:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	})
	if err != nil {
		t.Fatalf("failed registering asset: %v", err)
	}

	page := 12
	ev, err := svc.RecordEvidence(ctx, &evidence.Evidence{
		AssetID: asset.ID,
		Page:    &page,
		Text:    "OEM Part Number 84154233 - Fits 2026 Apex Ridge",
		Extractor: evidence.ExtractorIdentity{
			Name:    "product-identity-extractor",
			Version: "v1.0",
		},
	})
	if err != nil {
		t.Fatalf("failed recording evidence: %v", err)
	}

	t.Run("Propose and Attest Claim", func(t *testing.T) {
		claim, err := svc.ProposeClaim(ctx, &evidence.Claim{
			SubjectType: "candidate_product",
			SubjectID:   "cand-101",
			FieldPath:   "identifiers.oem_part_number",
			Value: product.TypedValue{
				Type:      "string",
				StringVal: "84154233",
			},
			Confidence:  0.92,
			EvidenceIDs: []string{ev.ID},
		})
		if err != nil {
			t.Fatalf("propose claim failed: %v", err)
		}

		att, err := svc.AttestClaim(ctx, &evidence.Attestation{
			ClaimID: claim.ID,
			Status:  evidence.AttestationApproved,
			Validator: evidence.ValidatorIdentity{
				Type: "rule_engine",
				ID:   "oem_format_checker",
			},
		})
		if err != nil {
			t.Fatalf("attest claim failed: %v", err)
		}
		if att.Status != evidence.AttestationApproved {
			t.Errorf("expected Approved, got %s", att.Status)
		}
	})
}
