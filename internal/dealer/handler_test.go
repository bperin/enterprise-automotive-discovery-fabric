package dealer_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/dealer"
)

func TestDealerHandler(t *testing.T) {
	repo := dealer.NewMemoryRepository()
	handler := dealer.NewHandler(repo)
	authenticator := auth.NewMockAuthenticator()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authenticator)

	t.Run("Create Dealer HTTP", func(t *testing.T) {
		reqBody := dealer.Dealer{
			ID:             "dealer-austin-02",
			Name:           "Apex Motors Austin",
			BrandIDs:       []string{"ApexMotors"},
			ServiceEnabled: true,
			PartsEnabled:   true,
		}
		raw, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/dealers", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Get Dealer HTTP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/dealers/dealer-austin-01", nil)
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("List Dealers HTTP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/dealers", nil)
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}
