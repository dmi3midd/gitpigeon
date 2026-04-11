package scanner

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"gitpigeon/internal/domain"
	githubapi "gitpigeon/internal/github"
	"gitpigeon/internal/notifier"
)

// Scanner periodically checks for new releases and notifies subscribers.
type Scanner struct {
	repoRepo domain.RepositoryRepo
	subRepo  domain.SubscriptionRepo
	ghClient githubapi.Client
	notifier notifier.Notifier
	interval time.Duration
}

// NewScanner creates a new Scanner instance.
func NewScanner(
	repoRepo domain.RepositoryRepo,
	subRepo domain.SubscriptionRepo,
	ghClient githubapi.Client,
	notifier notifier.Notifier,
	intervalMinutes int,
) *Scanner {
	return &Scanner{
		repoRepo: repoRepo,
		subRepo:  subRepo,
		ghClient: ghClient,
		notifier: notifier,
		interval: time.Duration(intervalMinutes) * time.Minute,
	}
}

// Start begins the periodic scanning loop. It blocks until the context is cancelled.
// Should be called in a separate goroutine.
func (s *Scanner) Start(ctx context.Context) {
	log.Printf("scanner: starting with interval %s", s.interval)

	// Run an initial scan immediately on start
	s.scan(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("scanner: stopping")
			return
		case <-ticker.C:
			s.scan(ctx)
		}
	}
}

// scan performs a single scan pass across all tracked repositories.
func (s *Scanner) scan(ctx context.Context) {
	log.Println("scanner: starting scan pass")

	repos, err := s.repoRepo.ListAll(ctx)
	if err != nil {
		log.Printf("scanner: failed to list repositories: %v", err)
		return
	}

	if len(repos) == 0 {
		log.Println("scanner: no repositories to scan")
		return
	}

	for _, repo := range repos {
		select {
		case <-ctx.Done():
			return
		default:
			s.checkRepository(ctx, repo)
		}
	}

	log.Printf("scanner: scan pass complete, checked %d repositories", len(repos))
}

// checkRepository checks a single repository for a new release.
func (s *Scanner) checkRepository(ctx context.Context, repo domain.Repository) {
	release, err := s.ghClient.GetLatestRelease(ctx, repo.Owner, repo.Name)
	if err != nil {
		if errors.Is(err, githubapi.ErrNoRelease) {
			return // Repository has no releases, skip silently
		}
		if errors.Is(err, githubapi.ErrRateLimitExceeded) {
			log.Printf("scanner: rate limit exceeded, pausing scan")
			return
		}
		log.Printf("scanner: failed to get latest release for %s/%s: %v", repo.Owner, repo.Name, err)
		return
	}

	// First time seeing this repo — save the current tag without notifying
	if repo.LastSeenTag == "" {
		log.Printf("scanner: first scan for %s/%s, saving tag %s", repo.Owner, repo.Name, release.TagName)
		if err := s.repoRepo.UpdateLastSeenTag(ctx, repo.ID, release.TagName); err != nil {
			log.Printf("scanner: failed to update last seen tag for %s/%s: %v", repo.Owner, repo.Name, err)
		}
		return
	}

	// No new release
	if release.TagName == repo.LastSeenTag {
		return
	}

	// New release detected!
	log.Printf("scanner: new release %s found for %s/%s (previous: %s)",
		release.TagName, repo.Owner, repo.Name, repo.LastSeenTag)

	// Update the last seen tag
	if err := s.repoRepo.UpdateLastSeenTag(ctx, repo.ID, release.TagName); err != nil {
		log.Printf("scanner: failed to update last seen tag for %s/%s: %v", repo.Owner, repo.Name, err)
		return
	}

	// Notify all confirmed subscribers
	s.notifySubscribers(ctx, repo, release)
}

// notifySubscribers sends email notifications to all confirmed subscribers of a repository.
func (s *Scanner) notifySubscribers(ctx context.Context, repo domain.Repository, release *githubapi.Release) {
	repoFullName := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)

	subs, err := s.subRepo.ListConfirmedByRepositoryID(ctx, repo.ID)
	if err != nil {
		log.Printf("scanner: failed to list subscribers for %s: %v", repoFullName, err)
		return
	}

	if len(subs) == 0 {
		return
	}

	msg := &notifier.Notification{
		RepoName:    repoFullName,
		TagName:     release.TagName,
		ReleaseName: release.ReleaseName,
		ReleaseURL:  release.URL,
		PublishedAt: release.PublishedAt.Format(time.RFC3339),
	}

	for _, sub := range subs {
		if err := s.notifier.Notify(msg, sub.Email); err != nil {
			log.Printf("scanner: failed to send notification to %s for %s: %v", sub.Email, repoFullName, err)
			continue
		}
		log.Printf("scanner: notified %s about %s@%s", sub.Email, repoFullName, release.TagName)
	}
}
