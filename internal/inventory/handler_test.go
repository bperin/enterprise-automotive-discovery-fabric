package inventory_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/inventory"
)

func TestInventoryHandler(t *testing.T) {
	repo := inventory.NewMemoryRepository()
	svc := inventory.NewService(repo)
	handler := inventory.NewHandler(svc)
	authenticator := auth.NewMockAuthenticator()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authenticator)

	t.Run("Create Location HTTP", func(t *testing.T) {
		reqBody := map[string]any{
			"name":        "Dallas Hub",
			"postal_code": "75201",
		}
		raw, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/inventory/locations", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Record Observation HTTP", func(t *testing.T) {
		reqBody := map[string]any{
			"product_id":   "wheel-102",
			"location_id":  "loc-123",
			"quantity":     10,
			"availability": "in_stock",
		}
		raw, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/inventory/observations", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}
