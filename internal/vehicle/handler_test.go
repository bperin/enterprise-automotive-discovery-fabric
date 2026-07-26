package vehicle_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/vehicle"
)

func TestVehicleHandler(t *testing.T) {
	repo := vehicle.NewMemoryRepository()
	handler := vehicle.NewHandler(repo)
	authenticator := auth.NewMockAuthenticator()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authenticator)

	t.Run("Create Vehicle HTTP", func(t *testing.T) {
		reqBody := vehicle.VehicleConfiguration{
			ID:         "veh-2026-apex-ridge-1500",
			Year:       2026,
			Make:       "ApexMotors",
			Model:      "Ridge 1500",
			Trim:       "Trail Z71",
			Engine:     "5.3L V8",
			Drivetrain: "4WD",
		}
		raw, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/vehicles", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Get Vehicle HTTP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/vehicles/veh-2026-apex-ridge-1500", nil)
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}
