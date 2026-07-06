# Overview

Acme GitHub Range Cursor is a synthetic connector used as a conformance v2 self-test bundle for
the `github_date_range` `param_format` (N1, wave0 REVIEW.md re-review): its incremental cursor
field is a JSON NUMBER on the wire (Unix seconds), so the max-observed cursor
`cursor_advances` re-reads with is a bare digit string — exactly the app-persisted cursor
shape (`internal/app/sync_modes.go` `recordCursor` -> `toComparableString`) that the engine's
`formatParam` normalizes to a UTC RFC3339 `>=` qualifier for `github_date_range`. Locks in that
`formatCursorForAssertion`'s `github_date_range` branch performs the SAME digits->RFC3339
normalization as the engine, rather than returning `">=" + value` verbatim.

## Auth setup

No auth required; public synthetic API.

## Streams notes

`events` is incremental on `created` (a Unix-seconds integer) and has no pagination.

## Write actions & risks

None; read-only bundle.

## Known limits

None; this is test fixture data.
