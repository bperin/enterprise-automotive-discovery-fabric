package vehicle

import (
	"context"
	"sync"
)

type MemoryRepository struct {
	mu            sync.RWMutex
	configs map[string]*VehicleConfiguration
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		configs: make(map[string]*VehicleConfiguration),
	}
}

func (r *MemoryRepository) GetConfiguration(ctx context.Context, id string) (*VehicleConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.configs[id]
	if !ok {
		return nil, ErrVehicleNotFound
	}
	return v, nil
}

func (r *MemoryRepository) FindConfigurations(ctx context.Context, make, model string, year int) ([]*VehicleConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*VehicleConfiguration
	for _, v := range r.configs {
		if (make == "" || v.Make == make) && (model == "" || v.Model == model) && (year == 0 || v.Year == year) {
			result = append(result, v)
		}
	}
	return result, nil
}

func (r *MemoryRepository) SaveConfiguration(ctx context.Context, config *VehicleConfiguration) error {
	if config == nil || config.ID == "" {
		return ErrInvalidVehicle
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs[config.ID] = config
	return nil
}
