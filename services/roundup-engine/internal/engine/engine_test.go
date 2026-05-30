package engine

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/roundup-engine/internal/repository"
)

type mockRoundUpRepo struct {
	getUserPreferencesFn    func(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error)
	getDailyRoundUpTotalFn  func(ctx context.Context, userID uuid.UUID) (float64, error)
	createRoundUpFn         func(ctx context.Context, ru *repository.RoundUpRecord) error
	updateRoundUpStatusFn   func(ctx context.Context, id uuid.UUID, status string) error
}

func (m *mockRoundUpRepo) GetUserPreferences(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error) {
	return m.getUserPreferencesFn(ctx, userID)
}
func (m *mockRoundUpRepo) GetDailyRoundUpTotal(ctx context.Context, userID uuid.UUID) (float64, error) {
	return m.getDailyRoundUpTotalFn(ctx, userID)
}
func (m *mockRoundUpRepo) CreateRoundUp(ctx context.Context, ru *repository.RoundUpRecord) error {
	return m.createRoundUpFn(ctx, ru)
}
func (m *mockRoundUpRepo) UpdateRoundUpStatus(ctx context.Context, id uuid.UUID, status string) error {
	if m.updateRoundUpStatusFn != nil {
		return m.updateRoundUpStatusFn(ctx, id, status)
	}
	return nil
}

type mockREProducer struct {
	publishFn func(ctx context.Context, msg kafka.Message) error
}

func (m *mockREProducer) Publish(ctx context.Context, msg kafka.Message) error {
	return m.publishFn(ctx, msg)
}
func (m *mockREProducer) Close() error { return nil }

func fequal(a, b, eps float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < eps
}

func TestCalculate_Simple(t *testing.T) {
	result := Calculate(7.12, 1.00, 1)
	if !fequal(result, 0.88, 0.001) {
		t.Fatalf("expected ~0.88 for 7.12, got %g", result)
	}
}

func TestCalculate_ExactDollar(t *testing.T) {
	result := Calculate(10.00, 1.00, 1)
	if !fequal(result, 0, 0.001) {
		t.Fatalf("expected 0 for exact dollar, got %g", result)
	}
}

func TestCalculate_WithMultiplier(t *testing.T) {
	result := Calculate(5.50, 1.00, 3)
	expected := 0.50 * 3
	if !fequal(result, expected, 0.001) {
		t.Fatalf("expected %g (0.50 * 3), got %g", expected, result)
	}
}

func TestCalculate_CustomNearest(t *testing.T) {
	result := Calculate(7.25, 0.50, 1)
	if !fequal(result, 0.25, 0.001) {
		t.Fatalf("expected 0.25 for 7.25 with nearest 0.50, got %g", result)
	}
}

func TestCalculate_NegativeAmount(t *testing.T) {
	result := Calculate(-5.00, 1.00, 1)
	if !fequal(result, 0, 0.001) {
		t.Fatalf("expected 0 for negative amount, got %g", result)
	}
}

func TestCalculate_ZeroNearest(t *testing.T) {
	result := Calculate(7.12, 0, 1)
	if !fequal(result, 0.88, 0.001) {
		t.Fatalf("expected ~0.88 with default nearest, got %g", result)
	}
}

func TestCalculate_ZeroMultiplier(t *testing.T) {
	result := Calculate(7.12, 1.00, 0)
	if !fequal(result, 0.88, 0.001) {
		t.Fatalf("expected ~0.88 with default multiplier, got %g", result)
	}
}

