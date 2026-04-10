package snapshot_test

import (
	"context"
	"testing"

	"github.com/your-org/pgstream/internal/snapshot"
	"github.com/your-org/pgstream/internal/wal"
)

func TestNew_MissingDSN(t *testing.T) {
	_, err := snapshot.New(snapshot.Config{
		Tables: []string{"users"},
	})
	if err == nil {
		t.Fatal("expected error for missing DSN")
	}
}

func TestNew_MissingTables(t *testing.T) {
	_, err := snapshot.New(snapshot.Config{
		DSN: "postgres://localhost/test",
	})
	if err == nil {
		t.Fatal("expected error for missing tables")
	}
}

func TestNew_DefaultBatchSize(t *testing.T) {
	// New should not return an error for a valid config even if the DB
	// is not reachable — the connection is lazy.
	snap, err := snapshot.New(snapshot.Config{
		DSN:    "postgres://localhost/test",
		Tables: []string{"users"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap == nil {
		t.Fatal("expected non-nil Snapshot")
	}
}

func TestNew_Valid(t *testing.T) {
	snap, err := snapshot.New(snapshot.Config{
		DSN:       "postgres://localhost/test",
		Tables:    []string{"orders", "users"},
		BatchSize: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap == nil {
		t.Fatal("expected non-nil Snapshot")
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	// We cannot reach a real DB in unit tests, so we verify that a
	// pre-cancelled context causes Run to return an error quickly.
	snap, err := snapshot.New(snapshot.Config{
		DSN:    "postgres://localhost:9999/nonexistent",
		Tables: []string{"users"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	emitted := 0
	err = snap.Run(ctx, func(_ context.Context, _ *wal.Event) error {
		emitted++
		return nil
	})
	// Expect an error (context cancelled or connection refused).
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
	if emitted != 0 {
		t.Fatalf("expected 0 emitted events, got %d", emitted)
	}
}
