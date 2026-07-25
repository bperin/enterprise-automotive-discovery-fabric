// Package vehicle implements the vehicle model, trim, and configuration domain,
// providing structured representation for vehicles, trims, pricing, and attribute specifications.
package vehicle

import (
	"context"
	"errors"
	"time"

	"enterprise-search/internal/product"
)

// VehicleModel represents a marketed vehicle model line.
type VehicleModel struct {
	ID          string `json:"id"`
	BrandID     string `json:"brand_id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// VehicleTrim represents a specific trim level for a vehicle model year.
type VehicleTrim struct {
	ID         string                        `json:"id"`
	ModelID    string                        `json:"model_id"`
	ModelYear  int                           `json:"model_year"`
	Name       string                        `json:"name"`
	MSRP       float64                       `json:"msrp"`
	Currency   string                        `json:"currency"`
	Attributes map[string]product.TypedValue `json:"attributes,omitempty"`
}

// VehicleConfiguration is the DDD Aggregate Root representing a fully configured vehicle.
type VehicleConfiguration struct {
	ID         string                        `json:"id"`
	TrimID     string                        `json:"trim_id"`
	Year       int                           `json:"year"`
	Make       string                        `json:"make"`
	Model      string                        `json:"model"`
	Trim       string                        `json:"trim"`
	Engine     string                        `json:"engine"`
	Drivetrain string                        `json:"drivetrain"`
	Cab        string                        `json:"cab,omitempty"`
	Bed        string                        `json:"bed,omitempty"`
	Options    []string                      `json:"options,omitempty"`
	Attributes map[string]product.TypedValue `json:"attributes,omitempty"`
	CreatedAt  time.Time                     `json:"created_at"`
	UpdatedAt  time.Time                     `json:"updated_at"`
}

var (
	ErrVehicleNotFound = errors.New("vehicle configuration not found")
	ErrInvalidVehicle  = errors.New("invalid vehicle configuration")
)

// Repository is the DDD consumer port for vehicle configurations and models.
type Repository interface {
	GetConfiguration(ctx context.Context, id string) (*VehicleConfiguration, error)
	FindConfigurations(ctx context.Context, make, model string, year int) ([]*VehicleConfiguration, error)
	SaveConfiguration(ctx context.Context, config *VehicleConfiguration) error
}
