// Package filter implements table-level allow/deny filtering for WAL events.
//
// A Filter can be configured with an allow list, a deny list, or both.
// When an allow list is provided, only tables explicitly listed are processed.
// When a deny list is provided, any matching table is skipped regardless of
// the allow list. Deny rules always take precedence.
//
// Table names are matched case-insensitively in "schema.table" format.
//
// Example usage:
//
//	f := filter.New(filter.Config{
//		AllowTables: []string{"public.orders", "public.products"},
//		DenyTables:  []string{"public.audit_log"},
//	})
//
//	if f.Allow(event.Schema, event.Table) {
//		// process event
//	}
package filter
