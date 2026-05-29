package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/wallet/internal/service"
)

type WalletHandler struct {
	svc *service.WalletService
}

func NewWalletHandler(svc *service.WalletService) *WalletHandler {
	return &WalletHandler{svc: svc}
}

func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
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

	wallet, err := h.svc.GetOrCreateWallet(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "wallet error")
		return
	}

	writeJSON(w, http.StatusOK, wallet)
}

func (h *WalletHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
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

	entries, err := h.svc.GetTransactions(r.Context(), userID, 1, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query error")
		return
	}

	writeJSON(w, http.StatusOK, entries)
}

func (h *WalletHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID      string         `json:"user_id"`
		Amount      float64        `json:"amount"`
		Destination map[string]any `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	if err := h.svc.Withdraw(r.Context(), userID, req.Amount, req.Destination); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "withdrawal_initiated"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
