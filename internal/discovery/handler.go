package discovery

import (
	"encoding/json"
	"net/http"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/http/response"
)

// Handler exposes the Unified Answer and Discovery Gateway REST endpoint in DDD.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches the authenticated discovery endpoint to an http.ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authenticator auth.Authenticator) {
	searchAuth := auth.Middleware(authenticator, "search")

	mux.Handle("POST /v1/discovery/query", searchAuth(http.HandlerFunc(h.ProcessQuery)))
}

func (h *Handler) ProcessQuery(w http.ResponseWriter, r *http.Request) {
	var req DiscoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Failed to parse discovery request JSON")
		return
	}

	if principal, ok := auth.PrincipalFromContext(r.Context()); ok {
		req.User = principal
	}

	resp, err := h.service.ProcessQuery(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "discovery_failed", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, resp)
}
