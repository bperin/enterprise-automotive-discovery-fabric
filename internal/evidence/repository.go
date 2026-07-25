package evidence

import (
	"context"
	"sync"
	"time"
)

// MemoryRepository is an in-memory implementation of the evidence Repository port.
type MemoryRepository struct {
	mu           sync.RWMutex
	assets       map[string]*Asset
	evidence     map[string]*Evidence
	claims       map[string]*Claim
	attestations map[string]*Attestation
	bySubject    map[string][]string // subject_id -> claim IDs
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		assets:       make(map[string]*Asset),
		evidence:     make(map[string]*Evidence),
		claims:       make(map[string]*Claim),
		attestations: make(map[string]*Attestation),
		bySubject:    make(map[string][]string),
	}
}

func (r *MemoryRepository) SaveAsset(ctx context.Context, asset *Asset) error {
	if asset.ID == "" {
		return ErrInvalidEvidence
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if asset.IngestedAt.IsZero() {
		asset.IngestedAt = time.Now().UTC()
	}

	cp := *asset
	r.assets[asset.ID] = &cp
	return nil
}

func (r *MemoryRepository) GetAsset(ctx context.Context, id string) (*Asset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.assets[id]
	if !ok {
		return nil, ErrAssetNotFound
	}
	cp := *a
	return &cp, nil
}

func (r *MemoryRepository) SaveEvidence(ctx context.Context, ev *Evidence) error {
	if ev.ID == "" || ev.AssetID == "" {
		return ErrInvalidEvidence
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = time.Now().UTC()
	}

	cp := *ev
	r.evidence[ev.ID] = &cp
	return nil
}

func (r *MemoryRepository) GetEvidence(ctx context.Context, id string) (*Evidence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ev, ok := r.evidence[id]
	if !ok {
		return nil, ErrEvidenceNotFound
	}
	cp := *ev
	return &cp, nil
}

func (r *MemoryRepository) SaveClaim(ctx context.Context, claim *Claim) error {
	if claim.ID == "" || claim.SubjectID == "" {
		return ErrInvalidEvidence
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if claim.CreatedAt.IsZero() {
		claim.CreatedAt = time.Now().UTC()
	}

	cp := *claim
	r.claims[claim.ID] = &cp
	r.bySubject[claim.SubjectID] = appendIfMissing(r.bySubject[claim.SubjectID], claim.ID)
	return nil
}

func (r *MemoryRepository) GetClaimsBySubject(ctx context.Context, subjectID string) ([]*Claim, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.bySubject[subjectID]
	var res []*Claim
	for _, id := range ids {
		if c, ok := r.claims[id]; ok {
			cp := *c
			res = append(res, &cp)
		}
	}
	return res, nil
}

func (r *MemoryRepository) SaveAttestation(ctx context.Context, att *Attestation) error {
	if att.ID == "" || att.ClaimID == "" {
		return ErrInvalidEvidence
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if att.CreatedAt.IsZero() {
		att.CreatedAt = time.Now().UTC()
	}

	cp := *att
	r.attestations[att.ClaimID] = &cp
	return nil
}

func (r *MemoryRepository) GetAttestationForClaim(ctx context.Context, claimID string) (*Attestation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	att, ok := r.attestations[claimID]
	if !ok {
		return nil, ErrEvidenceNotFound
	}
	cp := *att
	return &cp, nil
}

func appendIfMissing(slice []string, val string) []string {
	for _, item := range slice {
		if item == val {
			return slice
		}
	}
	return append(slice, val)
}