func TestHandleTransactionSettled_Success(t *testing.T) {
	var published bool
	engine := NewRoundUpEngine(&mockRoundUpRepo{
		getUserPreferencesFn: func(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error) {
			return &repository.UserPreference{
				RoundToNearest:  1.00,
				MaxDailyRoundup: 5.00,
				Multiplier:      1,
				AutoInvest:      true,
			}, nil
		},
		getDailyRoundUpTotalFn: func(ctx context.Context, userID uuid.UUID) (float64, error) {
			return 0, nil
		},
		createRoundUpFn: func(ctx context.Context, ru *repository.RoundUpRecord) error {
			return nil
		},
	}, &mockREProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			published = true
			if msg.Topic != event.TopicRoundUpCalculated {
				t.Fatalf("expected topic %s, got %s", event.TopicRoundUpCalculated, msg.Topic)
			}
			return nil
		},
	})

	err := engine.HandleTransaction(context.Background(), event.TopicTransactionSettled, uuid.New().String(), mustJSON(event.TransactionSettled{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		Amount:        12.34,
		Currency:      "USD",
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !published {
		t.Fatal("expected event to be published")
	}
}

func TestHandleTransactionSettled_DailyCapExceeded(t *testing.T) {
	engine := NewRoundUpEngine(&mockRoundUpRepo{
		getUserPreferencesFn: func(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error) {
			return &repository.UserPreference{
				RoundToNearest:  1.00,
				MaxDailyRoundup: 5.00,
				Multiplier:      1,
			}, nil
		},
		getDailyRoundUpTotalFn: func(ctx context.Context, userID uuid.UUID) (float64, error) {
			return 4.80, nil
		},
		createRoundUpFn: func(ctx context.Context, ru *repository.RoundUpRecord) error {
			if !fequal(ru.RoundUpAmount, 0.20, 0.001) {
				t.Fatalf("expected capped roundup ~0.20, got %g", ru.RoundUpAmount)
			}
			return nil
		},
	}, &mockREProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			return nil
		},
	})

	err := engine.HandleTransaction(context.Background(), event.TopicTransactionSettled, uuid.New().String(), mustJSON(event.TransactionSettled{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		Amount:        7.12,
		Currency:      "USD",
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHandleTransactionSettled_FullyCapped(t *testing.T) {
	engine := NewRoundUpEngine(&mockRoundUpRepo{
		getUserPreferencesFn: func(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error) {
			return &repository.UserPreference{
				RoundToNearest:  1.00,
				MaxDailyRoundup: 5.00,
				Multiplier:      1,
			}, nil
		},
		getDailyRoundUpTotalFn: func(ctx context.Context, userID uuid.UUID) (float64, error) {
			return 5.00, nil
		},
	}, &mockREProducer{})

	err := engine.HandleTransaction(context.Background(), event.TopicTransactionSettled, uuid.New().String(), mustJSON(event.TransactionSettled{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		Amount:        7.12,
		Currency:      "USD",
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCalculate_RoundUpZeroAmount(t *testing.T) {
	result := Calculate(0.00, 1.00, 1)
	if !fequal(result, 0, 0.001) {
		t.Fatalf("expected 0 for 0.00 amount, got %g", result)
	}
}

func TestCalculate_LargeAmount(t *testing.T) {
	result := Calculate(999999.99, 1.00, 1)
	if !fequal(result, 0.01, 0.001) {
		t.Fatalf("expected ~0.01 for 999999.99, got %g", result)
	}
}

func TestCalculate_LargeAmountWithMultiplier(t *testing.T) {
	result := Calculate(1234567.89, 0.05, 10)
	expected := 0.01 * 10
	if !fequal(result, expected, 0.001) {
		t.Fatalf("expected %g for large amount with custom nearest and multiplier, got %g", expected, result)
	}
}

func TestCalculate_ZeroAmountWithMultiplier(t *testing.T) {
	result := Calculate(0.00, 1.00, 5)
	if !fequal(result, 0, 0.001) {
		t.Fatalf("expected 0 for zero amount regardless of multiplier, got %g", result)
	}
}

func TestHandleTransactionSettled_ZeroRoundUp(t *testing.T) {
	var published bool
	engine := NewRoundUpEngine(&mockRoundUpRepo{
		getUserPreferencesFn: func(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error) {
			return &repository.UserPreference{
				RoundToNearest:  1.00,
				MaxDailyRoundup: 5.00,
				Multiplier:      1,
			}, nil
		},
		getDailyRoundUpTotalFn: func(ctx context.Context, userID uuid.UUID) (float64, error) {
			return 0, nil
		},
	}, &mockREProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			published = true
			return nil
		},
	})

	err := engine.HandleTransaction(context.Background(), event.TopicTransactionSettled, uuid.New().String(), mustJSON(event.TransactionSettled{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		Amount:        0.00,
		Currency:      "USD",
	}))
	if err != nil {
		t.Fatalf("expected no error for zero amount, got %v", err)
	}
	if published {
		t.Fatal("expected no event published for zero round-up")
	}
}

func TestHandleTransactionSettled_NegativeAmount(t *testing.T) {
	var published bool
	engine := NewRoundUpEngine(&mockRoundUpRepo{
		getUserPreferencesFn: func(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error) {
			return &repository.UserPreference{
				RoundToNearest:  1.00,
				MaxDailyRoundup: 5.00,
				Multiplier:      1,
			}, nil
		},
		getDailyRoundUpTotalFn: func(ctx context.Context, userID uuid.UUID) (float64, error) {
			return 0, nil
		},
	}, &mockREProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			published = true
			return nil
		},
	})

	err := engine.HandleTransaction(context.Background(), event.TopicTransactionSettled, uuid.New().String(), mustJSON(event.TransactionSettled{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		Amount:        -50.00,
		Currency:      "USD",
	}))
	if err != nil {
		t.Fatalf("expected no error for negative amount, got %v", err)
	}
	if published {
		t.Fatal("expected no event published for negative amount")
	}
}

func TestHandleTransactionSettled_VeryLargeAmount(t *testing.T) {
	var published bool
	var capturedRoundUp float64
	engine := NewRoundUpEngine(&mockRoundUpRepo{
		getUserPreferencesFn: func(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error) {
			return &repository.UserPreference{
				RoundToNearest:  0.01,
				MaxDailyRoundup: 1000000.00,
				Multiplier:      2,
			}, nil
		},
		getDailyRoundUpTotalFn: func(ctx context.Context, userID uuid.UUID) (float64, error) {
			return 0, nil
		},
		createRoundUpFn: func(ctx context.Context, ru *repository.RoundUpRecord) error {
			capturedRoundUp = ru.RoundUpAmount
			return nil
		},
	}, &mockREProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			published = true
			return nil
		},
	})

	err := engine.HandleTransaction(context.Background(), event.TopicTransactionSettled, uuid.New().String(), mustJSON(event.TransactionSettled{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		Amount:        99999999.99,
		Currency:      "USD",
	}))
	if err != nil {
		t.Fatalf("expected no error for large amount, got %v", err)
	}
	if !published {
		t.Fatal("expected event to be published for large amount")
	}
	if capturedRoundUp <= 0 {
		t.Fatalf("expected positive round-up for large amount, got %g", capturedRoundUp)
	}
}

func TestHandleTransactionSettled_PublishFailure(t *testing.T) {
	var statusUpdated bool
	var updatedStatus string
	engine := NewRoundUpEngine(&mockRoundUpRepo{
		getUserPreferencesFn: func(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error) {
			return &repository.UserPreference{
				RoundToNearest:  1.00,
				MaxDailyRoundup: 5.00,
				Multiplier:      1,
			}, nil
		},
		getDailyRoundUpTotalFn: func(ctx context.Context, userID uuid.UUID) (float64, error) {
			return 0, nil
		},
		createRoundUpFn: func(ctx context.Context, ru *repository.RoundUpRecord) error {
			return nil
		},
		updateRoundUpStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
			statusUpdated = true
			updatedStatus = status
			return nil
		},
	}, &mockREProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			return errors.New("kafka unavailable")
		},
	})

	err := engine.HandleTransaction(context.Background(), event.TopicTransactionSettled, uuid.New().String(), mustJSON(event.TransactionSettled{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		Amount:        7.12,
		Currency:      "USD",
	}))
	if err == nil {
		t.Fatal("expected error for publish failure")
	}
	if !statusUpdated {
		t.Fatal("expected UpdateRoundUpStatus to be called on publish failure")
	}
	if updatedStatus != "failed" {
		t.Fatalf("expected status 'failed', got %q", updatedStatus)
	}
}

