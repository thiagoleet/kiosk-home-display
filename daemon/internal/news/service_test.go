package news

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func openTestRepository(t *testing.T) *SQLiteRepository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	// Every connection to :memory: is a separate database.
	db.SetMaxOpenConns(1)

	t.Cleanup(func() { db.Close() })

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return NewSQLiteRepository(db)
}

type stubFetcher struct {
	articles map[string][]Article
	failures map[string]error

	// fetch, when set, replaces the maps above.
	fetch func(ctx context.Context, source Source) ([]Article, error)
}

func (f *stubFetcher) Fetch(
	ctx context.Context,
	source Source,
) ([]Article, error) {
	if f.fetch != nil {
		return f.fetch(ctx, source)
	}

	if err := f.failures[source.FeedURL]; err != nil {
		return nil, err
	}

	return f.articles[source.FeedURL], nil
}

func newStubFetcher() *stubFetcher {
	return &stubFetcher{
		articles: map[string][]Article{},
		failures: map[string]error{},
	}
}

var (
	sourceA = Source{Name: "A", FeedURL: "https://a.example/feed", Category: "general"}
	sourceB = Source{Name: "B", FeedURL: "https://b.example/feed", Category: "science"}
	sourceC = Source{Name: "C", FeedURL: "https://c.example/feed", Category: "tech"}
)

var base = time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)

// feedArticles builds count articles for a source, the first one published at
// newest and each following one step earlier, as feeds list them.
func feedArticles(
	source Source,
	prefix string,
	count int,
	newest time.Time,
	step time.Duration,
) []Article {
	articles := make([]Article, 0, count)

	for i := range count {
		link := fmt.Sprintf("https://%s.example/%s-%03d", source.Name, prefix, i)

		articles = append(articles, Article{
			ID:           articleID(source.FeedURL, "", link),
			Title:        fmt.Sprintf("%s %s %d", source.Name, prefix, i),
			URL:          link,
			CanonicalURL: normalizeURL(link),
			PublishedAt:  newest.Add(-time.Duration(i) * step),
		})
	}

	return articles
}

func newTestService(
	t *testing.T,
	repository *SQLiteRepository,
	fetcher Fetcher,
	config Config,
) *Service {
	t.Helper()

	config.Enabled = true

	if config.DisplayLimit == 0 {
		config.DisplayLimit = 20
	}

	if config.Retention.PerSource == 0 {
		config.Retention.PerSource = 30
	}

	if config.Retention.Total == 0 {
		config.Retention.Total = 100
	}

	if config.Retention.DisabledGrace == 0 {
		config.Retention.DisabledGrace = 7 * 24 * time.Hour
	}

	service := NewService(repository, fetcher, config)
	service.clock = func() time.Time { return base }

	if err := service.Init(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}

	return service
}

// stored counts the rows each source has in the database, listed or not.
func stored(t *testing.T, repository *SQLiteRepository) map[string]int {
	t.Helper()

	rows, err := repository.db.Query(`
		SELECT s.name, COUNT(a.id)
		FROM news_sources s
		LEFT JOIN news_articles a ON a.source_id = s.id
		GROUP BY s.id
	`)
	if err != nil {
		t.Fatalf("count: %v", err)
	}

	defer rows.Close()

	counts := map[string]int{}

	for rows.Next() {
		var name string
		var count int

		if err := rows.Scan(&name, &count); err != nil {
			t.Fatalf("scan: %v", err)
		}

		counts[name] = count
	}

	return counts
}

func list(t *testing.T, service *Service, limit int) Feed {
	t.Helper()

	feed, err := service.List(context.Background(), limit)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	return feed
}

func TestHighVolumeSourceDoesNotEvictOthers(t *testing.T) {
	repository := openTestRepository(t)
	fetcher := newStubFetcher()

	// A publishes a hundred stories in the last hours, B a few, days ago:
	// ordered by date alone, A would fill the whole retention.
	fetcher.articles[sourceA.FeedURL] = feedArticles(sourceA, "busy", 100, base, time.Minute)
	fetcher.articles[sourceB.FeedURL] = feedArticles(sourceB, "quiet", 5, base.Add(-72*time.Hour), time.Hour)

	service := newTestService(t, repository, fetcher, Config{
		Sources:     []Source{sourceA, sourceB},
		MaxPerFetch: 30,
		Retention:   Retention{PerSource: 30, Total: 100},
	})

	service.Refresh(context.Background())

	counts := stored(t, repository)

	if counts["A"] != 30 || counts["B"] != 5 {
		t.Fatalf("stored = %v, want A:30 B:5", counts)
	}
}

