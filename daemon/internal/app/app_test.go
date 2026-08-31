package app

import (
	"context"
	"testing"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/config"
)

func TestAppStopsWhenContextIsCancelled(t *testing.T) {
	application, err := New(config.Default())
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)

	go func() {
		done <- application.Run(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

	case <-time.After(time.Second):
		t.Fatal("application did not stop after context cancellation")
	}
}

func TestAppStartsWithLinuxDisplayMode(t *testing.T) {
	cfg := config.Default()
	cfg.Display.Mode = "linux"

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	defer application.db.Close()
}
