package product_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/product"
)

func TestProductHandler(t *testing.T) {
	repo := product.NewMemoryRepository()
	svc := product.NewService(repo)
	handler := product.NewHandler(svc)
	authenticator := auth.NewMockAuthenticator()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authenticator)

	t.Run("Create Product HTTP", func(t *testing.T) {
		body := map[string]any{
			"canonical_name": "Test Wheel 18-inch",
			"category":       "wheels",
			"brand":          "CustomBrand",
		}
		raw, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/products", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var created product.Product
		if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
			t.Fatalf("failed decoding JSON response: %v", err)
		}
		if created.CanonicalName != "Test Wheel 18-inch" {
			t.Errorf("expected 'Test Wheel 18-inch', got '%s'", created.CanonicalName)
		}
	})

	t.Run("List Products HTTP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/products?limit=10", nil)
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", rec.Code)
		}
	})
}
