// Copyright (c) 2026 RoundPenny. All rights reserved.

package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/roundup-platform/services/admin/internal/service"
)

type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		claims, err := h.svc.ValidateToken(token)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		if claims["role"] != "admin" {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	resp, err := h.svc.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			respond(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		slog.Error("login error", "error", err)
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	respond(w, http.StatusOK, resp)
}

func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusNoContent, nil)
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		slog.Error("get stats error", "error", err)
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	respond(w, http.StatusOK, stats)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	users, total, err := h.svc.ListUsers(r.Context(), page, pageSize)
	if err != nil {
		slog.Error("list users error", "error", err)
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	respond(w, http.StatusOK, map[string]interface{}{
		"data":  users,
		"total": total,
		"page":  page,
	})
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		slog.Error("get user error", "error", err)
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if user == nil {
		respond(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	respond(w, http.StatusOK, user)
}

func (h *AdminHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		KYCStatus string `json:"kyc_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := h.svc.UpdateUserStatus(r.Context(), id, body.KYCStatus); err != nil {
		slog.Error("update user status error", "error", err)
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *AdminHandler) ListMerchants(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	merchants, total, err := h.svc.ListMerchants(r.Context(), page, pageSize)
	if err != nil {
		slog.Error("list merchants error", "error", err)
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	respond(w, http.StatusOK, map[string]interface{}{
		"data":  merchants,
		"total": total,
		"page":  page,
	})
}

func (h *AdminHandler) GetMerchant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	merchant, err := h.svc.GetMerchant(r.Context(), id)
	if err != nil {
		slog.Error("get merchant error", "error", err)
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if merchant == nil {
		respond(w, http.StatusNotFound, map[string]string{"error": "merchant not found"})
		return
	}
	respond(w, http.StatusOK, merchant)
}

func (h *AdminHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	txs, total, err := h.svc.ListTransactions(r.Context(), page, pageSize)
	if err != nil {
		slog.Error("list transactions error", "error", err)
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	respond(w, http.StatusOK, map[string]interface{}{
		"data":  txs,
		"total": total,
		"page":  page,
	})
}

func (h *AdminHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tx, err := h.svc.GetTransaction(r.Context(), id)
	if err != nil {
		slog.Error("get transaction error", "error", err)
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if tx == nil {
		respond(w, http.StatusNotFound, map[string]string{"error": "transaction not found"})
		return
	}
	respond(w, http.StatusOK, tx)
}

func (h *AdminHandler) ListFraudAlerts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	alerts, total, err := h.svc.ListFraudAlerts(r.Context(), page, pageSize)
	if err != nil {
		slog.Error("list fraud alerts error", "error", err)
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	respond(w, http.StatusOK, map[string]interface{}{
		"data":  alerts,
		"total": total,
		"page":  page,
	})
}

func (h *AdminHandler) ReviewFraudAlert(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := h.svc.ReviewFraudAlert(r.Context(), id, body.Status); err != nil {
		slog.Error("review fraud alert error", "error", err)
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *AdminHandler) ListKYCSubmissions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	subs, total, err := h.svc.ListKYCSubmissions(r.Context(), page, pageSize)
	if err != nil {
		slog.Error("list kyc submissions error", "error", err)
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	respond(w, http.StatusOK, map[string]interface{}{
		"data":  subs,
		"total": total,
		"page":  page,
	})
}

func (h *AdminHandler) ReviewKYCSubmission(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status          string `json:"status"`
		RejectionReason string `json:"rejection_reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := h.svc.ReviewKYCSubmission(r.Context(), id, body.Status, body.RejectionReason); err != nil {
		slog.Error("review kyc submission error", "error", err)
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "updated"})
}
