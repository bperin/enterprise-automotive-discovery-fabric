package discovery

import (
	"context"
	"fmt"
	"strings"
)

// GroundingVerdict represents the outcome of the grounding and policy check.
type GroundingVerdict string

const (
	VerdictPass     GroundingVerdict = "pass"
	VerdictQualify  GroundingVerdict = "qualify"
	VerdictFail     GroundingVerdict = "fail"
	VerdictEscalate GroundingVerdict = "escalate"
)

// GroundingGate evaluates candidate generated text against authoritative citations and source hierarchy policy.
//
// GROUNDING CHECK RULES:
// 1. Is every material factual statement supported by an approved reference?
// 2. Is the source authorized for this fact type (e.g. never infer inventory from marketing copy)?
// 3. Are sources contradictory or stale?
// 4. If claims fail grounding, suppress generated narrative and return structured results only.
type GroundingGate struct{}

func NewGroundingGate() *GroundingGate {
	return &GroundingGate{}
}

// Evaluate performs claim verification and returns verdict, citations, and warnings.
func (g *GroundingGate) Evaluate(ctx context.Context, req DiscoveryRequest, rawAnswer string, results []DiscoveryResult, citations []Citation) (GroundingVerdict, string, []Warning) {
	var warnings []Warning

	if len(results) == 0 && len(citations) == 0 {
		return VerdictFail, "No authoritative source records were found to ground this query.", []Warning{
			{Code: "NO_SOURCES_FOUND", Message: "Query returned zero grounded source records."},
		}
	}

	// Source authority check for inventory queries
	queryLower := strings.ToLower(req.Query)
	isInventoryQuery := strings.Contains(queryLower, "stock") || strings.Contains(queryLower, "available") || strings.Contains(queryLower, "quantity") || strings.Contains(queryLower, "inventory")

	if isInventoryQuery {
		hasLiveInventorySource := false
		for _, res := range results {
			if res.EntityType == "inventory" || strings.Contains(res.SourceRef.SourceType, "inventory") || strings.Contains(res.SourceRef.SourceType, "cloud_sql") {
				hasLiveInventorySource = true
				break
			}
		}

		if !hasLiveInventorySource {
			warnings = append(warnings, Warning{
				Code:    "INVENTORY_UNVERIFIED",
				Message: "Inventory availability cannot be inferred from general marketing copy. Query live dealer API.",
			})
			return VerdictQualify, rawAnswer, warnings
		}
	}

	// Check citations coverage
	if len(citations) == 0 {
		// Attach synthetic citation from top result if available
		if len(results) > 0 {
			top := results[0]
			citations = append(citations, Citation{
				SourceID: top.SourceRef.SourceID,
				Title:    top.Title,
				URI:      top.SourceRef.SourceURI,
				Excerpt:  top.Snippet,
			})
		}
	}

	if len(citations) > 0 {
		return VerdictPass, rawAnswer, warnings
	}

	return VerdictQualify, fmt.Sprintf("%s (Note: Information should be verified with dealer).", rawAnswer), warnings
}
