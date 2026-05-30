// Copyright (c) 2026 RoundPenny. All rights reserved.

package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/roundup-platform/services/fraud/internal/service"
)

var errNoAuth = errors.New("missing or invalid authorization")

type FraudHandler struct {
	ruleSvc  *service.FraudService
	alertSvc *service.FraudService
	jwtSecret []byte
}

func NewFraudHandler(svc *service.FraudService, jwtSecret string) *FraudHandler {
	return &FraudHandler{ruleSvc: svc, alertSvc: svc, jwtSecret: []byte(jwtSecret)}
}

func (h *FraudHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	_, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req service.CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.ruleSvc.CreateRule(r.Context(), req)
	if err != nil {
		if err == service.ErrInvalidRuleName || err == service.ErrInvalidRuleType {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *FraudHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rule id")
		return
	}

	resp, err := h.ruleSvc.GetRule(r.Context(), id)
	if err != nil {
		if err == service.ErrRuleNotFound {
			writeError(w, http.StatusNotFound, "rule not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get rule")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *FraudHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	_, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rule id")
		return
	}

	var req service.UpdateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.ruleSvc.UpdateRule(r.Context(), id, req)
	if err != nil {
		if err == service.ErrRuleNotFound {
			writeError(w, http.StatusNotFound, "rule not found")
			return
		}
		if err == service.ErrInvalidRuleName || err == service.ErrInvalidRuleType {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *FraudHandler) ListRules(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.ruleSvc.ListRules(r.Context(), page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list rules")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *FraudHandler) CreateAlert(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	_ = userID

	var req service.CreateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.alertSvc.CreateAlert(r.Context(), req)
	if err != nil {
		if err == service.ErrInvalidSeverity {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *FraudHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert id")
		return
	}

	resp, err := h.alertSvc.GetAlert(r.Context(), id)
	if err != nil {
		if err == service.ErrAlertNotFound {
			writeError(w, http.StatusNotFound, "alert not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get alert")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *FraudHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.alertSvc.ListAlerts(r.Context(), page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *FraudHandler) UpdateAlertStatus(w http.ResponseWriter, r *http.Request) {
	_, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert id")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.alertSvc.UpdateAlertStatus(r.Context(), id, req.Status)
	if err != nil {
		if err == service.ErrAlertNotFound {
			writeError(w, http.StatusNotFound, "alert not found")
			return
		}
		if err == service.ErrInvalidStatus {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *FraudHandler) GetAlertsByUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
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

	resp, err := h.alertSvc.GetAlertsByUser(r.Context(), userID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get alerts")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *FraudHandler) GetAlertsBySeverity(w http.ResponseWriter, r *http.Request) {
	severity := r.PathValue("severity")

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

	resp, err := h.alertSvc.GetAlertsBySeverity(r.Context(), severity, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get alerts")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func extractUserID(r *http.Request, jwtSecret []byte) (uuid.UUID, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return uuid.Nil, errNoAuth
	}

	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	if tokenStr == auth {
		return uuid.Nil, errNoAuth
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errNoAuth
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errNoAuth
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, errNoAuth
	}

	return uuid.Parse(sub)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
