// Package connsdk provides reusable building blocks for hand-written connectors:
// an HTTP requester with retry/rate-limit handling, pluggable authenticators,
// paginators, response extraction, schema inference, and incremental cursor
// helpers.
//
// connsdk is a leaf package: it deliberately does NOT import the parent
// connectors package, so any per-system connector package (github, stripe, ...)
// can depend on both connsdk and connectors without creating an import cycle.
//
// Records are plain map[string]any, which is the exact underlying type of
// connectors.Record, so converting at the connector boundary is free:
//
//	for _, rec := range pageRecords { emit(connectors.Record(rec)) }
package connsdk

// Record is a single emitted row. Its underlying type matches connectors.Record.
type Record = map[string]any
