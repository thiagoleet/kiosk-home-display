package activity

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

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

func (r *SQLiteRepository) Create(
	ctx context.Context,
	activity events.Activity,
) error {
	const query = `
		INSERT INTO activities (
			id,
			context,
			type,
			title,
			description,
			timestamp
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		activity.ID,
		activity.Context,
		activity.Type,
		activity.Title,
		activity.Description,
		activity.Timestamp.UTC().Format(time.RFC3339Nano),
	)

	if err != nil {
		return fmt.Errorf(
			"failed to create activity: %w",
			err,
		)
	}

	return nil
}

func (r *SQLiteRepository) List(
	ctx context.Context,
	limit int,
) ([]events.Activity, error) {
	if limit <= 0 {
		limit = 5
	}

	const query = `
		SELECT
			id,
			context,
			type,
			title,
			description,
			timestamp
		FROM activities
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list activities: %w",
			err,
		)
	}

	defer rows.Close()

	activities := make(
		[]events.Activity,
		0,
		limit,
	)

	for rows.Next() {
		var activity events.Activity
		var timestamp string

		if err := rows.Scan(
			&activity.ID,
			&activity.Context,
			&activity.Type,
			&activity.Title,
			&activity.Description,
			&timestamp,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan activity: %w",
				err,
			)
		}

		activity.Timestamp, err =
			time.Parse(
				time.RFC3339Nano,
				timestamp,
			)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to parse activity timestamp: %w",
				err,
			)
		}

		activities = append(
			activities,
			activity,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate activities: %w",
			err,
		)
	}

	return activities, nil
}
