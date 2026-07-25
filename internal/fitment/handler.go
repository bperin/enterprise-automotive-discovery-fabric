package fitment

import (
	"encoding/json"
	"net/http"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/http/response"
)

// Handler exposes Fitment REST API endpoints in DDD.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches authenticated fitment endpoints to an http.ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authenticator auth.Authenticator) {
	readAuth := auth.Middleware(authenticator, "read")
	writeAuth := auth.Middleware(authenticator, "write")

	mux.Handle("GET /v1/products/{id}/fitments", readAuth(http.HandlerFunc(h.GetProductFitments)))
	mux.Handle("POST /v1/fitment/verify", readAuth(http.HandlerFunc(h.VerifyFitment)))
	mux.Handle("POST /v1/fitment/assertions", writeAuth(http.HandlerFunc(h.CreateAssertion)))
}

type VerifyRequest struct {
	ProductID string `json:"product_id"`
	VehicleID string `json:"vehicle_id"`
}

func (h *Handler) VerifyFitment(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON body")
		return
	}

	result, err := h.service.VerifyFitment(r.Context(), req.ProductID, req.VehicleID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) GetProductFitments(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("id")
	if productID == "" {
		productID = r.URL.Query().Get("id")
	}
	if productID == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "Product ID required")
		return
	}

	fitments, err := h.service.GetFitmentsByProduct(r.Context(), productID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"product_id": productID,
		"fitments":   fitments,
		"count":      len(fitments),
	})
}

func (h *Handler) CreateAssertion(w http.ResponseWriter, r *http.Request) {
	var req FitmentAssertion
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON body")
		return
	}

	created, err := h.service.CreateAssertion(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_assertion", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, created)
}
