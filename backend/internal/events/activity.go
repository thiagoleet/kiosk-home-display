package events

import (
	"time"

	"github.com/google/uuid"
)

type Activity struct {
	ID          string              `json:"id"`
	Context     NotificationContext `json:"context"`
	Type        Type                `json:"type"`
	Title       string              `json:"title"`
	Description string              `json:"description,omitempty"`
	Timestamp   time.Time           `json:"timestamp"`
}

func NewActivity(
	context NotificationContext,
	eventType Type,
	title string,
	description string,
) Activity {
	return Activity{
		ID:          uuid.NewString(),
		Context:     context,
		Type:        eventType,
		Title:       title,
		Description: description,
		Timestamp:   time.Now(),
	}
}

func PublishActivity(
	bus *Bus,
	activity Activity,
) {
	bus.Publish(Event{
		Type: EventActivity,
		Data: activity,
	})
}
