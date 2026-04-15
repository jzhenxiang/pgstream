package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePageParams_Defaults(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	p := ParsePageParams(r)
	if p.Page != defaultPage {
		t.Fatalf("expected page %d, got %d", defaultPage, p.Page)
	}
	if p.Size != defaultPageSize {
		t.Fatalf("expected size %d, got %d", defaultPageSize, p.Size)
	}
}

func TestParsePageParams_Custom(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page=3&size=50", nil)
	p := ParsePageParams(r)
	if p.Page != 3 {
		t.Fatalf("expected page 3, got %d", p.Page)
	}
	if p.Size != 50 {
		t.Fatalf("expected size 50, got %d", p.Size)
	}
}

func TestParsePageParams_CapsSize(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?size=9999", nil)
	p := ParsePageParams(r)
	if p.Size != maxPageSize {
		t.Fatalf("expected size capped at %d, got %d", maxPageSize, p.Size)
	}
}

func TestParsePageParams_InvalidFallsBack(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page=abc&size=-1", nil)
	p := ParsePageParams(r)
	if p.Page != defaultPage {
		t.Fatalf("expected default page, got %d", p.Page)
	}
	if p.Size != defaultPageSize {
		t.Fatalf("expected default size, got %d", p.Size)
	}
}

func TestPageParams_Offset(t *testing.T) {
	cases := []struct {
		page, size, want int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 10, 20},
		{0, 20, 0},
	}
	for _, c := range cases {
		p := PageParams{Page: c.page, Size: c.size}
		if got := p.Offset(); got != c.want {
			t.Errorf("Offset() page=%d size=%d: got %d, want %d", c.page, c.size, got, c.want)
		}
	}
}

func TestParsePageParams_ZeroPage_FallsBack(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page=0", nil)
	p := ParsePageParams(r)
	if p.Page != defaultPage {
		t.Fatalf("expected default page for zero value, got %d", p.Page)
	}
}
