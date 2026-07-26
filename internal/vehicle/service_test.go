package vehicle_test

import (
	"context"
	"testing"

	"enterprise-search/internal/vehicle"
)

func TestVehicleRepository(t *testing.T) {
	repo := vehicle.NewMemoryRepository()
	ctx := context.Background()

	v := &vehicle.VehicleConfiguration{
		ID:         "veh-2026-apex-ridge-1500",
		Year:       2026,
		Make:       "ApexMotors",
		Model:      "Ridge 1500",
		Trim:       "Trail Z71",
		Engine:     "5.3L V8",
		Drivetrain: "4WD",
	}

	if err := repo.SaveConfiguration(ctx, v); err != nil {
		t.Fatalf("failed to save vehicle config: %v", err)
	}

	fetched, err := repo.GetConfiguration(ctx, "veh-2026-apex-ridge-1500")
	if err != nil {
		t.Fatalf("failed to get vehicle config: %v", err)
	}

	if fetched.Model != "Ridge 1500" {
		t.Errorf("expected Ridge 1500, got %s", fetched.Model)
	}

	configs, err := repo.FindConfigurations(ctx, "ApexMotors", "Ridge 1500", 2026)
	if err != nil {
		t.Fatalf("failed to find configurations: %v", err)
	}

	if len(configs) != 1 {
		t.Errorf("expected 1 matching configuration, got %d", len(configs))
	}
}
