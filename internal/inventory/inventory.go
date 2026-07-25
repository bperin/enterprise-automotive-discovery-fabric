package inventory

import (
	"context"
	"errors"
	"time"

	"enterprise-search/internal/fitment"
	"enterprise-search/internal/product"
)

// AvailabilityStatus classifies current inventory status in DDD.
type AvailabilityStatus string

const (
	InStock      AvailabilityStatus = "in_stock"
	OutOfStock   AvailabilityStatus = "out_of_stock"
	Backordered  AvailabilityStatus = "backordered"
	Discontinued AvailabilityStatus = "discontinued"
	StatusUnknown AvailabilityStatus = "unknown"
)

// Money represents currency amounts in DDD.
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"` // e.g. "USD"
}

// Location represents a physical dealer, distributor, or warehouse location.
type Location struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	PostalCode string  `json:"postal_code"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

// InventoryObservation is the DDD Aggregate Root representing a time-stamped stock observation.
// Because dealer feeds and supplier feeds may disagree on stock quantities, inventory is modeled as
// time-bound observations rather than mutable current balances.
// In production, observations stream via Cloud Pub/Sub CDC into Cloud SQL / BigQuery.
// In local development, observations are stored in memory or SQLite.
type InventoryObservation struct {
	ID                 string                    `json:"id"`
	ProductID          string                    `json:"product_id"`
	LocationID         string                    `json:"location_id"`
	SourceRef          product.SourceReference   `json:"source_ref"`
	Quantity           *int                      `json:"quantity,omitempty"`
	Availability       AvailabilityStatus        `json:"availability"`
	Price              *Money                    `json:"price,omitempty"`
	ObservedAt         time.Time                 `json:"observed_at"`
	ReceivedAt         time.Time                 `json:"received_at"`
	ExpiresAt          *time.Time                `json:"expires_at,omitempty"`
	Authority          fitment.AuthorityLevel    `json:"authority"`
	Confidence         float64                   `json:"confidence"`
	VerificationStatus fitment.VerificationStatus `json:"verification_status"`
}

var (
	ErrInventoryNotFound = errors.New("inventory observation not found")
	ErrInvalidInventory  = errors.New("invalid inventory observation")
)

// Repository is the DDD consumer port for inventory observation persistence.
type Repository interface {
	GetObservation(ctx context.Context, id string) (*InventoryObservation, error)
	FindObservationsByProduct(ctx context.Context, productID string) ([]*InventoryObservation, error)
	SaveObservation(ctx context.Context, obs *InventoryObservation) error
	GetLocation(ctx context.Context, id string) (*Location, error)
	SaveLocation(ctx context.Context, loc *Location) error
}
