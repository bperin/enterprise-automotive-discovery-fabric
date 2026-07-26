package content_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/content"
)

func TestContentHandler(t *testing.T) {
	repo := content.NewMemoryRepository()
	handler := content.NewHandler(repo)
	authenticator := auth.NewMockAuthenticator()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authenticator)

	t.Run("Create Content HTTP", func(t *testing.T) {
		reqBody := content.ContentItem{
			ID:          "content-2026-ridge-launch",
			BrandID:     "ApexMotors",
			ContentType: content.TypeModelPage,
			Title:       "2026 Apex Ridge Trail Launch Announcement",
			Body:        "Introducing the 2026 Apex Ridge Trail with 7,000 lbs max towing capacity.",
			Status:      content.StatusPublished,
		}
		raw, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/content", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Get Content HTTP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/content/content-2026-ridge-launch", nil)
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}
