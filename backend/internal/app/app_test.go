package app

import (
	"context"
	"testing"
	"time"
)

func TestAppStopsWhenContextIsCancelled(t *testing.T) {
	application := New()

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
