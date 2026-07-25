// Package discovery provides the Unified Answer and Discovery Gateway in DDD.
//
// ARCHITECTURAL TARGET & PURPOSE (from update.md):
// Replaces a fragmented collection of marketing-site search tools, brand-specific chatbots,
// vendor-specific RAG backends, and inconsistent knowledge stores with one unified, company-owned
// GCP-native answer and discovery gateway.
//
// Every brand website (e.g. ApexMotors, NovaMotors, Support Portal, Dealer Portal) calls this gateway.
// This establishes:
//   - One OIDC authentication boundary
//   - One permission & access control policy layer
//   - One source-authority precedence model (Specs DB > Approved Doc RAG > Website; Live API > CDC > Cached)
//   - One grounding & contradiction validation gate
//   - One shared response contract with evidence citations & suggested actions
package discovery

import (
	"enterprise-search/internal/auth"
	"enterprise-search/internal/product"
)

// FactType classifies information categories for source authority resolution.
type FactType string

const (
	FactVehicleSpecs FactType = "vehicle_specs"
	FactInventory    FactType = "inventory"
	FactWarranty     FactType = "warranty"
	FactPricing      FactType = "pricing"
	FactGeneralCopy  FactType = "general_marketing"
)

// VehicleContext carries vehicle trim and configuration details in discovery requests.
type VehicleContext struct {
	Year       int    `json:"year,omitempty"`
	Make       string `json:"make,omitempty"`
	Model      string `json:"model,omitempty"`
	Trim       string `json:"trim,omitempty"`
	Engine     string `json:"engine,omitempty"`
	Drivetrain string `json:"drivetrain,omitempty"`
}

// GeoContext carries geographical and postal details for location-aware queries.
type GeoContext struct {
	PostalCode  string  `json:"postal_code,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	RadiusMiles float64 `json:"radius_miles,omitempty"`
}

// Citation cites the exact authoritative source record supporting a response claim.
type Citation struct {
	SourceID string `json:"source_id"`
	Title    string `json:"title"`
	URI      string `json:"uri"`
	Excerpt  string `json:"excerpt,omitempty"`
	Page     *int   `json:"page,omitempty"`
}

// ActionType classifies call-to-action interactions.
type ActionType string

const (
	ActionLink        ActionType = "link"
	ActionDealerNav   ActionType = "dealer_navigation"
	ActionScheduleTest ActionType = "schedule_test_drive"
	ActionContactForm ActionType = "contact_form"
)

// SuggestedAction provides a structured call-to-action in discovery responses.
type SuggestedAction struct {
	Label      string     `json:"label"`
	URL        string     `json:"url"`
	ActionType ActionType `json:"action_type"`
}

// Warning highlights quality issues such as stale inventory, unverified fitment, or missing citations.
type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DiscoveryRequest is the unified input contract for all brand marketing properties.
type DiscoveryRequest struct {
	Query          string          `json:"query"`
	Brand          string          `json:"brand"`          // e.g. "ApexMotors", "NovaMotors"
	Channel        string          `json:"channel"`        // e.g. "public_web", "owner_support", "dealer_portal"
	Locale         string          `json:"locale"`         // e.g. "en_US"
	User           *auth.Principal `json:"user,omitempty"`
	Vehicle        *VehicleContext `json:"vehicle,omitempty"`
	Location       *GeoContext     `json:"location,omitempty"`
	ConversationID string          `json:"conversation_id,omitempty"`
}

// DiscoveryResult represents an individual normalized result item.
type DiscoveryResult struct {
	EntityID     string                  `json:"entity_id"`
	EntityType   string                  `json:"entity_type"` // e.g. "product", "document", "dealer"
	Title        string                  `json:"title"`
	Snippet      string                  `json:"snippet"`
	Score        float64                 `json:"score"`
	Fields       map[string]any          `json:"fields,omitempty"`
	SourceRef    product.SourceReference `json:"source_ref"`
	AccessPolicy auth.AccessPolicy       `json:"access_policy"`
}

// DiscoveryResponse is the unified response contract returned by the discovery gateway.
type DiscoveryResponse struct {
	Answer    string            `json:"answer"`
	Results   []DiscoveryResult `json:"results"`
	Citations []Citation        `json:"citations"`
	Actions   []SuggestedAction `json:"actions"`
	Warnings  []Warning         `json:"warnings"`
	Grounded  bool              `json:"grounded"`
	TraceID   string            `json:"trace_id"`
}
