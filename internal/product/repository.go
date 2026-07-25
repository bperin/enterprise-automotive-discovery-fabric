package product

import (
	"context"
	"strings"
	"sync"
	"time"
)

// MemoryRepository is an in-memory implementation of the Product Repository port.
// In production, this would be backed by PostgreSQL / Cloud SQL via sqlc or pgx.
type MemoryRepository struct {
	mu          sync.RWMutex
	products    map[string]*Product
	identifiers map[string]string // normalized identifier -> product ID
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		products:    make(map[string]*Product),
		identifiers: make(map[string]string),
	}
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (*Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.products[id]
	if !ok {
		return nil, ErrProductNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *MemoryRepository) GetByIdentifier(ctx context.Context, normalizedVal string) (*Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	norm := strings.ToLower(strings.TrimSpace(normalizedVal))
	id, ok := r.identifiers[norm]
	if !ok {
		return nil, ErrProductNotFound
	}
	p, ok := r.products[id]
	if !ok {
		return nil, ErrProductNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *MemoryRepository) Save(ctx context.Context, p *Product) error {
	if p.ID == "" {
		return ErrInvalidProduct
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	cp := *p
	r.products[p.ID] = &cp

	for _, ident := range p.Identifiers {
		norm := strings.ToLower(strings.TrimSpace(ident.Normalized))
		if norm != "" {
			r.identifiers[norm] = p.ID
		}
	}
	return nil
}

func (r *MemoryRepository) List(ctx context.Context, limit, offset int) ([]*Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Product
	idx := 0
	for _, p := range r.products {
		if idx >= offset && (limit <= 0 || len(result) < limit) {
			cp := *p
			result = append(result, &cp)
		}
		idx++
	}
	return result, nil
}
