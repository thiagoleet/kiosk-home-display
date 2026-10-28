package news_test

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/activity"
	"github.com/thiagoleet/kiosk-home-display/internal/database"
	"github.com/thiagoleet/kiosk-home-display/internal/events"
	"github.com/thiagoleet/kiosk-home-display/internal/news"
)

type cyclingFetcher struct {
	mu    sync.Mutex
	cycle int
}

// Fetch returns a new batch on every call, so every refresh writes and prunes.
func (f *cyclingFetcher) Fetch(
	ctx context.Context,
	source news.Source,
) ([]news.Article, error) {
	f.mu.Lock()
	f.cycle++
	cycle := f.cycle
	f.mu.Unlock()

	now := time.Now()
	articles := make([]news.Article, 0, 20)

	for i := range 20 {
		link := fmt.Sprintf("%s/%d-%d", source.FeedURL, cycle, i)

		articles = append(articles, news.Article{
			ID:           link,
			Title:        link,
			URL:          link,
			CanonicalURL: link,
			PublishedAt:  now.Add(-time.Duration(i) * time.Second),
		})
	}

	return articles, nil
}

// TestRefreshAlongsideReadersAndActivities runs refresh cycles against the
// database the daemon actually opens, while the endpoint reads it and the
// activity manager writes to it, and checks that none of them fail.
func TestRefreshAlongsideReadersAndActivities(t *testing.T) {
	db, err := database.Open(database.Config{
		Path: filepath.Join(t.TempDir(), "data", "kiosk.db"),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	defer db.Close()

	if err := database.Migrate(db.DB); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	sources := []news.Source{
		{Name: "A", FeedURL: "https://a.example/feed", Category: "general"},
		{Name: "B", FeedURL: "https://b.example/feed", Category: "general"},
		{Name: "C", FeedURL: "https://c.example/feed", Category: "general"},
	}

	service := news.NewService(
		news.NewSQLiteRepository(db.DB),
		&cyclingFetcher{},
		news.Config{
			Enabled:      true,
			Sources:      sources,
			DisplayLimit: 20,
			MaxPerFetch:  20,
			Retention: news.Retention{
				PerSource:     15,
				Total:         30,
				DisabledGrace: time.Hour,
			},
		},
	)

	if err := service.Init(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}

	activities := activity.NewSQLiteRepository(db.DB)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	errs := make(chan error, 1000)

	wg.Go(func() {
		for range 20 {
			service.Refresh(context.Background())
		}

		cancel()
	})

	for reader := range 4 {
		wg.Go(func() {
			for ctx.Err() == nil {
				feed, err := service.List(context.Background(), 20)
				if err != nil {
					errs <- fmt.Errorf("reader %d: %w", reader, err)

					return
				}

				if len(feed.Articles) > 20 || len(feed.Sources) != len(sources) {
					errs <- fmt.Errorf("reader %d: inconsistent feed: %d articles, %d sources",
						reader, len(feed.Articles), len(feed.Sources))

					return
				}
			}
		})
	}

	wg.Go(func() {
		for i := 0; ctx.Err() == nil; i++ {
			if err := activities.Create(
				context.Background(),
				events.NewActivity(
					events.NotificationContextSystem,
					events.EventActivity,
					fmt.Sprintf("activity %d", i),
					"",
				),
			); err != nil {
				errs <- fmt.Errorf("activity %d: %w", i, err)

				return
			}
		}
	})

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}

	feed, err := service.List(context.Background(), 30)
	if err != nil {
		t.Fatalf("final list: %v", err)
	}

	for _, status := range feed.Sources {
		if status.Status != news.SourceStatusOK {
			t.Errorf("source %s ended as %s", status.Name, status.Status)
		}
	}

	if feed.UpdatedAt == nil {
		t.Errorf("updatedAt never set")
	}
}
