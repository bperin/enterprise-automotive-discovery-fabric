package evidence

import (
	"context"
	"errors"
	"time"

	"enterprise-search/internal/fitment"
	"enterprise-search/internal/product"
)

// BoundingBox represents a normalized bounding region on a page [0.0 - 1.0].
type BoundingBox struct {
	XMin float64 `json:"xmin"`
	YMin float64 `json:"ymin"`
	XMax float64 `json:"xmax"`
	YMax float64 `json:"ymax"`
}

// ExtractorIdentity identifies the model and prompt version that produced evidence.
type ExtractorIdentity struct {
	Name    string `json:"name"`    // e.g. "product-identity-extractor"
	Version string `json:"version"` // e.g. "v1.2.0"
}

// Asset represents an unparsed PDF, image, or document file in DDD.
type Asset struct {
	ID         string                  `json:"id"`
	URI        string                  `json:"uri"`
	MediaType  string                  `json:"media_type"` // e.g. "application/pdf", "image/png"
	SHA256     string                  `json:"sha256"`
	PageCount  *int                    `json:"page_count,omitempty"`
	Width      *int                    `json:"width,omitempty"`
	Height     *int                    `json:"height,omitempty"`
	SourceRef  product.SourceReference `json:"source_ref"`
	IngestedAt time.Time               `json:"ingested_at"`
}

// Evidence represents a specific page/region text or snippet extracted from an asset in DDD.
type Evidence struct {
	ID          string            `json:"id"`
	AssetID     string            `json:"asset_id"`
	Page        *int              `json:"page,omitempty"`
	Region      *BoundingBox      `json:"region,omitempty"`
	Text        string            `json:"text,omitempty"`
	ImageURI    string            `json:"image_uri,omitempty"`
	ContentHash string            `json:"content_hash"`
	Extractor   ExtractorIdentity `json:"extractor"`
	CreatedAt   time.Time         `json:"created_at"`
}

// Claim is the DDD Aggregate Root for probabilistic assertions extracted by AI models.
// Extractors return Claims, which are validated before publication to canonical product entities.
type Claim struct {
	ID              string                 `json:"id"`
	SubjectType     string                 `json:"subject_type"` // e.g., "candidate_product", "fitment"
	SubjectID       string                 `json:"subject_id"`
	FieldPath       string                 `json:"field_path"` // e.g., "identifiers.oem_part_number"
	Value           product.TypedValue     `json:"value"`
	Confidence      float64                `json:"confidence"`
	Authority       fitment.AuthorityLevel `json:"authority"`
	SourceRef       product.SourceReference `json:"source_ref"`
	EvidenceIDs     []string               `json:"evidence_ids"`
	ExtractionRunID string                 `json:"extraction_run_id"`
	CreatedAt       time.Time              `json:"created_at"`
}

// AttestationStatus defines the publication gate for candidate claims.
type AttestationStatus string

const (
	AttestationApproved    AttestationStatus = "approved"
	AttestationQuarantined AttestationStatus = "quarantined"
	AttestationRejected   AttestationStatus = "rejected"
)

// ValidatorIdentity records whether a claim was verified deterministically or by a human operator.
type ValidatorIdentity struct {
	Type string `json:"type"` // "rule_engine", "human_reviewer", "cross_source_match"
	ID   string `json:"id"`   // user ID or rule name
}

// Attestation records the formal validation decision for a claim in DDD.
type Attestation struct {
	ID          string            `json:"id"`
	ClaimID     string            `json:"claim_id"`
	Status      AttestationStatus `json:"status"`
	ReasonCodes []string          `json:"reason_codes,omitempty"`
	Notes       string            `json:"notes,omitempty"`
	EvidenceIDs []string          `json:"evidence_ids,omitempty"`
	Validator   ValidatorIdentity `json:"validator"`
	CreatedAt   time.Time         `json:"created_at"`
}

var (
	ErrAssetNotFound    = errors.New("asset not found")
	ErrEvidenceNotFound = errors.New("evidence not found")
	ErrClaimNotFound    = errors.New("claim not found")
	ErrInvalidEvidence  = errors.New("invalid evidence data")
)

// Repository is the DDD consumer port for assets, evidence, claims, and attestations.
type Repository interface {
	SaveAsset(ctx context.Context, asset *Asset) error
	GetAsset(ctx context.Context, id string) (*Asset, error)
	SaveEvidence(ctx context.Context, ev *Evidence) error
	GetEvidence(ctx context.Context, id string) (*Evidence, error)
	SaveClaim(ctx context.Context, claim *Claim) error
	GetClaimsBySubject(ctx context.Context, subjectID string) ([]*Claim, error)
	SaveAttestation(ctx context.Context, att *Attestation) error
	GetAttestationForClaim(ctx context.Context, claimID string) (*Attestation, error)
}
