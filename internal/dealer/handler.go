// Package dealer exposes REST HTTP API handlers for managing dealer organizations,
// location updates, and inventory write operations.
package dealer

import (
	"encoding/json"
	"net/http"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/http/response"
)

// Handler exposes Dealer REST API endpoints in DDD.
type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes attaches authenticated dealer endpoints to an http.ServeMux.
// @Summary Register dealer routes
// @Security OAuth2Auth[read,write]
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authenticator auth.Authenticator) {
	readAuth := auth.Middleware(authenticator, "read")
	writeAuth := auth.Middleware(authenticator, "write")

	mux.Handle("GET /v1/dealers", readAuth(http.HandlerFunc(h.ListDealers)))
	mux.Handle("GET /v1/dealers/{id}", readAuth(http.HandlerFunc(h.GetDealer)))
	mux.Handle("POST /v1/dealers", writeAuth(http.HandlerFunc(h.CreateDealer)))
}

// GetDealer handles GET /v1/dealers/{id} requests.
// @Summary Get dealer by ID
// @Security OAuth2Auth[read]
// @Success 200 {object} dealer.Dealer
// @Router /v1/dealers/{id} [get]
func (h *Handler) GetDealer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "Dealer ID is required")
		return
	}

	d, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "not_found", "Dealer not found")
		return
	}

	response.JSON(w, http.StatusOK, d)
}

// CreateDealer handles POST /v1/dealers write requests.
// @Summary Create or update dealer organization
// @Security OAuth2Auth[write]
// @Success 201 {object} dealer.Dealer
// @Router /v1/dealers [post]
func (h *Handler) CreateDealer(w http.ResponseWriter, r *http.Request) {
	var req Dealer
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON body")
		return
	}

	if err := h.repo.Save(r.Context(), &req); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, req)
}

// ListDealers handles GET /v1/dealers requests.
// @Summary List all registered dealers
// @Security OAuth2Auth[read]
// @Success 200 {array} dealer.Dealer
// @Router /v1/dealers [get]
func (h *Handler) ListDealers(w http.ResponseWriter, r *http.Request) {
	dealers, err := h.repo.List(r.Context(), 20, 0)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, dealers)
}
