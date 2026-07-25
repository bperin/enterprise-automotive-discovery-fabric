package fitment

import (
	"context"
	"errors"
	"time"

	"enterprise-search/internal/product"
)

// CompatibilityStatus defines the compatibility determination in DDD.
type CompatibilityStatus string

const (
	DirectFit      CompatibilityStatus = "direct_fit"
	ConditionalFit CompatibilityStatus = "conditional_fit"
	NotCompatible  CompatibilityStatus = "not_compatible"
	UnknownFitment CompatibilityStatus = "unknown"
)

// AuthorityLevel defines source authority hierarchy in DDD.
type AuthorityLevel string

const (
	AuthAuthoritative   AuthorityLevel = "authoritative"
	AuthLicensedCurated AuthorityLevel = "licensed_curated"
	AuthVerifiedDerived AuthorityLevel = "verified_derived"
	AuthDerived         AuthorityLevel = "derived"
	AuthUserSupplied    AuthorityLevel = "user_supplied"
	AuthUnknown         AuthorityLevel = "unknown"
)

// VerificationStatus tracks the audit status of a fitment assertion.
type VerificationStatus string

const (
	VerifiedBySourceContract   VerificationStatus = "verified_by_source_contract"
	VerifiedByRule             VerificationStatus = "verified_by_rule"
	VerifiedByCrossSourceMatch VerificationStatus = "verified_by_cross_source_match"
	VerifiedByHuman            VerificationStatus = "verified_by_human"
	ModelExtracted             VerificationStatus = "model_extracted"
	VerificationConflicting    VerificationStatus = "conflicting"
	VerificationStale          VerificationStatus = "stale"
	VerificationRejected       VerificationStatus = "rejected"
)

// FitmentCondition represents explicit requirements for conditional fitment.
type FitmentCondition struct {
	Name  string `json:"name"`  // e.g. "lift_kit_required"
	Value string `json:"value"` // e.g. "2_inch_minimum"
}

// VehicleConfiguration represents a specific vehicle trim/year/make/model in DDD.
type VehicleConfiguration struct {
	ID         string                        `json:"id"`
	Year       int                           `json:"year"`
	Make       string                        `json:"make"`
	Model      string                        `json:"model"`
	Trim       string                        `json:"trim,omitempty"`
	Engine     string                        `json:"engine,omitempty"`
	Drivetrain string                        `json:"drivetrain,omitempty"`
	BodyStyle  string                        `json:"body_style,omitempty"`
	Attributes map[string]product.TypedValue `json:"attributes,omitempty"`
}

// FitmentAssertion is the DDD Aggregate Root for fitment compatibility claims.
// In production, fitments are evaluated against ACES/PIES catalog standards in Cloud SQL / BigQuery.
// In local development, fitments are stored in memory or SQLite.
type FitmentAssertion struct {
	ID                     string                  `json:"id"`
	ProductID              string                  `json:"product_id"`
	VehicleConfigurationID string                  `json:"vehicle_configuration_id"`
	Compatibility          CompatibilityStatus     `json:"compatibility"`
	Conditions             []FitmentCondition      `json:"conditions,omitempty"`
	SourceRef              product.SourceReference `json:"source_ref"`
	Authority              AuthorityLevel          `json:"authority"`
	Confidence             float64                 `json:"confidence"`
	EvidenceIDs            []string                `json:"evidence_ids,omitempty"`
	ObservedAt             time.Time               `json:"observed_at"`
	VerificationStatus     VerificationStatus      `json:"verification_status"`
}

var (
	ErrFitmentNotFound = errors.New("fitment assertion not found")
	ErrInvalidFitment  = errors.New("invalid fitment assertion")
)

// Repository is the DDD consumer port for fitment assertions and vehicle configs.
type Repository interface {
	GetAssertion(ctx context.Context, id string) (*FitmentAssertion, error)
	FindFitmentsForVehicle(ctx context.Context, vehicleID string) ([]*FitmentAssertion, error)
	FindFitmentsForProduct(ctx context.Context, productID string) ([]*FitmentAssertion, error)
	SaveAssertion(ctx context.Context, assertion *FitmentAssertion) error
	GetVehicleConfig(ctx context.Context, id string) (*VehicleConfiguration, error)
	SaveVehicleConfig(ctx context.Context, config *VehicleConfiguration) error
}
