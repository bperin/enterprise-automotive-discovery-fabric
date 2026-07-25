package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"enterprise-search/internal/evidence"
	"enterprise-search/internal/fitment"
	"enterprise-search/internal/infra/postgres/sqlc"
	"enterprise-search/internal/inventory"
	"enterprise-search/internal/product"
	"github.com/jackc/pgx/v5/pgtype"
)

// SQLCStore wraps generated sqlc.Queries to provide type-safe PostgreSQL repositories in DDD.
type SQLCStore struct {
	q sqlc.Querier
}

func NewSQLCStore(q sqlc.Querier) *SQLCStore {
	return &SQLCStore{q: q}
}

// Ensure compile-time interface implementation assertions
var _ product.Repository = (*SQLCStore)(nil)
var _ fitment.Repository = (*SQLCStore)(nil)
var _ inventory.Repository = (*SQLCStore)(nil)
var _ evidence.Repository = (*SQLCStore)(nil)

// Helper functions for pgx/v5 pgtype conversions

func toPgTimestamptz(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func fromPgTimestamptz(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

func toPgInt4(val *int) pgtype.Int4 {
	if val == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*val), Valid: true}
}

func fromPgInt4(i pgtype.Int4) *int {
	if !i.Valid {
		return nil
	}
	v := int(i.Int32)
	return &v
}

func toPgFloat8(val *float64) pgtype.Float8 {
	if val == nil {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: *val, Valid: true}
}

// --- Product Repository Implementation ---

func (s *SQLCStore) GetByID(ctx context.Context, id string) (*product.Product, error) {
	row, err := s.q.GetProductByID(ctx, id)
	if err != nil {
		return nil, product.ErrProductNotFound
	}
	return mapSQLCProductToDomain(row), nil
}

func (s *SQLCStore) GetByIdentifier(ctx context.Context, normalizedVal string) (*product.Product, error) {
	row, err := s.q.GetProductByIdentifier(ctx, normalizedVal)
	if err != nil {
		return nil, product.ErrProductNotFound
	}
	return mapSQLCProductToDomain(row), nil
}

func (s *SQLCStore) Save(ctx context.Context, p *product.Product) error {
	attrBytes, _ := json.Marshal(p.Attributes)

	arg := sqlc.UpsertProductParams{
		ID:             p.ID,
		ManufacturerID: p.ManufacturerID,
		CanonicalName:  p.CanonicalName,
		Description:    p.Description,
		Category:       string(p.Category),
		Brand:          p.Brand,
		Oem:            p.OEM,
		Attributes:     attrBytes,
		Publication:    string(p.Publication),
	}

	_, err := s.q.UpsertProduct(ctx, arg)
	if err != nil {
		return fmt.Errorf("upsert product: %w", err)
	}

	for _, ident := range p.Identifiers {
		_ = s.q.InsertProductIdentifier(ctx, sqlc.InsertProductIdentifierParams{
			ProductID:  p.ID,
			Type:       string(ident.Type),
			Value:      ident.Value,
			Normalized: ident.Normalized,
			SourceID:   ident.SourceID,
		})
	}
	return nil
}

