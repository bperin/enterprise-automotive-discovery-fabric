// Package content implements the marketing content, CMS items, support articles,
// and documentation lifecycle management for multi-brand automotive discovery.
package content

import (
	"context"
	"errors"
	"time"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/product"
)

type ContentType string

const (
	TypeModelPage       ContentType = "model_page"
	TypeFeaturePage     ContentType = "feature_page"
	TypeCampaign        ContentType = "campaign"
	TypeOffer           ContentType = "offer"
	TypeFAQ             ContentType = "faq"
	TypeSupportArticle  ContentType = "support_article"
	TypeWarrantyDoc     ContentType = "warranty_document"
	TypeOwnersManual    ContentType = "owners_manual"
	TypeInstallationGuide ContentType = "installation_guide"
	TypeLegalDisclaimer ContentType = "legal_disclaimer"
)

type ContentStatus string

const (
	StatusDraft     ContentStatus = "draft"
	StatusPublished ContentStatus = "published"
	StatusExpired   ContentStatus = "expired"
	StatusArchived  ContentStatus = "archived"
)

// ContentItem is the DDD Aggregate Root representing a published content asset.
type ContentItem struct {
	ID            string                  `json:"id"`
	BrandID       string                  `json:"brand_id"`
	ContentType   ContentType             `json:"content_type"`
	Title         string                  `json:"title"`
	Body          string                  `json:"body"`
	Locale        string                  `json:"locale"`
	Audience      []string                `json:"audience,omitempty"`
	VehicleIDs    []string                `json:"vehicle_ids,omitempty"`
	ProductIDs    []string                `json:"product_ids,omitempty"`
	EffectiveFrom *time.Time              `json:"effective_from,omitempty"`
	EffectiveTo   *time.Time              `json:"effective_to,omitempty"`
	Status        ContentStatus           `json:"status"`
	SourceRef     product.SourceReference `json:"source_ref"`
	AssetIDs      []string                `json:"asset_ids,omitempty"`
	AccessPolicy  auth.AccessPolicy       `json:"access_policy"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

var (
	ErrContentNotFound = errors.New("content item not found")
	ErrInvalidContent  = errors.New("invalid content item")
)

// Repository is the DDD consumer port for content items.
type Repository interface {
	GetByID(ctx context.Context, id string) (*ContentItem, error)
	FindByBrand(ctx context.Context, brandID string, contentType ContentType) ([]*ContentItem, error)
	Save(ctx context.Context, item *ContentItem) error
}
