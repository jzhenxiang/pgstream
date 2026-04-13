package pause_test

import (
	"context"
	"testing"
	"time"

	"pgstream/internal/pause"
)

func TestNew_InitiallyRunning(t *testing.T) {
	ctrl := pause.New()
	if ctrl.IsPaused() {
		t.Fatal("expected controller to be running after New")
	}
}

func TestPause_SetsPausedState(t *testing.T) {
	ctrl := pause.New()
	ctrl.Pause()
	if !ctrl.IsPaused() {
		t.Fatal("expected controller to be paused")
	}
}

func TestResume_ClearsPausedState(t *testing.T) {
	ctrl := pause.New()
	ctrl.Pause()
	ctrl.Resume()
	if ctrl.IsPaused() {
		t.Fatal("expected controller to be running after Resume")
	}
}

func TestPause_IsIdempotent(t *testing.T) {
	ctrl := pause.New()
	ctrl.Pause()
	ctrl.Pause() // second call should not panic or deadlock
	if !ctrl.IsPaused() {
		t.Fatal("expected controller to remain paused")
	}
}

func TestResume_IsIdempotent(t *testing.T) {
	ctrl := pause.New()
	ctrl.Resume() // no-op when not paused
	if ctrl.IsPaused() {
		t.Fatal("expected controller to remain running")
	}
}

func TestWait_NotPaused_ReturnsImmediately(t *testing.T) {
	ctrl := pause.New()
	ctx := context.Background()
	if err := ctrl.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWait_ContextCancelled_ReturnsError(t *testing.T) {
	ctrl := pause.New()
	ctrl.Pause()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := ctrl.Wait(ctx); err == nil {
		t.Fatal("expected error when context is cancelled")
	}
}

func TestWait_UnblocksOnResume(t *testing.T) {
	ctrl := pause.New()
	ctrl.Pause()

	done := make(chan error, 1)
	go func() {
		done <- ctrl.Wait(context.Background())
	}()

	time.Sleep(20 * time.Millisecond)
	ctrl.Resume()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error after resume: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Wait did not unblock after Resume")
	}
}
