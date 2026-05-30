package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/roundup-platform/services/merchant/internal/service"
)

type MerchantHandler struct {
	svc       *service.MerchantService
	jwtSecret []byte
}

func NewMerchantHandler(svc *service.MerchantService, jwtSecret string) *MerchantHandler {
	return &MerchantHandler{svc: svc, jwtSecret: []byte(jwtSecret)}
}

func (h *MerchantHandler) CreateMerchant(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	_ = userID

	var req service.CreateMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.CreateMerchant(r.Context(), req)
	if err != nil {
		if err == service.ErrInvalidName || err == service.ErrInvalidEmail {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *MerchantHandler) GetMerchant(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid merchant id")
		return
	}

	resp, err := h.svc.GetMerchant(r.Context(), id)
	if err != nil {
		if err == service.ErrMerchantNotFound {
			writeError(w, http.StatusNotFound, "merchant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get merchant")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *MerchantHandler) UpdateMerchant(w http.ResponseWriter, r *http.Request) {
	_, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid merchant id")
		return
	}

	var req service.UpdateMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.UpdateMerchant(r.Context(), id, req)
	if err != nil {
		if err == service.ErrMerchantNotFound {
			writeError(w, http.StatusNotFound, "merchant not found")
			return
		}
		if err == service.ErrInvalidName || err == service.ErrInvalidEmail {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *MerchantHandler) DeleteMerchant(w http.ResponseWriter, r *http.Request) {
	_, err := extractUserID(r, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid merchant id")
		return
	}

	if err := h.svc.DeleteMerchant(r.Context(), id); err != nil {
		if err == service.ErrMerchantNotFound {
			writeError(w, http.StatusNotFound, "merchant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MerchantHandler) ListMerchants(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.svc.ListMerchants(r.Context(), page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list merchants")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *MerchantHandler) SearchMerchants(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "search query is required")
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

	resp, err := h.svc.SearchMerchants(r.Context(), query, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to search merchants")
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
