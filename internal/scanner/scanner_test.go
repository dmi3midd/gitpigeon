package scanner

import (
	"context"
	"errors"
	"testing"
	"time"

	"gitpigeon/internal/domain"
	githubapi "gitpigeon/internal/github"
	"gitpigeon/internal/mocks"
	"gitpigeon/internal/notifier"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestScanner(
	repoRepo *mocks.MockRepositoryRepo,
	subRepo *mocks.MockSubscriptionRepo,
	ghClient *mocks.MockGitHubClient,
	notif *mocks.MockNotifier,
) *Scanner {
	return NewScanner(repoRepo, subRepo, ghClient, notif, 15)
}

// --- checkRepository tests ---

func TestCheckRepository_NewRelease_NotifiesSubscribers(t *testing.T) {
	repo := domain.Repository{ID: 1, Owner: "golang", Name: "go", LastSeenTag: "v1.25"}

	var updatedTag string
	repoRepo := &mocks.MockRepositoryRepo{
		UpdateLastSeenTagFn: func(ctx context.Context, id int, tag string) error {
			updatedTag = tag
			return nil
		},
	}
	subRepo := &mocks.MockSubscriptionRepo{
		ListConfirmedByRepositoryIDFn: func(ctx context.Context, repositoryID int) ([]domain.Subscription, error) {
			return []domain.Subscription{
				{ID: 1, Email: "user1@example.com"},
				{ID: 2, Email: "user2@example.com"},
			}, nil
		},
	}
	ghClient := &mocks.MockGitHubClient{
		GetLatestReleaseFn: func(ctx context.Context, owner, repoName string) (*githubapi.Release, error) {
			return &githubapi.Release{
				TagName:     "v1.26",
				ReleaseName: "Go 1.26",
				URL:         "https://github.com/golang/go/releases/tag/v1.26",
				PublishedAt: time.Now(),
			}, nil
		},
	}
	notif := &mocks.MockNotifier{}

	s := newTestScanner(repoRepo, subRepo, ghClient, notif)
	s.checkRepository(context.Background(), repo)

	assert.Equal(t, "v1.26", updatedTag)
	require.Len(t, notif.Calls, 2)
	assert.Equal(t, "user1@example.com", notif.Calls[0].To)
	assert.Equal(t, "user2@example.com", notif.Calls[1].To)
	assert.Equal(t, "v1.26", notif.Calls[0].Msg.TagName)
}

func TestCheckRepository_SameTag_NoNotification(t *testing.T) {
	repo := domain.Repository{ID: 1, Owner: "golang", Name: "go", LastSeenTag: "v1.25"}

	ghClient := &mocks.MockGitHubClient{
		GetLatestReleaseFn: func(ctx context.Context, owner, repoName string) (*githubapi.Release, error) {
			return &githubapi.Release{TagName: "v1.25"}, nil
		},
	}
	notif := &mocks.MockNotifier{}

	s := newTestScanner(nil, nil, ghClient, notif)
	s.checkRepository(context.Background(), repo)

	assert.Empty(t, notif.Calls)
}

func TestCheckRepository_FirstScan_SavesTagWithoutNotifying(t *testing.T) {
	repo := domain.Repository{ID: 1, Owner: "golang", Name: "go", LastSeenTag: ""}

	var updatedTag string
	repoRepo := &mocks.MockRepositoryRepo{
		UpdateLastSeenTagFn: func(ctx context.Context, id int, tag string) error {
			updatedTag = tag
			return nil
		},
	}
	ghClient := &mocks.MockGitHubClient{
		GetLatestReleaseFn: func(ctx context.Context, owner, repoName string) (*githubapi.Release, error) {
			return &githubapi.Release{TagName: "v1.25"}, nil
		},
	}
	notif := &mocks.MockNotifier{}

	s := newTestScanner(repoRepo, nil, ghClient, notif)
	s.checkRepository(context.Background(), repo)

	assert.Equal(t, "v1.25", updatedTag)
	assert.Empty(t, notif.Calls, "should not send notifications on first scan")
}

func TestCheckRepository_NoRelease_Skips(t *testing.T) {
	repo := domain.Repository{ID: 1, Owner: "golang", Name: "go", LastSeenTag: "v1.25"}

	ghClient := &mocks.MockGitHubClient{
		GetLatestReleaseFn: func(ctx context.Context, owner, repoName string) (*githubapi.Release, error) {
			return nil, githubapi.ErrNoRelease
		},
	}
	notif := &mocks.MockNotifier{}

	s := newTestScanner(nil, nil, ghClient, notif)
	s.checkRepository(context.Background(), repo)

	assert.Empty(t, notif.Calls)
}

func TestCheckRepository_RateLimit_Skips(t *testing.T) {
	repo := domain.Repository{ID: 1, Owner: "golang", Name: "go", LastSeenTag: "v1.25"}

	ghClient := &mocks.MockGitHubClient{
		GetLatestReleaseFn: func(ctx context.Context, owner, repoName string) (*githubapi.Release, error) {
			return nil, githubapi.ErrRateLimitExceeded
		},
	}
	notif := &mocks.MockNotifier{}

	s := newTestScanner(nil, nil, ghClient, notif)
	s.checkRepository(context.Background(), repo)

	assert.Empty(t, notif.Calls)
}

func TestCheckRepository_NotifyError_ContinuesWithOthers(t *testing.T) {
	repo := domain.Repository{ID: 1, Owner: "golang", Name: "go", LastSeenTag: "v1.25"}

	repoRepo := &mocks.MockRepositoryRepo{
		UpdateLastSeenTagFn: func(ctx context.Context, id int, tag string) error {
			return nil
		},
	}
	subRepo := &mocks.MockSubscriptionRepo{
		ListConfirmedByRepositoryIDFn: func(ctx context.Context, repositoryID int) ([]domain.Subscription, error) {
			return []domain.Subscription{
				{ID: 1, Email: "fail@example.com"},
				{ID: 2, Email: "success@example.com"},
			}, nil
		},
	}
	ghClient := &mocks.MockGitHubClient{
		GetLatestReleaseFn: func(ctx context.Context, owner, repoName string) (*githubapi.Release, error) {
			return &githubapi.Release{TagName: "v1.26", PublishedAt: time.Now()}, nil
		},
	}
	callCount := 0
	notif := &mocks.MockNotifier{
		NotifyFn: func(msg *notifier.Notification, to string) error {
			callCount++
			if callCount == 1 {
				return errors.New("smtp error")
			}
			return nil
		},
	}

	s := newTestScanner(repoRepo, subRepo, ghClient, notif)
	s.checkRepository(context.Background(), repo)

	// Both subscribers should be attempted even if first fails
	assert.Len(t, notif.Calls, 2)
}

// --- scan tests ---

func TestScan_EmptyRepos_NoOp(t *testing.T) {
	repoRepo := &mocks.MockRepositoryRepo{
		ListAllFn: func(ctx context.Context) ([]domain.Repository, error) {
			return []domain.Repository{}, nil
		},
	}
	notif := &mocks.MockNotifier{}

	s := newTestScanner(repoRepo, nil, nil, notif)
	s.scan(context.Background())

	assert.Empty(t, notif.Calls)
}

func TestScan_ListAllError_Returns(t *testing.T) {
	repoRepo := &mocks.MockRepositoryRepo{
		ListAllFn: func(ctx context.Context) ([]domain.Repository, error) {
			return nil, errors.New("db error")
		},
	}

	s := newTestScanner(repoRepo, nil, nil, nil)
	// Should not panic
	s.scan(context.Background())
}

func TestScan_CancelledContext_StopsEarly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	checkedRepos := 0
	repoRepo := &mocks.MockRepositoryRepo{
		ListAllFn: func(ctx context.Context) ([]domain.Repository, error) {
			return []domain.Repository{
				{ID: 1, Owner: "a", Name: "b", LastSeenTag: "v1"},
				{ID: 2, Owner: "c", Name: "d", LastSeenTag: "v1"},
			}, nil
		},
	}
	ghClient := &mocks.MockGitHubClient{
		GetLatestReleaseFn: func(ctx context.Context, owner, repoName string) (*githubapi.Release, error) {
			checkedRepos++
			cancel() // Cancel after first repo
			return &githubapi.Release{TagName: "v1"}, nil
		},
	}

	s := newTestScanner(repoRepo, nil, ghClient, nil)
	s.scan(ctx)

	assert.Equal(t, 1, checkedRepos, "should stop after context cancellation")
}

// --- notifySubscribers tests ---

func TestNotifySubscribers_NoSubscribers_NoOp(t *testing.T) {
	repo := domain.Repository{ID: 1, Owner: "golang", Name: "go"}
	release := &githubapi.Release{TagName: "v1.26", PublishedAt: time.Now()}

	subRepo := &mocks.MockSubscriptionRepo{
		ListConfirmedByRepositoryIDFn: func(ctx context.Context, repositoryID int) ([]domain.Subscription, error) {
			return []domain.Subscription{}, nil
		},
	}
	notif := &mocks.MockNotifier{}

	s := newTestScanner(nil, subRepo, nil, notif)
	s.notifySubscribers(context.Background(), repo, release)

	assert.Empty(t, notif.Calls)
}
