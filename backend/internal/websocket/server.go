package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type Server struct {
	bus *events.Bus

	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewServer(bus *events.Bus) *Server {
	return &Server{
		bus:     bus,
		clients: make(map[*Client]struct{}),
	}
}

func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(s.handleConnection)
}

func (s *Server) handleConnection(
	w http.ResponseWriter,
	r *http.Request,
) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("websocket connection failed: %v", err)
		return
	}

	client := NewClient(conn)

	s.addClient(client)
	client.Start()

	defer func() {
		s.removeClient(client)
		client.Close()
	}()

	log.Println("WebSocket client connected")

	<-r.Context().Done()

	log.Println("WebSocket client disconnected")
}

func (s *Server) addClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[client] = struct{}{}
}

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, client)
}

func (s *Server) Start() {
	s.bus.Subscribe(events.EventIdleTimeout, s.handleEvent)
	s.bus.Subscribe(events.EventScheduleOn, s.handleEvent)
	s.bus.Subscribe(events.EventScheduleOff, s.handleEvent)

	s.bus.Subscribe(events.EventDisplayWake, s.handleEvent)
	s.bus.Subscribe(events.EventDisplaySleep, s.handleEvent)

	s.bus.Subscribe(events.EventPrinterStarted, s.handleEvent)
	s.bus.Subscribe(events.EventPrinterCompleted, s.handleEvent)

	s.bus.Subscribe(events.EventNotification, s.handleEvent)
}

func (s *Server) handleEvent(event events.Event) {
	message := Message{
		Type: string(event.Type),
		Data: event.Data,
	}

	s.broadcast(message)
}

func (s *Server) broadcast(message Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for client := range s.clients {
		client.Send(message)
	}
}
