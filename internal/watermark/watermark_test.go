package watermark_test

import (
	"testing"

	"github.com/your-org/pgstream/internal/watermark"
)

func TestNew_InitialConfirmed(t *testing.T) {
	wm := watermark.New(100)
	if got := wm.Confirmed(); got != 100 {
		t.Fatalf("expected confirmed=100, got %d", got)
	}
}

func TestLSN_String(t *testing.T) {
	lsn := watermark.LSN(0x0000000100000001)
	if s := lsn.String(); s == "" {
		t.Fatal("expected non-empty LSN string")
	}
}

func TestTrack_IncreasesPendingCount(t *testing.T) {
	wm := watermark.New(0)
	wm.Track(10)
	wm.Track(20)
	if got := wm.PendingCount(); got != 2 {
		t.Fatalf("expected pending=2, got %d", got)
	}
}

func TestConfirm_AdvancesWhenNoPending(t *testing.T) {
	wm := watermark.New(0)
	wm.Track(50)
	wm.Confirm(50)
	if got := wm.Confirmed(); got != 50 {
		t.Fatalf("expected confirmed=50, got %d", got)
	}
	if got := wm.PendingCount(); got != 0 {
		t.Fatalf("expected pending=0, got %d", got)
	}
}

func TestConfirm_DoesNotAdvancePastPending(t *testing.T) {
	wm := watermark.New(0)
	wm.Track(10)
	wm.Track(20)
	// Confirm the higher one first — confirmed should stay below 10.
	wm.Confirm(20)
	if got := wm.Confirmed(); got >= 20 {
		t.Fatalf("expected confirmed < 20 while lsn 10 is pending, got %d", got)
	}
}

func TestConfirm_AdvancesAfterAllPendingCleared(t *testing.T) {
	wm := watermark.New(0)
	wm.Track(10)
	wm.Track(20)
	wm.Confirm(20)
	wm.Confirm(10)
	if got := wm.Confirmed(); got != 20 {
		t.Fatalf("expected confirmed=20, got %d", got)
	}
}

func TestConfirm_OutOfOrderDelivery(t *testing.T) {
	wm := watermark.New(0)
	for _, lsn := range []watermark.LSN{1, 2, 3, 4, 5} {
		wm.Track(lsn)
	}
	// Confirm out of order.
	wm.Confirm(3)
	wm.Confirm(5)
	wm.Confirm(1)
	wm.Confirm(4)
	wm.Confirm(2)
	if got := wm.Confirmed(); got != 5 {
		t.Fatalf("expected confirmed=5, got %d", got)
	}
	if got := wm.PendingCount(); got != 0 {
		t.Fatalf("expected pending=0, got %d", got)
	}
}
