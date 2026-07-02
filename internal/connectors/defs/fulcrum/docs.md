# Overview

Fulcrum is a mobile data-collection platform. This bundle reads Fulcrum forms, records, projects,
choice lists, and classification sets through the Fulcrum REST API v2
(`https://api.fulcrumapp.com/api/v2`), migrating `internal/connectors/fulcrum` (the hand-written
legacy connector, which stays registered and unchanged until wave6's registry flip) to a
declarative bundle at capability parity. Fulcrum is read-only here — no write actions.

## Auth setup

Provide a Fulcrum REST API v2 token via the `api_key` secret. It is sent as the `X-ApiToken`
header (`auth: [{"mode": "api_key_header", "header": "X-ApiToken", "value": "{{ secrets.api_key
}}"}]`) and is never logged.

## Streams notes

All 5 streams (`forms`, `records`, `projects`, `choice_lists`, `classification_sets`) share the
same shape: `GET /<resource>.json`, records at the top-level `<resource>` array key, primary key
`["id"]`, cursor field `["updated_at"]` (informational — Fulcrum has no server-side incremental
filter parameter in this API, so no `incremental` block is declared; a full sync runs every time,
matching legacy exactly).

Pagination is Fulcrum's page-number convention: `page`/`per_page` query params, `page_size: 100`
(matches legacy's default `page_size` config value). Legacy's own stop condition compares the
response body's `current_page` against `total_pages` (in addition to an empty-page check); the
engine's `page_number` paginator only supports the short-page stop rule (`recordCount <
page_size`). See Known limits for why this is a data-safe deviation, not a behavior change.

## Write actions & risks

None. Fulcrum is exposed read-only (`capabilities.write: false`), matching legacy's
`Capabilities{Write: false}` and its `Write` method returning `connectors.ErrUnsupportedOperation`
unconditionally.

## Known limits

- **Pagination stop-condition deviation (data-safe)**: legacy stops when `current_page >=
  total_pages` OR a page returns zero records. The engine's `page_number` pagination type has no
  `stop_path`-equivalent body-field check (`internal/connectors/engine/paginate.go`'s
  `PageNumberPaginator.Next` stops only on a short page, `recordCount < PageSize`). For any real
  Fulcrum response, the two stop conditions coincide except in the corner case where the last page
  happens to be exactly full (a multiple-of-100 total record count): legacy would stop
  immediately (current==total), while the engine would issue one additional request that Fulcrum
  answers with a genuinely empty page, which the engine's short-page check (`0 < 100`) then stops
  on. This produces at most one extra, harmless, empty-page HTTP request — it never omits,
  duplicates, or reorders any emitted record, so it does not change accepted-input behavior under
  the conventions.md §5 meta-rule. Declaring a `page_number`+`stop_path`-body-field combination is
  not supported by the current dialect (only the `cursor` pagination type has `stop_path`); closing
  this fully would be an `ENGINE_GAP` (extending `page_number` with an optional `stop_path`), not
  something this bundle can express today.
- No incremental sync: Fulcrum's list endpoints have no documented server-side "updated since"
  filter, matching legacy (`InitialState` always returns an empty cursor; `CursorFields` on the
  legacy stream catalog are informational only, not enforced). `client_filtered` incremental was
  considered and rejected: adding client-side cursor filtering here would be new behavior legacy
  never had, not a migration.
- `page_size`/`max_pages` are declared in `spec.json` for documentation/discoverability (matching
  legacy's accepted config surface) but are NOT wired into `streams.json`'s `base.pagination`
  block as runtime-config-driven overrides — the engine's `PaginationSpec.PageSize` is a static
  bundle-authored integer with no per-request config lookup mechanism (the same gap documented for
  searxng's `page_size`/`max_pages` in `docs/migration/conventions.md`). The bundle bakes in
  `page_size: 100`, matching legacy's own default. `max_pages` has no bundle-level cap declared
  (Fulcrum's total record counts are typically small); the engine has no config-driven
  `max_pages` wiring at all for the declarative path today.
