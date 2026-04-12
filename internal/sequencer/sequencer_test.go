package sequencer_test

import (
	"sync"
	"testing"

	"pgstream/internal/sequencer"
	"pgstream/internal/wal"
)

func TestNew_DefaultField(t *testing.T) {
	s := sequencer.New(sequencer.Config{})
	if s == nil {
		t.Fatal("expected non-nil sequencer")
	}
}

func TestNext_NilEvent_ReturnsError(t *testing.T) {
	s := sequencer.New(sequencer.Config{})
	_, err := s.Next(nil)
	if err == nil {
		t.Fatal("expected error for nil event")
	}
}

func TestNext_StampsSequenceNumber(t *testing.T) {
	s := sequencer.New(sequencer.Config{})
	event := &wal.Event{}

	out, err := s.Next(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := out.Metadata["_seq"]
	if !ok {
		t.Fatal("expected _seq key in metadata")
	}
	if val.(uint64) != 1 {
		t.Fatalf("expected seq=1, got %v", val)
	}
}

func TestNext_Increments(t *testing.T) {
	s := sequencer.New(sequencer.Config{})

	for i := uint64(1); i <= 5; i++ {
		out, err := s.Next(&wal.Event{})
		if err != nil {
			t.Fatalf("unexpected error at i=%d: %v", i, err)
		}
		got := out.Metadata["_seq"].(uint64)
		if got != i {
			t.Fatalf("expected %d, got %d", i, got)
		}
	}
}

func TestNext_CustomField(t *testing.T) {
	s := sequencer.New(sequencer.Config{Field: "seq_num"})
	out, _ := s.Next(&wal.Event{})
	if _, ok := out.Metadata["seq_num"]; !ok {
		t.Fatal("expected seq_num key in metadata")
	}
}

func TestCurrent_BeforeAnyNext_ReturnsZero(t *testing.T) {
	s := sequencer.New(sequencer.Config{})
	if s.Current() != 0 {
		t.Fatalf("expected 0, got %d", s.Current())
	}
}

func TestReset_ZerosCounter(t *testing.T) {
	s := sequencer.New(sequencer.Config{})
	s.Next(&wal.Event{}) //nolint:errcheck
	s.Reset()
	if s.Current() != 0 {
		t.Fatalf("expected 0 after reset, got %d", s.Current())
	}
}

func TestConcurrentNext_NoDataRace(t *testing.T) {
	s := sequencer.New(sequencer.Config{})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Next(&wal.Event{}) //nolint:errcheck
		}()
	}
	wg.Wait()
	if s.Current() != 50 {
		t.Fatalf("expected 50, got %d", s.Current())
	}
}
