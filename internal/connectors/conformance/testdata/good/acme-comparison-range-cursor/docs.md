# Overview

Acme Comparison Range Cursor is a synthetic connector used as a conformance v2 self-test bundle
for generic `rfc3339_utc` plus `operator_prefix` incremental formatting. Its incremental cursor
field is a JSON NUMBER on the wire (Unix seconds), so the max-observed cursor `cursor_advances`
re-reads with is a bare digit string — exactly the app-persisted cursor shape
(`internal/app/sync_modes.go` `recordCursor` -> `toComparableString`) that the engine normalizes to
a UTC RFC3339 comparison qualifier.

## Auth setup

No auth required; public synthetic API.

## Streams notes

`events` is incremental on `created` (a Unix-seconds integer) and has no pagination.

## Write actions & risks

None; read-only bundle.

## Known limits

None; this is test fixture data.
