package fitment

import (
	"context"
	"sync"
	"time"
)

// MemoryRepository implements the fitment Repository port in memory.
// In production, this would be backed by PostgreSQL or BigQuery fitment queries.
type MemoryRepository struct {
	mu            sync.RWMutex
	assertions    map[string]*FitmentAssertion
	vehicles      map[string]*VehicleConfiguration
	byProduct     map[string][]string // product_id -> assertion IDs
	byVehicle     map[string][]string // vehicle_id -> assertion IDs
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		assertions: make(map[string]*FitmentAssertion),
		vehicles:   make(map[string]*VehicleConfiguration),
		byProduct:  make(map[string][]string),
		byVehicle:  make(map[string][]string),
	}
}

func (r *MemoryRepository) GetAssertion(ctx context.Context, id string) (*FitmentAssertion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.assertions[id]
	if !ok {
		return nil, ErrFitmentNotFound
	}
	cp := *a
	return &cp, nil
}

func (r *MemoryRepository) FindFitmentsForVehicle(ctx context.Context, vehicleID string) ([]*FitmentAssertion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byVehicle[vehicleID]
	var res []*FitmentAssertion
	for _, id := range ids {
		if a, ok := r.assertions[id]; ok {
			cp := *a
			res = append(res, &cp)
		}
	}
	return res, nil
}

func (r *MemoryRepository) FindFitmentsForProduct(ctx context.Context, productID string) ([]*FitmentAssertion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byProduct[productID]
	var res []*FitmentAssertion
	for _, id := range ids {
		if a, ok := r.assertions[id]; ok {
			cp := *a
			res = append(res, &cp)
		}
	}
	return res, nil
}

func (r *MemoryRepository) SaveAssertion(ctx context.Context, assertion *FitmentAssertion) error {
	if assertion.ID == "" || assertion.ProductID == "" || assertion.VehicleConfigurationID == "" {
		return ErrInvalidFitment
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if assertion.ObservedAt.IsZero() {
		assertion.ObservedAt = time.Now().UTC()
	}

	cp := *assertion
	r.assertions[assertion.ID] = &cp

	r.byProduct[assertion.ProductID] = appendIfMissing(r.byProduct[assertion.ProductID], assertion.ID)
	r.byVehicle[assertion.VehicleConfigurationID] = appendIfMissing(r.byVehicle[assertion.VehicleConfigurationID], assertion.ID)
	return nil
}

func (r *MemoryRepository) GetVehicleConfig(ctx context.Context, id string) (*VehicleConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.vehicles[id]
	if !ok {
		return nil, ErrFitmentNotFound
	}
	cp := *v
	return &cp, nil
}

func (r *MemoryRepository) SaveVehicleConfig(ctx context.Context, config *VehicleConfiguration) error {
	if config.ID == "" {
		return ErrInvalidFitment
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *config
	r.vehicles[config.ID] = &cp
	return nil
}

func appendIfMissing(slice []string, val string) []string {
	for _, item := range slice {
		if item == val {
			return slice
		}
	}
	return append(slice, val)
}
