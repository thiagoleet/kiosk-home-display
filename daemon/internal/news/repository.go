package news

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Schema is applied by database.Migrate alongside the other tables.
const Schema = `
	CREATE TABLE IF NOT EXISTS news_sources (
		id              INTEGER PRIMARY KEY,
		name            TEXT NOT NULL,
		feed_url        TEXT NOT NULL UNIQUE,
		category        TEXT NOT NULL DEFAULT 'general',
		enabled         INTEGER NOT NULL DEFAULT 1,
		last_fetched_at TEXT,
		last_error      TEXT
	);

	CREATE TABLE IF NOT EXISTS news_articles (
		id           TEXT PRIMARY KEY,
		source_id    INTEGER NOT NULL,
		title        TEXT NOT NULL,
		summary      TEXT,
		url          TEXT NOT NULL,
		image_url    TEXT,
		published_at TEXT,
		fetched_at   TEXT NOT NULL,
		FOREIGN KEY (source_id) REFERENCES news_sources(id)
	);

	CREATE INDEX IF NOT EXISTS idx_news_articles_published_at
	ON news_articles(published_at DESC);
`

type Repository interface {
	SyncSources(ctx context.Context, sources []Source) ([]Source, error)
	SaveArticles(ctx context.Context, sourceID int64, articles []Article, fetchedAt time.Time) error
	MarkFetched(ctx context.Context, sourceID int64, fetchedAt time.Time) error
	MarkFailed(ctx context.Context, sourceID int64, message string) error
	Prune(ctx context.Context, keep int) error
	List(ctx context.Context, limit int) ([]Article, error)
	UpdatedAt(ctx context.Context) (*time.Time, error)
}

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(
	db *sql.DB,
) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

// SyncSources makes the table match the configuration and returns the enabled
// sources with their row IDs. A source dropped from the configuration is
// disabled rather than deleted, so its articles stay referenced but stop being
// listed, and come back if it is configured again.
func (r *SQLiteRepository) SyncSources(
	ctx context.Context,
	sources []Source,
) ([]Source, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to begin news source sync: %w",
			err,
		)
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE news_sources SET enabled = 0`,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to disable news sources: %w",
			err,
		)
	}

	const upsert = `
		INSERT INTO news_sources (name, feed_url, category, enabled)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(feed_url) DO UPDATE SET
			name = excluded.name,
			category = excluded.category,
			enabled = 1
	`

	for _, source := range sources {
		if _, err := tx.ExecContext(
			ctx,
			upsert,
			source.Name,
			source.FeedURL,
			source.Category,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to save news source %q: %w",
				source.Name,
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf(
			"failed to commit news source sync: %w",
			err,
		)
	}

	return r.enabledSources(ctx)
}

func (r *SQLiteRepository) enabledSources(
	ctx context.Context,
) ([]Source, error) {
	const query = `
		SELECT id, name, feed_url, category, last_fetched_at, last_error
		FROM news_sources
		WHERE enabled = 1
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list news sources: %w",
			err,
		)
	}

	defer rows.Close()

	var sources []Source

	for rows.Next() {
		var source Source
		var lastFetchedAt sql.NullString
		var lastError sql.NullString

		if err := rows.Scan(
			&source.ID,
			&source.Name,
			&source.FeedURL,
			&source.Category,
			&lastFetchedAt,
			&lastError,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan news source: %w",
				err,
			)
		}

		source.Enabled = true
		source.LastError = lastError.String
		source.LastFetchedAt = parseOptionalTime(lastFetchedAt)

		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate news sources: %w",
			err,
		)
	}

	return sources, nil
}

// SaveArticles upserts a source's articles in one transaction. An article seen
// again keeps its original publication time, so a feed that rewrites its dates
// on every fetch does not reshuffle the slideshow.
func (r *SQLiteRepository) SaveArticles(
	ctx context.Context,
	sourceID int64,
	articles []Article,
	fetchedAt time.Time,
) error {
	if len(articles) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(
			"failed to begin saving news articles: %w",
			err,
		)
	}

	defer tx.Rollback()

	const upsert = `
		INSERT INTO news_articles (
			id,
			source_id,
			title,
			summary,
			url,
			image_url,
			published_at,
			fetched_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			summary = excluded.summary,
			url = excluded.url,
			image_url = excluded.image_url,
			fetched_at = excluded.fetched_at
	`

	fetched := formatTime(fetchedAt)

	for _, article := range articles {
		published := article.PublishedAt

		// Without a date the article is dated when it was first seen, which
		// is also when it first appeared in the feed to within one interval.
		if published.IsZero() {
			published = fetchedAt
		}

		if _, err := tx.ExecContext(
			ctx,
			upsert,
			article.ID,
			sourceID,
			article.Title,
			nullString(article.Summary),
			article.URL,
			nullString(article.ImageURL),
			formatTime(published),
			fetched,
		); err != nil {
			return fmt.Errorf(
				"failed to save news article %q: %w",
				article.ID,
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"failed to commit news articles: %w",
			err,
		)
	}

	return nil
}

