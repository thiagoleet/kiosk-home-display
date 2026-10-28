package app

import (
	"context"
	"testing"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/config"
	"github.com/thiagoleet/kiosk-home-display/internal/news"
)

// isolateDatabase moves the test into an empty directory. New opens the
// database at the relative path ./data/kiosk.db, which would otherwise be
// created, and migrated, inside this package's source directory.
func isolateDatabase(t *testing.T) {
	t.Helper()

	t.Chdir(t.TempDir())
}

func TestAppStopsWhenContextIsCancelled(t *testing.T) {
	isolateDatabase(t)

	cfg := config.Default()

	// A new database is seeded with the default news sources, which Run would
	// start fetching from the internet.
	cfg.News.Enabled = false

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)

	go func() {
		done <- application.Run(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

	case <-time.After(time.Second):
		t.Fatal("application did not stop after context cancellation")
	}
}

func TestAppStartsWithLinuxDisplayMode(t *testing.T) {
	isolateDatabase(t)

	cfg := config.Default()
	cfg.Display.Mode = "linux"

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	defer application.db.Close()
}

// enabledFeeds lists the feed URLs the news service would collect.
func enabledFeeds(t *testing.T, application *App) []string {
	t.Helper()

	rows, err := application.db.Query(
		`SELECT feed_url FROM news_sources WHERE enabled = 1 ORDER BY id`,
	)
	if err != nil {
		t.Fatalf("query sources: %v", err)
	}

	defer rows.Close()

	var feeds []string

	for rows.Next() {
		var feed string

		if err := rows.Scan(&feed); err != nil {
			t.Fatalf("scan source: %v", err)
		}

		feeds = append(feeds, feed)
	}

	return feeds
}

// New is the daemon's startup path: it migrates the database before the news
// service reads it, so a fresh install collects the default sources without
// NEWS_SOURCES. Run is not called, so nothing is fetched.
func TestNewSeedsAndSchedulesDefaultSources(t *testing.T) {
	isolateDatabase(t)

	application, err := New(config.Default())
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	defer application.db.Close()

	if got := len(enabledFeeds(t, application)); got != 5 {
		t.Fatalf("enabled sources = %d, want 5", got)
	}

	if application.newsSched == nil {
		t.Fatal("expected the default sources to be scheduled")
	}
}

// NEWS_SOURCES, when set, still decides: the defaults are seeded but disabled
// by the sync that follows.
func TestNewsSourcesOverrideTheDefaults(t *testing.T) {
	isolateDatabase(t)

	cfg := config.Default()
	cfg.News.Sources = []news.Source{
		{Name: "Local", FeedURL: "https://local.example/feed", Category: "general"},
	}

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	defer application.db.Close()

	feeds := enabledFeeds(t, application)

	if len(feeds) != 1 || feeds[0] != "https://local.example/feed" {
		t.Fatalf("enabled = %v, want only the configured source", feeds)
	}

	var total int

	if err := application.db.QueryRow(
		`SELECT COUNT(*) FROM news_sources`,
	).Scan(&total); err != nil {
		t.Fatalf("count sources: %v", err)
	}

	if total != 6 {
		t.Fatalf("sources = %d, want 6 (5 defaults kept, disabled)", total)
	}
}
