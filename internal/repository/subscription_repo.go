package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"subscription-service/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("subscription not found")

type SubscriptionRepo struct {
	db *sqlx.DB
}

func NewSubscriptionRepo(db *sqlx.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	query := `
        INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := r.db.ExecContext(ctx, query, sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate)
	return err
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	var sub domain.Subscription
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at 
              FROM subscriptions WHERE id = $1`
	err := r.db.GetContext(ctx, &sub, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &sub, err
}

// Update, Delete, List аналогично (List с пагинацией)
// ...

func (r *SubscriptionRepo) SumCostInPeriod(ctx context.Context, from, to time.Time, userID *uuid.UUID, serviceName *string) (int, error) {
	query := `
        SELECT COALESCE(SUM(price), 0)
        FROM subscriptions
        WHERE start_date <= $1
          AND (end_date IS NULL OR end_date >= $2)
    `
	args := []interface{}{to, from} // to - конец периода (последний день), from - начало
	argPos := 3
	if userID != nil {
		query += " AND user_id = $" + strconv.Itoa(argPos)
		args = append(args, *userID)
		argPos++
	}
	if serviceName != nil {
		query += " AND service_name = $" + strconv.Itoa(argPos)
		args = append(args, *serviceName)
	}
	var sum int
	err := r.db.GetContext(ctx, &sum, query, args...)
	return sum, err
}

func (r *SubscriptionRepo) Update(ctx context.Context, sub *domain.Subscription) error {
	query := `
		UPDATE subscriptions
		SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.ExecContext(ctx, query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.ID)
	return err
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *SubscriptionRepo) List(ctx context.Context, limit, offset int) ([]domain.Subscription, error) {
	var subs []domain.Subscription
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	err := r.db.SelectContext(ctx, &subs, query, limit, offset)
	return subs, err
}
