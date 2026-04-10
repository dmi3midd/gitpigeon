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

// SubscribeResult is the response returned after a successful subscription.
type SubscribeResult struct {
	Message string `json:"message"`
}

// SubscriptionInfo is the response DTO for listing subscriptions.
type SubscriptionInfo struct {
	ID        int    `json:"id"`
	Repo      string `json:"repo"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// SubscriptionService contains the business logic for subscription management.
type SubscriptionService interface {
	// Subscribe creates a new subscription for the given email and repository.
	// Flow:
	//  1. Validate owner/repo format
	//  2. Verify repository exists on GitHub (404 / 429 handling)
	//  3. Find or create repository record in DB
	//  4. Generate confirmation and unsubscribe tokens
	//  5. Create subscription (confirmed=false)
	//  6. Send confirmation email
	Subscribe(ctx context.Context, email, repo string) (*Subscription, error)
	// Confirm confirms a subscription by its confirmation token.
	Confirm(ctx context.Context, token string) error
	// Unsubscribe removes a subscription by its unsubscribe token.
	Unsubscribe(ctx context.Context, token string) error
	// GetSubscriptions returns all active (confirmed) subscriptions for the given email.
	GetSubscriptions(ctx context.Context, email string) ([]Subscription, error)
}
