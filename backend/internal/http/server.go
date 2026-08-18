package http

import (
	"context"
	"fmt"
	"log"
	nethttp "net/http"

	"github.com/thiagoleet/kiosk-home-display/internal/display"
	"github.com/thiagoleet/kiosk-home-display/internal/websocket"
)

type Server struct {
	server *nethttp.Server
}

func NewServer(
	host string,
	port int,
	websocketServer *websocket.Server,
	displayManager *display.Manager,
) *Server {
	mux := nethttp.NewServeMux()

	displayHandler := NewDisplayHandler(
		displayManager,
	)

	mux.Handle(
		"/ws",
		websocketServer.Handler(),
	)

	mux.HandleFunc(
		"/health",
		healthHandler,
	)

	mux.HandleFunc(
		"/api/display/sleep",
		displayHandler.Sleep,
	)

	mux.HandleFunc(
		"/api/display/wake",
		displayHandler.Wake,
	)

	mux.HandleFunc(
		"/api/display/brightness",
		displayHandler.Brightness,
	)

	return &Server{
		server: &nethttp.Server{
			Addr:    fmt.Sprintf("%s:%d", host, port),
			Handler: mux,
		},
	}
}

func healthHandler(
	w nethttp.ResponseWriter,
	r *nethttp.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(nethttp.StatusOK)

	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) Start() error {
	log.Printf(
		"HTTP server listening on %s",
		s.server.Addr,
	)

	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
