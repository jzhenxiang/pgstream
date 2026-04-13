package router_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/pgstream/internal/router"
)

func TestIntegration_RateLimit_ConcurrentRequests(t *testing.T) {
	const (
		maxReqs    = 10
		goroutines = 20
		ip         = "10.0.0.99:1234"
	)

	mw := router.WithRateLimit(5*time.Second, maxReqs)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var (
		allowed  int64
		blocked  int64
		wg      sync.WaitGroup
	)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = ip
			h.ServeHTTP(rec, req)
			switch rec.Code {
			case http.StatusOK:
				atomic.AddInt64(&allowed, 1)
			case http.StatusTooManyRequests:
				atomic.AddInt64(&blocked, 1)
			}
		}()
	}

	wg.Wait()

	if got := atomic.LoadInt64(&allowed); got != maxReqs {
		t.Errorf("expected %d allowed, got %d", maxReqs, got)
	}
	if got := atomic.LoadInt64(&blocked); got != goroutines-maxReqs {
		t.Errorf("expected %d blocked, got %d", goroutines-maxReqs, got)
	}
}