func (r *SQLiteRepository) MarkFetched(
	ctx context.Context,
	sourceID int64,
	fetchedAt time.Time,
) error {
	const query = `
		UPDATE news_sources
		SET last_fetched_at = ?, last_error = NULL
		WHERE id = ?
	`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		formatTime(fetchedAt),
		sourceID,
	); err != nil {
		return fmt.Errorf(
			"failed to mark news source fetched: %w",
			err,
		)
	}

	return nil
}

// MarkFailed records the error and leaves last_fetched_at alone, so updatedAt
// keeps reporting the last time the content actually changed.
func (r *SQLiteRepository) MarkFailed(
	ctx context.Context,
	sourceID int64,
	message string,
) error {
	const query = `
		UPDATE news_sources
		SET last_error = ?
		WHERE id = ?
	`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		message,
		sourceID,
	); err != nil {
		return fmt.Errorf(
			"failed to mark news source failed: %w",
			err,
		)
	}

	return nil
}

// Prune keeps the newest articles and deletes the rest. Articles of disabled
// sources count too, so a removed feed ages out instead of lingering forever.
func (r *SQLiteRepository) Prune(
	ctx context.Context,
	keep int,
) error {
	const query = `
		DELETE FROM news_articles
		WHERE id NOT IN (
			SELECT id
			FROM news_articles
			ORDER BY published_at DESC, id
			LIMIT ?
		)
	`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		keep,
	); err != nil {
		return fmt.Errorf(
			"failed to prune news articles: %w",
			err,
		)
	}

	return nil
}

func (r *SQLiteRepository) List(
	ctx context.Context,
	limit int,
) ([]Article, error) {
	const query = `
		SELECT
			a.id,
			a.title,
			a.summary,
			s.name,
			s.category,
			a.image_url,
			a.url,
			a.published_at
		FROM news_articles a
		JOIN news_sources s ON s.id = a.source_id
		WHERE s.enabled = 1
		ORDER BY a.published_at DESC, a.id
		LIMIT ?
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list news articles: %w",
			err,
		)
	}

	defer rows.Close()

	articles := make([]Article, 0, limit)

	for rows.Next() {
		var article Article
		var summary sql.NullString
		var imageURL sql.NullString
		var publishedAt sql.NullString

		if err := rows.Scan(
			&article.ID,
			&article.Title,
			&summary,
			&article.Source,
			&article.Category,
			&imageURL,
			&article.URL,
			&publishedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan news article: %w",
				err,
			)
		}

		article.Summary = summary.String
		article.ImageURL = imageURL.String

		if published := parseOptionalTime(publishedAt); published != nil {
			article.PublishedAt = *published
		}

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate news articles: %w",
			err,
		)
	}

	return articles, nil
}

// UpdatedAt is the most recent successful fetch among the enabled sources.
func (r *SQLiteRepository) UpdatedAt(
	ctx context.Context,
) (*time.Time, error) {
	const query = `
		SELECT MAX(last_fetched_at)
		FROM news_sources
		WHERE enabled = 1
	`

	var value sql.NullString

	if err := r.db.QueryRowContext(
		ctx,
		query,
	).Scan(&value); err != nil {
		return nil, fmt.Errorf(
			"failed to read news update time: %w",
			err,
		)
	}

	return parseOptionalTime(value), nil
}

// Times are stored as fixed-width UTC RFC 3339 so that text ordering in
// SQLite matches chronological ordering.
const timeLayout = "2006-01-02T15:04:05.000000000Z"

func formatTime(value time.Time) string {
	return value.UTC().Format(timeLayout)
}

func parseOptionalTime(value sql.NullString) *time.Time {
	if !value.Valid || value.String == "" {
		return nil
	}

	parsed, err := time.Parse(time.RFC3339Nano, value.String)
	if err != nil {
		return nil
	}

	return &parsed
}

func nullString(value string) sql.NullString {
	return sql.NullString{
		String: value,
		Valid:  value != "",
	}
}
