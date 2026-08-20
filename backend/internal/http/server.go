package http

import (
	"context"
	"fmt"
	"log"
	nethttp "net/http"

	"github.com/thiagoleet/kiosk-home-display/internal/activity"
	"github.com/thiagoleet/kiosk-home-display/internal/display"
	"github.com/thiagoleet/kiosk-home-display/internal/events"
	"github.com/thiagoleet/kiosk-home-display/internal/i18n"
	"github.com/thiagoleet/kiosk-home-display/internal/printer"
	"github.com/thiagoleet/kiosk-home-display/internal/websocket"
)

type Server struct {
	server *nethttp.Server
}

func NewServer(
	host string,
	port int,
	bus *events.Bus,
	websocketServer *websocket.Server,
	displayManager *display.Manager,
	printerManager *printer.Manager,
	activityRepository activity.Repository,
	texts i18n.Catalog,
	allowedOrigins []string,
) *Server {
	mux := nethttp.NewServeMux()

	displayHandler := NewDisplayHandler(
		displayManager,
	)

	notificationHandler := NewNotificationHandler(bus, texts)

	printerHandler := NewPrinterHandler(
		printerManager,
	)

	activityHandler := NewActivityHandler(
		activityRepository,
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

	mux.HandleFunc(
		"/api/notifications/test",
		notificationHandler.Test,
	)

	mux.HandleFunc(
		"/api/printer/print",
		printerHandler.Print,
	)

	mux.HandleFunc(
		"/api/activities",
		activityHandler.List,
	)

	return &Server{
		server: &nethttp.Server{
			Addr:    fmt.Sprintf("%s:%d", host, port),
			Handler: corsMiddleware(allowedOrigins, mux),
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
