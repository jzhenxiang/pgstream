// Package validator implements rule-based validation for WAL change events.
//
// A Validator is configured with a set of Rules, each targeting a specific
// table. When an event arrives the Validator looks up the matching rule and
// checks that every column listed in RequiredColumns is present and non-nil in
// the event payload.
//
// Events whose table has no matching rule are passed through unchanged,
// making the validator opt-in on a per-table basis.
//
// Example usage:
//
//	v := validator.New(validator.Config{
//		Rules: []validator.Rule{
//			{Table: "public.orders", RequiredColumns: []string{"id", "status"}},
//		},
//	})
//	if err := v.Validate(event); err != nil {
//		// handle validation failure
//	}
package validator
