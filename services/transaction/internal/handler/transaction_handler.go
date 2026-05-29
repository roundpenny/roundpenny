package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/transaction/internal/service"
)

type TransactionHandler struct {
	svc *service.TransactionService
}

func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	tx, err := h.svc.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	tx, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	writeJSON(w, http.StatusOK, tx)
}

func (h *TransactionHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		writeError(w, http.StatusBadRequest, "user_id required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	txs, err := h.svc.ListByUser(r.Context(), userID, 1, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query failed")
		return
	}

	writeJSON(w, http.StatusOK, txs)
}

func (h *TransactionHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")

	body, err := json.Marshal(map[string]any{
		"provider": provider,
		"status":   "received",
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "marshal error")
		return
	}

	tx, err := h.svc.ProcessWebhook(r.Context(), provider, body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, tx)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
