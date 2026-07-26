// Package content exposes REST HTTP API handlers for CMS marketing items and documents.
package content

import (
	"encoding/json"
	"net/http"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/http/response"
)

// Handler exposes Content REST API endpoints in DDD.
type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes attaches authenticated content endpoints to an http.ServeMux.
// @Summary Register marketing content routes
// @Security OAuth2Auth[read,write]
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authenticator auth.Authenticator) {
	readAuth := auth.Middleware(authenticator, "read")
	writeAuth := auth.Middleware(authenticator, "write")

	mux.Handle("GET /v1/content/{id}", readAuth(http.HandlerFunc(h.GetContent)))
	mux.Handle("POST /v1/content", writeAuth(http.HandlerFunc(h.CreateContent)))
}

// GetContent handles GET /v1/content/{id} requests.
// @Summary Get marketing content item by ID
// @Security OAuth2Auth[read]
// @Success 200 {object} content.ContentItem
// @Router /v1/content/{id} [get]
func (h *Handler) GetContent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "Content ID is required")
		return
	}

	item, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "not_found", "Content item not found")
		return
	}

	response.JSON(w, http.StatusOK, item)
}

// CreateContent handles POST /v1/content write requests.
// @Summary Publish or update marketing content
// @Security OAuth2Auth[write]
// @Success 201 {object} content.ContentItem
// @Router /v1/content [post]
func (h *Handler) CreateContent(w http.ResponseWriter, r *http.Request) {
	var req ContentItem
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
