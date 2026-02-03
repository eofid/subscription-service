package repository

import (
	"context"
	"database/sql"
	"subscription-service/internal/domain"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, sub domain.Subscription) (int, error) {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var endDate *time.Time
	if sub.EndDate != nil {
		t := sub.EndDate.Time()
		endDate = &t
	}

	var id int
	err := r.db.QueryRowContext(ctx, query,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate.Time(), endDate,
	).Scan(&id)

	return id, err
}

func (r *Repository) GetAllByUserID(ctx context.Context, userID string) ([]domain.Subscription, error) {
	query := `SELECT id, service_name, price, start_date, end_date FROM subscriptions WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		var start time.Time
		var end *time.Time

		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &start, &end); err != nil {
			return nil, err
		}

		s.UserID = userID
		s.StartDate = domain.CustomDate(start)
		if end != nil {
			d := domain.CustomDate(*end)
			s.EndDate = &d
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *Repository) Update(ctx context.Context, sub domain.Subscription) error {
	query := `
		UPDATE subscriptions 
		SET service_name=$1, price=$2, start_date=$3, end_date=$4
		WHERE id=$5`

	var endDate *time.Time
	if sub.EndDate != nil {
		t := sub.EndDate.Time()
		endDate = &t
	}

	res, err := r.db.ExecContext(ctx, query,
		sub.ServiceName, sub.Price, sub.StartDate.Time(), endDate, sub.ID,
	)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM subscriptions WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}
