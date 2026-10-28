package news

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// schema creates the tables on a fresh database. Columns added after the first
// release are not listed here but in Migrate, which adds them to tables that
// already exist on a box.
const schema = `
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

	CREATE INDEX IF NOT EXISTS idx_news_articles_source
	ON news_articles(source_id, published_at DESC);

	-- A single row: when the last refresh cycle with at least one successful
	-- source completed.
	CREATE TABLE IF NOT EXISTS news_refreshes (
		id           INTEGER PRIMARY KEY CHECK (id = 1),
		completed_at TEXT NOT NULL
	);
`

// Migrate creates the news tables and brings tables created by an earlier
// release up to date. Every step is additive, so the other tables in the
// database, and the articles already stored, are left as they are.
func Migrate(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf(
			"failed to create news tables: %w",
			err,
		)
	}

	added, err := addColumn(
		db,
		"news_articles",
		"canonical_url",
		"TEXT NOT NULL DEFAULT ''",
	)
	if err != nil {
		return err
	}

	if added {
		if err := backfillCanonicalURLs(db); err != nil {
			return err
		}
	}

	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_news_articles_canonical_url
		ON news_articles(canonical_url)
	`); err != nil {
		return fmt.Errorf(
			"failed to index canonical urls: %w",
			err,
		)
	}

	return nil
}

// addColumn adds a column unless the table already has it, and reports
// whether it did.
func addColumn(
	db *sql.DB,
	table string,
	column string,
	definition string,
) (bool, error) {
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false, fmt.Errorf(
			"failed to read %s columns: %w",
			table,
			err,
		)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			index        int
			name         string
			kind         string
			notNull      int
			defaultValue sql.NullString
			primaryKey   int
		)

		if err := rows.Scan(
			&index,
			&name,
			&kind,
			&notNull,
			&defaultValue,
			&primaryKey,
		); err != nil {
			return false, fmt.Errorf(
				"failed to scan %s columns: %w",
				table,
				err,
			)
		}

		if name == column {
			return false, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf(
			"failed to iterate %s columns: %w",
			table,
			err,
		)
	}

	rows.Close()

	if _, err := db.Exec(
		`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + definition,
	); err != nil {
		return false, fmt.Errorf(
			"failed to add %s.%s: %w",
			table,
			column,
			err,
		)
	}

	return true, nil
}

// backfillCanonicalURLs fills the column for articles stored before it
// existed. The normalisation lives in Go, so it cannot be a single UPDATE.
func backfillCanonicalURLs(db *sql.DB) error {
	rows, err := db.Query(
		`SELECT id, url FROM news_articles WHERE canonical_url = ''`,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to read articles to backfill: %w",
			err,
		)
	}

	urls := map[string]string{}

	for rows.Next() {
		var id, link string

		if err := rows.Scan(&id, &link); err != nil {
			rows.Close()

			return fmt.Errorf(
				"failed to scan article to backfill: %w",
				err,
			)
		}

		urls[id] = normalizeURL(link)
	}

	if err := rows.Err(); err != nil {
		rows.Close()

		return fmt.Errorf(
			"failed to iterate articles to backfill: %w",
			err,
		)
	}

	rows.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf(
			"failed to begin canonical url backfill: %w",
			err,
		)
	}

	defer tx.Rollback()

	for id, canonical := range urls {
		if _, err := tx.Exec(
			`UPDATE news_articles SET canonical_url = ? WHERE id = ?`,
			canonical,
			id,
		); err != nil {
			return fmt.Errorf(
				"failed to backfill canonical url: %w",
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"failed to commit canonical url backfill: %w",
			err,
		)
	}

	return nil
}

// Retention bounds what the database keeps.
//
// PerSource is how many articles each enabled source keeps, newest first.
// Total caps all enabled sources together and is applied fairly: every source
// keeps its newest articles before any source keeps its older ones, so a
// source that publishes a hundred stories an hour cannot push out a quieter
// one, nor a source whose fetch is failing.
//
// Articles of disabled sources count towards neither, and are deleted once
// they have not been seen in a fetch for DisabledGrace. A source switched off
// by mistake and restored within that time comes back with its articles.
type Retention struct {
	PerSource     int
	Total         int
	DisabledGrace time.Duration
}

