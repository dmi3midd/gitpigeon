package mocks

import (
	"context"
	"sync"

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
// Thread-safe: Calls are protected by a mutex for use with async goroutines.
type MockNotifier struct {
	NotifyFn func(msg *notifier.Notification, to string) error

	mu    sync.Mutex
	Calls []NotifyCall
}

type NotifyCall struct {
	Msg *notifier.Notification
	To  string
}

func (m *MockNotifier) Notify(msg *notifier.Notification, to string) error {
	m.mu.Lock()
	m.Calls = append(m.Calls, NotifyCall{Msg: msg, To: to})
	m.mu.Unlock()
	if m.NotifyFn != nil {
		return m.NotifyFn(msg, to)
	}
	return nil
}

// GetCalls returns a thread-safe copy of all recorded calls.
func (m *MockNotifier) GetCalls() []NotifyCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]NotifyCall, len(m.Calls))
	copy(result, m.Calls)
	return result
}
