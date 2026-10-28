package database

import (
	"database/sql"
	"fmt"

	"github.com/thiagoleet/kiosk-home-display/internal/news"
)

// steps change data, and must run exactly once each, after the schema below
// exists. PRAGMA user_version counts how many have been applied, so a database
// created before steps existed reads 0 and runs them all, and a new step is
// appended to the end — never inserted or reordered.
//
// The schema itself stays idempotent and runs on every start; only what
// cannot be repeated safely belongs here.
var steps = []func(tx *sql.Tx) error{
	// 1: the default news sources.
	news.SeedDefaultSources,
}

func Migrate(db *sql.DB) error {
	const query = `
		CREATE TABLE IF NOT EXISTS activities (
			id TEXT PRIMARY KEY,
			context TEXT NOT NULL,
			type TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			timestamp TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_activities_timestamp
		ON activities(timestamp DESC);
	`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf(
			"failed to migrate database: %w",
			err,
		)
	}

	if err := news.Migrate(db); err != nil {
		return fmt.Errorf(
			"failed to migrate news tables: %w",
			err,
		)
	}

	if err := applySteps(db, steps); err != nil {
		return fmt.Errorf(
			"failed to migrate database data: %w",
			err,
		)
	}

	return nil
}

// applySteps runs the steps the database has not seen yet, each in its own
// transaction together with the version that records it, so a step that fails
// leaves neither its changes nor its version behind and runs again on the
// next start. A database already past the last step, written by a newer
// release, is left alone.
func applySteps(
	db *sql.DB,
	steps []func(tx *sql.Tx) error,
) error {
	var version int

	if err := db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fmt.Errorf(
			"failed to read schema version: %w",
			err,
		)
	}

	for index := version; index < len(steps); index++ {
		if err := applyStep(db, steps[index], index+1); err != nil {
			return fmt.Errorf(
				"step %d: %w",
				index+1,
				err,
			)
		}
	}

	return nil
}

func applyStep(
	db *sql.DB,
	step func(tx *sql.Tx) error,
	version int,
) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf(
			"failed to begin: %w",
			err,
		)
	}

	defer tx.Rollback()

	if err := step(tx); err != nil {
		return err
	}

	// PRAGMA takes no bound parameters; version is an int from this package.
	if _, err := tx.Exec(
		fmt.Sprintf(`PRAGMA user_version = %d`, version),
	); err != nil {
		return fmt.Errorf(
			"failed to record schema version: %w",
			err,
		)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"failed to commit: %w",
			err,
		)
	}

	return nil
}
