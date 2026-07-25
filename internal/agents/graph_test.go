package agents

import (
	"context"
	"testing"

	"enterprise-search/internal/search"
)

func TestGraphOrchestrator_ExecuteGraph(t *testing.T) {
	memEngine := search.NewMemorySearchEngine()
	searchSvc := search.NewService(memEngine)

	orch, err := NewGraphOrchestrator(searchSvc)
	if err != nil {
		t.Fatalf("failed to create GraphOrchestrator: %v", err)
	}

	state := &DiscoveryGraphState{
		Query:  "2026 Apex Ridge Trail towing capacity",
		Brand:  "ApexMotors",
		Channel: "public_web",
	}

	resState, err := orch.ExecuteGraph(context.Background(), state)
	if err != nil {
		t.Fatalf("ExecuteGraph failed: %v", err)
	}

	if resState.Intent != IntentProductSpec {
		t.Errorf("expected intent product_spec, got %s", resState.Intent)
	}

	if len(resState.Trace) == 0 {
		t.Errorf("expected execution trace steps, got none")
	}
}
