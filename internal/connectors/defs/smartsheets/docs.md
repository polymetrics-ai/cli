# Overview

Smartsheets is migrated as a Tier-2 bundle (declarative streams.json/spec.json/schemas plus one
hook interface). Legacy (`internal/connectors/smartsheets/smartsheets.go`) is a pure `connsdk`-HTTP
connector — it builds a `connsdk.Requester` with Bearer auth and issues plain `GET` requests
against the Smartsheet REST API v2; there is no SQL/queue/SDK protocol, no non-declarative auth
flow, and no write path (`Write` always returns `connectors.ErrUnsupportedOperation`). Despite the
wave2 catalog inventory mislabeling this connector's `runtime_kind` as `native_go`, legacy is a
plain HTTP connector, not a native-protocol one. The `sheets` stream is fully declarative; only
`sheet_rows` needs a Go hook (`hooks/smartsheets/hooks.go`, a single `StreamHook`, ~180 lines,
well under the Tier-2 line cap) because its per-row cell flattening depends on a sibling
`columns[]` array carried in the SAME page body as the rows themselves — see Known limits and
Streams notes below for exactly why this cannot be expressed with `records`/`computed_fields`
alone, and why this does not escalate past Tier 2 (a single hook interface, no auth/write
involvement, no connection-lifecycle management).

## Auth setup

Provide a Smartsheet API access token via the `access_token` secret; it is used only for Bearer
auth (`Authorization: Bearer <access_token>`) and is never logged. `base_url` defaults to
`https://api.smartsheet.com/2.0`, matching legacy's own default exactly. `spreadsheet_id` (the
sheet or report ID to read rows from) is required only for the `sheet_rows` stream, matching
legacy's own per-read validation (`sheet_rows` errors with a clear message when unset; `sheets`
needs no sheet ID at all).

## Streams notes

