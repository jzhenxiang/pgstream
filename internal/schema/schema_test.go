package schema

import (
	"fmt"
	"sync"
	"testing"
)

func sampleSchema(schema, table string) *TableSchema {
	return &TableSchema{
		Schema: schema,
		Table:  table,
		Columns: []Column{
			{Name: "id", Type: "int4", Position: 1},
			{Name: "name", Type: "text", Position: 2},
		},
	}
}

func TestNew_ReturnsEmptyCache(t *testing.T) {
	c := New()
	if c.Len() != 0 {
		t.Fatalf("expected 0 entries, got %d", c.Len())
	}
}

func TestSet_And_Get(t *testing.T) {
	c := New()
	ts := sampleSchema("public", "users")
	c.Set(ts)

	got, ok := c.Get("public", "users")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got.Key() != "public.users" {
		t.Fatalf("unexpected key: %s", got.Key())
	}
	if len(got.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(got.Columns))
	}
}

func TestSet_NilIsNoOp(t *testing.T) {
	c := New()
	c.Set(nil)
	if c.Len() != 0 {
		t.Fatal("nil Set should not add an entry")
	}
}

func TestGet_MissingEntry(t *testing.T) {
	c := New()
	_, ok := c.Get("public", "missing")
	if ok {
		t.Fatal("expected miss for unknown table")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	c := New()
	c.Set(sampleSchema("public", "orders"))
	c.Delete("public", "orders")
	_, ok := c.Get("public", "orders")
	if ok {
		t.Fatal("entry should have been deleted")
	}
}

func TestTableSchema_Key(t *testing.T) {
	ts := &TableSchema{Schema: "myschema", Table: "mytable"}
	expected := "myschema.mytable"
	if ts.Key() != expected {
		t.Fatalf("expected %q, got %q", expected, ts.Key())
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ts := sampleSchema("public", fmt.Sprintf("table%d", i))
			c.Set(ts)
			c.Get("public", fmt.Sprintf("table%d", i))
		}(i)
	}
	wg.Wait()
	if c.Len() == 0 {
		t.Fatal("expected entries after concurrent writes")
	}
}
