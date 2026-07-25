package search

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// MemorySearchEngine is an in-memory implementation of the SearchEngine port.
// In production, this would be replaced by Agent Search, Agent Retrieval, or BigQuery Search adapters.
//
// PERMISSION FILTERING RULE:
// The search engine MUST filter candidate records using q.Principal.CanAccess(doc.AccessPolicy) BEFORE
// returning them to callers or grounded LLM synthesis stages. Restricted commercial terms, internal legal
// documents, or competitor supplier pricing must never be retrieved for unauthorized principals.
type MemorySearchEngine struct {
	mu   sync.RWMutex
	docs map[string]SearchResult
}

func NewMemorySearchEngine() *MemorySearchEngine {
	return &MemorySearchEngine{
		docs: make(map[string]SearchResult),
	}
}

func (e *MemorySearchEngine) Search(ctx context.Context, q SearchQuery) (*SearchResponse, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var matched []SearchResult
	queryText := strings.ToLower(strings.TrimSpace(q.Text))
	queryTerms := strings.Fields(queryText)

	for _, doc := range e.docs {
		// 1. Mandatory Security & Authorization Pre-Filtering
		if q.Principal != nil {
			if !q.Principal.CanAccess(doc.AccessPolicy) {
				continue // Skip unauthorized records immediately
			}
		} else if doc.AccessPolicy.Visibility != "" && doc.AccessPolicy.Visibility != "public" {
			continue // Unauthenticated queries see only public records
		}

		matchScore := 0.0

		// 2. Check term matches across title, snippet, and string fields
		if len(queryTerms) > 0 {
			for _, term := range queryTerms {
				if strings.Contains(strings.ToLower(doc.Title), term) {
					matchScore += 1.0
				}
				if strings.Contains(strings.ToLower(doc.Snippet), term) {
					matchScore += 0.5
				}
				for _, fieldVal := range doc.Fields {
					if str, ok := fieldVal.(string); ok {
						if strings.Contains(strings.ToLower(str), term) {
							matchScore += 0.3
						}
					}
				}
			}
		} else {
			matchScore = 1.0
		}

		// 3. Apply exact structured filters
		passesFilters := true
		for _, filter := range q.Filters {
			if val, exists := doc.Fields[filter.Field]; exists {
				if filter.Operator == "=" && val != filter.Value {
					passesFilters = false
					break
				}
			}
		}

		if passesFilters && matchScore > 0 {
			docCopy := doc
			docCopy.Score = matchScore
			matched = append(matched, docCopy)
		}
	}

	// Apply paging
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	if len(matched) > pageSize {
		matched = matched[:pageSize]
	}

	traceID := "trace-" + uuid.New().String()[:8]
	answer := ""
	if len(matched) > 0 {
		answer = "Found matching records in discovery fabric."
	} else {
		answer = "No matching records found for query."
	}

	return &SearchResponse{
		Answer:     answer,
		Results:    matched,
		TotalCount: len(matched),
		TraceID:    traceID,
	}, nil
}

func (e *MemorySearchEngine) Upsert(ctx context.Context, doc SearchResult) error {
	if doc.EntityID == "" {
		return ErrInvalidQuery
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	e.docs[doc.EntityID] = doc
	return nil
}

func (e *MemorySearchEngine) Delete(ctx context.Context, documentIDs []string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, id := range documentIDs {
		delete(e.docs, id)
	}
	return nil
}

func (e *MemorySearchEngine) Health(ctx context.Context) error {
	return nil
}
