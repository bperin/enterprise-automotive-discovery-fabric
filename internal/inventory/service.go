package inventory

import (
	"context"
	"fmt"

	"enterprise-search/internal/fitment"
	"github.com/google/uuid"
)

// Service manages inventory observations and location availability in DDD.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RecordObservation(ctx context.Context, obs *InventoryObservation) (*InventoryObservation, error) {
	if obs.ProductID == "" || obs.LocationID == "" {
		return nil, fmt.Errorf("%w: ProductID and LocationID are required", ErrInvalidInventory)
	}

	if obs.ID == "" {
		obs.ID = "inv-" + uuid.New().String()[:8]
	}

	if obs.Availability == "" {
		if obs.Quantity != nil && *obs.Quantity > 0 {
			obs.Availability = InStock
		} else {
			obs.Availability = OutOfStock
		}
	}

	if obs.Authority == "" {
		obs.Authority = fitment.AuthAuthoritative
	}

	if obs.VerificationStatus == "" {
		obs.VerificationStatus = fitment.VerifiedBySourceContract
	}

	if err := s.repo.SaveObservation(ctx, obs); err != nil {
		return nil, err
	}

	return obs, nil
}

func (s *Service) GetInventoryForProduct(ctx context.Context, productID string) ([]*InventoryObservation, error) {
	if productID == "" {
		return nil, ErrInvalidInventory
	}
	return s.repo.FindObservationsByProduct(ctx, productID)
}

func (s *Service) GetFreshestObservation(ctx context.Context, productID, locationID string) (*InventoryObservation, error) {
	obsList, err := s.repo.FindObservationsByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	var freshest *InventoryObservation
	for _, obs := range obsList {
		if obs.LocationID == locationID {
			if freshest == nil || obs.ObservedAt.After(freshest.ObservedAt) {
				freshest = obs
			}
		}
	}

	if freshest == nil {
		return nil, ErrInventoryNotFound
	}
	return freshest, nil
}

func (s *Service) SaveLocation(ctx context.Context, loc *Location) (*Location, error) {
	if loc.Name == "" || loc.PostalCode == "" {
		return nil, fmt.Errorf("%w: Name and PostalCode required", ErrInvalidInventory)
	}
	if loc.ID == "" {
		loc.ID = "loc-" + uuid.New().String()[:8]
	}
	if err := s.repo.SaveLocation(ctx, loc); err != nil {
		return nil, err
	}
	return loc, nil
}

func (s *Service) GetLocation(ctx context.Context, id string) (*Location, error) {
	return s.repo.GetLocation(ctx, id)
}
