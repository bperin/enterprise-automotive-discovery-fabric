package eval_test

import (
	"context"
	"testing"

	"enterprise-search/internal/discovery"
	"enterprise-search/internal/eval"
	"enterprise-search/internal/search"
)

func TestEvaluationHarness(t *testing.T) {
	engine := search.NewMemorySearchEngine()
	searchSvc := search.NewService(engine)
	discoverySvc := discovery.NewService(searchSvc)
	harness := eval.NewHarness(discoverySvc)

	reports, err := harness.RunBenchmark(context.Background())
	if err != nil {
		t.Fatalf("RunBenchmark failed: %v", err)
	}

	if len(reports) != 8 {
		t.Errorf("expected 8 benchmark comparison reports, got %d", len(reports))
	}

	for _, rep := range reports {
		if rep.Unified_Status != "PASSED" {
			t.Errorf("case %s: expected PASSED status for unified platform", rep.CaseID)
		}
	}
}
