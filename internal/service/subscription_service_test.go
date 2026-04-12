package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gitpigeon/internal/domain"
	githubapi "gitpigeon/internal/github"
	"gitpigeon/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(
	subRepo *mocks.MockSubscriptionRepo,
	repoRepo *mocks.MockRepositoryRepo,
	ghClient *mocks.MockGitHubClient,
	notif *mocks.MockNotifier,
) *SubscriptionService {
	return NewSubscriptionService(subRepo, repoRepo, ghClient, notif, "http://localhost:8080")
}

// --- Subscribe tests ---

func TestSubscribe_Success(t *testing.T) {
	subRepo := &mocks.MockSubscriptionRepo{
		CreateFn: func(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
			sub.ID = 1
			sub.CreatedAt = "2026-04-12T00:00:00Z"
			return sub, nil
		},
	}
	repoRepo := &mocks.MockRepositoryRepo{
		GetByOwnerAndNameFn: func(ctx context.Context, owner, name string) (*domain.Repository, error) {
			return &domain.Repository{ID: 1, Owner: owner, Name: name}, nil
		},
	}
	ghClient := &mocks.MockGitHubClient{
		GetRepositoryFn: func(ctx context.Context, owner, repo string) (*githubapi.Repository, error) {
			return &githubapi.Repository{Owner: owner, Name: repo}, nil
		},
	}
	notif := &mocks.MockNotifier{}

	svc := newTestService(subRepo, repoRepo, ghClient, notif)
	result, err := svc.Subscribe(context.Background(), "user@example.com", "golang/go")

	require.NoError(t, err)
	assert.Contains(t, result.Message, "Confirmation email sent")

	// Give the async goroutine time to run
	time.Sleep(100 * time.Millisecond)
	calls := notif.GetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "user@example.com", calls[0].To)
}

func TestSubscribe_CreatesRepoInDBWhenNotExists(t *testing.T) {
	var createdRepo *domain.Repository
	subRepo := &mocks.MockSubscriptionRepo{
		CreateFn: func(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
			sub.ID = 1
			return sub, nil
		},
	}
	repoRepo := &mocks.MockRepositoryRepo{
		GetByOwnerAndNameFn: func(ctx context.Context, owner, name string) (*domain.Repository, error) {
			return nil, errors.New("not found")
		},
		CreateFn: func(ctx context.Context, repo *domain.Repository) (*domain.Repository, error) {
			createdRepo = repo
			repo.ID = 42
			return repo, nil
		},
	}
	ghClient := &mocks.MockGitHubClient{
		GetRepositoryFn: func(ctx context.Context, owner, repo string) (*githubapi.Repository, error) {
			return &githubapi.Repository{Owner: owner, Name: repo}, nil
		},
	}
	notif := &mocks.MockNotifier{}

	svc := newTestService(subRepo, repoRepo, ghClient, notif)
	result, err := svc.Subscribe(context.Background(), "user@example.com", "golang/go")

	require.NoError(t, err)
	assert.NotNil(t, result)
	require.NotNil(t, createdRepo)
	assert.Equal(t, "golang", createdRepo.Owner)
	assert.Equal(t, "go", createdRepo.Name)
}

func TestSubscribe_InvalidRepoFormat(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil)

	tests := []struct {
		name string
		repo string
	}{
		{"empty", ""},
		{"no slash", "golanggo"},
		{"trailing slash", "golang/"},
		{"leading slash", "/go"},
		{"extra slashes", "golang/go/extra"},
		{"whitespace", "golang/ go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Subscribe(context.Background(), "user@example.com", tt.repo)
			assert.ErrorIs(t, err, ErrInvalidRepoFormat)
		})
	}
}

func TestSubscribe_InvalidEmail(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil)

	tests := []struct {
		name  string
		email string
	}{
		{"empty", ""},
		{"no at", "userexample.com"},
		{"no domain", "user@"},
		{"no dot in domain", "user@example"},
		{"too short", "a@"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Subscribe(context.Background(), tt.email, "golang/go")
			assert.ErrorIs(t, err, ErrInvalidEmail)
		})
	}
}

