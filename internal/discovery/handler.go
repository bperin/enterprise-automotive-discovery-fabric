package discovery

import (
	"encoding/json"
	"net/http"

	"enterprise-search/internal/auth"
	"enterprise-search/internal/http/response"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev/demo
	},
}

// Handler exposes the Unified Answer and Discovery Gateway REST, SSE, and WSS endpoints in DDD.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches the authenticated discovery endpoints to an http.ServeMux.
// @Summary Query unified discovery gateway via REST, SSE, or WSS
// @Security OAuth2Auth[search]
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authenticator auth.Authenticator) {
	searchAuth := auth.Middleware(authenticator, "search")

	mux.Handle("POST /v1/discovery/query", searchAuth(http.HandlerFunc(h.ProcessQuery)))
	mux.Handle("POST /v1/discovery/stream", searchAuth(http.HandlerFunc(h.StreamQuery)))
	mux.Handle("GET /v1/discovery/ws", http.HandlerFunc(h.WebSocketQuery))
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

// WebSocketQuery handles GET /v1/discovery/ws WebSocket / WSS streaming connections.
// @Summary Real-time bidirectional streaming for ADK 2.0 graph workflow execution
// @Router /v1/discovery/ws [get]
func (h *Handler) WebSocketQuery(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var req DiscoveryRequest
		if err := json.Unmarshal(message, &req); err != nil {
			errData, _ := json.Marshal(map[string]string{"error": "Failed to parse discovery request JSON"})
			_ = conn.WriteMessage(websocket.TextMessage, errData)
			continue
		}

		// Send graph execution start status
		startData, _ := json.Marshal(map[string]string{"status": "graph_started", "query": req.Query})
		_ = conn.WriteMessage(websocket.TextMessage, startData)

		resp, err := h.service.ProcessQuery(r.Context(), req)
		if err != nil {
			errData, _ := json.Marshal(map[string]string{"error": err.Error()})
			_ = conn.WriteMessage(websocket.TextMessage, errData)
			continue
		}

		respData, _ := json.Marshal(resp)
		_ = conn.WriteMessage(websocket.TextMessage, respData)
	}
}
