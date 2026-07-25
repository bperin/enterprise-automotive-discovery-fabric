package evidence_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/evidence"
)

func TestEvidenceHandler(t *testing.T) {
	repo := evidence.NewMemoryRepository()
	svc := evidence.NewService(repo)
	handler := evidence.NewHandler(svc)
	authenticator := auth.NewMockAuthenticator()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authenticator)

	t.Run("Register Asset HTTP", func(t *testing.T) {
		reqBody := map[string]any{
			"uri":        "gs://automotive-assets/doc.pdf",
			"media_type": "application/pdf",
		}
		raw, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/assets", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Propose Claim HTTP", func(t *testing.T) {
		reqBody := map[string]any{
			"subject_type": "product",
			"subject_id":   "cand-200",
			"field_path":   "brand",
			"value": map[string]any{
				"type":       "string",
				"string_val": "ACDelco",
			},
		}
		raw, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/claims", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}