func TestSubscribe_RepoNotFoundOnGitHub(t *testing.T) {
	ghClient := &mocks.MockGitHubClient{
		GetRepositoryFn: func(ctx context.Context, owner, repo string) (*githubapi.Repository, error) {
			return nil, githubapi.ErrRepositoryNotFound
		},
	}
	svc := newTestService(nil, nil, ghClient, nil)

	_, err := svc.Subscribe(context.Background(), "user@example.com", "nonexistent/repo")
	assert.ErrorIs(t, err, ErrRepoNotFound)
}

func TestSubscribe_RateLimitExceeded(t *testing.T) {
	ghClient := &mocks.MockGitHubClient{
		GetRepositoryFn: func(ctx context.Context, owner, repo string) (*githubapi.Repository, error) {
			return nil, githubapi.ErrRateLimitExceeded
		},
	}
	svc := newTestService(nil, nil, ghClient, nil)

	_, err := svc.Subscribe(context.Background(), "user@example.com", "golang/go")
	assert.ErrorIs(t, err, ErrRateLimitExceeded)
}

func TestSubscribe_DuplicateSubscription(t *testing.T) {
	subRepo := &mocks.MockSubscriptionRepo{
		CreateFn: func(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
			return nil, errors.New("subscription create: UNIQUE constraint failed: subscriptions.repository_id, subscriptions.email")
		},
	}
	repoRepo := &mocks.MockRepositoryRepo{
		GetByOwnerAndNameFn: func(ctx context.Context, owner, name string) (*domain.Repository, error) {
			return &domain.Repository{ID: 1, Owner: owner, Name: name}, nil
		},
	}
	ghClient := &mocks.MockGitHubClient{
		GetRepositoryFn: func(ctx context.Context, owner, repo string) (*githubapi.Repository, error) {
			return &githubapi.Repository{Owner: owner, Name: repo}, nil
		},
	}
	svc := newTestService(subRepo, repoRepo, ghClient, nil)

	_, err := svc.Subscribe(context.Background(), "user@example.com", "golang/go")
	assert.ErrorIs(t, err, ErrSubscriptionExists)
}

// --- Confirm tests ---

func TestConfirm_Success(t *testing.T) {
	confirmed := false
	subRepo := &mocks.MockSubscriptionRepo{
		GetByConfirmTokenFn: func(ctx context.Context, token string) (*domain.Subscription, error) {
			return &domain.Subscription{ID: 1, Confirmed: false}, nil
		},
		ConfirmFn: func(ctx context.Context, id int) error {
			confirmed = true
			return nil
		},
	}
	svc := newTestService(subRepo, nil, nil, nil)

	err := svc.Confirm(context.Background(), "valid-token")
	require.NoError(t, err)
	assert.True(t, confirmed)
}

func TestConfirm_InvalidToken(t *testing.T) {
	subRepo := &mocks.MockSubscriptionRepo{
		GetByConfirmTokenFn: func(ctx context.Context, token string) (*domain.Subscription, error) {
			return nil, errors.New("not found")
		},
	}
	svc := newTestService(subRepo, nil, nil, nil)

	err := svc.Confirm(context.Background(), "invalid-token")
	assert.ErrorIs(t, err, ErrTokenNotFound)
}

func TestConfirm_AlreadyConfirmed(t *testing.T) {
	subRepo := &mocks.MockSubscriptionRepo{
		GetByConfirmTokenFn: func(ctx context.Context, token string) (*domain.Subscription, error) {
			return &domain.Subscription{ID: 1, Confirmed: true}, nil
		},
	}
	svc := newTestService(subRepo, nil, nil, nil)

	err := svc.Confirm(context.Background(), "some-token")
	assert.ErrorIs(t, err, ErrAlreadyConfirmed)
}

