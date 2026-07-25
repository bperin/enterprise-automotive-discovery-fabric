package product

import (
	"context"
	"errors"
	"time"

	"enterprise-search/internal/auth"
)

// ProductCategory classifies products within the discovery fabric.
type ProductCategory string

const (
	CategoryWheels     ProductCategory = "wheels"
	CategoryBrakes     ProductCategory = "brakes"
	CategoryLighting   ProductCategory = "lighting"
	CategorySuspension ProductCategory = "suspension"
	CategoryElectrical ProductCategory = "electrical"
	CategoryGeneric    ProductCategory = "generic"
)

// IdentifierType defines standard automotive part identifier schemes.
type IdentifierType string

const (
	IDOEMPartNumber       IdentifierType = "oem_part_number"
	IDSupplierSKU         IdentifierType = "supplier_sku"
	IDUPC                 IdentifierType = "upc"
	IDCrossReference      IdentifierType = "cross_reference"
	IDSupersededPartNum IdentifierType = "superseded_part_number"
)

// PublicationState controls whether a product is searchable or quarantined.
type PublicationState string

const (
	PublicationDraft                 PublicationState = "draft"
	PublicationSearchableWithWarning PublicationState = "searchable_with_warning"
	PublicationPublished             PublicationState = "published"
	PublicationQuarantined           PublicationState = "quarantined"
	PublicationRejected              PublicationState = "rejected"
)

// TypedValue represents a strongly typed product attribute.
type TypedValue struct {
	Type        string   `json:"type"` // "string", "number", "bool", "string_list"
	StringVal   string   `json:"string_val,omitempty"`
	NumberVal   float64  `json:"number_val,omitempty"`
	BoolVal     bool     `json:"bool_val,omitempty"`
	ListVal     []string `json:"list_val,omitempty"`
	UnitOfMeasure string `json:"uom,omitempty"`
}

// ProductIdentifier maps external source IDs to canonical normalized formats.
type ProductIdentifier struct {
	Type       IdentifierType `json:"type"`
	Value      string         `json:"value"`
	Normalized string         `json:"normalized"`
	SourceID   string         `json:"source_id"`
}

// SourceReference tracks provenance across external feeds.
type SourceReference struct {
	SourceID       string `json:"source_id"`
	SourceType     string `json:"source_type"` // e.g., "dealer_db", "pdf_catalog", "supplier_feed"
	SourceRecordID string `json:"source_record_id"`
	SourceURI      string `json:"source_uri"`
	Version        string `json:"version"`
}

// Product is the DDD Aggregate Root for canonical automotive parts.
// Every product includes an AccessPolicy ensuring search and discovery queries filter restricted commercial terms
// or confidential supplier attributes before returning results to clients or LLMs.
type Product struct {
	ID             string                `json:"id"`
	ManufacturerID string                `json:"manufacturer_id"`
	CanonicalName  string                `json:"canonical_name"`
	Description    string                `json:"description"`
	Category       ProductCategory       `json:"category"`
	Brand          string                `json:"brand"`
	OEM            bool                  `json:"oem"`
	Attributes     map[string]TypedValue `json:"attributes"`
	Identifiers    []ProductIdentifier   `json:"identifiers"`
	AssetIDs       []string              `json:"asset_ids"`
	SourceRefs     []SourceReference     `json:"source_refs"`
	Publication    PublicationState      `json:"publication"`
	AccessPolicy   auth.AccessPolicy     `json:"access_policy"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidProduct  = errors.New("invalid product data")
)

// Repository is the DDD consumer port for persistence operations on Products.
type Repository interface {
	GetByID(ctx context.Context, id string) (*Product, error)
	GetByIdentifier(ctx context.Context, normalizedVal string) (*Product, error)
	Save(ctx context.Context, p *Product) error
	List(ctx context.Context, limit, offset int) ([]*Product, error)
}
