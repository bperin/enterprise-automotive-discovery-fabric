package discovery

import (
	"context"
	"fmt"
	"strings"

	"enterprise-search/internal/agents"
	"enterprise-search/internal/product"
	"enterprise-search/internal/search"
	"github.com/google/uuid"
)

// Service provides the Unified Answer and Discovery Gateway logic in DDD backed by ADK 2.0 agent graph workflow.
type Service struct {
	searchSvc     *search.Service
	groundingGate *GroundingGate
	orchestrator  *agents.GraphOrchestrator
}

func NewService(searchSvc *search.Service) *Service {
	orch, _ := agents.NewGraphOrchestrator(searchSvc)
	return &Service{
		searchSvc:     searchSvc,
		groundingGate: NewGroundingGate(),
		orchestrator:  orch,
	}
}

// ProcessQuery runs the unified discovery gateway pipeline with ADK 2.0 multi-agent graph orchestration.
func (s *Service) ProcessQuery(ctx context.Context, req DiscoveryRequest) (*DiscoveryResponse, error) {
	if strings.TrimSpace(req.Query) == "" {
		return nil, fmt.Errorf("query string is required")
	}

	if req.Brand == "" {
		req.Brand = "ApexMotors"
	}
	if req.Channel == "" {
		req.Channel = "public_web"
	}

	postalCode := ""
	if req.Location != nil {
		postalCode = req.Location.PostalCode
	}

	// 1. Run ADK 2.0 Multi-Agent Graph Orchestration
	initialState := &agents.DiscoveryGraphState{
		Query:      req.Query,
		Brand:      req.Brand,
		Channel:    req.Channel,
		PostalCode: postalCode,
	}

	graphState, err := s.orchestrator.ExecuteGraph(ctx, initialState)
	if err != nil {
		return nil, fmt.Errorf("ADK 2.0 graph execution failed: %w", err)
	}

	// 2. Map SearchResults to DiscoveryResults & Citations
	var results []DiscoveryResult
	var citations []Citation

	for _, sr := range graphState.SearchResults {
		var srcRef product.SourceReference
		if len(sr.SourceRefs) > 0 {
			srcRef = sr.SourceRefs[0]
		} else {
			srcRef = product.SourceReference{
				SourceID:   "search_index",
				SourceType: "agent_search",
				SourceURI:  "https://discovery.fabric.io/catalog/" + sr.EntityID,
			}
		}

		dRes := DiscoveryResult{
			EntityID:     sr.EntityID,
			EntityType:   sr.EntityType,
			Title:        sr.Title,
			Snippet:      sr.Snippet,
			Score:        sr.Score,
			Fields:       sr.Fields,
			SourceRef:    srcRef,
			AccessPolicy: sr.AccessPolicy,
		}
		results = append(results, dRes)

		citations = append(citations, Citation{
			SourceID: srcRef.SourceID,
			Title:    sr.Title,
			URI:      srcRef.SourceURI,
			Excerpt:  sr.Snippet,
		})
	}

	// 3. Grounding Gate & Policy Check
	verdict, rawAnswer, gateWarnings := s.groundingGate.Evaluate(ctx, req, graphState.Answer, results, citations)

	var warnings []Warning
	for _, gw := range gateWarnings {
		warnings = append(warnings, gw)
	}

	grounded := graphState.Grounded && (verdict == VerdictPass || verdict == VerdictQualify)

	// 4. Generate Suggested Call-to-Action Interactions
	actions := s.generateSuggestedActions(req, results)

	traceID := "trace-agent-graph-" + uuid.New().String()[:8]

	return &DiscoveryResponse{
		Answer:    rawAnswer,
		Results:   results,
		Citations: citations,
		Actions:   actions,
		Warnings:  warnings,
		Grounded:  grounded,
		TraceID:   traceID,
	}, nil
}

func (s *Service) generateSuggestedActions(req DiscoveryRequest, results []DiscoveryResult) []SuggestedAction {
	var actions []SuggestedAction

	if req.Location != nil && req.Location.PostalCode != "" {
		actions = append(actions, SuggestedAction{
			Label:      fmt.Sprintf("Find Nearby Dealers near %s", req.Location.PostalCode),
			URL:        fmt.Sprintf("https://www.%s.com/dealers?postal_code=%s", strings.ToLower(req.Brand), req.Location.PostalCode),
			ActionType: ActionDealerNav,
		})
	}

	if len(results) > 0 {
		actions = append(actions, SuggestedAction{
			Label:      fmt.Sprintf("Schedule Test Drive for %s", results[0].Title),
			URL:        fmt.Sprintf("https://www.%s.com/test-drive?model=%s", strings.ToLower(req.Brand), results[0].EntityID),
			ActionType: ActionScheduleTest,
		})
	}

	return actions
}