func TestFailingSourceKeepsArticlesWhileOthersPublish(t *testing.T) {
	repository := openTestRepository(t)
	fetcher := newStubFetcher()

	fetcher.articles[sourceA.FeedURL] = feedArticles(sourceA, "c0", 30, base, time.Minute)
	fetcher.articles[sourceB.FeedURL] = feedArticles(sourceB, "old", 10, base.Add(-48*time.Hour), time.Hour)

	service := newTestService(t, repository, fetcher, Config{
		Sources:     []Source{sourceA, sourceB},
		MaxPerFetch: 30,
		Retention:   Retention{PerSource: 30, Total: 40},
	})

	service.Refresh(context.Background())

	fetcher.failures[sourceB.FeedURL] = errors.New("timeout")

	// A brings thirty new stories every cycle while B stays down.
	for cycle := 1; cycle <= 5; cycle++ {
		now := base.Add(time.Duration(cycle) * time.Hour)

		fetcher.articles[sourceA.FeedURL] = feedArticles(
			sourceA,
			fmt.Sprintf("c%d", cycle),
			30,
			now,
			time.Minute,
		)

		service.clock = func() time.Time { return now }
		service.Refresh(context.Background())
	}

	counts := stored(t, repository)

	if counts["B"] != 10 {
		t.Fatalf("failing source lost articles: %v", counts)
	}

	if counts["A"]+counts["B"] > 40 {
		t.Fatalf("total retention exceeded: %v", counts)
	}
}

func TestTotalRetentionIsSharedFairly(t *testing.T) {
	repository := openTestRepository(t)
	fetcher := newStubFetcher()

	// A is the newest throughout, C the oldest.
	fetcher.articles[sourceA.FeedURL] = feedArticles(sourceA, "a", 20, base, time.Minute)
	fetcher.articles[sourceB.FeedURL] = feedArticles(sourceB, "b", 20, base.Add(-24*time.Hour), time.Minute)
	fetcher.articles[sourceC.FeedURL] = feedArticles(sourceC, "c", 20, base.Add(-48*time.Hour), time.Minute)

	service := newTestService(t, repository, fetcher, Config{
		Sources:   []Source{sourceA, sourceB, sourceC},
		Retention: Retention{PerSource: 30, Total: 30},
	})

	service.Refresh(context.Background())

	counts := stored(t, repository)

	if counts["A"] != 10 || counts["B"] != 10 || counts["C"] != 10 {
		t.Fatalf("stored = %v, want 10 each", counts)
	}
}

func TestListInterleavesSources(t *testing.T) {
	repository := openTestRepository(t)
	fetcher := newStubFetcher()

	fetcher.articles[sourceA.FeedURL] = feedArticles(sourceA, "a", 5, base, time.Minute)
	fetcher.articles[sourceB.FeedURL] = feedArticles(sourceB, "b", 5, base.Add(-24*time.Hour), time.Minute)

	service := newTestService(t, repository, fetcher, Config{
		Sources: []Source{sourceA, sourceB},
	})

	service.Refresh(context.Background())

	var order []string

	for _, article := range list(t, service, 6).Articles {
		order = append(order, article.Source)
	}

	if fmt.Sprint(order) != "[A B A B A B]" {
		t.Fatalf("order = %v", order)
	}
}

