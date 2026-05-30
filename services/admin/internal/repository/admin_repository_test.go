package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockRow struct {
	scanFn func(dest ...any) error
}

func (m *mockRow) Scan(dest ...any) error {
	return m.scanFn(dest...)
}

type mockRows struct {
	closeFn func()
	errFn   func() error
	nextFn  func() bool
	scanFn  func(dest ...any) error
}

func (m *mockRows) Close() { m.closeFn() }
func (m *mockRows) Err() error {
	if m.errFn != nil {
		return m.errFn()
	}
	return nil
}
func (m *mockRows) Next() bool { return m.nextFn() }
func (m *mockRows) Scan(dest ...any) error { return m.scanFn(dest...) }
func (m *mockRows) CommandTag() pgconn.CommandTag { return pgconn.NewCommandTag("INSERT 0 1") }
func (m *mockRows) Values() ([]any, error) { return nil, nil }
func (m *mockRows) RawValues() [][]byte { return nil }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockRows) Conn() *pgx.Conn { return nil }

type mockPool struct {
	queryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
	queryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	execFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (m *mockPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.queryRowFn(ctx, sql, args...)
}
func (m *mockPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return m.queryFn(ctx, sql, args...)
}
func (m *mockPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return m.execFn(ctx, sql, args...)
}

func TestGetDashboardStats(t *testing.T) {
	callCount := 0
	pool := &mockPool{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			callCount++
			return &mockRow{
				scanFn: func(dest ...any) error {
					switch callCount {
					case 1:
						*dest[0].(*int) = 100 // TotalUsers
					case 2:
						*dest[0].(*int) = 15 // PendingKYC
					case 3:
						*dest[0].(*int) = 42 // ActiveMerchants
					case 4:
						*dest[0].(*int) = 5678 // TotalTransactions
					case 5:
						*dest[0].(*int) = 3 // OpenFraudAlerts
					case 6:
						*dest[0].(*int) = 8 // PendingPayments
					}
					return nil
				},
			}
		},
	}
	repo := NewAdminRepository(pool)

	stats, err := repo.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalUsers != 100 {
		t.Errorf("TotalUsers = %d, want 100", stats.TotalUsers)
	}
	if stats.PendingKYC != 15 {
		t.Errorf("PendingKYC = %d, want 15", stats.PendingKYC)
	}
	if stats.ActiveMerchants != 42 {
		t.Errorf("ActiveMerchants = %d, want 42", stats.ActiveMerchants)
	}
	if stats.TotalTransactions != 5678 {
		t.Errorf("TotalTransactions = %d, want 5678", stats.TotalTransactions)
	}
	if stats.OpenFraudAlerts != 3 {
		t.Errorf("OpenFraudAlerts = %d, want 3", stats.OpenFraudAlerts)
	}
	if stats.PendingPayments != 8 {
		t.Errorf("PendingPayments = %d, want 8", stats.PendingPayments)
	}
	if callCount != 6 {
		t.Errorf("expected 6 QueryRow calls, got %d", callCount)
	}
}

func TestGetDashboardStats_QueryError(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					return errors.New("db error")
				},
			}
		},
	}
	repo := NewAdminRepository(pool)

	_, err := repo.GetStats(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetUserByID(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	pool := &mockPool{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					*dest[0].(*string) = "user-1"
					*dest[1].(*string) = "alice@test.com"
					*dest[2].(*string) = "Alice Smith"
					*dest[3].(*string) = "+1234567890"
					*dest[4].(*bool) = true
					*dest[5].(*string) = "approved"
					*dest[6].(*bool) = false
					*dest[7].(*string) = "user"
					*dest[8].(*time.Time) = now
					*dest[9].(*time.Time) = now.Add(24 * time.Hour)
					return nil
				},
			}
		},
	}
	repo := NewAdminRepository(pool)

	user, err := repo.GetUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.ID != "user-1" {
		t.Errorf("ID = %q, want user-1", user.ID)
	}
	if user.Email != "alice@test.com" {
		t.Errorf("Email = %q, want alice@test.com", user.Email)
	}
	if user.FullName != "Alice Smith" {
		t.Errorf("FullName = %q, want Alice Smith", user.FullName)
	}
	if user.Phone != "+1234567890" {
		t.Errorf("Phone = %q, want +1234567890", user.Phone)
	}
	if !user.EmailVerified {
		t.Error("EmailVerified = false, want true")
	}
	if user.KYCStatus != "approved" {
		t.Errorf("KYCStatus = %q, want approved", user.KYCStatus)
	}
	if user.MFaEnabled {
		t.Error("MFaEnabled = true, want false")
	}
	if user.Role != "user" {
		t.Errorf("Role = %q, want user", user.Role)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}
	repo := NewAdminRepository(pool)

	user, err := repo.GetUser(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user != nil {
		t.Fatal("expected nil user, got non-nil")
	}
}

func TestListKycSubmissions(t *testing.T) {
	now := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	rowsReturned := 0
	pool := &mockPool{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					*dest[0].(*int) = 2
					return nil
				},
			}
		},
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return &mockRows{
				nextFn: func() bool {
					if rowsReturned < 2 {
						rowsReturned++
						return true
					}
					return false
				},
				scanFn: func(dest ...any) error {
					*dest[0].(*string) = "kyc-1"
					*dest[1].(*string) = "user-1"
					*dest[2].(*string) = "Bob Johnson"
					*dest[3].(*string) = "passport"
					*dest[4].(*string) = "AB123456"
					*dest[5].(*string) = "pending"
					*dest[6].(*string) = ""
					*dest[7].(*time.Time) = now
					*(dest[8].(**time.Time)) = &now
					return nil
				},
				closeFn: func() {},
			}, nil
		},
	}
	repo := NewAdminRepository(pool)

	subs, total, err := repo.ListKYCSubmissions(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(subs) != 2 {
		t.Errorf("len(subs) = %d, want 2", len(subs))
	}
	if len(subs) > 0 && subs[0].ID != "kyc-1" {
		t.Errorf("subs[0].ID = %q, want kyc-1", subs[0].ID)
	}
}

func TestListKycSubmissions_QueryError(t *testing.T) {
	pool := &mockPool{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					return nil
				},
			}
		},
		queryFn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return nil, errors.New("query error")
		},
	}
	repo := NewAdminRepository(pool)

	_, _, err := repo.ListKYCSubmissions(context.Background(), 0, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateUserStatus(t *testing.T) {
	pool := &mockPool{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}
	repo := NewAdminRepository(pool)

	err := repo.UpdateUserStatus(context.Background(), "user-1", "approved")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReviewKYCSubmission(t *testing.T) {
	pool := &mockPool{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}
	repo := NewAdminRepository(pool)

	err := repo.ReviewKYCSubmission(context.Background(), "kyc-1", "approved", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
