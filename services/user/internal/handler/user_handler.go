// Copyright (c) 2026 RoundPenny. All rights reserved.

package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/roundup-platform/services/user/internal/service"
)

type UserHandler struct {
	svc       *service.UserService
	jwtSecret []byte
}

func NewUserHandler(svc *service.UserService, jwtSecret string) *UserHandler {
	return &UserHandler{svc: svc, jwtSecret: []byte(jwtSecret)}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "profile not found")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req service.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if err := h.svc.UpdateProfile(r.Context(), userID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prefs, err := h.svc.GetPreferences(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "preferences not found")
		return
	}

	writeJSON(w, http.StatusOK, prefs)
}

func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req service.UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if err := h.svc.UpdatePreferences(r.Context(), userID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
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
