package fitment_test

import (
	"context"
	"testing"

	"enterprise-search/internal/fitment"
)

func TestFitmentService(t *testing.T) {
	repo := fitment.NewMemoryRepository()
	svc := fitment.NewService(repo)
	ctx := context.Background()

	vehicle, err := svc.SaveVehicleConfig(ctx, &fitment.VehicleConfiguration{
		Year:  2022,
		Make:  "ApexMotors",
		Model: "Ridge 1500",
		Trim:  "LT",
	})
	if err != nil {
		t.Fatalf("failed creating vehicle config: %v", err)
	}

	t.Run("Create and Verify Fitment Assertion", func(t *testing.T) {
		assertion, err := svc.CreateAssertion(ctx, &fitment.FitmentAssertion{
			ProductID:              "wheel-102",
			VehicleConfigurationID: vehicle.ID,
			Compatibility:          fitment.DirectFit,
			Authority:              fitment.AuthAuthoritative,
			Confidence:             1.0,
			VerificationStatus:     fitment.VerifiedBySourceContract,
		})
		if err != nil {
			t.Fatalf("failed creating fitment assertion: %v", err)
		}

		verified, err := svc.VerifyFitment(ctx, "wheel-102", vehicle.ID)
		if err != nil {
			t.Fatalf("verify fitment failed: %v", err)
		}
		if verified.Compatibility != fitment.DirectFit {
			t.Errorf("expected DirectFit, got %s", verified.Compatibility)
		}
		if verified.ID != assertion.ID {
			t.Errorf("expected assertion ID %s, got %s", assertion.ID, verified.ID)
		}
	})

	t.Run("Verify Unknown Fitment", func(t *testing.T) {
		unknown, err := svc.VerifyFitment(ctx, "unknown-part", vehicle.ID)
		if err != nil {
			t.Fatalf("verify fitment failed: %v", err)
		}
		if unknown.Compatibility != fitment.UnknownFitment {
			t.Errorf("expected UnknownFitment, got %s", unknown.Compatibility)
		}
	})
}