func TestMaxPerFetchKeepsTheNewest(t *testing.T) {
	repository := openTestRepository(t)
	fetcher := newStubFetcher()

	articles := feedArticles(sourceA, "a", 50, base, time.Minute)

	// Feeds are not always in order: put the oldest first.
	for i, j := 0, len(articles)-1; i < j; i, j = i+1, j-1 {
		articles[i], articles[j] = articles[j], articles[i]
	}

	fetcher.articles[sourceA.FeedURL] = articles

	service := newTestService(t, repository, fetcher, Config{
		Sources:     []Source{sourceA},
		MaxPerFetch: 10,
	})

	service.Refresh(context.Background())

	if counts := stored(t, repository); counts["A"] != 10 {
		t.Fatalf("stored = %v, want 10", counts)
	}

	feed := list(t, service, 20)

	if !feed.Articles[0].PublishedAt.Equal(base) ||
		!feed.Articles[9].PublishedAt.Equal(base.Add(-9*time.Minute)) {
		t.Fatalf("kept the wrong articles: first %v, last %v",
			feed.Articles[0].PublishedAt,
			feed.Articles[9].PublishedAt,
		)
	}
}

func TestDisabledSourceArticles(t *testing.T) {
	repository := openTestRepository(t)
	fetcher := newStubFetcher()

	fetcher.articles[sourceA.FeedURL] = feedArticles(sourceA, "a", 20, base.Add(-24*time.Hour), time.Minute)
	fetcher.articles[sourceB.FeedURL] = feedArticles(sourceB, "b", 20, base, time.Minute)

	service := newTestService(t, repository, fetcher, Config{
		Sources:   []Source{sourceA, sourceB},
		Retention: Retention{PerSource: 20, Total: 40},
	})

	service.Refresh(context.Background())

	// B is removed from the configuration. Its articles are newer than all
	// of A's, and must neither be listed nor take A's place in a total that
	// now only fits A.
	retention := Retention{PerSource: 20, Total: 20, DisabledGrace: 7 * 24 * time.Hour}

	service = newTestService(t, repository, fetcher, Config{
		Sources:   []Source{sourceA},
		Retention: retention,
	})

	service.clock = func() time.Time { return base.Add(time.Hour) }
	service.Refresh(context.Background())

	counts := stored(t, repository)

	if counts["A"] != 20 {
		t.Fatalf("disabled source consumed retention: %v", counts)
	}

	if counts["B"] != 20 {
		t.Fatalf("disabled source pruned inside the grace period: %v", counts)
	}

	for _, article := range list(t, service, 20).Articles {
		if article.Source == "B" {
			t.Fatalf("disabled source listed: %+v", article)
		}
	}

	service.clock = func() time.Time { return base.Add(8 * 24 * time.Hour) }
	service.Refresh(context.Background())

	if counts := stored(t, repository); counts["B"] != 0 || counts["A"] != 20 {
		t.Fatalf("after the grace period = %v, want A:20 B:0", counts)
	}
}

func TestSharedArticleSurvivesDisablingOneSource(t *testing.T) {
	repository := openTestRepository(t)
	fetcher := newStubFetcher()

	const link = "https://wire.example/story"

	shared := func(source Source, guid string, published time.Time) Article {
		return Article{
			ID:           articleID(source.FeedURL, guid, link),
			Title:        "Shared story",
			URL:          link + "?utm_source=" + source.Name,
			CanonicalURL: normalizeURL(link),
			PublishedAt:  published,
		}
	}

	fetcher.articles[sourceA.FeedURL] = []Article{shared(sourceA, "a-1", base)}
	fetcher.articles[sourceB.FeedURL] = []Article{shared(sourceB, "b-1", base.Add(-time.Minute))}

	service := newTestService(t, repository, fetcher, Config{
		Sources: []Source{sourceA, sourceB},
	})

	service.Refresh(context.Background())

	feed := list(t, service, 20)

	if len(feed.Articles) != 1 || feed.Articles[0].Source != "A" {
		t.Fatalf("expected one copy, from A: %+v", feed.Articles)
	}

	service = newTestService(t, repository, fetcher, Config{
		Sources: []Source{sourceB},
	})

	feed = list(t, service, 20)

	if len(feed.Articles) != 1 || feed.Articles[0].Source != "B" {
		t.Fatalf("expected B's copy after disabling A: %+v", feed.Articles)
	}
}

