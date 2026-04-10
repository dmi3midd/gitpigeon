package domain

import "context"

type Subscription struct {
	ID               int    `json:"id" db:"id"`
	RepositoryID     int    `json:"repository_id" db:"repository_id"`
	Email            string `json:"email" db:"email"`
	Confirmed        bool   `json:"confirmed" db:"confirmed"`
	ConfirmToken     string `json:"-" db:"confirm_token"`
	UnsubscribeToken string `json:"-" db:"unsubscribe_token"`
	CreatedAt        string `json:"created_at" db:"created_at"`
}

// SubscriptionRepo defines the interface for subscription persistence operations.
type SubscriptionRepo interface {
	// Create inserts a new subscription and returns it with the generated ID.
	Create(ctx context.Context, sub *Subscription) (*Subscription, error)
	// GetByID returns a subscription by its ID.
	GetByID(ctx context.Context, id int) (*Subscription, error)
	// GetByConfirmToken returns a subscription by its confirmation token.
	GetByConfirmToken(ctx context.Context, token string) (*Subscription, error)
	// GetByUnsubscribeToken returns a subscription by its unsubscribe token.
	GetByUnsubscribeToken(ctx context.Context, token string) (*Subscription, error)
	// ListByRepositoryID returns all confirmed subscriptions for a given repository.
	ListByRepositoryID(ctx context.Context, repositoryID int) ([]Subscription, error)
	// ListConfirmedByRepositoryID returns all confirmed subscriptions for a given repository.
	ListConfirmedByRepositoryID(ctx context.Context, repositoryID int) ([]Subscription, error)
	// ListByEmail returns all confirmed subscriptions for a given email.
	ListByEmail(ctx context.Context, email string) ([]Subscription, error)
	// Confirm marks a subscription as confirmed.
	Confirm(ctx context.Context, id int) error
	// Delete removes a subscription by its ID.
	Delete(ctx context.Context, id int) error
	// DeleteByRepositoryIDAndEmail removes a subscription by repository ID and email.
	DeleteByRepositoryIDAndEmail(ctx context.Context, repositoryID int, email string) error
}
