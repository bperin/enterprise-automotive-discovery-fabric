package evidence

import (
	"encoding/json"
	"net/http"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/http/response"
)

// Handler exposes Evidence and Asset REST API endpoints in DDD.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches authenticated asset and claim endpoints to an http.ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authenticator auth.Authenticator) {
	readAuth := auth.Middleware(authenticator, "read")
	writeAuth := auth.Middleware(authenticator, "write")

	mux.Handle("POST /v1/assets", writeAuth(http.HandlerFunc(h.RegisterAsset)))
	mux.Handle("GET /v1/assets/{id}", readAuth(http.HandlerFunc(h.GetAsset)))
	mux.Handle("POST /v1/claims", writeAuth(http.HandlerFunc(h.ProposeClaim)))
	mux.Handle("GET /v1/claims", readAuth(http.HandlerFunc(h.GetClaimsBySubject)))
	mux.Handle("POST /v1/attestations", writeAuth(http.HandlerFunc(h.AttestClaim)))
}

func (h *Handler) RegisterAsset(w http.ResponseWriter, r *http.Request) {
	var req Asset
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON body")
		return
	}

	created, err := h.service.RegisterAsset(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_asset", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, created)
}

func (h *Handler) GetAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "Asset ID required")
		return
	}

	asset, err := h.service.GetAsset(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "not_found", "Asset not found")
		return
	}

	response.JSON(w, http.StatusOK, asset)
}

func (h *Handler) ProposeClaim(w http.ResponseWriter, r *http.Request) {
	var req Claim
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON body")
		return
	}

	created, err := h.service.ProposeClaim(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_claim", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, created)
}

func (h *Handler) GetClaimsBySubject(w http.ResponseWriter, r *http.Request) {
	subjectID := r.URL.Query().Get("subject_id")
	if subjectID == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "Query param 'subject_id' required")
		return
	}

	claims, err := h.service.GetClaimsForSubject(r.Context(), subjectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"subject_id": subjectID,
		"claims":     claims,
		"count":      len(claims),
	})
}

func (h *Handler) AttestClaim(w http.ResponseWriter, r *http.Request) {
	var req Attestation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON body")
		return
	}

	created, err := h.service.AttestClaim(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_attestation", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, created)
}
