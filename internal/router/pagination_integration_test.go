package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/pgstream/internal/router"
)

func TestIntegration_Pagination_EndToEnd(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := router.ParsePageParams(r)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{
			"page":   p.Page,
			"size":   p.Size,
			"offset": p.Offset(),
		})
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/?page=2&size=25")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if body["page"] != 2 {
		t.Errorf("expected page 2, got %d", body["page"])
	}
	if body["size"] != 25 {
		t.Errorf("expected size 25, got %d", body["size"])
	}
	if body["offset"] != 25 {
		t.Errorf("expected offset 25, got %d", body["offset"])
	}
}

func TestIntegration_Pagination_DefaultsOnMissingParams(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := router.ParsePageParams(r)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"page": p.Page, "size": p.Size})
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if body["page"] != 1 {
		t.Errorf("expected default page 1, got %d", body["page"])
	}
	if body["size"] != 20 {
		t.Errorf("expected default size 20, got %d", body["size"])
	}
}
