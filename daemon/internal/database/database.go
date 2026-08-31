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

	db, err := sql.Open("sqlite", config.Path)
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
