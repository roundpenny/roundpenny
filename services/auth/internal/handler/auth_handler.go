// Copyright (c) 2026 RoundPenny. All rights reserved.

package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/auth/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "email, password, and full_name are required")
		return
	}

	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrAccountLocked) {
			writeError(w, http.StatusTooManyRequests, "account is temporarily locked")
			return
		}
		if errors.Is(err, service.ErrMFARequired) {
			writeJSON(w, http.StatusOK, resp)
			return
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MFAToken string `json:"mfa_token"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.VerifyMFA(r.Context(), req.MFAToken, req.Code)
	if err != nil {
		if errors.Is(err, service.ErrInvalidMFA) {
			writeError(w, http.StatusUnauthorized, "invalid MFA code")
			return
		}
		writeError(w, http.StatusInternalServerError, "MFA verification failed")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrTokenRevoked) {
			writeError(w, http.StatusUnauthorized, "refresh token has been revoked")
			return
		}
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.svc)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.svc.Logout(r.Context(), userID); err != nil {
		slog.Info("logout error", "error", err)
		writeError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.svc)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) SetupMFA(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.svc)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	secret, url, err := h.svc.SetupMFA(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"secret": secret,
		"url":    url,
	})
}

func (h *AuthHandler) EnableMFA(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.svc)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	backupCodes, err := h.svc.EnableMFA(r.Context(), userID, req.Code)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":      "MFA enabled successfully",
		"backup_codes": backupCodes,
	})
}

func (h *AuthHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.svc)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.svc.DisableMFA(r.Context(), userID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "MFA disabled"})
}

func (h *AuthHandler) OAuth(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	if provider != "google" && provider != "github" {
		writeError(w, http.StatusBadRequest, "unsupported provider")
		return
	}

	var req struct {
		Code        string `json:"code"`
		State       string `json:"state"`
		RedirectURI string `json:"redirect_uri"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	resp, err := h.svc.OAuthLogin(r.Context(), provider, req.Code, req.State, req.RedirectURI)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.svc)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.svc.SendEmailVerification(r.Context(), userID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "verification email sent"})
}

func (h *AuthHandler) ConfirmEmailVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	if err := h.svc.ConfirmEmailVerification(r.Context(), req.Token); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
}

func (h *AuthHandler) SubmitKYC(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.svc)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		FullName       string `json:"full_name"`
		DocumentType   string `json:"document_type"`
		DocumentNumber string `json:"document_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FullName == "" || req.DocumentType == "" || req.DocumentNumber == "" {
		writeError(w, http.StatusBadRequest, "full_name, document_type, and document_number are required")
		return
	}

	sub, err := h.svc.SubmitKYC(r.Context(), userID, req.FullName, req.DocumentType, req.DocumentNumber)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, sub)
}

func (h *AuthHandler) GetKYCStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.svc)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	status, err := h.svc.GetKYCStatus(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "no KYC submission found")
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func extractUserID(r *http.Request, svc *service.AuthService) (uuid.UUID, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return uuid.Nil, errors.New("no authorization header")
	}

	token := strings.TrimPrefix(auth, "Bearer ")
	if token == auth {
		return uuid.Nil, errors.New("invalid authorization format")
	}

	claims, err := svc.Authenticate(token)
	if err != nil {
		return uuid.Nil, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID in token")
	}

	return userID, nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
