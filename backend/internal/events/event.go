package events

type Type string

const (
	EventIdleTimeout Type = "idle.timeout"

	EventScheduleOn  Type = "schedule.on"
	EventScheduleOff Type = "schedule.off"

	EventPrinterStarted   Type = "printer.started"
	EventPrinterCompleted Type = "printer.completed"

	EventDisplayWake  Type = "display.wake"
	EventDisplaySleep Type = "display.sleep"

	EventNotification Type = "notification"
)

type Event struct {
	Type Type
	Data any
}
