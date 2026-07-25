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

// RegisterRoutes attaches the authenticated discovery endpoints to an http.ServeMux.
// @Summary Query unified discovery gateway via REST or SSE
// @Security OAuth2Auth[search]
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authenticator auth.Authenticator) {
	searchAuth := auth.Middleware(authenticator, "search")

	mux.Handle("POST /v1/discovery/query", searchAuth(http.HandlerFunc(h.ProcessQuery)))
	mux.Handle("POST /v1/discovery/stream", searchAuth(http.HandlerFunc(h.StreamQuery)))
}

// ProcessQuery handles POST /v1/discovery/query REST requests.
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

// StreamQuery handles POST /v1/discovery/stream Server-Sent Events (SSE) streaming requests.
// @Summary Stream graph workflow execution progress and final answer via SSE
// @Security OAuth2Auth[search]
// @Router /v1/discovery/stream [post]
func (h *Handler) StreamQuery(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "streaming_unsupported", "Server-Sent Events streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var req DiscoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eventData, _ := json.Marshal(map[string]string{"error": "Failed to parse discovery request JSON"})
		w.Write([]byte("event: error\ndata: " + string(eventData) + "\n\n"))
		flusher.Flush()
		return
	}

	if principal, ok := auth.PrincipalFromContext(r.Context()); ok {
		req.User = principal
	}

	// Send initial event
	startData, _ := json.Marshal(map[string]string{"status": "graph_started", "query": req.Query})
	w.Write([]byte("event: status\ndata: " + string(startData) + "\n\n"))
	flusher.Flush()

	resp, err := h.service.ProcessQuery(r.Context(), req)
	if err != nil {
		errData, _ := json.Marshal(map[string]string{"error": err.Error()})
		w.Write([]byte("event: error\ndata: " + string(errData) + "\n\n"))
		flusher.Flush()
		return
	}

	// Stream final result event
	respData, _ := json.Marshal(resp)
	w.Write([]byte("event: result\ndata: " + string(respData) + "\n\n"))
	flusher.Flush()
}
