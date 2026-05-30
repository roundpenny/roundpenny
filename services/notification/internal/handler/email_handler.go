package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/roundup-platform/services/notification/internal/service"
)

type EmailHandler struct {
	svc *service.EmailService
}

func NewEmailHandler(svc *service.EmailService) *EmailHandler {
	return &EmailHandler{svc: svc}
}

type sendEmailRequest struct {
	UserID  string `json:"user_id"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	HTML    string `json:"html"`
}

func (h *EmailHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	var req sendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.To == "" || req.Subject == "" {
		writeError(w, http.StatusBadRequest, "to and subject required")
		return
	}

	log, err := h.svc.Send(r.Context(), req.UserID, req.To, req.Subject, req.Body, req.HTML)
	if err != nil {
		slog.Error("send email error", "error", err)
		writeError(w, http.StatusInternalServerError, "send failed")
		return
	}
	writeJSON(w, http.StatusOK, log)
}

func (h *EmailHandler) ListEmails(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	logs, total, err := h.svc.List(r.Context(), page, pageSize)
	if err != nil {
		slog.Error("list emails error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  logs,
		"total": total,
		"page":  page,
	})
}
