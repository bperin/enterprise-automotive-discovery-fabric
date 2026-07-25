package fitment_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/fitment"
)

func TestFitmentHandler(t *testing.T) {
	repo := fitment.NewMemoryRepository()
	svc := fitment.NewService(repo)
	handler := fitment.NewHandler(svc)
	authenticator := auth.NewMockAuthenticator()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authenticator)

	t.Run("Create Assertion HTTP", func(t *testing.T) {
		reqBody := map[string]any{
			"product_id":               "wheel-101",
			"vehicle_configuration_id": "veh-2026-apex-ridge",
			"compatibility":            "direct_fit",
			"authority":                "authoritative",
			"confidence":               1.0,
		}
		raw, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/fitment/assertions", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Verify Fitment HTTP", func(t *testing.T) {
		reqBody := map[string]any{
			"product_id": "wheel-101",
			"vehicle_id": "veh-2026-apex-ridge",
		}
		raw, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/fitment/verify", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}
