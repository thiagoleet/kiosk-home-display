package http

import (
	"context"
	"fmt"
	"log"
	nethttp "net/http"

	"github.com/thiagoleet/kiosk-home-display/internal/websocket"
)

type Server struct {
	server *nethttp.Server
}

func NewServer(
	host string,
	port int,
	websocketServer *websocket.Server,
) *Server {
	mux := nethttp.NewServeMux()

	mux.Handle(
		"/ws",
		websocketServer.Handler(),
	)

	mux.HandleFunc(
		"/health",
		healthHandler,
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