- **`sheets`** — `GET /sheets`, page-number pagination (`type: page_number`, `page_param: page`,
  `size_param: pageSize`, `start_page: 1`), records at `data`, `projection: passthrough` (legacy's
  own `record()` copies every raw field verbatim — the catalog's declared
  `id`/`name`/`permalink`/`modifiedAt` fields are a documented SUBSET, not the actual emitted
  shape, so schema-mode projection would silently narrow legacy's real output). Legacy stops
  pagination when `page >= totalPages`; the engine's `page_number` paginator stops on a short page
  (`recordCount < pageSize`) instead. These agree in every case except the rare boundary where the
  last page is exactly full — see Known limits (same class as the jamf-pro parity-ledger entry in
  `conventions.md` §5, item 13: at most one harmless extra request, never a data change). The
  `pageSize` value itself is a fixed 100 (the engine's default page size, matching legacy's own
  fallback default), not config-driven — see Known limits.
- **`sheet_rows`** — `GET /sheets/{spreadsheet_id}?include=rows`, StreamHook-handled (see Overview).
  `hooks/smartsheets/hooks.go`'s `readRows` ports `smartsheets.go`'s `readRows`/`rowRecord`/
  `columnsByID` verbatim: each page decodes once, builds a `columnId -> title` map from that same
  page's `columns[]` array, and flattens every row's `cells[]` into dynamically-named top-level
  fields (one per sheet column, keyed by the column's title — e.g. `"Name"` — or `cell_<columnId>`
  when a column has no title), alongside the fixed `sheet_id`/`sheet_name`/`row_id`/`row_number`/
  `modified_at`/`cells` fields the declarative schema declares. No incremental filtering is
  modeled: `modified_at` is published as `x-cursor-field` for manifest-surface parity (matching
  legacy's own `CursorFields: []string{"modified_at"}`), but neither connector actually filters or
  advances reads by it — every read is a full sweep of every row on every sync, exactly matching
  legacy (§8 rule 2 of `conventions.md`: bare CursorFields with no server-side filter or
  client-side comparison means no `incremental` block is declared).

## Write actions & risks

None. Legacy's `Write` always returns `connectors.ErrUnsupportedOperation`; this bundle declares
`capabilities.write: false` and ships no `writes.json`, matching legacy exactly (the wave2 catalog
inventory's `runtime_kind: native_go` label is a mislabel — see Overview).

## Known limits

- **`sheet_rows`' dynamic per-column flattening is an `ENGINE_GAP` for the declarative dialect
  specifically (not a blocker: expressed via the sanctioned Tier-2 `StreamHook` seam instead).**
  `computed_fields` can only ever populate a fixed, statically-declared set of output field names;
  it has no primitive for "look up a per-record foreign key (`cell.columnId`) against a sibling
  array carried elsewhere in the SAME page body (`columns[]`), and use the looked-up value ITSELF
  as an output field NAME" — the real column set (and therefore the real per-record field set) is
  only known at read time, per sheet, never statically at bundle-author time. `records.keyed_object`
  does not apply either (it explodes a keyed OBJECT's values into one record per key; here the
  shape needing resolution is an array of `{columnId, value}` cell objects against a same-page
  sibling lookup table, not a keyed record collection). If 3+ additional wave connectors need this
  exact "flatten an array using a sibling lookup table for field names" shape, the `ENGINE_GAP`
  recurrence rule (`conventions.md` §6) would promote this to a real engine feature; this is
  occurrence #1.
- **The `sheet_rows` declarative `streams.json` entry is never live-dispatched** while a
  `smartsheets` hook set is registered — it carries a `conformance.skip_dynamic` marker naming
  `hooks/smartsheets/hooks_test.go` as the authoritative substitute; conformance's dynamic
  (fixture-replay) checks skip this stream outright rather than exercising a declarative shape
  that could never reproduce the column-flatten behavior. The declarative `records: {path: "rows"}`
  entry is kept so the bundle stays uniform (every stream still declares a full shape, per
  `conventions.md` §1's Tier-2 rule) and so a future engine extension closing this gap has a
  ready-made declarative shape to switch back to.
- **`sheets`' short-page pagination stop can issue one harmless extra request** versus legacy's
  exact `page >= totalPages` stop, on the boundary where the final page is exactly full-sized. See
  Streams notes above and `conventions.md` §5 item 13 (jamf-pro) for the identical, already-accepted
  reasoning.
- **`sheets`' `page_size` config override is not modeled.** Legacy's `pageQuery` helper reads
  `config.page_size` for BOTH streams (falling back to 100 when unset/invalid); this bundle's
  `sheet_rows` hook honors it (`hooks/smartsheets/hooks.go`'s `pageSize`), but the declarative
  `sheets` stream's `page_number` pagination spec is a fixed literal in `streams.json` with no
  config-templated page-size knob (mirrors the aircall/bitly `next_url`
  page-size-is-not-config-driven limitation already ledgered in `conventions.md`) — `sheets` always
  requests `pageSize=100` regardless of a configured `page_size`, matching legacy's own default
  behavior for any caller who never overrode it.
- **`access_token`'s legacy multi-source fallback (`credentials.access_token` in either Secrets or
  Config) is not modeled.** This bundle declares only the primary `secrets.access_token` key (the
  first candidate legacy's own `accessToken` helper tries); a caller relying solely on the
  undocumented `credentials.access_token` fallback key must switch to `access_token`.
- **Legacy's `mode: fixture` credential-free affordance is NOT part of this bundle**, matching the
  monday-pilot precedent (`conventions.md`-referenced `docs/migration` note): legacy's
  `readFixture`/`fixtureMode` (`smartsheets.go:197-213,295-297`) emit synthetic records without any
  network call when `config.mode == "fixture"` — a legacy-only testing convenience, not part of the
  live record shape. Parity is asserted against legacy's LIVE (`httptest`-driven) read path only.
- **`cells`' declared schema type is `["array","null"]`, not legacy's catalog-declared
  `"object"`.** Legacy's own `Catalog()` (`smartsheets.go:63`) types `cells` as `"object"`, but the
  actual emitted value (`row["cells"]`, copied verbatim in `rowRecord`) is the raw JSON ARRAY of
  cell objects — a pre-existing inconsistency in legacy's own catalog declaration, not introduced by
  this migration. This bundle's schema type matches the REAL emitted data shape rather than
  reproducing legacy's own aspirational/incorrect catalog type.
