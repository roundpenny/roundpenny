package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	stripepkg "github.com/roundup-platform/pkg/stripe"
	"github.com/roundup-platform/services/payment-gateway/internal/service"
)

type PaymentHandler struct {
	svc          *service.PaymentService
	stripeClient *stripepkg.StripeClient
}

func NewPaymentHandler(svc *service.PaymentService, stripeClient *stripepkg.StripeClient) *PaymentHandler {
	return &PaymentHandler{svc: svc, stripeClient: stripeClient}
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID        string         `json:"user_id"`
		Amount        float64        `json:"amount"`
		Currency      string         `json:"currency"`
		PaymentMethod string         `json:"payment_method"`
		Description   string         `json:"description"`
		Metadata      map[string]any `json:"metadata"`
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

	resp, err := h.svc.CreatePayment(r.Context(), service.CreatePaymentRequest{
		UserID:        userID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		Description:   req.Description,
		Metadata:      req.Metadata,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment id")
		return
	}

	resp, err := h.svc.GetPayment(r.Context(), id)
	if err != nil {
		if err == service.ErrPaymentNotFound {
			writeError(w, http.StatusNotFound, "payment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get payment")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *PaymentHandler) ListUserPayments(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.svc.ListUserPayments(r.Context(), userID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list payments")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *PaymentHandler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment id")
		return
	}

	var req struct {
		StripePaymentMethodID string `json:"stripe_payment_method_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.ConfirmPayment(r.Context(), id, req.StripePaymentMethodID)
	if err != nil {
		if err == service.ErrPaymentNotFound {
			writeError(w, http.StatusNotFound, "payment not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := h.stripeClient.ConstructWebhookEvent(payload, sigHeader)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var intent struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
			writeError(w, http.StatusBadRequest, "invalid event data")
			return
		}

		payment, err := h.svc.GetPaymentByStripeIntent(r.Context(), intent.ID)
		if err != nil {
			writeError(w, http.StatusNotFound, "payment not found for intent")
			return
		}

		txID := uuid.New()
		if _, err := h.svc.SucceedPayment(r.Context(), payment.ID, txID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to confirm payment")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "processed"})

	case "payment_intent.payment_failed":
		var intent struct {
			ID               string `json:"id"`
			LastPaymentError *struct {
				Message string `json:"message"`
			} `json:"last_payment_error"`
		}
		if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
			writeError(w, http.StatusBadRequest, "invalid event data")
			return
		}

		errorMessage := ""
		if intent.LastPaymentError != nil {
			errorMessage = intent.LastPaymentError.Message
		}

		payment, err := h.svc.GetPaymentByStripeIntent(r.Context(), intent.ID)
		if err != nil {
			writeError(w, http.StatusNotFound, "payment not found for intent")
			return
		}

		if _, err := h.svc.FailPayment(r.Context(), payment.ID, errorMessage); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to fail payment")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "processed"})

	default:
		writeJSON(w, http.StatusOK, map[string]string{"status": "unhandled"})
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
