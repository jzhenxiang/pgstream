package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/pgstream/internal/router"
)

func TestIntegration_VersionEndpoint_RegisteredOnRouter(t *testing.T) {
	hc := newHC()
	info := &router.BuildInfo{
		Version:   "0.9.0",
		Commit:    "deadbeef",
		BuildTime: "2024-06-01T12:00:00Z",
	}

	mux, err := router.New(hc, &router.Options{Version: info})
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/version")
	if err != nil {
		t.Fatalf("GET /version: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got router.BuildInfo
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Version != "0.9.0" {
		t.Errorf("version mismatch: got %q", got.Version)
	}
	if got.Commit != "deadbeef" {
		t.Errorf("commit mismatch: got %q", got.Commit)
	}
}
