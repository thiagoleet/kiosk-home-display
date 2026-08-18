package events

import "testing"

func TestBusPublishesEvent(t *testing.T) {
	bus := NewBus()

	called := false

	bus.Subscribe(EventNotification, func(event Event) {
		called = true
	})

	bus.Publish(Event{
		Type: EventNotification,
		Data: "Hello",
	})

	if !called {
		t.Errorf("Expected handler to be called, but it was not")
	}
}

func TestBusDoesNotCallHandlerForDifferentEvent(t *testing.T) {
	bus := NewBus()

	called := false

	bus.Subscribe(EventNotification, func(event Event) {
		called = true
	})

	bus.Publish(Event{
		Type: EventPrinterCompleted,
	})

	if called {
		t.Fatal("handler should not have been called")
	}
}
