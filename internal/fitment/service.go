package fitment

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service provides fitment verification and assertion management use cases in DDD.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) VerifyFitment(ctx context.Context, productID, vehicleID string) (*FitmentAssertion, error) {
	if productID == "" || vehicleID == "" {
		return nil, ErrInvalidFitment
	}

	assertions, err := s.repo.FindFitmentsForProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	for _, a := range assertions {
		if a.VehicleConfigurationID == vehicleID {
			return a, nil
		}
	}

	return &FitmentAssertion{
		ProductID:              productID,
		VehicleConfigurationID: vehicleID,
		Compatibility:          UnknownFitment,
		Authority:              AuthUnknown,
		Confidence:             0.0,
		VerificationStatus:     VerificationStale,
	}, nil
}

func (s *Service) CreateAssertion(ctx context.Context, a *FitmentAssertion) (*FitmentAssertion, error) {
	if a.ProductID == "" || a.VehicleConfigurationID == "" {
		return nil, fmt.Errorf("%w: ProductID and VehicleConfigurationID are required", ErrInvalidFitment)
	}

	if a.ID == "" {
		a.ID = "fit-" + uuid.New().String()[:8]
	}

	if a.Compatibility == "" {
		a.Compatibility = UnknownFitment
	}

	if a.Authority == "" {
		a.Authority = AuthDerived
	}

	if err := s.repo.SaveAssertion(ctx, a); err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Service) GetFitmentsByProduct(ctx context.Context, productID string) ([]*FitmentAssertion, error) {
	return s.repo.FindFitmentsForProduct(ctx, productID)
}

func (s *Service) GetFitmentsByVehicle(ctx context.Context, vehicleID string) ([]*FitmentAssertion, error) {
	return s.repo.FindFitmentsForVehicle(ctx, vehicleID)
}

func (s *Service) SaveVehicleConfig(ctx context.Context, v *VehicleConfiguration) (*VehicleConfiguration, error) {
	if v.Make == "" || v.Model == "" || v.Year <= 0 {
		return nil, fmt.Errorf("%w: Make, Model, Year required", ErrInvalidFitment)
	}
	if v.ID == "" {
		v.ID = fmt.Sprintf("veh-%d-%s-%s", v.Year, v.Make, v.Model)
	}
	if err := s.repo.SaveVehicleConfig(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}
