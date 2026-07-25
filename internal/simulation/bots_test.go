package simulation_test

import (
	"context"
	"testing"

	"enterprise-search/internal/simulation"
)

func TestLegacyBotSuite(t *testing.T) {
	suite := simulation.NewLegacyBotSuite()
	ctx := context.Background()

	t.Run("Bot A Website Search Failures", func(t *testing.T) {
		res := suite.QueryBotA_WebsiteOnly(ctx, "2026 Apex Hauler towing capacity")
		if len(res.Failures) == 0 {
			t.Errorf("expected Bot A to record failures")
		}
	})

	t.Run("Bot B Stale RAG Failures", func(t *testing.T) {
		res := suite.QueryBotB_StaleRAG(ctx, "Apex Hauler towing capacity")
		if len(res.Failures) == 0 {
			t.Errorf("expected Bot B to record stale RAG failures")
		}
	})

	t.Run("Bot C Third-Party Hallucination Failures", func(t *testing.T) {
		res := suite.QueryBotC_ThirdParty(ctx, "Does warranty cover commercial tire punctures?")
		if len(res.Failures) == 0 {
			t.Errorf("expected Bot C to record hallucination failures")
		}
	})
}
