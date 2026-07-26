package content

import (
	"context"
	"sync"
)

type MemoryRepository struct {
	mu    sync.RWMutex
	items map[string]*ContentItem
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		items: make(map[string]*ContentItem),
	}
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (*ContentItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok {
		return nil, ErrContentNotFound
	}
	return item, nil
}

func (r *MemoryRepository) FindByBrand(ctx context.Context, brandID string, contentType ContentType) ([]*ContentItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ContentItem
	for _, item := range r.items {
		if item.BrandID == brandID && (contentType == "" || item.ContentType == contentType) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *MemoryRepository) Save(ctx context.Context, item *ContentItem) error {
	if item == nil || item.ID == "" {
		return ErrInvalidContent
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return nil
}
