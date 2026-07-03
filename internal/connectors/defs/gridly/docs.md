# Overview

Gridly is a declarative-HTTP Tier-1 migration. It reads Gridly views, per-view grid records, and
per-view branches through the read-only Gridly REST API v1 (`GET https://api.gridly.com/v1/...`).
This bundle targets full read-parity with `internal/connectors/gridly` (the hand-written connector
it migrates); the legacy package stays registered and unchanged until wave6's registry flip. Legacy
is a pure `connsdk`-based HTTP connector with no custom auth/stream/write hooks, so Tier 1
(declarative bundle, zero Go) applies — see `docs/migration/conventions.md` §1's tier ladder.

## Auth setup

Provide a Gridly API key via the `api_key` secret; it is sent as the `Authorization` header in the
form `ApiKey <key>` (matching legacy's `connsdk.APIKeyHeader("Authorization", token, "ApiKey ")`,
`gridly.go:166`) and is never logged. `base_url` defaults to `https://api.gridly.com/v1` and may be
overridden for tests/proxies.

## Streams notes

- `views` — `GET views`, single unfanned-out stream listing every view in the account, records at
  the response root (`records.path: ""`), primary key `["id"]`. Paginated with `page_number`
  (`page`/`pageSize` query params, matching legacy's `harvest` loop), short-page stop at
  `page_size` (default 100, matching legacy's `defaultPageSize`).
- `records` — `GET views/{{ fanout.id }}/records`, one request-sequence per configured view id
  (`fan_out.ids_from.config_key: view_ids`, forwarded as a `path_var` — legacy's
  `strings.ReplaceAll(endpoint.resource, "{view}", url.PathEscape(viewID))`,
  `gridly.go:103`), records at the response root, primary key `["view_id", "id"]`. The fan-out id
  is stamped onto every record's `view_id` field (`fan_out.stamp_field`), matching legacy's
  `harvest`'s `if viewID != "" { rec["view_id"] = viewID }` injection (`gridly.go:124-126`).
- `branches` — `GET views/{{ fanout.id }}/branches`, identical per-view fan-out shape to `records`,
  primary key `["view_id", "id"]`.

Both fan-out streams require `view_ids` (comma-separated, split/trimmed/empty-entries-dropped,
matching legacy's `splitList`, `gridly.go:225`) — legacy itself hard-errors
(`"gridly %s stream requires config view_ids"`, `gridly.go:99`) when `view_ids` is unset for a
per-view stream, so this is not a scope narrowing, just the same required-input shape expressed
declaratively.

None of the 3 streams are incremental — legacy publishes no `CursorFields` on any stream and the
Gridly v1 API exposes no server-side updated-since filter legacy uses.

## Write actions & risks

None. Legacy's `Write` unconditionally returns `connectors.ErrUnsupportedOperation`
(`gridly.go:256-258`); `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **Per-column `cell_<column_id>` flattened fields are NOT modeled (ACCEPTABLE, documented
  deviation).** Legacy's `gridRecord` (`streams.go:32-46`) walks each record's raw `cells` array
  and, for every cell, derives a sanitized key (`cellKey`: lowercase, non-alphanumeric runs become
  `_`, trimmed, falls back to `"value"` if empty) prefixed with `cell_`, then sets
  `rec[cellKey] = cell["value"]` — i.e. it fans a fixed array into a DYNAMIC, per-record set of
  top-level fields whose very NAMES are derived from data (`cells[i].columnId`), not from any
  static schema. The engine's `computed_fields` dialect (`docs/migration/conventions.md` §3) is a
  map of FIXED output-field-name → template resolved against the raw record — every field name is
  declared once in `streams.json`/known at bundle-load time; there is no primitive to derive a
  field's own NAME from record data at read time (the closest existing dynamic-shape primitive,
  `records.keyed_object`, explodes a keyed OBJECT into multiple RECORDS, not one record's array
  into multiple FIELDS on that same record). Modeling this would need a new engine primitive
  (e.g. a "flatten array `<path>` into `cell_<sanitize(item.columnId)>` = `item.value` fields"
  dialect addition) that does not exist today; per `docs/migration/conventions.md` §6, this is
  filed as `ENGINE_GAP` rather than faked. **The raw `cells` array is preserved verbatim** on every
  `records` stream record (`records.json`'s `cells` field, typed `["array","null"]`, exactly
  mirroring legacy's own `rec["cells"] = item["cells"]` line, `streams.go:33`) — no data is lost,
  only the legacy convenience of per-column top-level fields is not reproduced. A capability-
  expansion pass adding a dedicated flatten primitive to the engine dialect (or, short of that, a
  Tier-2 `RecordHook` — a legitimate escalation since this is genuinely record-shape derivation
  Go, not templating — see `docs/migration/conventions.md` §1's Tier-2 table) would close this.
- **View-id auto-discovery is not modeled.** Legacy requires `view_ids` to be explicitly configured
  for the `records`/`branches` streams (no auto-discovery fallback exists in legacy itself —
  `gridly.go:97-100` hard-errors when unset) — this bundle reproduces that exact required-input
  shape, not a narrowing.
- **`max_pages`'s `all`/`unlimited` string aliases are folded into the plain unbounded (`0`/absent)
  case.** Legacy's `maxPages` (`gridly.go:198-208`) accepts `""`, `"all"`, or `"unlimited"` as
  synonyms for "unbounded" in addition to a literal `0`; the engine's `PaginationSpec.MaxPages` is
  a plain integer with `<= 0` meaning unbounded (`docs/migration/conventions.md` §3) — the
  literal strings `all`/`unlimited` are not themselves parsed by the engine. `max_pages`'s
  `spec.json` description documents which literal values a caller should actually send.
