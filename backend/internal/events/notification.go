package events

type NotificationLevel string

const (
	NotificationInfo    NotificationLevel = "info"
	NotificationSuccess NotificationLevel = "success"
	NotificationWarning NotificationLevel = "warning"
	NotificationError   NotificationLevel = "error"
)

type Notification struct {
	Title   string            `json:"title"`
	Message string            `json:"message"`
	Level   NotificationLevel `json:"level"`
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
