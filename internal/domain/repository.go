package domain

import "context"

type Repository struct {
	ID          int    `json:"id" db:"id"`
	Owner       string `json:"owner" db:"owner"`
	Name        string `json:"name" db:"name"`
	LastSeenTag string `json:"last_seen_tag" db:"last_seen_tag"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

// RepositoryRepo defines the interface for repository persistence operations.
type RepositoryRepo interface {
	// Create inserts a new repository and returns it with the generated ID.
	Create(ctx context.Context, repo *Repository) (*Repository, error)
	// GetByID returns a repository by its ID.
	GetByID(ctx context.Context, id int) (*Repository, error)
	// GetByOwnerAndName returns a repository by owner and name.
	GetByOwnerAndName(ctx context.Context, owner, name string) (*Repository, error)
	// ListAll returns all repositories.
	ListAll(ctx context.Context) ([]Repository, error)
	// UpdateLastSeenTag updates the last_seen_tag and updated_at for a repository.
	UpdateLastSeenTag(ctx context.Context, id int, tag string) error
	// Delete removes a repository by its ID.
	Delete(ctx context.Context, id int) error
}
