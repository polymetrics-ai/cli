# Overview

NinjaOne RMM is a read-only connector for the NinjaOne v2 remote monitoring and management REST
API. This bundle migrates `internal/connectors/ninjaone-rmm` (the hand-written legacy connector) to
a declarative Tier-1 bundle at full capability parity: it reads the same five streams
(`organizations`, `devices`, `locations`, `activities`, `policies`), using the same Bearer auth and
NinjaOne's after-cursor pagination. The legacy package stays registered and unchanged until the
wave6 registry flip.

## Auth setup

Provide a NinjaOne RMM API token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(secret)`. Never logged.

## Streams notes

`organizations`, `devices`, `locations`, and `activities` share an identical shape: `GET` against
the matching NinjaOne v2 endpoint, records at the response root (a bare top-level JSON array —
`records.path: "."`, the engine's documented root-selector spelling; a plain `""` also selects root
for record extraction via `connsdk.RecordsAt`/`selectPath`, but the `cursor`+`last_record_field`
paginator's own last-record lookup (`lastRecordCursor.Next`) additionally treats an EMPTY
`recordsPath` string as "unset" and silently substitutes `"data"` — which does not exist on a bare
top-level array and would silently stop pagination after page 1. `"."` avoids that fallback so the
paginator finds the last record's `id` from the true root array, while extraction behaves
identically either way), pagination via NinjaOne's `after=<last entity id>` convention
(`pagination.type: cursor` with `last_record_field: id`, `cursor_param: after`) — the next page's
`after` value is the numeric `id` of the last record on the current page, matching legacy's
`harvest` loop. `policies` is unpaginated (legacy's `endpoint.paginated == false` branch): it
overrides `pagination: {"type": "none"}` at the stream level and returns its full set in one
request.

Every record field that legacy renames from NinjaOne's camelCase wire shape to the emitted
snake_case name (`organizationId` -> `organization_id`, `nodeApprovalMode` -> `node_approval_mode`,
etc.) uses a `computed_fields` bare single-reference rename (e.g. `"organization_id": "{{
record.organizationId }}"`), which the engine's typed-extraction rule applies to: the raw JSON
value's native type (integer/boolean) is preserved, not stringified, matching legacy's
`connectors.Record` map assignment exactly (`item["organizationId"]` copied verbatim, no
`fmt.Sprintf`).

`activities`' `activityTime` is declared as `x-cursor-field` for manifest-surface parity with
legacy's `CursorFields: []string{"activityTime"}`, but neither legacy nor this bundle actually
filters or advances reads by it (no incremental request param exists in `ninjaone-rmm.go`) — both
connectors always perform a full stream read.

## Write actions & risks

None. NinjaOne RMM is exposed read-only (no safe reverse-ETL write actions in this stream set);
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`ErrUnsupportedOperation` Write stub.

## Known limits

- **Pagination stop signal is empty-page-only, not short-page.** Legacy's `harvest` loop stops
  pagination as soon as a page returns FEWER records than `page_size` (a "short page" — the typical
  final-page signal for an API that always fills full pages until the last one) OR an empty page.
  The engine's `cursor`+`last_record_field` paginator (`lastRecordCursor`) has no `page_size`-aware
  short-page check at all — it stops only when a page returns zero records, or (if declared)
  `stop_path` is falsy, which this bundle does not declare (NinjaOne has no boolean has-more
  indicator to key one off). Net effect: on a genuinely short-but-nonempty final page, this bundle
  issues exactly one additional request past what legacy would have made, which correctly returns
  an empty page and then stops — never a data or ordering difference, never a duplicate or dropped
  record, for any input legacy itself would accept. This is an efficiency-only deviation (one extra
  HTTP round trip at the tail of a sync), not an `emitted-record-data` deviation, so it is
  ACCEPTABLE per `docs/migration/conventions.md` §5's meta-rule; the engine dialect has no
  configurable short-page stop threshold for the `last_record_field` cursor variant to close this
  gap tighter today.
- `activityTime` (`activities`' cursor field) is declared for catalog/manifest parity only; neither
  connector filters or advances reads by it. Matches legacy's actual (non-)behavior exactly, not a
  narrowing.
- **`max_pages` is not runtime-configurable.** Legacy accepts a `max_pages` config override, but the
  engine's `PaginationSpec.MaxPages` is static and cannot be templated from config. The bundle
  leaves pagination unbounded, matching legacy's default, and does not declare a dead `max_pages`
  config key.
