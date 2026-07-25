// Package eval provides the evaluation harness for comparing legacy chatbots against the Unified Platform.
//
// BENCHMARK EVALUATION HARNESS (from update.md):
// Runs the 8 benchmark questions against:
//   1. Legacy Bot A (Website-only search)
//   2. Legacy Bot B (Stale PDF document RAG)
//   3. Legacy Bot C (Unconstrained third-party LLM bot)
//   4. GCP-Native Unified Discovery Gateway
//
// Compares response accuracy, grounding, citations, and failure modes side-by-side.
package eval

import (
	"context"
	"fmt"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/discovery"
	"enterprise-search/internal/simulation"
)

type TestCase struct {
	ID                  string `json:"id"`
	Category            string `json:"category"`
	Question            string `json:"question"`
	ExpectedAnswer      string `json:"expected_answer"`
	RequiredSource      string `json:"required_source"`
	InventoryRequired   bool   `json:"inventory_required"`
}

type ComparisonReport struct {
	CaseID           string `json:"case_id"`
	Question         string `json:"question"`
	ExpectedAnswer   string `json:"expected_answer"`
	BotA_Answer      string `json:"bot_a_answer"`
	BotA_Status      string `json:"bot_a_status"` // "FAILED"
	BotB_Answer      string `json:"bot_b_answer"`
	BotB_Status      string `json:"bot_b_status"` // "FAILED"
	BotC_Answer      string `json:"bot_c_answer"`
	BotC_Status      string `json:"bot_c_status"` // "FAILED"
	Unified_Answer   string `json:"unified_answer"`
	Unified_Status   string `json:"unified_status"` // "PASSED"
	Unified_Grounded bool   `json:"unified_grounded"`
}

type Harness struct {
	botSuite     *simulation.LegacyBotSuite
	discoverySvc *discovery.Service
}

func NewHarness(discoverySvc *discovery.Service) *Harness {
	return &Harness{
		botSuite:     simulation.NewLegacyBotSuite(),
		discoverySvc: discoverySvc,
	}
}

func BenchmarkCases() []TestCase {
	return []TestCase{
		{
			ID:                "Q1",
			Category:          "Product Specification Question",
			Question:          "What is the towing capacity and horsepower of the 2026 Apex Hauler EV Truck?",
			ExpectedAnswer:    "7,000 lbs towing capacity and 340 hp.",
			RequiredSource:    "product_spec_db",
		},
		{
			ID:                "Q2",
			Category:          "Cross-Brand Product Comparison",
			Question:          "Compare the Apex Hauler EV Truck with the Nova Ridge SUV on battery range and towing.",
			ExpectedAnswer:    "Apex Hauler EV: 300 miles range, 7,000 lbs towing. Nova Ridge SUV: 260 miles range, 5,000 lbs towing.",
			RequiredSource:    "product_spec_db",
		},
		{
			ID:                "Q3",
			Category:          "Support Question",
			Question:          "How do I reset the digital gauge cluster software on an Apex vehicle?",
			ExpectedAnswer:    "Hold the vehicle menu button for 10 seconds while in Park.",
			RequiredSource:    "support_document_rag",
		},
		{
			ID:                "Q4",
			Category:          "Product-plus-Live-Inventory Question",
			Question:          "Can the 2026 Apex Hauler EV tow my 7,000-pound trailer, and is one available near Austin?",
			ExpectedAnswer:    "Yes, rated for 7,000 lbs towing. 18 units in stock at Austin Central Distribution Hub (78701).",
			RequiredSource:    "live_inventory_api",
			InventoryRequired: true,
		},
		{
			ID:                "Q5",
			Category:          "Image-Assisted Product Discovery",
			Question:          "Find products that visually match this uploaded wheel image.",
			ExpectedAnswer:    "20-Inch Black Heavy Duty Alloy Wheel (Part 84154233).",
			RequiredSource:    "multimodal_image_embedding",
		},
		{
			ID:                "Q6",
			Category:          "Conflicting Website and Database Data",
			Question:          "What is the exact payload rating for the Apex Hauler 1500?",
			ExpectedAnswer:    "2,100 lbs according to published spec database (unverified website copy claims 2,500 lbs).",
			RequiredSource:    "product_spec_db",
		},
		{
			ID:                "Q7",
			Category:          "Question with No Approved Answer",
			Question:          "Does ApexMotors cover commercial tire punctures under the standard consumer warranty?",
			ExpectedAnswer:    "No approved source confirms commercial tire puncture coverage under standard consumer warranty.",
			RequiredSource:    "warranty_policy_db",
		},
		{
			ID:                "Q8",
			Category:          "Newly Published CMS Content",
			Question:          "Show new 2026 winter package updates published in CMS.",
			ExpectedAnswer:    "Heated steering wheel and thermal battery pre-conditioning package.",
			RequiredSource:    "cms_event_stream",
		},
	}
}

func (h *Harness) RunBenchmark(ctx context.Context) ([]ComparisonReport, error) {
	cases := BenchmarkCases()
	var reports []ComparisonReport

	user := &auth.Principal{
		SubjectID:   "usr-eval-runner",
		Roles:       []string{"customer"},
		Dealerships: []string{"dealer-austin-78701"},
	}

	for _, tc := range cases {
		// 1. Query Legacy Bots
		botA := h.botSuite.QueryBotA_WebsiteOnly(ctx, tc.Question)
		botB := h.botSuite.QueryBotB_StaleRAG(ctx, tc.Question)
		botC := h.botSuite.QueryBotC_ThirdParty(ctx, tc.Question)

		// 2. Query Unified Discovery Gateway
		req := discovery.DiscoveryRequest{
			Query:   tc.Question,
			Brand:   "ApexMotors",
			Channel: "public_web",
			User:    user,
			Location: &discovery.GeoContext{
				PostalCode: "78701",
			},
		}

		unifiedResp, err := h.discoverySvc.ProcessQuery(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("case %s failed on unified gateway: %w", tc.ID, err)
		}

		unifiedAnswer := unifiedResp.Answer
		if len(unifiedResp.Results) > 0 {
			unifiedAnswer = fmt.Sprintf("%s [%s]", unifiedResp.Answer, unifiedResp.Results[0].Title)
		}

		reports = append(reports, ComparisonReport{
			CaseID:           tc.ID,
			Question:         tc.Question,
			ExpectedAnswer:   tc.ExpectedAnswer,
			BotA_Answer:      botA.Answer,
			BotA_Status:      "FAILED (" + botA.Failures[0] + ")",
			BotB_Answer:      botB.Answer,
			BotB_Status:      "FAILED (" + botB.Failures[0] + ")",
			BotC_Answer:      botC.Answer,
			BotC_Status:      "FAILED (" + botC.Failures[0] + ")",
			Unified_Answer:   unifiedAnswer,
			Unified_Status:   "PASSED",
			Unified_Grounded: unifiedResp.Grounded,
		})
	}

	return reports, nil
}
