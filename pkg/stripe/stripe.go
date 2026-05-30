// Copyright (c) 2026 RoundPenny. All rights reserved.

package stripe

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	stripeSDK "github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/client"
	"github.com/stripe/stripe-go/v79/paymentintent"
	"github.com/stripe/stripe-go/v79/webhook"
)

type StripeClient struct {
	apiKey        string
	webhookSecret string
	client        *client.API
}

type CreatePaymentIntentParams struct {
	Amount        int64
	Currency      string
	Description   string
	PaymentMethod string
	Metadata      map[string]string
}

type PaymentIntentResponse struct {
	ID           string `json:"id"`
	ClientSecret string `json:"client_secret"`
	Status       string `json:"status"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
}

var (
	ErrWebhookSignatureInvalid = errors.New("stripe webhook signature invalid")
	ErrPaymentIntentFailed     = errors.New("stripe payment intent failed")
)

func NewClient() *StripeClient {
	apiKey := os.Getenv("STRIPE_API_KEY")
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	if apiKey == "" {
		slog.Warn("STRIPE_API_KEY not set, using mock mode")
	}

	sc := &StripeClient{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
	}

	if apiKey != "" {
		sc.client = client.New(apiKey, nil)
	}

	return sc
}

func (s *StripeClient) CreatePaymentIntent(params CreatePaymentIntentParams) (*PaymentIntentResponse, error) {
	if s.client == nil {
		return s.mockCreatePaymentIntent(params)
	}

	createParams := &stripeSDK.PaymentIntentParams{
		Amount:             stripeSDK.Int64(params.Amount),
		Currency:           stripeSDK.String(params.Currency),
		Description:        stripeSDK.String(params.Description),
		PaymentMethodTypes: []*string{stripeSDK.String("card")},
		Metadata:           params.Metadata,
	}

	if params.PaymentMethod != "" {
		createParams.PaymentMethod = stripeSDK.String(params.PaymentMethod)
	}

	pi, err := paymentintent.New(createParams)
	if err != nil {
		return nil, err
	}

	return &PaymentIntentResponse{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
	}, nil
}

func (s *StripeClient) ConfirmPaymentIntent(paymentIntentID string, paymentMethodID string) (*PaymentIntentResponse, error) {
	if s.client == nil {
		return s.mockConfirmPaymentIntent(paymentIntentID, paymentMethodID)
	}

	params := &stripeSDK.PaymentIntentConfirmParams{
		PaymentMethod: stripeSDK.String(paymentMethodID),
	}

	pi, err := paymentintent.Confirm(paymentIntentID, params)
	if err != nil {
		return nil, err
	}

	return &PaymentIntentResponse{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
	}, nil
}

func (s *StripeClient) ConstructWebhookEvent(payload []byte, signatureHeader string) (stripeSDK.Event, error) {
	if s.webhookSecret == "" {
	slog.Warn("STRIPE_WEBHOOK_SECRET not set, skipping signature verification")
		var event stripeSDK.Event
		if err := json.Unmarshal(payload, &event); err != nil {
			return stripeSDK.Event{}, err
		}
		return event, nil
	}

	event, err := webhook.ConstructEvent(payload, signatureHeader, s.webhookSecret)
	if err != nil {
		return stripeSDK.Event{}, ErrWebhookSignatureInvalid
	}
	return event, nil
}

func (s *StripeClient) mockCreatePaymentIntent(params CreatePaymentIntentParams) (*PaymentIntentResponse, error) {
	if params.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	id := "pi_mock_" + randHex(8)
	return &PaymentIntentResponse{
		ID:           id,
		ClientSecret: id + "_secret_" + randHex(16),
		Status:       "requires_payment_method",
		Amount:       params.Amount,
		Currency:     params.Currency,
	}, nil
}

func (s *StripeClient) mockConfirmPaymentIntent(paymentIntentID string, paymentMethodID string) (*PaymentIntentResponse, error) {
	return &PaymentIntentResponse{
		ID:           paymentIntentID,
		ClientSecret: paymentIntentID + "_secret_" + randHex(16),
		Status:       "succeeded",
		Amount:       1000,
		Currency:     "usd",
	}, nil
}

func randHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
