package news

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

var ErrDisabled = errors.New("news feed is not enabled")

type Config struct {
	Enabled bool
	Sources []Source

	// DisplayLimit is how many articles the endpoint returns when the caller
	// does not ask for a number. Retention.Total also caps what a caller may
	// ask for, since the database holds no more than that.
	DisplayLimit int

	// MaxPerFetch is how many articles a single fetch of one feed may store,
	// newest first. Portals that list a hundred items would otherwise write a
	// hundred rows every cycle only for retention to delete most of them.
	MaxPerFetch int

	Retention Retention
}

type Service struct {
	repository Repository
	fetcher    Fetcher
	config     Config
	clock      func() time.Time

	// refreshing keeps a slow cycle from overlapping the next tick.
	refreshing sync.Mutex
	sources    []Source
}

func NewService(
	repository Repository,
	fetcher Fetcher,
	config Config,
) *Service {
	return &Service{
		repository: repository,
		fetcher:    fetcher,
		config:     config,
		clock:      time.Now,
	}
}

func (s *Service) Enabled() bool {
	return s.config.Enabled
}

// Init writes the configured sources to the database and deletes the expired
// articles of disabled sources. It touches no network, so it is safe to run
// before the HTTP server starts.
//
// The cleanup runs even with news switched off, because then no refresh cycle
// ever prunes. It is the only cleanup in that case, so articles expire at the
// first daemon start after their grace period — a reboot, deploy or restart —
// rather than the moment it ends. With news off the source list is not
// synced, so sources still marked enabled keep their articles: switching the
// feature off is not taken as removing its sources.
func (s *Service) Init(ctx context.Context) error {
	if s.config.Enabled {
		sources, err := s.repository.SyncSources(ctx, s.config.Sources)
		if err != nil {
			return err
		}

		s.refreshing.Lock()
		s.sources = sources
		s.refreshing.Unlock()
	}

	// Housekeeping must not keep the kiosk from starting.
	if err := s.repository.PruneDisabled(
		ctx,
		s.config.Retention.DisabledGrace,
		s.clock(),
	); err != nil {
		log.Printf("[NEWS] %v", err)
	}

	return nil
}

// Refresh fetches every enabled source once. Each source is isolated: a feed
// that fails keeps the articles it already has, records its error, and does
// not stop the others from updating.
//
// The update time advances only when at least one source succeeded. A cycle
// cut short by shutdown ends silently: it records no failures, prunes
// nothing and leaves the update time alone, and the next start fetches again.
func (s *Service) Refresh(ctx context.Context) {
	if !s.config.Enabled {
		return
	}

	s.refreshing.Lock()
	defer s.refreshing.Unlock()

	updated := 0

	for _, source := range s.sources {
		if ctx.Err() != nil {
			return
		}

		if err := s.refreshSource(ctx, source); err != nil {
			if cancelled(ctx, err) {
				return
			}

			log.Printf(
				"[NEWS] %s: %v",
				source.Name,
				err,
			)

			if markErr := s.repository.MarkFailed(
				ctx,
				source.ID,
				err.Error(),
			); markErr != nil {
				log.Printf("[NEWS] %v", markErr)
			}

			continue
		}

		updated++
	}

	completedAt := s.clock()

	if err := s.repository.Prune(
		ctx,
		s.config.Retention,
		completedAt,
	); err != nil && !cancelled(ctx, err) {
		log.Printf("[NEWS] %v", err)
	}

	if updated > 0 {
		if err := s.repository.MarkRefreshed(
			ctx,
			completedAt,
		); err != nil && !cancelled(ctx, err) {
			log.Printf("[NEWS] %v", err)
		}
	}

	log.Printf(
		"[NEWS] refreshed %d of %d sources",
		updated,
		len(s.sources),
	)
}

func (s *Service) refreshSource(
	ctx context.Context,
	source Source,
) error {
	articles, err := s.fetcher.Fetch(ctx, source)
	if err != nil {
		return err
	}

	fetchedAt := s.clock()

	if err := s.repository.SaveFetch(
		ctx,
		source.ID,
		newest(articles, s.config.MaxPerFetch),
		fetchedAt,
	); err != nil {
		return fmt.Errorf(
			"save articles: %w",
			err,
		)
	}

	return nil
}

// newest keeps the limit most recent articles. Undated articles count as the
// newest, since they are dated when first stored, and keep their feed order.
func newest(articles []Article, limit int) []Article {
	if limit <= 0 || len(articles) <= limit {
		return articles
	}

	sorted := make([]Article, len(articles))
	copy(sorted, articles)

	sort.SliceStable(sorted, func(i, j int) bool {
		left, right := sorted[i].PublishedAt, sorted[j].PublishedAt

		if left.IsZero() || right.IsZero() {
			return left.IsZero() && !right.IsZero()
		}

		return left.After(right)
	})

	return sorted[:limit]
}

// cancelled reports whether err only reflects the cycle being stopped, which
// is not a fault of the source and must not be recorded as one.
func cancelled(ctx context.Context, err error) bool {
	return ctx.Err() != nil || errors.Is(err, context.Canceled)
}

// List returns the articles with the sources interleaved, the update time and
// the status of each source. A limit of zero or less means the configured
// display limit.
func (s *Service) List(
	ctx context.Context,
	limit int,
) (Feed, error) {
	if !s.config.Enabled {
		return Feed{}, ErrDisabled
	}

	if limit <= 0 {
		limit = s.config.DisplayLimit
	}

	if limit > s.config.Retention.Total {
		limit = s.config.Retention.Total
	}

	feed, err := s.repository.Snapshot(ctx, limit)
	if err != nil {
		return Feed{}, err
	}

	// With no source enabled the stored update time belongs to sources that
	// are gone, and would date an empty list.
	if len(feed.Sources) == 0 {
		feed.UpdatedAt = nil
	}

	return feed, nil
}
