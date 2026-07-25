package search

import (
	"context"
	"fmt"
	"strings"

	"enterprise-search/internal/product"
	"golang.org/x/sync/errgroup"
)

// Service coordinates discovery queries, multi-engine retrieval fusion, and indexing in DDD.
//
// GROUNDED QUERY GRAPH WORKFLOW PLACEMENT:
// The graph orchestrates search execution across authorized specialists without owning credential exchange or ETL:
//
//   Query
//     ↓
//   Identify Principal & Access Constraints (OIDC / Identity Claims)
//     ↓
//   Choose Relevant Sources & Retrieval Specialists
//     ↓
//   Run Authorized Searches in Parallel (Exact, Lexical, Semantic, Multimodal)
//     ↓
//   Reconcile Results & Verify Freshness / Fitment
//     ↓
//   Produce Evidence-Backed Grounded Response
type Service struct {
	engine SearchEngine
}

func NewService(engine SearchEngine) *Service {
	return &Service{engine: engine}
}

// ExecuteSearch runs the 5-stage authorized discovery workflow.
func (s *Service) ExecuteSearch(ctx context.Context, q SearchQuery) (*SearchResponse, error) {
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	if len(q.Modes) == 0 {
		q.Modes = []SearchMode{ModeExact, ModeLexical, ModeSemantic, ModeHybrid}
	}

	// 1. Identify User & Permissions Context
	// Principal and AccessPolicy are propagated into search backends.

	// 2. Run Authorized Searches in Parallel across requested modes
	g, gCtx := errgroup.WithContext(ctx)
	resultsChan := make(chan *SearchResponse, len(q.Modes))

	for _, mode := range q.Modes {
		m := mode
		g.Go(func() error {
			subQuery := q
			subQuery.Modes = []SearchMode{m}

			resp, err := s.engine.Search(gCtx, subQuery)
			if err != nil {
				return err
			}
			resultsChan <- resp
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		close(resultsChan)
		return nil, fmt.Errorf("search workflow failed: %w", err)
	}
	close(resultsChan)

	// 3. Reconcile Results & Deduplicate Candidate Sets
	dedupMap := make(map[string]SearchResult)
	var allWarnings []string
	traceID := ""

	for resp := range resultsChan {
		if traceID == "" {
			traceID = resp.TraceID
		}
		for _, w := range resp.Warnings {
			allWarnings = append(allWarnings, w)
		}
		for _, res := range resp.Results {
			if existing, found := dedupMap[res.EntityID]; !found || res.Score > existing.Score {
				dedupMap[res.EntityID] = res
			}
		}
	}

	var reconciled []SearchResult
	for _, res := range dedupMap {
		reconciled = append(reconciled, res)
	}

	// 4. Produce Evidence-Backed Response
	answer := ""
	if len(reconciled) > 0 {
		var titles []string
		for _, r := range reconciled {
			titles = append(titles, r.Title)
		}
		answer = fmt.Sprintf("Found %d authorized candidate records: %s", len(reconciled), strings.Join(titles, ", "))
	} else {
		answer = "No matching authorized records found."
	}

	return &SearchResponse{
		Answer:     answer,
		Results:    reconciled,
		TotalCount: len(reconciled),
		Warnings:   allWarnings,
		TraceID:    traceID,
	}, nil
}

func (s *Service) IndexProduct(ctx context.Context, p *product.Product) error {
	if p.ID == "" {
		return fmt.Errorf("%w: product ID required for indexing", ErrInvalidQuery)
	}

	fields := map[string]any{
		"category":        string(p.Category),
		"brand":           p.Brand,
		"oem":             p.OEM,
		"manufacturer_id": p.ManufacturerID,
		"publication":     string(p.Publication),
	}

	for _, ident := range p.Identifiers {
		fields[string(ident.Type)] = ident.Value
	}

	doc := SearchResult{
		EntityID:     p.ID,
		EntityType:   "product",
		Title:        p.CanonicalName,
		Snippet:      p.Description,
		Fields:       fields,
		SourceRefs:   p.SourceRefs,
		AccessPolicy: p.AccessPolicy,
	}

	return s.engine.Upsert(ctx, doc)
}

func (s *Service) Health(ctx context.Context) error {
	return s.engine.Health(ctx)
}
