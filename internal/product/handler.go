package product

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/http/response"
)

// Handler exposes Product REST API endpoints in DDD.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches authenticated product routes to a mux or chi router.
// @Summary List and manage automotive catalog products
// @Security OAuth2Auth[read,write]
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authenticator auth.Authenticator) {
	authMiddleware := auth.Middleware(authenticator, "read")
	writeAuthMiddleware := auth.Middleware(authenticator, "write")

	mux.Handle("GET /v1/products", authMiddleware(http.HandlerFunc(h.ListProducts)))
	mux.Handle("GET /v1/products/{id}", authMiddleware(http.HandlerFunc(h.GetProduct)))
	mux.Handle("GET /v1/products/lookup", authMiddleware(http.HandlerFunc(h.LookupIdentifier)))
	mux.Handle("POST /v1/products", writeAuthMiddleware(http.HandlerFunc(h.CreateProduct)))
}

// GetProduct handles GET /v1/products/{id} requests.
// @Summary Get product by ID
// @Security OAuth2Auth[read]
// @Success 200 {object} product.Product
// @Router /v1/products/{id} [get]
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		// Fallback for query param or url parsing
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "Product ID is required")
		return
	}

	p, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "Product not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, p)
}

func (h *Handler) LookupIdentifier(w http.ResponseWriter, r *http.Request) {
	identifier := r.URL.Query().Get("identifier")
	if identifier == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "Query param 'identifier' is required")
		return
	}

	p, err := h.service.LookupIdentifier(r.Context(), identifier)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "No product matching identifier")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, p)
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req Product
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON body")
		return
	}

	created, err := h.service.CreateOrUpdateProduct(r.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrInvalidProduct) || strings.Contains(err.Error(), "required") {
			response.Error(w, http.StatusBadRequest, "invalid_product", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, created)
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	products, err := h.service.ListProducts(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"products": products,
		"limit":    limit,
		"offset":   offset,
		"count":    len(products),
	})
}
