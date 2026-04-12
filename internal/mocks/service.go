package mocks

import (
	"context"

	"gitpigeon/internal/domain"
)

// MockSubscriptionService is a mock implementation of domain.SubscriptionService.
type MockSubscriptionService struct {
	SubscribeFn        func(ctx context.Context, email, repo string) (*domain.SubscribeResult, error)
	ConfirmFn          func(ctx context.Context, token string) error
	UnsubscribeFn      func(ctx context.Context, token string) error
	GetSubscriptionsFn func(ctx context.Context, email string) ([]domain.SubscriptionInfo, error)
}

func (m *MockSubscriptionService) Subscribe(ctx context.Context, email, repo string) (*domain.SubscribeResult, error) {
	return m.SubscribeFn(ctx, email, repo)
}
func (m *MockSubscriptionService) Confirm(ctx context.Context, token string) error {
	return m.ConfirmFn(ctx, token)
}
func (m *MockSubscriptionService) Unsubscribe(ctx context.Context, token string) error {
	return m.UnsubscribeFn(ctx, token)
}
func (m *MockSubscriptionService) GetSubscriptions(ctx context.Context, email string) ([]domain.SubscriptionInfo, error) {
	return m.GetSubscriptionsFn(ctx, email)
}
