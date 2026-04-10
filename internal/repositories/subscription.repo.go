package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"gitpigeon/internal/domain"

	"github.com/jmoiron/sqlx"
)

type SubscriptionRepo struct {
	db *sqlx.DB
}

// NewSubscriptionRepo creates a new SubscriptionRepo.
func NewSubscriptionRepo(db *sqlx.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
	query := `INSERT INTO subscriptions (repository_id, email) VALUES (:repository_id, :email) RETURNING id, created_at`

	rows, err := r.db.NamedQueryContext(ctx, query, sub)
	if err != nil {
		return nil, fmt.Errorf("subscription create: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(sub); err != nil {
			return nil, fmt.Errorf("subscription create scan: %w", err)
		}
	}

	return sub, nil
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id int) (*domain.Subscription, error) {
	query := `SELECT id, repository_id, email, created_at FROM subscriptions WHERE id = ?`

	var sub domain.Subscription
	if err := r.db.GetContext(ctx, &sub, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription not found: %w", err)
		}
		return nil, fmt.Errorf("subscription get by id: %w", err)
	}

	return &sub, nil
}

func (r *SubscriptionRepo) ListByRepositoryID(ctx context.Context, repositoryID int) ([]domain.Subscription, error) {
	query := `SELECT id, repository_id, email, created_at FROM subscriptions WHERE repository_id = ? ORDER BY id`

	var subs []domain.Subscription
	if err := r.db.SelectContext(ctx, &subs, query, repositoryID); err != nil {
		return nil, fmt.Errorf("subscription list by repository id: %w", err)
	}

	return subs, nil
}

func (r *SubscriptionRepo) ListByEmail(ctx context.Context, email string) ([]domain.Subscription, error) {
	query := `SELECT id, repository_id, email, created_at FROM subscriptions WHERE email = ? ORDER BY id`

	var subs []domain.Subscription
	if err := r.db.SelectContext(ctx, &subs, query, email); err != nil {
		return nil, fmt.Errorf("subscription list by email: %w", err)
	}

	return subs, nil
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM subscriptions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("subscription delete: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("subscription delete rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

func (r *SubscriptionRepo) DeleteByRepositoryIDAndEmail(ctx context.Context, repositoryID int, email string) error {
	query := `DELETE FROM subscriptions WHERE repository_id = ? AND email = ?`

	result, err := r.db.ExecContext(ctx, query, repositoryID, email)
	if err != nil {
		return fmt.Errorf("subscription delete by repo and email: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("subscription delete rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

