package githubapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/go-github/v84/github"
)

var (
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrNoRelease          = errors.New("no release found")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
)

// Client defines the interface for interacting with GitHub API
type Client interface {
	// GetRepository fetches repository info. It handles 404 and Rate limit separately.
	GetRepository(ctx context.Context, owner, repo string) (*Repository, error)
	// GetLatestRelease fetches the latest release for the specified repository.
	// It handles 404 and Rate limit separately.
	GetLatestRelease(ctx context.Context, owner, repo string) (*Release, error)
}

type client struct {
	ghClient *github.Client
}

// Repository contains information about a GitHub repository
type Repository struct {
	Owner       string
	Name        string
	Description string
}

// Release contains information about a GitHub release
type Release struct {
	TagName     string
	ReleaseName string
	URL         string
	PublishedAt time.Time
}

// NewClient creates a new GitHub API client.
// If personalAccessToken is provided, the client will make authenticated requests.
// Otherwise, it will make unauthenticated requests (subject to stricter rate limits).
func NewClient(personalAccessToken string) Client {
	var httpClient *http.Client

	if personalAccessToken != "" {
		httpClient = &http.Client{
			Transport: &transportWithToken{
				token: personalAccessToken,
				base:  http.DefaultTransport,
			},
		}
	}

	gh := github.NewClient(httpClient)

	return &client{
		ghClient: gh,
	}
}

// transportWithToken injects the Authorization header into every request
type transportWithToken struct {
	token string
	base  http.RoundTripper
}

func (t *transportWithToken) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(cloned)
}

// handleGitHubError maps common GitHub API errors to sentinel errors.
func handleGitHubError(err error, resp *github.Response, notFoundErr error) error {
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return notFoundErr
	}
	if _, ok := err.(*github.RateLimitError); ok {
		return ErrRateLimitExceeded
	}
	if _, ok := err.(*github.AbuseRateLimitError); ok {
		return ErrRateLimitExceeded
	}
	return err
}

func (c *client) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	repository, resp, err := c.ghClient.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, handleGitHubError(err, resp, ErrRepositoryNotFound)
	}

	return &Repository{
		Owner:       repository.GetOwner().GetLogin(),
		Name:        repository.GetName(),
		Description: repository.GetDescription(),
	}, nil
}

func (c *client) GetLatestRelease(ctx context.Context, owner, repo string) (*Release, error) {
	release, resp, err := c.ghClient.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, handleGitHubError(err, resp, ErrNoRelease)
	}

	var publishedAt time.Time
	if release.PublishedAt != nil {
		publishedAt = release.PublishedAt.Time
	}

	return &Release{
		TagName:     release.GetTagName(),
		ReleaseName: release.GetName(),
		URL:         release.GetHTMLURL(),
		PublishedAt: publishedAt,
	}, nil
}
