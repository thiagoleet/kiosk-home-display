package news

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var ErrDisabled = errors.New("news feed is not enabled")

type Config struct {
	Enabled bool
	Sources []Source

	// DisplayLimit is how many articles the endpoint returns when the caller
	// does not ask for a number. Retention is how many stay in the database,
	// and also caps what a caller may ask for.
	DisplayLimit int
	Retention    int
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

// Init writes the configured sources to the database. It touches no network,
// so it is safe to run before the HTTP server starts.
func (s *Service) Init(ctx context.Context) error {
	if !s.config.Enabled {
		return nil
	}

	sources, err := s.repository.SyncSources(ctx, s.config.Sources)
	if err != nil {
		return err
	}

	s.refreshing.Lock()
	s.sources = sources
	s.refreshing.Unlock()

	return nil
}

// Refresh fetches every enabled source once. Each source is isolated: a feed
// that fails keeps the articles it already has, records its error, and does
// not stop the others from updating.
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

	if err := s.repository.Prune(ctx, s.config.Retention); err != nil {
		log.Printf("[NEWS] %v", err)
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

	if err := s.repository.SaveArticles(
		ctx,
		source.ID,
		articles,
		fetchedAt,
	); err != nil {
		return fmt.Errorf(
			"save articles: %w",
			err,
		)
	}

	return s.repository.MarkFetched(ctx, source.ID, fetchedAt)
}

// List returns the newest articles. A limit of zero or less means the
// configured display limit, and anything beyond the retention is capped there
// since the database holds no more than that.
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

	if limit > s.config.Retention {
		limit = s.config.Retention
	}

	articles, err := s.repository.List(ctx, limit)
	if err != nil {
		return Feed{}, err
	}

	updatedAt, err := s.repository.UpdatedAt(ctx)
	if err != nil {
		return Feed{}, err
	}

	return Feed{
		Articles:  articles,
		UpdatedAt: updatedAt,
	}, nil
}
