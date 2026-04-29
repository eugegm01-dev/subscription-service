package service

import (
	"context"
	"errors"
	"subscription-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidDateRange = errors.New("end_date must be after start_date")
)

type SubscriptionService struct {
	repo domain.SubscriptionRepository
}

func NewSubscriptionService(repo domain.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Create(ctx context.Context, serviceName string, price int, userID uuid.UUID, startDate, endDate *time.Time) (*domain.Subscription, error) {
	if endDate != nil && endDate.Before(*startDate) {
		return nil, ErrInvalidDateRange
	}
	sub := &domain.Subscription{
		ID:          uuid.New(),
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   *startDate,
		EndDate:     endDate,
	}
	err := s.repo.Create(ctx, sub)
	return sub, err
}

func (s *SubscriptionService) TotalCost(ctx context.Context, from, to time.Time, userID *uuid.UUID, serviceName *string) (int, error) {
	return s.repo.SumCostInPeriod(ctx, from, to, userID, serviceName)
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) List(ctx context.Context, limit, offset int) ([]domain.Subscription, error) {
	return s.repo.List(ctx, limit, offset)
}
func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, serviceName *string, price *int, startDate, endDate *time.Time) (*domain.Subscription, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if serviceName != nil {
		existing.ServiceName = *serviceName
	}
	if price != nil {
		existing.Price = *price
	}
	if startDate != nil {
		existing.StartDate = *startDate
	}
	if endDate != nil {
		existing.EndDate = endDate
	}
	// Validate date range
	if existing.EndDate != nil && existing.EndDate.Before(existing.StartDate) {
		return nil, ErrInvalidDateRange
	}
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}
