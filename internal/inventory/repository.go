package inventory

import (
	"context"
	"sync"
	"time"
)

// MemoryRepository implements the inventory Repository port in memory.
// In production, this would be backed by PostgreSQL or BigQuery inventory tables.
type MemoryRepository struct {
	mu           sync.RWMutex
	observations map[string]*InventoryObservation
	locations    map[string]*Location
	byProduct    map[string][]string // product_id -> observation IDs
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		observations: make(map[string]*InventoryObservation),
		locations:    make(map[string]*Location),
		byProduct:    make(map[string][]string),
	}
}

func (r *MemoryRepository) GetObservation(ctx context.Context, id string) (*InventoryObservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.observations[id]
	if !ok {
		return nil, ErrInventoryNotFound
	}
	cp := *o
	return &cp, nil
}

func (r *MemoryRepository) FindObservationsByProduct(ctx context.Context, productID string) ([]*InventoryObservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byProduct[productID]
	var res []*InventoryObservation
	for _, id := range ids {
		if o, ok := r.observations[id]; ok {
			cp := *o
			res = append(res, &cp)
		}
	}
	return res, nil
}

func (r *MemoryRepository) SaveObservation(ctx context.Context, obs *InventoryObservation) error {
	if obs.ID == "" || obs.ProductID == "" || obs.LocationID == "" {
		return ErrInvalidInventory
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	if obs.ObservedAt.IsZero() {
		obs.ObservedAt = now
	}
	if obs.ReceivedAt.IsZero() {
		obs.ReceivedAt = now
	}

	cp := *obs
	r.observations[obs.ID] = &cp

	r.byProduct[obs.ProductID] = appendIfMissing(r.byProduct[obs.ProductID], obs.ID)
	return nil
}

func (r *MemoryRepository) GetLocation(ctx context.Context, id string) (*Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	l, ok := r.locations[id]
	if !ok {
		return nil, ErrInventoryNotFound
	}
	cp := *l
	return &cp, nil
}

func (r *MemoryRepository) SaveLocation(ctx context.Context, loc *Location) error {
	if loc.ID == "" {
		return ErrInvalidInventory
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *loc
	r.locations[loc.ID] = &cp
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
