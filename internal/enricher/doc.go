// Package enricher provides a metadata enrichment stage for the pgstream
// pipeline. It attaches configurable static fields, the processing hostname,
// and an enrichment timestamp to every WAL event, enabling downstream
// consumers to correlate events with their origin and processing time without
// requiring changes to the WAL reader or sink layers.
package enricher
