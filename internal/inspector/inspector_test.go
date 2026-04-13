package inspector_test

import (
	"testing"

	"github.com/your-org/pgstream/internal/inspector"
	"github.com/your-org/pgstream/internal/lsn"
)

func TestNew_InitialisesDefaults(t *testing.T) {
	insp := inspector.New()
	snap := insp.Snapshot()
	if snap.Received != 0 || snap.Processed != 0 || snap.Failed != 0 {
		t.Fatal("expected zero counters")
	}
	if !snap.SinkHealthy {
		t.Fatal("expected sink healthy by default")
	}
	if !snap.LastLSN.IsZero() {
		t.Fatal("expected zero LSN")
	}
}

func TestRecordReceived_IncrementsAndTracksLSN(t *testing.T) {
	insp := inspector.New()
	l := lsn.MustParse("0/1A2B3C")
	insp.RecordReceived(l)
	snap := insp.Snapshot()
	if snap.Received != 1 {
		t.Fatalf("expected 1 received, got %d", snap.Received)
	}
	if snap.LastLSN != l {
		t.Fatalf("expected LSN %s, got %s", l, snap.LastLSN)
	}
}

func TestRecordReceived_DoesNotRegressLSN(t *testing.T) {
	insp := inspector.New()
	high := lsn.MustParse("0/FFFFFF")
	low := lsn.MustParse("0/000001")
	insp.RecordReceived(high)
	insp.RecordReceived(low)
	snap := insp.Snapshot()
	if snap.LastLSN != high {
		t.Fatalf("LSN should not regress: got %s", snap.LastLSN)
	}
}

func TestRecordProcessed_Increments(t *testing.T) {
	insp := inspector.New()
	insp.RecordProcessed()
	insp.RecordProcessed()
	if insp.Snapshot().Processed != 2 {
		t.Fatal("expected 2 processed")
	}
}

func TestRecordFailed_MarksSinkUnhealthy(t *testing.T) {
	insp := inspector.New()
	insp.RecordFailed()
	snap := insp.Snapshot()
	if snap.Failed != 1 {
		t.Fatal("expected 1 failed")
	}
	if snap.SinkHealthy {
		t.Fatal("expected sink unhealthy after failure")
	}
}

func TestMarkSinkHealthy_RestoresFlag(t *testing.T) {
	insp := inspector.New()
	insp.RecordFailed()
	insp.MarkSinkHealthy()
	if !insp.Snapshot().SinkHealthy {
		t.Fatal("expected sink healthy after MarkSinkHealthy")
	}
}

func TestSnapshot_IsCopy(t *testing.T) {
	insp := inspector.New()
	s1 := insp.Snapshot()
	insp.RecordProcessed()
	s2 := insp.Snapshot()
	if s1.Processed == s2.Processed {
		t.Fatal("snapshots should be independent copies")
	}
}
