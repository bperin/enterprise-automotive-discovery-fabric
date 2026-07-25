package dealer

import (
	"context"
	"sync"
	"time"

	"enterprise-search/internal/auth"
)

type MemoryRepository struct {
	mu      sync.RWMutex
	dealers map[string]*Dealer
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		dealers: make(map[string]*Dealer),
	}
	repo.seedDefaultDealers()
	return repo
}

func (r *MemoryRepository) seedDefaultDealers() {
	now := time.Now()
	dealerID := "dealer-austin-01"
	r.dealers[dealerID] = &Dealer{
		ID:             dealerID,
		Name:           "Northstar Motors Austin",
		BrandIDs:       []string{"ApexMotors", "NovaMotors"},
		ServiceEnabled: true,
		PartsEnabled:   true,
		Locations: []DealerLocation{
			{
				ID:       "loc-austin-central",
				DealerID: dealerID,
				Name:     "Central Austin Dealership",
				Address: Address{
					Street:     "100 Congress Ave",
					City:       "Austin",
					State:      "TX",
					PostalCode: "78701",
					Country:    "USA",
				},
				Latitude:  30.2672,
				Longitude: -97.7431,
				Departments: []Department{
					{ID: "dept-1", Type: DeptSales, PhoneNumber: "512-555-0100", Hours: "Mon-Sat 9am-8pm"},
					{ID: "dept-2", Type: DeptParts, PhoneNumber: "512-555-0102", Hours: "Mon-Fri 8am-6pm"},
				},
			},
		},
		AccessPolicy: auth.AccessPolicy{
			Roles: []string{"anonymous_customer", "dealer_employee"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (*Dealer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.dealers[id]
	if !ok {
		return nil, ErrDealerNotFound
	}
	return d, nil
}

func (r *MemoryRepository) FindByPostalCode(ctx context.Context, postalCode string, radiusMiles float64) ([]*DealerLocation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var locations []*DealerLocation
	for _, dealer := range r.dealers {
		for i := range dealer.Locations {
			loc := &dealer.Locations[i]
			if loc.Address.PostalCode == postalCode {
				locations = append(locations, loc)
			}
		}
	}
	return locations, nil
}

func (r *MemoryRepository) Save(ctx context.Context, dealer *Dealer) error {
	if dealer == nil || dealer.ID == "" {
		return ErrInvalidDealer
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dealers[dealer.ID] = dealer
	return nil
}

func (r *MemoryRepository) List(ctx context.Context, limit, offset int) ([]*Dealer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*Dealer
	for _, d := range r.dealers {
		list = append(list, d)
	}
	return list, nil
}
