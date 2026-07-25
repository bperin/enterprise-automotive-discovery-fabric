// Package dealer implements the automotive dealer network and location domain model,
// tracking authorized brand sales, service departments, and dealer locations.
package dealer

import (
	"context"
	"errors"
	"time"

	"enterprise-search/internal/auth"
)

// DepartmentType classifies dealership functional departments.
type DepartmentType string

const (
	DeptSales     DepartmentType = "sales"
	DeptService   DepartmentType = "service"
	DeptParts     DepartmentType = "parts"
	DeptCollision DepartmentType = "collision"
	DeptFleet     DepartmentType = "fleet"
)

// Department represents an operational department within a dealer location.
type Department struct {
	ID          string         `json:"id"`
	Type        DepartmentType `json:"type"`
	PhoneNumber string         `json:"phone_number,omitempty"`
	Hours       string         `json:"hours,omitempty"`
}

// Address represents a physical mailing and geographical address.
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// DealerLocation represents a specific dealership facility within the dealer network.
type DealerLocation struct {
	ID          string       `json:"id"`
	DealerID    string       `json:"dealer_id"`
	Name        string       `json:"name"`
	Address     Address      `json:"address"`
	Latitude    float64      `json:"latitude"`
	Longitude   float64      `json:"longitude"`
	Departments []Department `json:"departments"`
}

// Dealer is the DDD Aggregate Root representing an automotive dealer organization.
type Dealer struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	BrandIDs       []string          `json:"brand_ids"`
	Locations      []DealerLocation  `json:"locations"`
	ServiceEnabled bool              `json:"service_enabled"`
	PartsEnabled   bool              `json:"parts_enabled"`
	AccessPolicy   auth.AccessPolicy `json:"access_policy"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

var (
	ErrDealerNotFound = errors.New("dealer organization not found")
	ErrInvalidDealer  = errors.New("invalid dealer record")
)

// Repository is the DDD consumer port for dealer network persistence.
type Repository interface {
	GetByID(ctx context.Context, id string) (*Dealer, error)
	FindByPostalCode(ctx context.Context, postalCode string, radiusMiles float64) ([]*DealerLocation, error)
	Save(ctx context.Context, dealer *Dealer) error
	List(ctx context.Context, limit, offset int) ([]*Dealer, error)
}
