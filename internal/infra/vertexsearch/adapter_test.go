package vertexsearch_test

import (
	"context"
	"testing"

	"enterprise-search/internal/infra/vertexsearch"
	"enterprise-search/internal/search"
)

func TestVertexSearchAdapterDeprecated(t *testing.T) {
	adapter := vertexsearch.NewAdapter(vertexsearch.Config{
		ProjectID:  "slap-agent-builder",
		Location:   "global",
		DataSpecID: "marketing-content-store",
	}, "mock-token")

	resp, err := adapter.Search(context.Background(), search.SearchQuery{
		Text:     "Apex Hauler EV",
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TraceID == "" {
		t.Errorf("expected non-empty trace ID")
	}
}
