package database

import (
	"database/sql"
	"fmt"
)

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

	return nil
}
