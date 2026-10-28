package app

import (
	"context"
	"testing"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/config"
)

// isolateDatabase moves the test into an empty directory. New opens the
// database at the relative path ./data/kiosk.db, which would otherwise be
// created, and migrated, inside this package's source directory.
func isolateDatabase(t *testing.T) {
	t.Helper()

	t.Chdir(t.TempDir())
}

func TestAppStopsWhenContextIsCancelled(t *testing.T) {
	isolateDatabase(t)

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
	isolateDatabase(t)

	cfg := config.Default()
	cfg.Display.Mode = "linux"

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	defer application.db.Close()
}
