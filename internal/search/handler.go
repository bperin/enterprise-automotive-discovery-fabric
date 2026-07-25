package search

import (
	"encoding/json"
	"net/http"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/http/response"
)

// Handler exposes Search REST API endpoints in DDD.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches authenticated search endpoints to an http.ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authenticator auth.Authenticator) {
	searchAuth := auth.Middleware(authenticator, "search")

	mux.Handle("POST /v1/search", searchAuth(http.HandlerFunc(h.Search)))
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	var req SearchQuery
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON request body")
		return
	}

	if principal, ok := auth.PrincipalFromContext(r.Context()); ok {
		req.Principal = principal
	}

	resp, err := h.service.ExecuteSearch(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "search_failed", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) DiscoveryQuery(w http.ResponseWriter, r *http.Request) {
	var req SearchQuery
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse JSON request body")
		return
	}

	if principal, ok := auth.PrincipalFromContext(r.Context()); ok {
		req.Principal = principal
	}

	resp, err := h.service.ExecuteSearch(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "discovery_failed", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, resp)
}
