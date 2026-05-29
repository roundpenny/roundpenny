package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/notification/internal/service"
)

type WebhookHandler struct {
	svc *service.WebhookService
}

func NewWebhookHandler(svc *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID      string   `json:"user_id"`
		URL         string   `json:"url"`
		Secret      string   `json:"secret"`
		Events      []string `json:"events"`
		Description string   `json:"description"`
		RetryCount  int      `json:"retry_count"`
		TimeoutMs   int      `json:"timeout_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	resp, err := h.svc.CreateWebhook(r.Context(), service.CreateWebhookRequest{
		UserID:      userID,
		URL:         req.URL,
		Secret:      req.Secret,
		Events:      req.Events,
		Description: req.Description,
		RetryCount:  req.RetryCount,
		TimeoutMs:   req.TimeoutMs,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *WebhookHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook id")
		return
	}

	resp, err := h.svc.GetWebhook(r.Context(), id)
	if err != nil {
		if err == service.ErrWebhookNotFound {
			writeError(w, http.StatusNotFound, "webhook not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get webhook")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *WebhookHandler) ListUserWebhooks(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	resp, err := h.svc.ListUserWebhooks(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list webhooks")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *WebhookHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook id")
		return
	}

	var req struct {
		URL         string `json:"url"`
		Secret      string `json:"secret"`
		Events      []string `json:"events"`
		IsActive    *bool  `json:"is_active"`
		Description string `json:"description"`
		RetryCount  int    `json:"retry_count"`
		TimeoutMs   int    `json:"timeout_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.UpdateWebhook(r.Context(), id, service.UpdateWebhookRequest{
		URL:         req.URL,
		Secret:      req.Secret,
		Events:      req.Events,
		IsActive:    req.IsActive,
		Description: req.Description,
		RetryCount:  req.RetryCount,
		TimeoutMs:   req.TimeoutMs,
	})
	if err != nil {
		if err == service.ErrWebhookNotFound {
			writeError(w, http.StatusNotFound, "webhook not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook id")
		return
	}

	if err := h.svc.DeleteWebhook(r.Context(), id); err != nil {
		if err == service.ErrWebhookNotFound {
			writeError(w, http.StatusNotFound, "webhook not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete webhook")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
