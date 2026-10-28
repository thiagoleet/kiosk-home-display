package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Database struct {
	*sql.DB
}

func Open(config Config) (*Database, error) {
	if err := ensureDirectory(config.Path); err != nil {
		return nil, fmt.Errorf(
			"failed to create database directory: %w",
			err,
		)
	}

	// The pool holds several connections to the same file, and SQLite refuses
	// a write outright while another connection is reading unless it is told
	// to wait. Without the timeout an activity or a news refresh that lands
	// during a read fails with SQLITE_BUSY instead of waiting a few ms.
	db, err := sql.Open(
		"sqlite",
		config.Path+"?_pragma=busy_timeout(5000)",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to open database: %w",
			err,
		)
	}

	if err := db.Ping(); err != nil {
		db.Close()

		return nil, fmt.Errorf(
			"failed to connect to database: %w",
			err,
		)
	}

	return &Database{
		DB: db,
	}, nil
}

func ensureDirectory(
	path string,
) error {
	directory := filepath.Dir(path)

	return os.MkdirAll(
		directory,
		0755,
	)
}
