package events

type Type string

const (
	EventIdleTimeout Type = "idle.timeout"

	EventScheduleOn  Type = "schedule.on"
	EventScheduleOff Type = "schedule.off"

	EventPrinterStarted   Type = "printer.started"
	EventPrinterCompleted Type = "printer.completed"

	EventDisplayWake         Type = "display.wake"
	EventDisplaySleep        Type = "display.sleep"
	EventDisplayStateChanged Type = "display.state_changed"

	EventNotification Type = "notification"

	EventActivity Type = "activity"
)

type Event struct {
	Type Type
	Data any
}
