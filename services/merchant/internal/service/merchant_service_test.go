package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/merchant/internal/repository"
)

type mockMerchantRepo struct {
	createFn func(ctx context.Context, m *repository.Merchant) error
	getByIDFn func(ctx context.Context, id uuid.UUID) (*repository.Merchant, error)
	updateFn func(ctx context.Context, m *repository.Merchant) error
	deleteFn func(ctx context.Context, id uuid.UUID) error
	searchFn func(ctx context.Context, query string, limit, offset int) ([]*repository.Merchant, error)
	listFn   func(ctx context.Context, limit, offset int) ([]*repository.Merchant, error)
}

func (m *mockMerchantRepo) Create(ctx context.Context, mr *repository.Merchant) error {
	return m.createFn(ctx, mr)
}
func (m *mockMerchantRepo) GetByID(ctx context.Context, id uuid.UUID) (*repository.Merchant, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockMerchantRepo) Update(ctx context.Context, mr *repository.Merchant) error {
	return m.updateFn(ctx, mr)
}
func (m *mockMerchantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFn(ctx, id)
}
func (m *mockMerchantRepo) Search(ctx context.Context, query string, limit, offset int) ([]*repository.Merchant, error) {
	return m.searchFn(ctx, query, limit, offset)
}
func (m *mockMerchantRepo) List(ctx context.Context, limit, offset int) ([]*repository.Merchant, error) {
	return m.listFn(ctx, limit, offset)
}

func TestCreateMerchant_Success(t *testing.T) {
	merchantID := uuid.New()
	now := time.Now()

	mock := &mockMerchantRepo{
		createFn: func(ctx context.Context, m *repository.Merchant) error {
			m.ID = merchantID
			m.CreatedAt = now
			m.UpdatedAt = now
			return nil
		},
	}

	svc := NewMerchantService(mock)
	resp, err := svc.CreateMerchant(context.Background(), CreateMerchantRequest{
		Name:  "Test Merchant",
		Email: "test@example.com",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != "Test Merchant" {
		t.Errorf("expected Test Merchant, got %s", resp.Name)
	}
	if resp.Status != "active" {
		t.Errorf("expected active, got %s", resp.Status)
	}
	if resp.FeePercentage != 0.50 {
		t.Errorf("expected 0.50, got %f", resp.FeePercentage)
	}
}

func TestCreateMerchant_InvalidName(t *testing.T) {
	svc := NewMerchantService(&mockMerchantRepo{})
	_, err := svc.CreateMerchant(context.Background(), CreateMerchantRequest{
		Email: "test@example.com",
	})
	if err != ErrInvalidName {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestCreateMerchant_InvalidEmail(t *testing.T) {
	svc := NewMerchantService(&mockMerchantRepo{})
	_, err := svc.CreateMerchant(context.Background(), CreateMerchantRequest{
		Name: "Test Merchant",
	})
	if err != ErrInvalidEmail {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestGetMerchant_NotFound(t *testing.T) {
	mock := &mockMerchantRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Merchant, error) {
			return nil, nil
		},
	}

	svc := NewMerchantService(mock)
	_, err := svc.GetMerchant(context.Background(), uuid.New())
	if err != ErrMerchantNotFound {
		t.Errorf("expected ErrMerchantNotFound, got %v", err)
	}
}

func TestGetMerchant_Success(t *testing.T) {
	merchantID := uuid.New()

	mock := &mockMerchantRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Merchant, error) {
			return &repository.Merchant{
				ID:     merchantID,
				Name:   "Test Merchant",
				Email:  "test@example.com",
				Status: "active",
			}, nil
		},
	}

	svc := NewMerchantService(mock)
	resp, err := svc.GetMerchant(context.Background(), merchantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != merchantID {
		t.Errorf("expected %v, got %v", merchantID, resp.ID)
	}
	if resp.Status != "active" {
		t.Errorf("expected active, got %s", resp.Status)
	}
}

func TestUpdateMerchant_Success(t *testing.T) {
	merchantID := uuid.New()

	mock := &mockMerchantRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Merchant, error) {
			return &repository.Merchant{
				ID:     merchantID,
				Name:   "Old Name",
				Email:  "old@example.com",
				Status: "active",
			}, nil
		},
		updateFn: func(ctx context.Context, m *repository.Merchant) error {
			return nil
		},
	}

	newName := "New Name"
	svc := NewMerchantService(mock)
	resp, err := svc.UpdateMerchant(context.Background(), merchantID, UpdateMerchantRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != "New Name" {
		t.Errorf("expected New Name, got %s", resp.Name)
	}
}

func TestUpdateMerchant_NotFound(t *testing.T) {
	mock := &mockMerchantRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Merchant, error) {
			return nil, nil
		},
	}

	svc := NewMerchantService(mock)
	_, err := svc.UpdateMerchant(context.Background(), uuid.New(), UpdateMerchantRequest{})
	if err != ErrMerchantNotFound {
		t.Errorf("expected ErrMerchantNotFound, got %v", err)
	}
}

func TestDeleteMerchant_Success(t *testing.T) {
	merchantID := uuid.New()

	mock := &mockMerchantRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Merchant, error) {
			return &repository.Merchant{
				ID:     merchantID,
				Name:   "Test",
				Email:  "test@example.com",
				Status: "active",
			}, nil
		},
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	svc := NewMerchantService(mock)
	err := svc.DeleteMerchant(context.Background(), merchantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteMerchant_NotFound(t *testing.T) {
	mock := &mockMerchantRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Merchant, error) {
			return nil, nil
		},
	}

	svc := NewMerchantService(mock)
	err := svc.DeleteMerchant(context.Background(), uuid.New())
	if err != ErrMerchantNotFound {
		t.Errorf("expected ErrMerchantNotFound, got %v", err)
	}
}

func TestListMerchants_DefaultPagination(t *testing.T) {
	mock := &mockMerchantRepo{
		listFn: func(ctx context.Context, limit, offset int) ([]*repository.Merchant, error) {
			if limit != 20 || offset != 0 {
				t.Errorf("expected limit=20, offset=0, got limit=%d offset=%d", limit, offset)
			}
			return []*repository.Merchant{}, nil
		},
	}

	svc := NewMerchantService(mock)
	resp, err := svc.ListMerchants(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list")
	}
}

func TestSearchMerchants_Success(t *testing.T) {
	merchantID := uuid.New()

	mock := &mockMerchantRepo{
		searchFn: func(ctx context.Context, query string, limit, offset int) ([]*repository.Merchant, error) {
			if query != "test" {
				t.Errorf("expected test, got %s", query)
			}
			return []*repository.Merchant{
				{ID: merchantID, Name: "Test Merchant", Email: "test@example.com", Status: "active"},
			}, nil
		},
	}

	svc := NewMerchantService(mock)
	resp, err := svc.SearchMerchants(context.Background(), "test", 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp))
	}
	if resp[0].Name != "Test Merchant" {
		t.Errorf("expected Test Merchant, got %s", resp[0].Name)
	}
}
