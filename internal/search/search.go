package search

import (
	"context"
	"errors"
	"time"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/product"
)

// SearchMode defines the retrieval strategy in DDD.
type SearchMode string

const (
	ModeExact      SearchMode = "exact"
	ModeLexical    SearchMode = "lexical"
	ModeSemantic   SearchMode = "semantic"
	ModeMultimodal SearchMode = "multimodal"
	ModeHybrid     SearchMode = "hybrid"
)

// Filter represents a structured query constraint.
type Filter struct {
	Field    string `json:"field"`    // e.g. "year", "category", "location_id"
	Operator string `json:"operator"` // "=", "!=", ">", "<", "in", "contains"
	Value    any    `json:"value"`
}

// FacetRequest requests facet distribution for a field.
type FacetRequest struct {
	Field string `json:"field"`
	Size  int    `json:"size"`
}

// FacetValue represents one bucket in a facet result.
type FacetValue struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// FacetResult contains facet counts for a specific field.
type FacetResult struct {
	Field  string       `json:"field"`
	Values []FacetValue `json:"values"`
}

// SortField defines sorting criteria.
type SortField struct {
	Field      string `json:"field"`
	Descending bool   `json:"descending"`
}

// QueryContext carries contextual signals such as uploaded images or user location.
type QueryContext struct {
	ImageURI       string  `json:"image_uri,omitempty"`
	PostalCode     string  `json:"postal_code,omitempty"`
	MaxRadiusMiles float64 `json:"max_radius_miles,omitempty"`
}

// SearchQuery represents a unified discovery request contract in DDD.
// Every SearchQuery carries the user's OIDC Principal. All retrieval backends must apply access constraints
// against the indexed record's AccessPolicy BEFORE returning results.
type SearchQuery struct {
	Text      string          `json:"text"`
	Filters   []Filter        `json:"filters,omitempty"`
	Facets    []FacetRequest  `json:"facets,omitempty"`
	Sort      []SortField     `json:"sort,omitempty"`
	Cursor    string          `json:"cursor,omitempty"`
	PageSize  int             `json:"page_size"`
	Modes     []SearchMode    `json:"modes,omitempty"`
	Principal *auth.Principal `json:"principal,omitempty"`
	Context   QueryContext    `json:"context,omitempty"`
}

// FreshnessMetadata indicates source timestamp and staleness warnings.
type FreshnessMetadata struct {
	LastObservedAt time.Time `json:"last_observed_at"`
	SourceID       string    `json:"source_id"`
	IsStale        bool      `json:"is_stale"`
}

// ResultWarning highlights issues such as stale inventory or uncertain fitment.
type ResultWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// EvidenceReference cites the exact asset page or region supporting a result.
type EvidenceReference struct {
	EvidenceID string `json:"evidence_id"`
	AssetID    string `json:"asset_id"`
	Page       *int   `json:"page,omitempty"`
	Excerpt    string `json:"excerpt,omitempty"`
}

// SearchResult represents a normalized item in the search response.
type SearchResult struct {
	EntityID        string                  `json:"entity_id"`
	EntityType      string                  `json:"entity_type"` // e.g. "product", "document"
	Title           string                  `json:"title"`
	Snippet         string                  `json:"snippet"`
	Score           float64                 `json:"score"`
	RetrievalScores map[string]float64      `json:"retrieval_scores,omitempty"`
	Fields          map[string]any          `json:"fields,omitempty"`
	Facets          map[string][]string     `json:"facets,omitempty"`
	Evidence        []EvidenceReference     `json:"evidence,omitempty"`
	SourceRefs      []product.SourceReference `json:"source_refs,omitempty"`
	Freshness       *FreshnessMetadata      `json:"freshness,omitempty"`
	Warnings        []ResultWarning         `json:"warnings,omitempty"`
	AccessPolicy    auth.AccessPolicy       `json:"access_policy"`
}

// SearchResponse represents the grounded discovery result returned to clients in DDD.
type SearchResponse struct {
	Answer     string              `json:"answer,omitempty"`
	Results    []SearchResult      `json:"results"`
	Facets     []FacetResult       `json:"facets,omitempty"`
	TotalCount int                 `json:"total_count"`
	Warnings   []string            `json:"warnings,omitempty"`
	TraceID    string              `json:"trace_id"`
}

var (
	ErrEngineUnavailable = errors.New("search engine unavailable")
	ErrInvalidQuery      = errors.New("invalid search query")
)

// SearchEngine is the DDD consumer port for underlying search providers.
// In production, adapters exist for Agent Search, Agent Retrieval, and BigQuery Search.
// In local development, an in-memory/SQLite engine implements this interface.
type SearchEngine interface {
	Search(ctx context.Context, query SearchQuery) (*SearchResponse, error)
	Upsert(ctx context.Context, doc SearchResult) error
	Delete(ctx context.Context, documentIDs []string) error
	Health(ctx context.Context) error
}
