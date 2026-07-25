package evidence

import (
	"context"
	"fmt"

	"enterprise-search/internal/fitment"
	"github.com/google/uuid"
)

// Service provides asset registration, claim extraction tracking, and attestation workflow use cases in DDD.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterAsset(ctx context.Context, asset *Asset) (*Asset, error) {
	if asset.URI == "" {
		return nil, fmt.Errorf("%w: URI required", ErrInvalidEvidence)
	}

	if asset.ID == "" {
		asset.ID = "asset-" + uuid.New().String()[:8]
	}

	if err := s.repo.SaveAsset(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *Service) RecordEvidence(ctx context.Context, ev *Evidence) (*Evidence, error) {
	if ev.AssetID == "" {
		return nil, fmt.Errorf("%w: AssetID required", ErrInvalidEvidence)
	}

	if ev.ID == "" {
		ev.ID = "ev-" + uuid.New().String()[:8]
	}

	if err := s.repo.SaveEvidence(ctx, ev); err != nil {
		return nil, err
	}
	return ev, nil
}

func (s *Service) ProposeClaim(ctx context.Context, claim *Claim) (*Claim, error) {
	if claim.SubjectID == "" || claim.FieldPath == "" {
		return nil, fmt.Errorf("%w: SubjectID and FieldPath required", ErrInvalidEvidence)
	}

	if claim.ID == "" {
		claim.ID = "claim-" + uuid.New().String()[:8]
	}

	if claim.Authority == "" {
		claim.Authority = fitment.AuthDerived
	}

	if err := s.repo.SaveClaim(ctx, claim); err != nil {
		return nil, err
	}
	return claim, nil
}

func (s *Service) AttestClaim(ctx context.Context, att *Attestation) (*Attestation, error) {
	if att.ClaimID == "" {
		return nil, fmt.Errorf("%w: ClaimID required", ErrInvalidEvidence)
	}

	if att.ID == "" {
		att.ID = "att-" + uuid.New().String()[:8]
	}

	if err := s.repo.SaveAttestation(ctx, att); err != nil {
		return nil, err
	}
	return att, nil
}

func (s *Service) GetClaimsForSubject(ctx context.Context, subjectID string) ([]*Claim, error) {
	return s.repo.GetClaimsBySubject(ctx, subjectID)
}

func (s *Service) GetEvidence(ctx context.Context, id string) (*Evidence, error) {
	return s.repo.GetEvidence(ctx, id)
}

func (s *Service) GetAsset(ctx context.Context, id string) (*Asset, error) {
	return s.repo.GetAsset(ctx, id)
}