func TestChangedGUIDReplacesTheStoredCopy(t *testing.T) {
	repository := openTestRepository(t)
	fetcher := newStubFetcher()

	const link = "https://a.example/story"

	story := func(guid string) Article {
		return Article{
			ID:           articleID(sourceA.FeedURL, guid, link),
			Title:        "Story",
			URL:          link,
			CanonicalURL: normalizeURL(link),
			PublishedAt:  base,
		}
	}

	fetcher.articles[sourceA.FeedURL] = []Article{story("first")}

	service := newTestService(t, repository, fetcher, Config{
		Sources: []Source{sourceA},
	})

	service.Refresh(context.Background())

	fetcher.articles[sourceA.FeedURL] = []Article{story("second")}
	service.Refresh(context.Background())

	if counts := stored(t, repository); counts["A"] != 1 {
		t.Fatalf("stored = %v, want a single copy", counts)
	}
}

func TestUpdatedAtAdvancesOnlyWithASuccessfulSource(t *testing.T) {
	repository := openTestRepository(t)
	fetcher := newStubFetcher()

	fetcher.failures[sourceA.FeedURL] = errors.New("down")
	fetcher.failures[sourceB.FeedURL] = errors.New("down")

	service := newTestService(t, repository, fetcher, Config{
		Sources: []Source{sourceA, sourceB},
	})

	service.Refresh(context.Background())

	if feed := list(t, service, 20); feed.UpdatedAt != nil {
		t.Fatalf("updatedAt set although every source failed: %v", feed.UpdatedAt)
	}

	// A recovers.
	delete(fetcher.failures, sourceA.FeedURL)
	fetcher.articles[sourceA.FeedURL] = feedArticles(sourceA, "a", 3, base, time.Minute)

	succeededAt := base.Add(time.Hour)
	service.clock = func() time.Time { return succeededAt }
	service.Refresh(context.Background())

	if feed := list(t, service, 20); feed.UpdatedAt == nil || !feed.UpdatedAt.Equal(succeededAt) {
		t.Fatalf("updatedAt = %v, want %v", feed.UpdatedAt, succeededAt)
	}

	// Everything fails again.
	fetcher.failures[sourceA.FeedURL] = errors.New("down")
	service.clock = func() time.Time { return base.Add(2 * time.Hour) }
	service.Refresh(context.Background())

	feed := list(t, service, 20)

	if feed.UpdatedAt == nil || !feed.UpdatedAt.Equal(succeededAt) {
		t.Fatalf("updatedAt moved on a failed cycle: %v", feed.UpdatedAt)
	}

	if len(feed.Articles) != 3 {
		t.Fatalf("articles lost on a failed cycle: %d", len(feed.Articles))
	}
}

func TestCancelledCycleIsSilent(t *testing.T) {
	repository := openTestRepository(t)
	fetcher := newStubFetcher()

	fetcher.articles[sourceA.FeedURL] = feedArticles(sourceA, "a", 3, base, time.Minute)
	fetcher.articles[sourceB.FeedURL] = feedArticles(sourceB, "b", 3, base, time.Minute)

	service := newTestService(t, repository, fetcher, Config{
		Sources: []Source{sourceA, sourceB},
	})

	service.Refresh(context.Background())

	before := list(t, service, 20)

	// Shutdown arrives while A is being fetched.
	ctx, cancel := context.WithCancel(context.Background())

	fetcher.fetch = func(ctx context.Context, source Source) ([]Article, error) {
		cancel()

		return nil, fmt.Errorf("request feed: %w", ctx.Err())
	}

	service.clock = func() time.Time { return base.Add(time.Hour) }
	service.Refresh(ctx)

	after := list(t, service, 20)

	if !after.UpdatedAt.Equal(*before.UpdatedAt) {
		t.Fatalf("updatedAt moved on a cancelled cycle: %v", after.UpdatedAt)
	}

	for _, status := range after.Sources {
		if status.Status != SourceStatusOK {
			t.Fatalf("cancellation recorded as a failure: %+v", status)
		}
	}
}

