package database

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/thiagoleet/kiosk-home-display/internal/news"
)

type sourceRow struct {
	ID       int64
	Name     string
	FeedURL  string
	Category string
	Enabled  bool
}

var wantDefaults = []sourceRow{
	{Name: "Agência Brasil", FeedURL: "https://agenciabrasil.ebc.com.br/feed/ultimasnoticias/feed.xml", Category: "general", Enabled: true},
	{Name: "G1", FeedURL: "https://g1.globo.com/rss/g1/", Category: "general", Enabled: true},
	{Name: "NASA", FeedURL: "https://www.nasa.gov/feed/", Category: "science", Enabled: true},
	{Name: "BBC News Brasil", FeedURL: "https://feeds.bbci.co.uk/portuguese/rss.xml", Category: "general", Enabled: true},
	{Name: "ACidade ON Ribeirão Preto", FeedURL: "https://www.acidadeon.com/ribeiraopreto/feed/", Category: "general", Enabled: true},
}

// openTestDatabase opens a database file of its own, as the daemon does, so
// the pool's several connections share it.
func openTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Open(Config{
		Path: filepath.Join(t.TempDir(), "kiosk.db"),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	return db.DB
}

// createPreviousSchema builds the tables as the release before the migration
// steps left them: the same idempotent schema, and user_version 0.
func createPreviousSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM news_sources`); err != nil {
		t.Fatalf("clear sources: %v", err)
	}

	if _, err := db.Exec(`PRAGMA user_version = 0`); err != nil {
		t.Fatalf("reset version: %v", err)
	}
}

func sources(t *testing.T, db *sql.DB) []sourceRow {
	t.Helper()

	rows, err := db.Query(
		`SELECT id, name, feed_url, category, enabled FROM news_sources ORDER BY id`,
	)
	if err != nil {
		t.Fatalf("query sources: %v", err)
	}

	defer rows.Close()

	var result []sourceRow

	for rows.Next() {
		var row sourceRow

		if err := rows.Scan(&row.ID, &row.Name, &row.FeedURL, &row.Category, &row.Enabled); err != nil {
			t.Fatalf("scan source: %v", err)
		}

		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("iterate sources: %v", err)
	}

	return result
}

// withoutIDs drops the row IDs, which depend on insertion order only.
func withoutIDs(rows []sourceRow) []sourceRow {
	result := make([]sourceRow, len(rows))

	for i, row := range rows {
		row.ID = 0
		result[i] = row
	}

	return result
}

func version(t *testing.T, db *sql.DB) int {
	t.Helper()

	var value int

	if err := db.QueryRow(`PRAGMA user_version`).Scan(&value); err != nil {
		t.Fatalf("read version: %v", err)
	}

	return value
}

func count(t *testing.T, db *sql.DB, query string) int {
	t.Helper()

	var value int

	if err := db.QueryRow(query).Scan(&value); err != nil {
		t.Fatalf("%s: %v", query, err)
	}

	return value
}

func TestMigrateSeedsDefaultSourcesOnAFreshDatabase(t *testing.T) {
	db := openTestDatabase(t)

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if got := withoutIDs(sources(t, db)); !reflect.DeepEqual(got, wantDefaults) {
		t.Fatalf("sources = %+v\nwant %+v", got, wantDefaults)
	}

	if got := version(t, db); got != len(steps) {
		t.Fatalf("version = %d, want %d", got, len(steps))
	}
}

func TestMigrateTwiceDoesNotDuplicateSources(t *testing.T) {
	db := openTestDatabase(t)

	for range 3 {
		if err := Migrate(db); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}

	if got := len(sources(t, db)); got != len(wantDefaults) {
		t.Fatalf("sources = %d, want %d", got, len(wantDefaults))
	}
}

// A default removed from the table stays removed: the seed runs once, not on
// every start.
func TestMigrateDoesNotRestoreARemovedDefault(t *testing.T) {
	db := openTestDatabase(t)

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if _, err := db.Exec(
		`DELETE FROM news_sources WHERE feed_url = ?`,
		"https://feeds.bbci.co.uk/portuguese/rss.xml",
	); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate again: %v", err)
	}

	if got := len(sources(t, db)); got != len(wantDefaults)-1 {
		t.Fatalf("sources = %d, want %d", got, len(wantDefaults)-1)
	}
}

// A box upgraded from the previous release has its sources synced from
// NEWS_SOURCES, articles that reference them, and activities. The seed adds
// only the defaults it lacks and changes nothing it already has.
func TestMigrateUpgradesAnExistingDatabase(t *testing.T) {
	db := openTestDatabase(t)

	createPreviousSchema(t, db)

	// G1 is configured with a name of its own and was last disabled by a
	// sync; Local is a source the defaults do not have.
	if _, err := db.Exec(`
		INSERT INTO news_sources (name, feed_url, category, enabled) VALUES
			('G1 Custom', 'https://g1.globo.com/rss/g1/', 'custom', 0),
			('Local', 'https://local.example/feed', 'general', 1);

		INSERT INTO news_articles (id, source_id, title, url, fetched_at, canonical_url) VALUES
			('g1-1', 1, 'G1 story', 'https://g1.example/1', '2026-10-27T10:00:00Z', 'https://g1.example/1'),
			('local-1', 2, 'Local story', 'https://local.example/1', '2026-10-27T10:00:00Z', 'https://local.example/1');

		INSERT INTO news_refreshes (id, completed_at) VALUES (1, '2026-10-27T10:00:00Z');

		INSERT INTO activities (id, context, type, title, description, timestamp) VALUES
			('activity-1', 'printer', 'print.completed', 'Printed', 'doc.pdf', '2026-10-27T09:00:00Z');
	`); err != nil {
		t.Fatalf("seed previous release: %v", err)
	}

	before := sources(t, db)

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	after := sources(t, db)

	if !reflect.DeepEqual(after[:len(before)], before) {
		t.Fatalf("existing sources changed:\nbefore %+v\nafter  %+v", before, after[:len(before)])
	}

	added := withoutIDs(after[len(before):])

	var wantAdded []sourceRow

	for _, row := range wantDefaults {
		if row.FeedURL != "https://g1.globo.com/rss/g1/" {
			wantAdded = append(wantAdded, row)
		}
	}

	if !reflect.DeepEqual(added, wantAdded) {
		t.Fatalf("added = %+v\nwant %+v", added, wantAdded)
	}

	if got := count(t, db, `SELECT COUNT(*) FROM news_articles WHERE id IN ('g1-1', 'local-1')`); got != 2 {
		t.Fatalf("articles = %d, want 2", got)
	}

	if got := count(t, db, `SELECT COUNT(*) FROM news_refreshes`); got != 1 {
		t.Fatalf("refreshes = %d, want 1", got)
	}

	var title, description string

	if err := db.QueryRow(
		`SELECT title, description FROM activities WHERE id = 'activity-1'`,
	).Scan(&title, &description); err != nil {
		t.Fatalf("activity: %v", err)
	}

	if title != "Printed" || description != "doc.pdf" {
		t.Fatalf("activity = %q %q", title, description)
	}

	if got := version(t, db); got != len(steps) {
		t.Fatalf("version = %d, want %d", got, len(steps))
	}
}

// A step that fails rolls back with its version, so it runs again on the next
// start instead of being recorded as done.
func TestFailedStepLeavesNothingBehind(t *testing.T) {
	db := openTestDatabase(t)

	createPreviousSchema(t, db)

	failing := func(tx *sql.Tx) error {
		if err := news.SeedDefaultSources(tx); err != nil {
			return err
		}

		return errors.New("boom")
	}

	if err := applySteps(db, []func(*sql.Tx) error{failing}); err == nil {
		t.Fatal("expected the step to fail")
	}

	if got := len(sources(t, db)); got != 0 {
		t.Fatalf("sources = %d, want 0", got)
	}

	if got := version(t, db); got != 0 {
		t.Fatalf("version = %d, want 0", got)
	}

	if err := applySteps(db, steps); err != nil {
		t.Fatalf("retry: %v", err)
	}

	if got := len(sources(t, db)); got != len(wantDefaults) {
		t.Fatalf("sources = %d, want %d", got, len(wantDefaults))
	}
}

// A database written by a newer release has steps this one does not know;
// none of the older ones runs again.
func TestMigrateLeavesANewerDatabaseAlone(t *testing.T) {
	db := openTestDatabase(t)

	createPreviousSchema(t, db)

	newer := len(steps) + 1

	if _, err := db.Exec(
		fmt.Sprintf(`PRAGMA user_version = %d`, newer),
	); err != nil {
		t.Fatalf("set version: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if got := len(sources(t, db)); got != 0 {
		t.Fatalf("sources = %d, want 0", got)
	}

	if got := version(t, db); got != newer {
		t.Fatalf("version = %d, want %d", got, newer)
	}
}
