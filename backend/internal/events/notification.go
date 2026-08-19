package events

import "github.com/google/uuid"

type NotificationLevel string

const (
	NotificationInfo    NotificationLevel = "info"
	NotificationSuccess NotificationLevel = "success"
	NotificationWarning NotificationLevel = "warning"
	NotificationError   NotificationLevel = "error"
)

type Notification struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Message  string            `json:"message"`
	Level    NotificationLevel `json:"level"`
	Duration int               `json:"duration"`
}

func PublishNotification(
	bus *Bus,
	notification Notification,
) {
	bus.Publish(Event{
		Type: EventNotification,
		Data: notification,
	})
}

func NewNotification(
	title string,
	message string,
	level NotificationLevel,
	duration int,
) Notification {
	return Notification{
		ID:       uuid.NewString(),
		Title:    title,
		Message:  message,
		Level:    level,
		Duration: duration,
	}
}
