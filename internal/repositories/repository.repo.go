package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitpigeon/internal/domain"

	"github.com/jmoiron/sqlx"
)

type RepositoryRepo struct {
	db *sqlx.DB
}

// NewRepositoryRepo creates a new RepositoryRepo.
func NewRepositoryRepo(db *sqlx.DB) *RepositoryRepo {
	return &RepositoryRepo{db: db}
}

func (r *RepositoryRepo) Create(ctx context.Context, repo *domain.Repository) (*domain.Repository, error) {
	query := `INSERT INTO repositories (owner, name, last_seen_tag) VALUES (:owner, :name, :last_seen_tag) RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, repo)
	if err != nil {
		return nil, fmt.Errorf("repository create: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		if err := rows.StructScan(repo); err != nil {
			return nil, fmt.Errorf("repository create scan: %w", err)
		}
	}

	return repo, nil
}

func (r *RepositoryRepo) GetByID(ctx context.Context, id int) (*domain.Repository, error) {
	query := `SELECT id, owner, name, last_seen_tag, created_at, updated_at FROM repositories WHERE id = ?`

	var repo domain.Repository
	if err := r.db.GetContext(ctx, &repo, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("repository not found: %w", err)
		}
		return nil, fmt.Errorf("repository get by id: %w", err)
	}

	return &repo, nil
}

func (r *RepositoryRepo) GetByOwnerAndName(ctx context.Context, owner, name string) (*domain.Repository, error) {
	query := `SELECT id, owner, name, last_seen_tag, created_at, updated_at FROM repositories WHERE owner = ? AND name = ?`

	var repo domain.Repository
	if err := r.db.GetContext(ctx, &repo, query, owner, name); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("repository not found: %w", err)
		}
		return nil, fmt.Errorf("repository get by owner and name: %w", err)
	}

	return &repo, nil
}

// ListAll returns all repositories from the database.
func (r *RepositoryRepo) ListAll(ctx context.Context) ([]domain.Repository, error) {
	query := `SELECT id, owner, name, last_seen_tag, created_at, updated_at FROM repositories ORDER BY id`

	var repos []domain.Repository
	if err := r.db.SelectContext(ctx, &repos, query); err != nil {
		return nil, fmt.Errorf("repository list all: %w", err)
	}

	return repos, nil
}

func (r *RepositoryRepo) UpdateLastSeenTag(ctx context.Context, id int, tag string) error {
	query := `UPDATE repositories SET last_seen_tag = ?, updated_at = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, tag, time.Now().UTC().Format(time.DateTime), id)
	if err != nil {
		return fmt.Errorf("repository update last seen tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository update last seen tag rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("repository not found")
	}

	return nil
}

func (r *RepositoryRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM repositories WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository delete: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository delete rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("repository not found")
	}

	return nil
}
