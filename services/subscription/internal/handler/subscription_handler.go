// Copyright (c) 2026 RoundPenny. All rights reserved.

package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/subscription/internal/service"
)

type SubscriptionHandler struct {
	svc *service.SubscriptionService
}

func NewSubscriptionHandler(svc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		PlanID string `json:"plan_id"`
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
	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid plan_id")
		return
	}

	resp, err := h.svc.CreateSubscription(r.Context(), service.CreateSubscriptionRequest{
		UserID: userID,
		PlanID: planID,
	})
	if err != nil {
		if errors.Is(err, service.ErrPlanNotFound) {
			writeError(w, http.StatusNotFound, "plan not found")
			return
		}
		if errors.Is(err, service.ErrAlreadyActive) {
			writeError(w, http.StatusConflict, "user already has an active subscription")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	subID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	resp, err := h.svc.CancelSubscription(r.Context(), subID, userID)
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		if errors.Is(err, service.ErrInvalidTransition) {
			writeError(w, http.StatusBadRequest, "subscription cannot be cancelled")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to cancel subscription")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *SubscriptionHandler) GetCurrentSubscription(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	resp, err := h.svc.GetCurrentSubscription(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			writeError(w, http.StatusNotFound, "no active subscription")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get subscription")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *SubscriptionHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.svc.ListPlans(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list plans")
		return
	}

	writeJSON(w, http.StatusOK, plans)
}

func (h *SubscriptionHandler) GetBillingHistory(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	page := 1
	pageSize := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 && n <= 100 {
			pageSize = n
		}
	}

	records, err := h.svc.GetBillingHistory(r.Context(), userID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get billing history")
		return
	}

	writeJSON(w, http.StatusOK, records)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
