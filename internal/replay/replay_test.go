package replay_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pgstream/pgstream/internal/offset"
	"github.com/pgstream/pgstream/internal/replay"
	"github.com/pgstream/pgstream/internal/wal"
)

type mockReader struct {
	msgs []*wal.Message
	idx  int
	err  error
}

func (m *mockReader) Read(_ context.Context) (*wal.Message, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.idx >= len(m.msgs) {
		return nil, context.Canceled
	}
	msg := m.msgs[m.idx]
	m.idx++
	return msg, nil
}

type mockSink struct {
	sent  []*wal.Event
	sendErr error
}

func (m *mockSink) Send(_ context.Context, e *wal.Event) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, e)
	return nil
}

func newTestOffset(t *testing.T) *offset.Offset {
	t.Helper()
	dir := t.TempDir()
	off, err := offset.New(offset.Config{FilePath: filepath.Join(dir, "offset")})
	if err != nil {
		t.Fatalf("offset.New: %v", err)
	}
	return off
}

func TestNew_NilReader_ReturnsError(t *testing.T) {
	off := newTestOffset(t)
	_, err := replay.New(nil, &mockSink{}, off, nil)
	if err == nil {
		t.Fatal("expected error for nil reader")
	}
}

func TestNew_NilSink_ReturnsError(t *testing.T) {
	off := newTestOffset(t)
	_, err := replay.New(&mockReader{}, nil, off, nil)
	if err == nil {
		t.Fatal("expected error for nil sink")
	}
}

func TestNew_NilOffset_ReturnsError(t *testing.T) {
	_, err := replay.New(&mockReader{}, &mockSink{}, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil offset")
	}
}

func TestNew_Valid(t *testing.T) {
	off := newTestOffset(t)
	r, err := replay.New(&mockReader{}, &mockSink{}, off, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil replayer")
	}
}

func TestRun_SendsEventsAndCommitsOffset(t *testing.T) {
	event := &wal.Event{Table: "users"}
	reader := &mockReader{
		msgs: []*wal.Message{
			{LSN: "0/1", Event: event},
		},
	}
	sink := &mockSink{}
	off := newTestOffset(t)

	r, _ := replay.New(reader, sink, off, nil)
	err := r.Run(context.Background())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if len(sink.sent) != 1 {
		t.Fatalf("expected 1 event sent, got %d", len(sink.sent))
	}
}

func TestRun_SinkError_ReturnsError(t *testing.T) {
	reader := &mockReader{
		msgs: []*wal.Message{{LSN: "0/1", Event: &wal.Event{}}},
	}
	sink := &mockSink{sendErr: errors.New("send failed")}
	off := newTestOffset(t)

	r, _ := replay.New(reader, sink, off, nil)
	err := r.Run(context.Background())
	if err == nil {
		t.Fatal("expected error from sink")
	}
}

func TestRun_ContextCancelled_StopsImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	off := newTestOffset(t)
	r, _ := replay.New(&mockReader{}, &mockSink{}, off, nil)
	err := r.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRun_ReadError_ReturnsError(t *testing.T) {
	reader := &mockReader{err: errors.New("read failed")}
	off := newTestOffset(t)

	r, _ := replay.New(reader, &mockSink{}, off, nil)
	err := r.Run(context.Background())
	if err == nil {
		t.Fatal("expected error from reader")
	}
}

func TestConfig_Validate_MissingOffsetFile(t *testing.T) {
	c := &replay.Config{}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for missing offset_file")
	}
}

func TestConfig_Validate_Valid(t *testing.T) {
	c := &replay.Config{OffsetFile: filepath.Join(os.TempDir(), "offset")}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