// --- Unsubscribe tests ---

func TestUnsubscribe_Success(t *testing.T) {
	deleted := false
	subRepo := &mocks.MockSubscriptionRepo{
		GetByUnsubscribeTokenFn: func(ctx context.Context, token string) (*domain.Subscription, error) {
			return &domain.Subscription{ID: 5}, nil
		},
		DeleteFn: func(ctx context.Context, id int) error {
			assert.Equal(t, 5, id)
			deleted = true
			return nil
		},
	}
	svc := newTestService(subRepo, nil, nil, nil)

	err := svc.Unsubscribe(context.Background(), "valid-unsub-token")
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestUnsubscribe_InvalidToken(t *testing.T) {
	subRepo := &mocks.MockSubscriptionRepo{
		GetByUnsubscribeTokenFn: func(ctx context.Context, token string) (*domain.Subscription, error) {
			return nil, errors.New("not found")
		},
	}
	svc := newTestService(subRepo, nil, nil, nil)

	err := svc.Unsubscribe(context.Background(), "bad-token")
	assert.ErrorIs(t, err, ErrTokenNotFound)
}

// --- GetSubscriptions tests ---

func TestGetSubscriptions_Success(t *testing.T) {
	subRepo := &mocks.MockSubscriptionRepo{
		ListByEmailFn: func(ctx context.Context, email string) ([]domain.Subscription, error) {
			return []domain.Subscription{
				{ID: 1, RepositoryID: 10, Email: email, CreatedAt: "2026-04-12T00:00:00Z"},
				{ID: 2, RepositoryID: 20, Email: email, CreatedAt: "2026-04-12T01:00:00Z"},
			}, nil
		},
	}
	repoRepo := &mocks.MockRepositoryRepo{
		GetByIDFn: func(ctx context.Context, id int) (*domain.Repository, error) {
			repos := map[int]*domain.Repository{
				10: {ID: 10, Owner: "golang", Name: "go"},
				20: {ID: 20, Owner: "docker", Name: "compose"},
			}
			return repos[id], nil
		},
	}
	svc := newTestService(subRepo, repoRepo, nil, nil)

	subs, err := svc.GetSubscriptions(context.Background(), "user@example.com")
	require.NoError(t, err)
	require.Len(t, subs, 2)
	assert.Equal(t, "golang/go", subs[0].Repo)
	assert.Equal(t, "docker/compose", subs[1].Repo)
}

func TestGetSubscriptions_EmptyList(t *testing.T) {
	subRepo := &mocks.MockSubscriptionRepo{
		ListByEmailFn: func(ctx context.Context, email string) ([]domain.Subscription, error) {
			return []domain.Subscription{}, nil
		},
	}
	svc := newTestService(subRepo, nil, nil, nil)

	subs, err := svc.GetSubscriptions(context.Background(), "user@example.com")
	require.NoError(t, err)
	assert.Empty(t, subs)
}

func TestGetSubscriptions_InvalidEmail(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil)

	_, err := svc.GetSubscriptions(context.Background(), "bad-email")
	assert.ErrorIs(t, err, ErrInvalidEmail)
}

// --- Helpers tests ---

func TestParseRepoFullName(t *testing.T) {
	tests := []struct {
		input     string
		owner     string
		repo      string
		expectErr bool
	}{
		{"golang/go", "golang", "go", false},
		{"docker/compose", "docker", "compose", false},
		{"invalid", "", "", true},
		{"", "", "", true},
		{"/repo", "", "", true},
		{"owner/", "", "", true},
		{"a/b/c", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			owner, repo, err := parseRepoFullName(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.owner, owner)
				assert.Equal(t, tt.repo, repo)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"a@b.c", true},
		{"test+tag@gmail.com", true},
		{"", false},
		{"noatsign", false},
		{"@example.com", false},
		{"user@", false},
		{"user@nodot", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			assert.Equal(t, tt.valid, isValidEmail(tt.email))
		})
	}
}
