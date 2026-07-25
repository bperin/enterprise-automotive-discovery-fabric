package product

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Service implements use cases for product ingestion, lookup, and publication management in DDD.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetProduct(ctx context.Context, id string) (*Product, error) {
	if id == "" {
		return nil, ErrInvalidProduct
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) LookupIdentifier(ctx context.Context, rawIdentifier string) (*Product, error) {
	norm := normalizeIdentifier(rawIdentifier)
	if norm == "" {
		return nil, ErrInvalidProduct
	}
	return s.repo.GetByIdentifier(ctx, norm)
}

func (s *Service) CreateOrUpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	if p.CanonicalName == "" {
		return nil, fmt.Errorf("%w: canonical_name is required", ErrInvalidProduct)
	}

	if p.ID == "" {
		p.ID = "prod-" + uuid.New().String()[:8]
	}

	if p.Publication == "" {
		p.Publication = PublicationDraft
	}

	// Normalize identifiers
	for i := range p.Identifiers {
		if p.Identifiers[i].Normalized == "" {
			p.Identifiers[i].Normalized = normalizeIdentifier(p.Identifiers[i].Value)
		}
	}

	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ListProducts(ctx context.Context, limit, offset int) ([]*Product, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}

func normalizeIdentifier(val string) string {
	cleaned := strings.ReplaceAll(val, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	return strings.ToLower(strings.TrimSpace(cleaned))
}
