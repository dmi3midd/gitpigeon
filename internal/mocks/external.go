package mocks

import (
	"context"

	githubapi "gitpigeon/internal/github"
	"gitpigeon/internal/notifier"
)

// MockGitHubClient is a mock implementation of githubapi.Client.
type MockGitHubClient struct {
	GetRepositoryFn    func(ctx context.Context, owner, repo string) (*githubapi.Repository, error)
	GetLatestReleaseFn func(ctx context.Context, owner, repo string) (*githubapi.Release, error)
}

func (m *MockGitHubClient) GetRepository(ctx context.Context, owner, repo string) (*githubapi.Repository, error) {
	return m.GetRepositoryFn(ctx, owner, repo)
}
func (m *MockGitHubClient) GetLatestRelease(ctx context.Context, owner, repo string) (*githubapi.Release, error) {
	return m.GetLatestReleaseFn(ctx, owner, repo)
}

// MockNotifier is a mock implementation of notifier.Notifier.
type MockNotifier struct {
	NotifyFn func(msg *notifier.Notification, to string) error
	Calls    []NotifyCall
}

type NotifyCall struct {
	Msg *notifier.Notification
	To  string
}

func (m *MockNotifier) Notify(msg *notifier.Notification, to string) error {
	m.Calls = append(m.Calls, NotifyCall{Msg: msg, To: to})
	if m.NotifyFn != nil {
		return m.NotifyFn(msg, to)
	}
	return nil
}
