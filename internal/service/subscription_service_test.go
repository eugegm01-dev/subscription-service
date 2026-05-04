package service

import (
	"context"
	"subscription-service/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *MockRepo) Update(ctx context.Context, sub *domain.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) List(ctx context.Context, limit, offset int) ([]domain.Subscription, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]domain.Subscription), args.Error(1)
}

func (m *MockRepo) SumCostInPeriod(ctx context.Context, from, to time.Time, userID *uuid.UUID, serviceName *string) (int, error) {
	args := m.Called(ctx, from, to, userID, serviceName)
	return args.Int(0), args.Error(1)
}

func TestCreate_Valid(t *testing.T) {
	repo := new(MockRepo)
	svc := NewSubscriptionService(repo)

	now := time.Now()
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)

	sub, err := svc.Create(context.Background(), "Netflix", 500, uuid.New(), &now, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Netflix", sub.ServiceName)
	repo.AssertExpectations(t)
}
