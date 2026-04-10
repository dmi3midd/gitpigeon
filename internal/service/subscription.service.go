package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"gitpigeon/internal/domain"
	githubapi "gitpigeon/internal/github"
	"gitpigeon/internal/notifier"
)

var (
	ErrInvalidRepoFormat  = errors.New("invalid repository format, expected owner/repo")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrRepoNotFound       = errors.New("repository not found on GitHub")
	ErrRateLimitExceeded  = errors.New("GitHub API rate limit exceeded, please try again later")
	ErrSubscriptionExists = errors.New("subscription already exists for this email and repository")
	ErrTokenNotFound      = errors.New("token not found")
	ErrAlreadyConfirmed   = errors.New("subscription already confirmed")
)

// SubscriptionService contains the business logic for subscription management.
type SubscriptionService struct {
	subRepo    domain.SubscriptionRepo
	repoRepo   domain.RepositoryRepo
	ghClient   githubapi.Client
	notifier   notifier.Notifier
	appBaseURL string
}

// NewSubscriptionService creates a new SubscriptionService.
func NewSubscriptionService(
	subRepo domain.SubscriptionRepo,
	repoRepo domain.RepositoryRepo,
	ghClient githubapi.Client,
	notifier notifier.Notifier,
	appBaseURL string,
) *SubscriptionService {
	return &SubscriptionService{
		subRepo:    subRepo,
		repoRepo:   repoRepo,
		ghClient:   ghClient,
		notifier:   notifier,
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
	}
}

func (s *SubscriptionService) Subscribe(ctx context.Context, email, repoFullName string) (*domain.SubscribeResult, error) {
	owner, name, err := parseRepoFullName(repoFullName)
	if err != nil {
		return nil, ErrInvalidRepoFormat
	}

	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	// Verify repository exists on GitHub
	_, err = s.ghClient.GetRepository(ctx, owner, name)
	if err != nil {
		if errors.Is(err, githubapi.ErrRepositoryNotFound) {
			return nil, ErrRepoNotFound
		}
		if errors.Is(err, githubapi.ErrRateLimitExceeded) {
			return nil, ErrRateLimitExceeded
		}
		return nil, fmt.Errorf("failed to verify repository: %w", err)
	}

	// Find or create repository in DB
	repo, err := s.repoRepo.GetByOwnerAndName(ctx, owner, name)
	if err != nil {
		// Repository not found in DB — create it
		repo = &domain.Repository{
			Owner: owner,
			Name:  name,
		}
		repo, err = s.repoRepo.Create(ctx, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to create repository: %w", err)
		}
	}

	// Generate secure tokens
	confirmToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate confirm token: %w", err)
	}
	unsubscribeToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate unsubscribe token: %w", err)
	}

	sub := &domain.Subscription{
		RepositoryID:     repo.ID,
		Email:            email,
		Confirmed:        false,
		ConfirmToken:     confirmToken,
		UnsubscribeToken: unsubscribeToken,
	}

	sub, err = s.subRepo.Create(ctx, sub)
	if err != nil {
		// Check if it's a unique constraint violation (duplicate subscription)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrSubscriptionExists
		}
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Send confirmation email (async — don't block the response)
	go func() {
		confirmURL := fmt.Sprintf("%s/api/confirm/%s", s.appBaseURL, confirmToken)
		msg := &notifier.Notification{
			RepoName:    fmt.Sprintf("%s/%s", owner, name),
			TagName:     "subscription confirmation",
			ReleaseName: "Please confirm your subscription",
			ReleaseURL:  confirmURL,
		}
		if notifyErr := s.notifier.Notify(msg, email); notifyErr != nil {
			log.Printf("failed to send confirmation email to %s: %v", email, notifyErr)
		}
	}()

	return &domain.SubscribeResult{
		Message: "Confirmation email sent. Please check your inbox.",
	}, nil
}

func (s *SubscriptionService) Confirm(ctx context.Context, token string) error {
	sub, err := s.subRepo.GetByConfirmToken(ctx, token)
	if err != nil {
		return ErrTokenNotFound
	}

	if sub.Confirmed {
		return ErrAlreadyConfirmed
	}

	if err := s.subRepo.Confirm(ctx, sub.ID); err != nil {
		return fmt.Errorf("failed to confirm subscription: %w", err)
	}

	return nil
}

func (s *SubscriptionService) Unsubscribe(ctx context.Context, token string) error {
	sub, err := s.subRepo.GetByUnsubscribeToken(ctx, token)
	if err != nil {
		return ErrTokenNotFound
	}

	if err := s.subRepo.Delete(ctx, sub.ID); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

func (s *SubscriptionService) GetSubscriptions(ctx context.Context, email string) ([]domain.SubscriptionInfo, error) {
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	subs, err := s.subRepo.ListByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	result := make([]domain.SubscriptionInfo, 0, len(subs))
	for _, sub := range subs {
		repo, err := s.repoRepo.GetByID(ctx, sub.RepositoryID)
		if err != nil {
			log.Printf("failed to get repository %d for subscription %d: %v", sub.RepositoryID, sub.ID, err)
			continue
		}

		result = append(result, domain.SubscriptionInfo{
			ID:        sub.ID,
			Repo:      fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
			Email:     sub.Email,
			CreatedAt: sub.CreatedAt,
		})
	}

	return result, nil
}

// parseRepoFullName splits "owner/repo" into owner and repo.
func parseRepoFullName(fullName string) (owner, repo string, err error) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", ErrInvalidRepoFormat
	}

	// Reject names with extra slashes or whitespace
	if strings.Contains(parts[1], "/") || strings.ContainsAny(fullName, " \t\n") {
		return "", "", ErrInvalidRepoFormat
	}

	return parts[0], parts[1], nil
}

// isValidEmail performs basic email validation.
func isValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	at := strings.LastIndex(email, "@")
	if at < 1 || at >= len(email)-1 {
		return false
	}

	domain := email[at+1:]
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

// generateToken creates a cryptographically secure random token.
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
