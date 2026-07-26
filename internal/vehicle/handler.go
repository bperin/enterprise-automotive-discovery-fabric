// Package vehicle exposes REST HTTP API handlers for vehicle configurations and models.
package vehicle

import (
	"encoding/json"
	"net/http"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/http/response"
)

// Handler exposes Vehicle REST API endpoints in DDD.
type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes attaches authenticated vehicle endpoints to an http.ServeMux.
// @Summary Register vehicle routes
// @Security OAuth2Auth[read,write]
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authenticator auth.Authenticator) {
	readAuth := auth.Middleware(authenticator, "read")
	writeAuth := auth.Middleware(authenticator, "write")

	mux.Handle("GET /v1/vehicles/{id}", readAuth(http.HandlerFunc(h.GetVehicle)))
	mux.Handle("POST /v1/vehicles", writeAuth(http.HandlerFunc(h.CreateVehicle)))
}

// GetVehicle handles GET /v1/vehicles/{id} requests.
// @Summary Get vehicle configuration by ID
// @Security OAuth2Auth[read]
// @Success 200 {object} vehicle.VehicleConfiguration
// @Router /v1/vehicles/{id} [get]
func (h *Handler) GetVehicle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "Vehicle ID is required")
		return
	}

	v, err := h.repo.GetConfiguration(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "not_found", "Vehicle configuration not found")
		return
	}

	response.JSON(w, http.StatusOK, v)
}

// CreateVehicle handles POST /v1/vehicles write requests.
// @Summary Create or update vehicle configuration
// @Security OAuth2Auth[write]
// @Success 201 {object} vehicle.VehicleConfiguration
// @Router /v1/vehicles [post]
func (h *Handler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var req VehicleConfiguration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON body")
		return
	}

	if err := h.repo.SaveConfiguration(r.Context(), &req); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, req)
}
