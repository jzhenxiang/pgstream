// Package redactor implements field-level value redaction for WAL events.
//
// Rules are matched by table name (exact or wildcard "*") and one of three
// strategies is applied to each matching column value:
//
//   - blank   – replaces the value with an empty string (default)
//   - hash    – replaces the value with a truncated SHA-256 hex digest
//   - partial – masks all but the first and last characters with asterisks
//
// Example usage:
//
//	cfg := redactor.Config{
//		Rules: []redactor.Rule{
//			{Table: "users", Columns: []string{"email", "phone"}, Strategy: redactor.StrategyHash},
//			{Table: "*",     Columns: []string{"ssn"},            Strategy: redactor.StrategyBlank},
//		},
//	}
//	r := redactor.New(cfg)
//	redacted := r.Apply(event)
package redactor
