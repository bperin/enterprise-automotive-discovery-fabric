package inventory

import (
	"encoding/json"
	"net/http"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/http/response"
)

// Handler exposes Inventory REST API endpoints in DDD.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches authenticated inventory endpoints to an http.ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authenticator auth.Authenticator) {
	readAuth := auth.Middleware(authenticator, "read")
	writeAuth := auth.Middleware(authenticator, "write")

	mux.Handle("GET /v1/products/{id}/inventory", readAuth(http.HandlerFunc(h.GetProductInventory)))
	mux.Handle("POST /v1/inventory/observations", writeAuth(http.HandlerFunc(h.RecordObservation)))
	mux.Handle("POST /v1/inventory/locations", writeAuth(http.HandlerFunc(h.CreateLocation)))
}

func (h *Handler) GetProductInventory(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("id")
	if productID == "" {
		productID = r.URL.Query().Get("id")
	}
	if productID == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "Product ID is required")
		return
	}

	observations, err := h.service.GetInventoryForProduct(r.Context(), productID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"product_id":   productID,
		"observations": observations,
		"count":        len(observations),
	})
}

func (h *Handler) RecordObservation(w http.ResponseWriter, r *http.Request) {
	var req InventoryObservation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON body")
		return
	}

	recorded, err := h.service.RecordObservation(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_observation", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, recorded)
}

func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var req Location
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON body")
		return
	}

	created, err := h.service.SaveLocation(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_location", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, created)
}
