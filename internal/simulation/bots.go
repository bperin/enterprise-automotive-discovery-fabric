// Package simulation provides intentionally broken legacy chatbot implementations.
//
// LEGACY FRAGMENTED BOT SIMULATION (from update.md):
// Demonstrates why separate, vendor-specific chatbots on different brand properties fail:
//   - Bot A (Website-only search): Crawls HTML text; misses structured product spec DB and live stock APIs.
//   - Bot B (Stale Document RAG): Queries static PDF manual index; returns outdated 2024 specifications.
//   - Bot C (Third-Party Vendor Bot): Unconstrained LLM API; hallucinates specs and approves non-existent warranties.
package simulation

import (
	"context"
	"strings"
)

type BotResponse struct {
	BotName  string   `json:"bot_name"`  // "Bot A (Website Search)", "Bot B (Stale RAG)", "Bot C (Third-Party Bot)"
	Answer   string   `json:"answer"`
	Source   string   `json:"source"`
	Failures []string `json:"failures"`  // Highlighted failure reasons
}

type LegacyBotSuite struct{}

func NewLegacyBotSuite() *LegacyBotSuite {
	return &LegacyBotSuite{}
}

func (s *LegacyBotSuite) QueryBotA_WebsiteOnly(ctx context.Context, question string) BotResponse {
	q := strings.ToLower(question)
	failures := []string{
		"Lacks access to structured product specification database",
		"Lacks access to live dealer inventory API",
	}

	if strings.Contains(q, "towing") || strings.Contains(q, "capacity") {
		return BotResponse{
			BotName:  "Bot A (Website Search)",
			Answer:   "According to the 2024 marketing blog post, Apex trucks have strong towing capabilities. Please consult your owner manual for exact figures.",
			Source:   "ApexMotors Marketing Web Crawler",
			Failures: append(failures, "Returned generic marketing copy without exact rated specification"),
		}
	}

	if strings.Contains(q, "available") || strings.Contains(q, "austin") || strings.Contains(q, "inventory") {
		return BotResponse{
			BotName:  "Bot A (Website Search)",
			Answer:   "Please contact your local dealer to inquire about vehicle stock in Austin.",
			Source:   "ApexMotors Static Site Index",
			Failures: append(failures, "Cannot query live inventory API"),
		}
	}

	return BotResponse{
		BotName:  "Bot A (Website Search)",
		Answer:   "Search results for query: " + question,
		Source:   "Website Crawler",
		Failures: failures,
	}
}

func (s *LegacyBotSuite) QueryBotB_StaleRAG(ctx context.Context, question string) BotResponse {
	q := strings.ToLower(question)
	failures := []string{
		"Indexed against outdated 2024 PDF documentation corpus",
		"Stale content not updated upon CMS publication",
	}

	if strings.Contains(q, "towing") || strings.Contains(q, "apex hauler") {
		return BotResponse{
			BotName:  "Bot B (Stale RAG)",
			Answer:   "The 2024 Apex Hauler 1500 has a maximum towing capacity of 5,000 lbs and 280 horsepower.",
			Source:   "2024_Apex_Product_Catalog.pdf",
			Failures: append(failures, "Returned outdated 2024 specs (2026 EV model is rated for 7,000 lbs and 340 hp)"),
		}
	}

	if strings.Contains(q, "payload") {
		return BotResponse{
			BotName:  "Bot B (Stale RAG)",
			Answer:   "The payload rating is 2,500 lbs based on the early brochure draft.",
			Source:   "2024_Brochure_Draft.pdf",
			Failures: append(failures, "Returned draft brochure spec contradicting verified spec database (2,100 lbs)"),
		}
	}

	return BotResponse{
		BotName:  "Bot B (Stale RAG)",
		Answer:   "Retrieved document excerpt for " + question,
		Source:   "Stale RAG Corpus",
		Failures: failures,
	}
}

func (s *LegacyBotSuite) QueryBotC_ThirdParty(ctx context.Context, question string) BotResponse {
	q := strings.ToLower(question)
	failures := []string{
		"Unconstrained generative LLM without grounding gate or source hierarchy",
		"Hallucinates features and unauthorized policy promises",
	}

	if strings.Contains(q, "towing") || strings.Contains(q, "trailer") {
		return BotResponse{
			BotName:  "Bot C (3rd Party Bot)",
			Answer:   "Yes! The 2026 Apex Hauler EV Truck can easily tow up to 12,000 lbs and includes a free trailer hitch package.",
			Source:   "3rd Party Chatbot Vendor LLM",
			Failures: append(failures, "Hallucinated 12,000 lbs towing capacity (actual rating 7,000 lbs) and invented free package"),
		}
	}

	if strings.Contains(q, "warranty") || strings.Contains(q, "tire") {
		return BotResponse{
			BotName:  "Bot C (3rd Party Bot)",
			Answer:   "ApexMotors fully covers all commercial tire punctures and rim replacements under standard warranty at no cost.",
			Source:   "3rd Party Chatbot Vendor LLM",
			Failures: append(failures, "Invented non-existent warranty coverage creating legal and financial liability"),
		}
	}

	return BotResponse{
		BotName:  "Bot C (3rd Party Bot)",
		Answer:   "I am happy to assist! " + question,
		Source:   "Generic Chatbot",
		Failures: failures,
	}
}