func (s *SQLCStore) List(ctx context.Context, limit, offset int) ([]*product.Product, error) {
	rows, err := s.q.ListProducts(ctx, sqlc.ListProductsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var res []*product.Product
	for _, row := range rows {
		res = append(res, mapSQLCProductToDomain(row))
	}
	return res, nil
}

func mapSQLCProductToDomain(row sqlc.Product) *product.Product {
	var attrs map[string]product.TypedValue
	if len(row.Attributes) > 0 {
		_ = json.Unmarshal(row.Attributes, &attrs)
	}

	return &product.Product{
		ID:             row.ID,
		ManufacturerID: row.ManufacturerID,
		CanonicalName:  row.CanonicalName,
		Description:    row.Description,
		Category:       product.ProductCategory(row.Category),
		Brand:          row.Brand,
		OEM:            row.Oem,
		Attributes:     attrs,
		Publication:    product.PublicationState(row.Publication),
		CreatedAt:      fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:      fromPgTimestamptz(row.UpdatedAt),
	}
}

// --- Fitment Repository Implementation ---

func (s *SQLCStore) GetAssertion(ctx context.Context, id string) (*fitment.FitmentAssertion, error) {
	row, err := s.q.GetFitmentAssertion(ctx, id)
	if err != nil {
		return nil, fitment.ErrFitmentNotFound
	}
	return mapSQLCFitmentToDomain(row), nil
}

func (s *SQLCStore) FindFitmentsForVehicle(ctx context.Context, vehicleID string) ([]*fitment.FitmentAssertion, error) {
	rows, err := s.q.ListFitmentsByVehicle(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	var res []*fitment.FitmentAssertion
	for _, row := range rows {
		res = append(res, mapSQLCFitmentToDomain(row))
	}
	return res, nil
}

func (s *SQLCStore) FindFitmentsForProduct(ctx context.Context, productID string) ([]*fitment.FitmentAssertion, error) {
	rows, err := s.q.ListFitmentsByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	var res []*fitment.FitmentAssertion
	for _, row := range rows {
		res = append(res, mapSQLCFitmentToDomain(row))
	}
	return res, nil
}

func (s *SQLCStore) SaveAssertion(ctx context.Context, a *fitment.FitmentAssertion) error {
	condBytes, _ := json.Marshal(a.Conditions)
	evBytes, _ := json.Marshal(a.EvidenceIDs)

	_, err := s.q.UpsertFitmentAssertion(ctx, sqlc.UpsertFitmentAssertionParams{
		ID:                     a.ID,
		ProductID:              a.ProductID,
		VehicleConfigurationID: a.VehicleConfigurationID,
		Compatibility:          string(a.Compatibility),
		Conditions:             condBytes,
		SourceID:               a.SourceRef.SourceID,
		SourceType:             a.SourceRef.SourceType,
		SourceRecordID:         a.SourceRef.SourceRecordID,
		Authority:              string(a.Authority),
		Confidence:             a.Confidence,
		EvidenceIds:            evBytes,
		ObservedAt:             toPgTimestamptz(a.ObservedAt),
		VerificationStatus:     string(a.VerificationStatus),
	})
	return err
}

func (s *SQLCStore) GetVehicleConfig(ctx context.Context, id string) (*fitment.VehicleConfiguration, error) {
	row, err := s.q.GetVehicleConfig(ctx, id)
	if err != nil {
		return nil, fitment.ErrFitmentNotFound
	}
	var attrs map[string]product.TypedValue
	if len(row.Attributes) > 0 {
		_ = json.Unmarshal(row.Attributes, &attrs)
	}
	return &fitment.VehicleConfiguration{
		ID:         row.ID,
		Year:       int(row.Year),
		Make:       row.Make,
		Model:      row.Model,
		Trim:       row.Trim,
		Engine:     row.Engine,
		Drivetrain: row.Drivetrain,
		BodyStyle:  row.BodyStyle,
		Attributes: attrs,
	}, nil
}

func (s *SQLCStore) SaveVehicleConfig(ctx context.Context, cfg *fitment.VehicleConfiguration) error {
	attrBytes, _ := json.Marshal(cfg.Attributes)
	_, err := s.q.UpsertVehicleConfig(ctx, sqlc.UpsertVehicleConfigParams{
		ID:         cfg.ID,
		Year:       int32(cfg.Year),
		Make:       cfg.Make,
		Model:      cfg.Model,
		Trim:       cfg.Trim,
		Engine:     cfg.Engine,
		Drivetrain: cfg.Drivetrain,
		BodyStyle:  cfg.BodyStyle,
		Attributes: attrBytes,
	})
	return err
}

func mapSQLCFitmentToDomain(row sqlc.FitmentAssertion) *fitment.FitmentAssertion {
	var conds []fitment.FitmentCondition
	if len(row.Conditions) > 0 {
		_ = json.Unmarshal(row.Conditions, &conds)
	}
	var evIDs []string
	if len(row.EvidenceIds) > 0 {
		_ = json.Unmarshal(row.EvidenceIds, &evIDs)
	}

	return &fitment.FitmentAssertion{
		ID:                     row.ID,
		ProductID:              row.ProductID,
		VehicleConfigurationID: row.VehicleConfigurationID,
		Compatibility:          fitment.CompatibilityStatus(row.Compatibility),
		Conditions:             conds,
		SourceRef: product.SourceReference{
			SourceID:       row.SourceID,
			SourceType:     row.SourceType,
			SourceRecordID: row.SourceRecordID,
		},
		Authority:          fitment.AuthorityLevel(row.Authority),
		Confidence:         row.Confidence,
		EvidenceIDs:        evIDs,
		ObservedAt:         fromPgTimestamptz(row.ObservedAt),
		VerificationStatus: fitment.VerificationStatus(row.VerificationStatus),
	}
}

// --- Inventory Repository Implementation ---

func (s *SQLCStore) GetObservation(ctx context.Context, id string) (*inventory.InventoryObservation, error) {
	row, err := s.q.GetInventoryObservation(ctx, id)
	if err != nil {
		return nil, inventory.ErrInventoryNotFound
	}

	qty := fromPgInt4(row.Quantity)
	var price *inventory.Money
	if row.PriceAmount.Valid {
		price = &inventory.Money{
			Amount:   row.PriceAmount.Float64,
			Currency: row.PriceCurrency,
		}
	}

	return &inventory.InventoryObservation{
		ID:           row.ID,
		ProductID:    row.ProductID,
		LocationID:   row.LocationID,
		Quantity:     qty,
		Availability: inventory.AvailabilityStatus(row.Availability),
		Price:        price,
		ObservedAt:   fromPgTimestamptz(row.ObservedAt),
		ReceivedAt:   fromPgTimestamptz(row.ReceivedAt),
		Authority:    fitment.AuthorityLevel(row.Authority),
		Confidence:   row.Confidence,
		SourceRef: product.SourceReference{
			SourceID:       row.SourceID,
			SourceType:     row.SourceType,
			SourceRecordID: row.SourceRecordID,
		},
	}, nil
}

func (s *SQLCStore) FindObservationsByProduct(ctx context.Context, productID string) ([]*inventory.InventoryObservation, error) {
	rows, err := s.q.ListInventoryByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	var res []*inventory.InventoryObservation
	for _, row := range rows {
		qty := fromPgInt4(row.Quantity)
		var price *inventory.Money
		if row.PriceAmount.Valid {
			price = &inventory.Money{
				Amount:   row.PriceAmount.Float64,
				Currency: row.PriceCurrency,
			}
		}

		res = append(res, &inventory.InventoryObservation{
			ID:           row.ID,
			ProductID:    row.ProductID,
			LocationID:   row.LocationID,
			Quantity:     qty,
			Availability: inventory.AvailabilityStatus(row.Availability),
			Price:        price,
			ObservedAt:   fromPgTimestamptz(row.ObservedAt),
			ReceivedAt:   fromPgTimestamptz(row.ReceivedAt),
			Authority:    fitment.AuthorityLevel(row.Authority),
			Confidence:   row.Confidence,
			SourceRef: product.SourceReference{
				SourceID:       row.SourceID,
				SourceType:     row.SourceType,
				SourceRecordID: row.SourceRecordID,
			},
		})
	}
	return res, nil
}

func (s *SQLCStore) SaveObservation(ctx context.Context, obs *inventory.InventoryObservation) error {
	var priceAmt *float64
	curr := "USD"
	if obs.Price != nil {
		priceAmt = &obs.Price.Amount
		if obs.Price.Currency != "" {
			curr = obs.Price.Currency
		}
	}

	var expiresAt pgtype.Timestamptz
	if obs.ExpiresAt != nil {
		expiresAt = toPgTimestamptz(*obs.ExpiresAt)
	}

	_, err := s.q.UpsertInventoryObservation(ctx, sqlc.UpsertInventoryObservationParams{
		ID:                 obs.ID,
		ProductID:          obs.ProductID,
		LocationID:         obs.LocationID,
		Quantity:           toPgInt4(obs.Quantity),
		Availability:       string(obs.Availability),
		PriceAmount:        toPgFloat8(priceAmt),
		PriceCurrency:      curr,
		ObservedAt:         toPgTimestamptz(obs.ObservedAt),
		ReceivedAt:         toPgTimestamptz(obs.ReceivedAt),
		ExpiresAt:          expiresAt,
		Authority:          string(obs.Authority),
		Confidence:         obs.Confidence,
		VerificationStatus: string(obs.VerificationStatus),
		SourceID:           obs.SourceRef.SourceID,
		SourceType:         obs.SourceRef.SourceType,
		SourceRecordID:     obs.SourceRef.SourceRecordID,
	})
	return err
}

func (s *SQLCStore) GetLocation(ctx context.Context, id string) (*inventory.Location, error) {
	loc, err := s.q.GetLocation(ctx, id)
	if err != nil {
		return nil, inventory.ErrInventoryNotFound
	}
	return &inventory.Location{
		ID:         loc.ID,
		Name:       loc.Name,
		PostalCode: loc.PostalCode,
		Latitude:   loc.Latitude,
		Longitude:  loc.Longitude,
	}, nil
}

func (s *SQLCStore) SaveLocation(ctx context.Context, loc *inventory.Location) error {
	_, err := s.q.UpsertLocation(ctx, sqlc.UpsertLocationParams{
		ID:         loc.ID,
		Name:       loc.Name,
		PostalCode: loc.PostalCode,
		Latitude:   loc.Latitude,
		Longitude:  loc.Longitude,
	})
	return err
}

// --- Evidence Repository Implementation ---

func (s *SQLCStore) SaveAsset(ctx context.Context, asset *evidence.Asset) error {
	_, err := s.q.UpsertAsset(ctx, sqlc.UpsertAssetParams{
		ID:             asset.ID,
		Uri:            asset.URI,
		MediaType:      asset.MediaType,
		Sha256:         asset.SHA256,
		PageCount:      toPgInt4(asset.PageCount),
		Width:          toPgInt4(asset.Width),
		Height:         toPgInt4(asset.Height),
		SourceID:       asset.SourceRef.SourceID,
		SourceType:     asset.SourceRef.SourceType,
		SourceRecordID: asset.SourceRef.SourceRecordID,
		IngestedAt:     toPgTimestamptz(asset.IngestedAt),
	})
	return err
}

func (s *SQLCStore) GetAsset(ctx context.Context, id string) (*evidence.Asset, error) {
	row, err := s.q.GetAsset(ctx, id)
	if err != nil {
		return nil, evidence.ErrAssetNotFound
	}
	return &evidence.Asset{
		ID:         row.ID,
		URI:        row.Uri,
		MediaType:  row.MediaType,
		SHA256:     row.Sha256,
		IngestedAt: fromPgTimestamptz(row.IngestedAt),
	}, nil
}

func (s *SQLCStore) SaveEvidence(ctx context.Context, ev *evidence.Evidence) error {
	regBytes, _ := json.Marshal(ev.Region)
	_, err := s.q.InsertEvidence(ctx, sqlc.InsertEvidenceParams{
		ID:               ev.ID,
		AssetID:          ev.AssetID,
		Page:             toPgInt4(ev.Page),
		Region:           regBytes,
		Text:             ev.Text,
		ImageUri:         ev.ImageURI,
		ContentHash:      ev.ContentHash,
		ExtractorName:    ev.Extractor.Name,
		ExtractorVersion: ev.Extractor.Version,
		CreatedAt:        toPgTimestamptz(ev.CreatedAt),
	})
	return err
}

func (s *SQLCStore) GetEvidence(ctx context.Context, id string) (*evidence.Evidence, error) {
	row, err := s.q.GetEvidenceByID(ctx, id)
	if err != nil {
		return nil, evidence.ErrEvidenceNotFound
	}
	var bbox *evidence.BoundingBox
	if len(row.Region) > 0 {
		_ = json.Unmarshal(row.Region, &bbox)
	}
	return &evidence.Evidence{
		ID:          row.ID,
		AssetID:     row.AssetID,
		Page:        fromPgInt4(row.Page),
		Region:      bbox,
		Text:        row.Text,
		ImageURI:    row.ImageUri,
		ContentHash: row.ContentHash,
		Extractor: evidence.ExtractorIdentity{
			Name:    row.ExtractorName,
			Version: row.ExtractorVersion,
		},
		CreatedAt: fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (s *SQLCStore) SaveClaim(ctx context.Context, claim *evidence.Claim) error {
	valBytes, _ := json.Marshal(claim.Value)
	evBytes, _ := json.Marshal(claim.EvidenceIDs)

	_, err := s.q.InsertClaim(ctx, sqlc.InsertClaimParams{
		ID:              claim.ID,
		SubjectType:     claim.SubjectType,
		SubjectID:       claim.SubjectID,
		FieldPath:       claim.FieldPath,
		Value:           valBytes,
		Confidence:      claim.Confidence,
		Authority:       string(claim.Authority),
		SourceID:        claim.SourceRef.SourceID,
		SourceType:      claim.SourceRef.SourceType,
		SourceRecordID:  claim.SourceRef.SourceRecordID,
		EvidenceIds:     evBytes,
		ExtractionRunID: claim.ExtractionRunID,
		CreatedAt:       toPgTimestamptz(time.Now().UTC()),
	})
	return err
}

func (s *SQLCStore) GetClaimsBySubject(ctx context.Context, subjectID string) ([]*evidence.Claim, error) {
	rows, err := s.q.ListClaimsBySubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}
	var res []*evidence.Claim
	for _, row := range rows {
		var val product.TypedValue
		if len(row.Value) > 0 {
			_ = json.Unmarshal(row.Value, &val)
		}
		var evIDs []string
		if len(row.EvidenceIds) > 0 {
			_ = json.Unmarshal(row.EvidenceIds, &evIDs)
		}

		res = append(res, &evidence.Claim{
			ID:          row.ID,
			SubjectType: row.SubjectType,
			SubjectID:   row.SubjectID,
			FieldPath:   row.FieldPath,
			Value:       val,
			Confidence:  row.Confidence,
			Authority:   fitment.AuthorityLevel(row.Authority),
			EvidenceIDs: evIDs,
			CreatedAt:   fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

func (s *SQLCStore) SaveAttestation(ctx context.Context, att *evidence.Attestation) error {
	reasons, _ := json.Marshal(att.ReasonCodes)
	evs, _ := json.Marshal(att.EvidenceIDs)
	_, err := s.q.UpsertAttestation(ctx, sqlc.UpsertAttestationParams{
		ID:            att.ID,
		ClaimID:       att.ClaimID,
		Status:        string(att.Status),
		ReasonCodes:   reasons,
		Notes:         att.Notes,
		EvidenceIds:   evs,
		ValidatorType: string(att.Validator.Type),
		ValidatorID:   att.Validator.ID,
		CreatedAt:     toPgTimestamptz(att.CreatedAt),
	})
	return err
}

func (s *SQLCStore) GetAttestationForClaim(ctx context.Context, claimID string) (*evidence.Attestation, error) {
	row, err := s.q.GetAttestationForClaim(ctx, claimID)
	if err != nil {
		return nil, evidence.ErrEvidenceNotFound
	}
	var reasons []string
	if len(row.ReasonCodes) > 0 {
		_ = json.Unmarshal(row.ReasonCodes, &reasons)
	}
	var evs []string
	if len(row.EvidenceIds) > 0 {
		_ = json.Unmarshal(row.EvidenceIds, &evs)
	}
	return &evidence.Attestation{
		ID:          row.ID,
		ClaimID:     row.ClaimID,
		Status:      evidence.AttestationStatus(row.Status),
		ReasonCodes: reasons,
		Notes:       row.Notes,
		EvidenceIDs: evs,
		Validator: evidence.ValidatorIdentity{
			Type: row.ValidatorType,
			ID:   row.ValidatorID,
		},
		CreatedAt: fromPgTimestamptz(row.CreatedAt),
	}, nil
}
