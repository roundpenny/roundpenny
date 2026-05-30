package event

import "time"

type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Subject   string    `json:"subject"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

type TransactionSettled struct {
	TransactionID string  `json:"transaction_id"`
	UserID        string  `json:"user_id"`
	MerchantID    string  `json:"merchant_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	ExternalTxID  string  `json:"external_tx_id"`
}

type RoundUpCalculated struct {
	TransactionID string  `json:"transaction_id"`
	UserID        string  `json:"user_id"`
	OriginalAmount float64 `json:"original_amount"`
	RoundedAmount  float64 `json:"rounded_amount"`
	RoundUpAmount  float64 `json:"round_up_amount"`
	Currency       string  `json:"currency"`
}

type WalletCredited struct {
	UserID    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Reference string  `json:"reference"`
	Balance   float64 `json:"balance"`
}

type FeeCharged struct {
	TransactionID string  `json:"transaction_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	FeeType       string  `json:"fee_type"`
}

type InvestmentCreated struct {
	UserID    string  `json:"user_id"`
	Portfolio string  `json:"portfolio"`
	Amount    float64 `json:"amount"`
}

type SubscriptionCreated struct {
	SubscriptionID string  `json:"subscription_id"`
	UserID         string  `json:"user_id"`
	PlanID         string  `json:"plan_id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
}

type SubscriptionCancelled struct {
	SubscriptionID string `json:"subscription_id"`
	UserID         string `json:"user_id"`
}

type SubscriptionRenewed struct {
	SubscriptionID string  `json:"subscription_id"`
	UserID         string  `json:"user_id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
}

type PaymentFailed struct {
	SubscriptionID string  `json:"subscription_id"`
	UserID         string  `json:"user_id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
}

const (
	TopicTransactionSettled  = "tx.settled"
	TopicRoundUpCalculated   = "roundup.calculated"
	TopicWalletCredited      = "wallet.credited"
	TopicFeeCharged          = "fee.charged"
	TopicInvestmentCreated   = "investment.created"
	TopicSubscriptionCreated = "subscription.created"
	TopicSubscriptionCancelled = "subscription.cancelled"
	TopicSubscriptionRenewed = "subscription.renewed"
	TopicPaymentFailed       = "subscription.payment_failed"
)
