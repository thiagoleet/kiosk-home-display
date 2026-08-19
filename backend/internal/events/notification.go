package events

import "github.com/google/uuid"

type NotificationContext string

const (
	NotificationContextPrinter NotificationContext = "printer"
	NotificationContextSystem  NotificationContext = "system"
	NotificationContextNetwork NotificationContext = "network"
	NotificationContextDisplay NotificationContext = "display"
)

type NotificationLevel string

const (
	NotificationInfo    NotificationLevel = "info"
	NotificationSuccess NotificationLevel = "success"
	NotificationWarning NotificationLevel = "warning"
	NotificationError   NotificationLevel = "error"
)

type Notification struct {
	ID       string              `json:"id"`
	Context  NotificationContext `json:"context"`
	Title    string              `json:"title"`
	Message  string              `json:"message"`
	Level    NotificationLevel   `json:"level"`
	Duration int                 `json:"duration"`
}

func NewNotification(
	context NotificationContext,
	title string,
	message string,
	level NotificationLevel,
	duration int,
) Notification {
	return Notification{
		ID:       uuid.NewString(),
		Context:  context,
		Title:    title,
		Message:  message,
		Level:    level,
		Duration: duration,
	}
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
