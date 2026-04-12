// Package normalizer provides lightweight, rule-based field normalization
// for WAL events in the pgstream pipeline.
//
// Rules are matched by table name (or wildcard "*") and column name.
// Supported normalization modes:
//
//   - lowercase  – converts string values to lower case
//   - uppercase  – converts string values to upper case
//   - trim       – removes leading and trailing whitespace
//   - trimspace  – alias for trim
//
// The Normalizer never mutates the original event; it returns a cloned
// copy with the transformed fields applied.
package normalizer
