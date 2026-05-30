// Copyright (c) 2026 RoundPenny. All rights reserved.

package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/roundup-platform/services/analytics/internal/service"
)

var errNoAuth = errors.New("unauthorized")

type AnalyticsHandler struct {
	svc       *service.AnalyticsService
	jwtSecret []byte
}

func NewAnalyticsHandler(svc *service.AnalyticsService, jwtSecret string) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc, jwtSecret: []byte(jwtSecret)}
}

func (h *AnalyticsHandler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req service.TrackEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.UserID = userID

	resp, err := h.svc.TrackEvent(r.Context(), req)
	if err != nil {
		if err == service.ErrInvalidType {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *AnalyticsHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.svc.GetUserEvents(r.Context(), userID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list events")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AnalyticsHandler) GetDailyStats(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")
	if startStr == "" || endStr == "" {
		writeError(w, http.StatusBadRequest, "start_date and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_date format (use YYYY-MM-DD)")
		return
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end_date format (use YYYY-MM-DD)")
		return
	}

	endDate = endDate.Add(24 * time.Hour)

	resp, err := h.svc.GetDailyStats(r.Context(), startDate, endDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get stats")
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
