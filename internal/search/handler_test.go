package search_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/search"
)

func TestSearchHandler(t *testing.T) {
	engine := search.NewMemorySearchEngine()
	svc := search.NewService(engine)
	handler := search.NewHandler(svc)
	authenticator := auth.NewMockAuthenticator()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authenticator)

	t.Run("POST /v1/search HTTP", func(t *testing.T) {
		reqBody := map[string]any{
			"text":      "heavy duty wheels",
			"page_size": 10,
		}
		raw, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/search", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp search.SearchResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed decoding search response: %v", err)
		}
		if resp.TraceID == "" {
			t.Errorf("expected non-empty trace_id")
		}
	})
}
