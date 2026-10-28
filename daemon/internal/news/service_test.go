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

	if _, err := db.Exec(Schema); err != nil {
		t.Fatalf("schema: %v", err)
	}

	return NewSQLiteRepository(db)
}

type stubFetcher struct {
	articles map[string][]Article
	failures map[string]error
}

func (f *stubFetcher) Fetch(
	ctx context.Context,
	source Source,
) ([]Article, error) {
	if err := f.failures[source.FeedURL]; err != nil {
		return nil, err
	}

	return f.articles[source.FeedURL], nil
}

func testArticle(id string, published time.Time) Article {
	return Article{
		ID:          id,
		Title:       "Title " + id,
		URL:         "https://example.com/" + id,
		PublishedAt: published,
	}
}

func TestServiceKeepsArticlesOfFailingSource(t *testing.T) {
	repository := openTestRepository(t)

	sources := []Source{
		{Name: "A", FeedURL: "https://a.example/feed", Category: "general"},
		{Name: "B", FeedURL: "https://b.example/feed", Category: "science"},
	}

	base := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)

	fetcher := &stubFetcher{
		articles: map[string][]Article{
			"https://a.example/feed": {testArticle("a1", base)},
			"https://b.example/feed": {testArticle("b1", base.Add(time.Hour))},
		},
		failures: map[string]error{},
	}

	service := NewService(repository, fetcher, Config{
		Enabled:      true,
		Sources:      sources,
		DisplayLimit: 20,
		Retention:    100,
	})

	service.clock = func() time.Time { return base.Add(2 * time.Hour) }

	if err := service.Init(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}

	service.Refresh(context.Background())

	// B goes down on the next cycle.
	fetcher.failures["https://b.example/feed"] = errors.New("timeout")
	fetcher.articles["https://a.example/feed"] = append(
		fetcher.articles["https://a.example/feed"],
		testArticle("a2", base.Add(90*time.Minute)),
	)

	service.clock = func() time.Time { return base.Add(3 * time.Hour) }

	service.Refresh(context.Background())

	feed, err := service.List(context.Background(), 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	var ids []string

	for _, article := range feed.Articles {
		ids = append(ids, article.ID)
	}

	if fmt.Sprint(ids) != "[a2 b1 a1]" {
		t.Fatalf("articles = %v", ids)
	}

	if feed.Articles[1].Source != "B" || feed.Articles[1].Category != "science" {
		t.Errorf("source not joined: %+v", feed.Articles[1])
	}

	if feed.UpdatedAt == nil || !feed.UpdatedAt.Equal(base.Add(3*time.Hour)) {
		t.Errorf("updatedAt = %v", feed.UpdatedAt)
	}

	enabled, err := repository.enabledSources(context.Background())
	if err != nil {
		t.Fatalf("sources: %v", err)
	}

	if enabled[1].LastError != "timeout" {
		t.Errorf("last error = %q", enabled[1].LastError)
	}

	if !enabled[1].LastFetchedAt.Equal(base.Add(2 * time.Hour)) {
		t.Errorf("failed source last fetch moved: %v", enabled[1].LastFetchedAt)
	}
}

func TestServicePrunesAndLimits(t *testing.T) {
	repository := openTestRepository(t)

	base := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)

	var articles []Article

	for i := 0; i < 10; i++ {
		articles = append(articles, testArticle(
			fmt.Sprintf("n%02d", i),
			base.Add(time.Duration(i)*time.Minute),
		))
	}

	source := Source{Name: "A", FeedURL: "https://a.example/feed", Category: "general"}

	service := NewService(
		repository,
		&stubFetcher{articles: map[string][]Article{source.FeedURL: articles}},
		Config{
			Enabled:      true,
			Sources:      []Source{source},
			DisplayLimit: 3,
			Retention:    5,
		},
	)

	if err := service.Init(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}

	service.Refresh(context.Background())

	feed, err := service.List(context.Background(), 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(feed.Articles) != 3 || feed.Articles[0].ID != "n09" {
		t.Fatalf("default limit: %+v", feed.Articles)
	}

	feed, err = service.List(context.Background(), 50)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(feed.Articles) != 5 || feed.Articles[4].ID != "n05" {
		t.Fatalf("retention: got %d articles", len(feed.Articles))
	}
}

func TestSyncSourcesDisablesRemovedSources(t *testing.T) {
	repository := openTestRepository(t)
	ctx := context.Background()

	first := []Source{
		{Name: "A", FeedURL: "https://a.example/feed", Category: "general"},
		{Name: "B", FeedURL: "https://b.example/feed", Category: "general"},
	}

	if _, err := repository.SyncSources(ctx, first); err != nil {
		t.Fatalf("sync: %v", err)
	}

	enabled, err := repository.SyncSources(ctx, first[1:])
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
