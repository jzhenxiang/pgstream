// Package projector provides column-level projection for WAL events.
//
// A Projector is configured with a map of table names to column allow-lists.
// When Apply is called on an event whose table matches a rule, only the listed
// columns are retained in the returned event copy. Tables without a matching
// rule are passed through unchanged.
//
// Example usage:
//
//	p, err := projector.New(projector.Config{
//		Rules: map[string][]string{
//			"public.orders": {"id", "status", "total"},
//		},
//	})
//	out := p.Apply(event)
package projector