func TestSourceStatuses(t *testing.T) {
	repository := openTestRepository(t)
	fetcher := newStubFetcher()

	fetcher.articles[sourceA.FeedURL] = feedArticles(sourceA, "a", 1, base, time.Minute)
	fetcher.failures[sourceB.FeedURL] = errors.New("dial tcp 10.0.0.1:443: connection refused")

	service := newTestService(t, repository, fetcher, Config{
		Sources: []Source{sourceA, sourceB},
	})

	if feed := list(t, service, 20); feed.Sources[0].Status != SourceStatusPending {
		t.Fatalf("before any fetch = %+v", feed.Sources)
	}

	service.Refresh(context.Background())

	feed := list(t, service, 20)

	if len(feed.Sources) != 2 {
		t.Fatalf("sources = %+v", feed.Sources)
	}

	a, b := feed.Sources[0], feed.Sources[1]

	if a.Name != "A" || a.Status != SourceStatusOK || a.LastSuccessAt == nil || !a.LastSuccessAt.Equal(base) {
		t.Errorf("A = %+v", a)
	}

	if b.Name != "B" || b.Status != SourceStatusError || b.LastSuccessAt != nil {
		t.Errorf("B = %+v", b)
	}
}

func TestMigrateUpgradesTheFirstReleaseSchema(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	db.SetMaxOpenConns(1)

	defer db.Close()

	// The tables as the first release created them, next to an unrelated one.
	if _, err := db.Exec(`
		CREATE TABLE activities (id TEXT PRIMARY KEY, title TEXT NOT NULL);
		INSERT INTO activities VALUES ('act-1', 'kept');

		CREATE TABLE news_sources (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			feed_url TEXT NOT NULL UNIQUE,
			category TEXT NOT NULL DEFAULT 'general',
			enabled INTEGER NOT NULL DEFAULT 1,
			last_fetched_at TEXT,
			last_error TEXT
		);

		CREATE TABLE news_articles (
			id TEXT PRIMARY KEY,
			source_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			summary TEXT,
			url TEXT NOT NULL,
			image_url TEXT,
			published_at TEXT,
			fetched_at TEXT NOT NULL,
			FOREIGN KEY (source_id) REFERENCES news_sources(id)
		);

		INSERT INTO news_sources (id, name, feed_url) VALUES (1, 'A', 'https://a.example/feed');
		INSERT INTO news_articles VALUES (
			'old-id', 1, 'Old', NULL, 'https://A.example/story/?utm_source=rss', NULL,
			'2026-10-27T10:00:00.000000000Z', '2026-10-27T10:00:00.000000000Z'
		);
	`); err != nil {
		t.Fatalf("seed: %v", err)
	}

	for range 2 {
		if err := Migrate(db); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}

	var canonical string

	if err := db.QueryRow(
		`SELECT canonical_url FROM news_articles WHERE id = 'old-id'`,
	).Scan(&canonical); err != nil {
		t.Fatalf("read: %v", err)
	}

	if canonical != "https://a.example/story" {
		t.Errorf("canonical_url = %q", canonical)
	}

	var title string

	if err := db.QueryRow(
		`SELECT title FROM activities WHERE id = 'act-1'`,
	).Scan(&title); err != nil || title != "kept" {
		t.Errorf("activities touched: %q, %v", title, err)
	}

	// The first fetch under the new ID scheme replaces the old row instead of
	// keeping both.
	repository := NewSQLiteRepository(db)

	link := "https://a.example/story"

	if err := repository.SaveFetch(
		context.Background(),
		1,
		[]Article{{
			ID:           articleID(sourceA.FeedURL, "", link),
			Title:        "Old",
			URL:          link,
			CanonicalURL: normalizeURL(link),
			PublishedAt:  base,
		}},
		base,
	); err != nil {
		t.Fatalf("save: %v", err)
	}

	var count int

	if err := db.QueryRow(`SELECT COUNT(*) FROM news_articles`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("articles after the first fetch = %d, %v", count, err)
	}
}

func TestSyncSourcesDisablesRemovedSources(t *testing.T) {
	repository := openTestRepository(t)
	ctx := context.Background()

	if _, err := repository.SyncSources(ctx, []Source{sourceA, sourceB}); err != nil {
		t.Fatalf("sync: %v", err)
	}

	enabled, err := repository.SyncSources(ctx, []Source{sourceB})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}

	if len(enabled) != 1 || enabled[0].Name != "B" {
		t.Fatalf("enabled = %+v", enabled)
	}
}

func TestServiceListWhenDisabled(t *testing.T) {
	service := NewService(nil, nil, Config{})

	if _, err := service.List(context.Background(), 0); !errors.Is(err, ErrDisabled) {
		t.Fatalf("expected ErrDisabled, got %v", err)
	}
}
