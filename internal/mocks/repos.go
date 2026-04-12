package mocks

import (
	"context"

	"gitpigeon/internal/domain"
)

// MockSubscriptionRepo is a mock implementation of domain.SubscriptionRepo.
type MockSubscriptionRepo struct {
	CreateFn                       func(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error)
	GetByIDFn                      func(ctx context.Context, id int) (*domain.Subscription, error)
	GetByConfirmTokenFn            func(ctx context.Context, token string) (*domain.Subscription, error)
	GetByUnsubscribeTokenFn        func(ctx context.Context, token string) (*domain.Subscription, error)
	ListByRepositoryIDFn           func(ctx context.Context, repositoryID int) ([]domain.Subscription, error)
	ListConfirmedByRepositoryIDFn  func(ctx context.Context, repositoryID int) ([]domain.Subscription, error)
	ListByEmailFn                  func(ctx context.Context, email string) ([]domain.Subscription, error)
	ConfirmFn                      func(ctx context.Context, id int) error
	DeleteFn                       func(ctx context.Context, id int) error
	DeleteByRepositoryIDAndEmailFn func(ctx context.Context, repositoryID int, email string) error
}

func (m *MockSubscriptionRepo) Create(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
	return m.CreateFn(ctx, sub)
}
func (m *MockSubscriptionRepo) GetByID(ctx context.Context, id int) (*domain.Subscription, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *MockSubscriptionRepo) GetByConfirmToken(ctx context.Context, token string) (*domain.Subscription, error) {
	return m.GetByConfirmTokenFn(ctx, token)
}
func (m *MockSubscriptionRepo) GetByUnsubscribeToken(ctx context.Context, token string) (*domain.Subscription, error) {
	return m.GetByUnsubscribeTokenFn(ctx, token)
}
func (m *MockSubscriptionRepo) ListByRepositoryID(ctx context.Context, repositoryID int) ([]domain.Subscription, error) {
	return m.ListByRepositoryIDFn(ctx, repositoryID)
}
func (m *MockSubscriptionRepo) ListConfirmedByRepositoryID(ctx context.Context, repositoryID int) ([]domain.Subscription, error) {
	return m.ListConfirmedByRepositoryIDFn(ctx, repositoryID)
}
func (m *MockSubscriptionRepo) ListByEmail(ctx context.Context, email string) ([]domain.Subscription, error) {
	return m.ListByEmailFn(ctx, email)
}
func (m *MockSubscriptionRepo) Confirm(ctx context.Context, id int) error {
	return m.ConfirmFn(ctx, id)
}
func (m *MockSubscriptionRepo) Delete(ctx context.Context, id int) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockSubscriptionRepo) DeleteByRepositoryIDAndEmail(ctx context.Context, repositoryID int, email string) error {
	return m.DeleteByRepositoryIDAndEmailFn(ctx, repositoryID, email)
}

// MockRepositoryRepo is a mock implementation of domain.RepositoryRepo.
type MockRepositoryRepo struct {
	CreateFn            func(ctx context.Context, repo *domain.Repository) (*domain.Repository, error)
	GetByIDFn           func(ctx context.Context, id int) (*domain.Repository, error)
	GetByOwnerAndNameFn func(ctx context.Context, owner, name string) (*domain.Repository, error)
	ListAllFn           func(ctx context.Context) ([]domain.Repository, error)
	UpdateLastSeenTagFn func(ctx context.Context, id int, tag string) error
	DeleteFn            func(ctx context.Context, id int) error
}

func (m *MockRepositoryRepo) Create(ctx context.Context, repo *domain.Repository) (*domain.Repository, error) {
	return m.CreateFn(ctx, repo)
}
func (m *MockRepositoryRepo) GetByID(ctx context.Context, id int) (*domain.Repository, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *MockRepositoryRepo) GetByOwnerAndName(ctx context.Context, owner, name string) (*domain.Repository, error) {
	return m.GetByOwnerAndNameFn(ctx, owner, name)
}
func (m *MockRepositoryRepo) ListAll(ctx context.Context) ([]domain.Repository, error) {
	return m.ListAllFn(ctx)
}
func (m *MockRepositoryRepo) UpdateLastSeenTag(ctx context.Context, id int, tag string) error {
	return m.UpdateLastSeenTagFn(ctx, id, tag)
}
func (m *MockRepositoryRepo) Delete(ctx context.Context, id int) error {
	return m.DeleteFn(ctx, id)
}
