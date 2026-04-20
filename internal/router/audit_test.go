package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandlerAudit(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithAudit_NilSink_IsNoOp(t *testing.T) {
	h := WithAudit(nil)(http.HandlerFunc(okHandlerAudit))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithAudit_RecordsEntry(t *testing.T) {
	sink := NewInMemoryAuditSink(10).(*inMemoryAuditSink)
	h := WithAudit(sink)(http.HandlerFunc(okHandlerAudit))
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	entries := sink.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", e.Method)
	}
	if e.Path != "/ping" {
		t.Errorf("expected /ping, got %s", e.Path)
	}
	if e.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", e.StatusCode)
	}
}

func TestWithAudit_RecordsStatusCode(t *testing.T) {
	sink := NewInMemoryAuditSink(10).(*inMemoryAuditSink)
	notFound := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	h := WithAudit(sink)(notFound)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/missing", nil))

	if got := sink.Entries()[0].StatusCode; got != http.StatusNotFound {
		t.Errorf("expected 404, got %d", got)
	}
}

func TestInMemoryAuditSink_EvictsOldest(t *testing.T) {
	sink := NewInMemoryAuditSink(3).(*inMemoryAuditSink)
	h := WithAudit(sink)(http.HandlerFunc(okHandlerAudit))
	for i := 0; i < 5; i++ {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if got := len(sink.Entries()); got != 3 {
		t.Errorf("expected 3 entries after eviction, got %d", got)
	}
}

func TestWithAudit_Entries_ReturnsCopy(t *testing.T) {
	sink := NewInMemoryAuditSink(10).(*inMemoryAuditSink)
	h := WithAudit(sink)(http.HandlerFunc(okHandlerAudit))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	a := sink.Entries()
	a[0].Path = "/mutated"
	b := sink.Entries()
	if b[0].Path == "/mutated" {
		t.Error("Entries should return a copy, not a reference")
	}
}