func TestHandleTransactionSettled_ConcurrentRoundups(t *testing.T) {
	var mu sync.Mutex
	createCount := 0

	engine := NewRoundUpEngine(&mockRoundUpRepo{
		getUserPreferencesFn: func(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error) {
			return &repository.UserPreference{
				RoundToNearest:  1.00,
				MaxDailyRoundup: 50.00,
				Multiplier:      1,
			}, nil
		},
		getDailyRoundUpTotalFn: func(ctx context.Context, userID uuid.UUID) (float64, error) {
			return 0, nil
		},
		createRoundUpFn: func(ctx context.Context, ru *repository.RoundUpRecord) error {
			mu.Lock()
			createCount++
			mu.Unlock()
			return nil
		},
	}, &mockREProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			return nil
		},
	})

	n := 20
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			_ = engine.HandleTransaction(context.Background(), event.TopicTransactionSettled, uuid.New().String(), mustJSON(event.TransactionSettled{
				TransactionID: uuid.New().String(),
				UserID:        uuid.New().String(),
				Amount:        10.01 + float64(i),
				Currency:      "USD",
			}))
		}(i)
	}

	wg.Wait()
	if createCount != n {
		t.Fatalf("expected %d round-ups created, got %d", n, createCount)
	}
}

func TestHandleTransactionSettled_ExactDollar(t *testing.T) {
	engine := NewRoundUpEngine(&mockRoundUpRepo{
		getUserPreferencesFn: func(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error) {
			return &repository.UserPreference{
				RoundToNearest:  1.00,
				MaxDailyRoundup: 5.00,
				Multiplier:      1,
			}, nil
		},
		getDailyRoundUpTotalFn: func(ctx context.Context, userID uuid.UUID) (float64, error) {
			return 0, nil
		},
	}, &mockREProducer{})

	err := engine.HandleTransaction(context.Background(), event.TopicTransactionSettled, uuid.New().String(), mustJSON(event.TransactionSettled{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		Amount:        10.00,
		Currency:      "USD",
	}))
	if err != nil {
		t.Fatalf("expected no error for exact dollar, got %v", err)
	}
}

func mustJSON(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}
