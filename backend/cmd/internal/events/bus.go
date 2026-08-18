package events

import "sync"

type Handler func(Event)

type Bus struct {
	mu       sync.RWMutex
	handlers map[Type][]Handler
}

func NewBus() *Bus {
	return &Bus{
		handlers: make(map[Type][]Handler),
	}
}

func (b *Bus) Subscribe(eventType Type, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *Bus) Publish(event Event) {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[event.Type]...)
	b.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}