type Repository interface {
	SyncSources(ctx context.Context, sources []Source) ([]Source, error)
	SaveFetch(ctx context.Context, sourceID int64, articles []Article, fetchedAt time.Time) error
	MarkFailed(ctx context.Context, sourceID int64, message string) error
	MarkRefreshed(ctx context.Context, completedAt time.Time) error
	Prune(ctx context.Context, retention Retention, now time.Time) error
	Snapshot(ctx context.Context, limit int) (Feed, error)
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

// SaveFetch stores a successful fetch: the source's articles and its success
// time, in one transaction, so a reader never sees the source marked fresh
// without the articles that made it so.
//
// An article seen again keeps its original publication time, so a feed that
// rewrites its dates on every fetch does not reshuffle the slideshow.
func (r *SQLiteRepository) SaveFetch(
	ctx context.Context,
	sourceID int64,
	articles []Article,
	fetchedAt time.Time,
) error {
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
			canonical_url,
			image_url,
			published_at,
			fetched_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			source_id = excluded.source_id,
			title = excluded.title,
			summary = excluded.summary,
			url = excluded.url,
			canonical_url = excluded.canonical_url,
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
			article.CanonicalURL,
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

		// One row per story per source. A feed that changes an item's GUID,
		// or a row stored under the previous ID scheme, would otherwise leave
		// a second copy taking a retention slot.
		if _, err := tx.ExecContext(
			ctx,
			`DELETE FROM news_articles
			WHERE source_id = ? AND canonical_url = ? AND id <> ?`,
			sourceID,
			article.CanonicalURL,
			article.ID,
		); err != nil {
			return fmt.Errorf(
				"failed to remove stale copies of %q: %w",
				article.ID,
				err,
			)
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE news_sources
		SET last_fetched_at = ?, last_error = NULL
		WHERE id = ?`,
		fetched,
		sourceID,
	); err != nil {
		return fmt.Errorf(
			"failed to mark news source fetched: %w",
			err,
		)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"failed to commit news articles: %w",
			err,
		)
	}

	return nil
}

// MarkFailed records the error and leaves last_fetched_at alone, so the
// source still reports when it last succeeded.
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

func (r *SQLiteRepository) MarkRefreshed(
	ctx context.Context,
	completedAt time.Time,
) error {
	const query = `
		INSERT INTO news_refreshes (id, completed_at)
		VALUES (1, ?)
		ON CONFLICT(id) DO UPDATE SET
			completed_at = excluded.completed_at
	`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		formatTime(completedAt),
	); err != nil {
		return fmt.Errorf(
			"failed to record news refresh: %w",
			err,
		)
	}

	return nil
}

// Prune applies the retention policy in one transaction.
func (r *SQLiteRepository) Prune(
	ctx context.Context,
	retention Retention,
	now time.Time,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(
			"failed to begin pruning news articles: %w",
			err,
		)
	}

	defer tx.Rollback()

	const disabled = `
		DELETE FROM news_articles
		WHERE source_id IN (SELECT id FROM news_sources WHERE enabled = 0)
		AND fetched_at < ?
	`

	if _, err := tx.ExecContext(
		ctx,
		disabled,
		formatTime(now.Add(-retention.DisabledGrace)),
	); err != nil {
		return fmt.Errorf(
			"failed to prune articles of disabled sources: %w",
			err,
		)
	}

	// turn is an article's position within its own source, newest first.
	// Ordering everything by turn before date is what makes the total fair:
	// the cut falls on the deepest turns, whichever source they belong to.
	const enabled = `
		WITH ranked AS (
			SELECT
				a.id,
				a.published_at,
				ROW_NUMBER() OVER (
					PARTITION BY a.source_id
					ORDER BY a.published_at DESC, a.id
				) AS turn
			FROM news_articles a
			JOIN news_sources s ON s.id = a.source_id
			WHERE s.enabled = 1
		),
		positioned AS (
			SELECT
				id,
				turn,
				ROW_NUMBER() OVER (
					ORDER BY turn, published_at DESC, id
				) AS position
			FROM ranked
		)
		DELETE FROM news_articles
		WHERE id IN (
			SELECT id
			FROM positioned
			WHERE turn > ? OR position > ?
		)
	`

	if _, err := tx.ExecContext(
		ctx,
		enabled,
		retention.PerSource,
		retention.Total,
	); err != nil {
		return fmt.Errorf(
			"failed to prune news articles: %w",
			err,
		)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"failed to commit news pruning: %w",
			err,
		)
	}

	return nil
}

// Snapshot reads everything the endpoint returns inside one transaction, so a
// refresh committing in between cannot pair new articles with an old update
// time or source status.
func (r *SQLiteRepository) Snapshot(
	ctx context.Context,
	limit int,
) (Feed, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Feed{}, fmt.Errorf(
			"failed to begin reading news: %w",
			err,
		)
	}

	defer tx.Rollback()

	articles, err := listArticles(ctx, tx, limit)
	if err != nil {
		return Feed{}, err
	}

	updatedAt, err := refreshedAt(ctx, tx)
	if err != nil {
		return Feed{}, err
	}

	sources, err := sourceStatuses(ctx, tx)
	if err != nil {
		return Feed{}, err
	}

	if err := tx.Commit(); err != nil {
		return Feed{}, fmt.Errorf(
			"failed to finish reading news: %w",
			err,
		)
	}

	return Feed{
		Articles:  articles,
		UpdatedAt: updatedAt,
		Sources:   sources,
	}, nil
}

// listArticles interleaves the sources: the newest article of every source,
// then the second newest of every source, and so on, each round newest first.
// A story carried by more than one source is listed once, from whichever copy
// is newest.
func listArticles(
	ctx context.Context,
	tx *sql.Tx,
	limit int,
) ([]Article, error) {
	const query = `
		WITH visible AS (
			SELECT
				a.id,
				a.source_id,
				a.title,
				a.summary,
				s.name,
				s.category,
				a.image_url,
				a.url,
				a.published_at,
				ROW_NUMBER() OVER (
					PARTITION BY CASE
						WHEN a.canonical_url = '' THEN a.id
						ELSE a.canonical_url
					END
					ORDER BY a.published_at DESC, a.source_id, a.id
				) AS copy
			FROM news_articles a
			JOIN news_sources s ON s.id = a.source_id
			WHERE s.enabled = 1
		),
		ranked AS (
			SELECT
				*,
				ROW_NUMBER() OVER (
					PARTITION BY source_id
					ORDER BY published_at DESC, id
				) AS turn
			FROM visible
			WHERE copy = 1
		)
		SELECT
			id,
			title,
			summary,
			name,
			category,
			image_url,
			url,
			published_at
		FROM ranked
		ORDER BY turn, published_at DESC, id
		LIMIT ?
	`

	rows, err := tx.QueryContext(
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

func refreshedAt(
	ctx context.Context,
	tx *sql.Tx,
) (*time.Time, error) {
	var value sql.NullString

	err := tx.QueryRowContext(
		ctx,
		`SELECT completed_at FROM news_refreshes WHERE id = 1`,
	).Scan(&value)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf(
			"failed to read news update time: %w",
			err,
		)
	}

	return parseOptionalTime(value), nil
}

func sourceStatuses(
	ctx context.Context,
	tx *sql.Tx,
) ([]SourceStatus, error) {
	const query = `
		SELECT name, category, last_fetched_at, last_error
		FROM news_sources
		WHERE enabled = 1
		ORDER BY id
	`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read news source statuses: %w",
			err,
		)
	}

	defer rows.Close()

	statuses := []SourceStatus{}

	for rows.Next() {
		var status SourceStatus
		var lastFetchedAt sql.NullString
		var lastError sql.NullString

		if err := rows.Scan(
			&status.Name,
			&status.Category,
			&lastFetchedAt,
			&lastError,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan news source status: %w",
				err,
			)
		}

		status.LastSuccessAt = parseOptionalTime(lastFetchedAt)

		switch {
		case lastError.Valid:
			status.Status = SourceStatusError

		case status.LastSuccessAt != nil:
			status.Status = SourceStatusOK

		default:
			status.Status = SourceStatusPending
		}

		statuses = append(statuses, status)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate news source statuses: %w",
			err,
		)
	}

	return statuses, nil
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
