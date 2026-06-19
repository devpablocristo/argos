package chat

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/devpablocristo/argos/core/internal/axis"
)

type axisClient interface {
	Chat(ctx context.Context, req axis.ChatRequest) (axis.ChatResponse, error)
	ListChatConversations(ctx context.Context, limit int) (axis.ChatConversationListResult, error)
	GetChatConversation(ctx context.Context, id string) (axis.ChatConversationDetail, error)
}

type Handler struct {
	axis axisClient
}

func NewHandler(axisClient axisClient) *Handler {
	return &Handler{axis: axisClient}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/chat", h.send)
	mux.HandleFunc("GET /v1/chat/conversations", h.listConversations)
	mux.HandleFunc("GET /v1/chat/conversations/{id}", h.getConversation)
}

func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	if h.axis == nil {
		writeError(w, http.StatusServiceUnavailable, "axis_not_configured", "Axis Companion is not configured")
		return
	}
	var req axis.ChatRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if strings.TrimSpace(req.Message) == "" && len(req.ConfirmedActions) == 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "message is required")
		return
	}
	req.ProductSurface = "argos"
	out, err := h.axis.Chat(r.Context(), req)
	if err != nil {
		if errors.Is(err, axis.ErrNotConfigured) {
			writeError(w, http.StatusServiceUnavailable, "axis_not_configured", "Axis Companion is not configured")
			return
		}
		writeError(w, http.StatusBadGateway, "axis_chat_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) listConversations(w http.ResponseWriter, r *http.Request) {
	if h.axis == nil {
		writeJSON(w, http.StatusOK, axis.ChatConversationListResult{Items: []axis.ChatConversationSummary{}})
		return
	}
	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100 {
			writeError(w, http.StatusBadRequest, "validation_error", "limit must be between 1 and 100")
			return
		}
		limit = parsed
	}
	out, err := h.axis.ListChatConversations(r.Context(), limit)
	if err != nil {
		if errors.Is(err, axis.ErrNotConfigured) {
			writeJSON(w, http.StatusOK, axis.ChatConversationListResult{Items: []axis.ChatConversationSummary{}})
			return
		}
		writeError(w, http.StatusBadGateway, "axis_chat_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) getConversation(w http.ResponseWriter, r *http.Request) {
	if h.axis == nil {
		writeError(w, http.StatusServiceUnavailable, "axis_not_configured", "Axis Companion is not configured")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "conversation id is required")
		return
	}
	out, err := h.axis.GetChatConversation(r.Context(), id)
	if err != nil {
		if errors.Is(err, axis.ErrNotConfigured) {
			writeError(w, http.StatusServiceUnavailable, "axis_not_configured", "Axis Companion is not configured")
			return
		}
		writeError(w, http.StatusBadGateway, "axis_chat_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"code": code, "message": message})
}
