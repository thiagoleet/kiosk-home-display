package activity

import (
	"context"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type Repository interface {
	Create(
		ctx context.Context,
		activity events.Activity,
	) error

	List(
		ctx context.Context,
		limit int,
	) ([]events.Activity, error)

	DeleteOlderThan(
		ctx context.Context,
		before time.Time,
	) error
}
