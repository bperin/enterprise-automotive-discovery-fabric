package discovery_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/discovery"
	"enterprise-search/internal/search"
)

func TestDiscoveryHandler(t *testing.T) {
	engine := search.NewMemorySearchEngine()
	searchSvc := search.NewService(engine)
	discoverySvc := discovery.NewService(searchSvc)
	handler := discovery.NewHandler(discoverySvc)
	authenticator := auth.NewMockAuthenticator()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authenticator)

	t.Run("POST /v1/discovery/query HTTP", func(t *testing.T) {
		reqBody := map[string]any{
			"query":   "2026 Apex Hauler towing capacity",
			"brand":   "ApexMotors",
			"channel": "public_web",
		}
		raw, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/discovery/query", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp discovery.DiscoveryResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed decoding discovery response: %v", err)
		}
		if resp.TraceID == "" {
			t.Errorf("expected non-empty trace_id")
		}
	})
}
