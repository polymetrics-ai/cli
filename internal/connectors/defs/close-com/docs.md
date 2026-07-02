# Overview

Close.com (Close CRM) is a wave2 fan-out declarative-HTTP migration. It reads Close leads,
contacts, opportunities, activities, and users through the Close REST API
(`GET https://api.close.com/api/v1/...`). This bundle targets capability parity with
`internal/connectors/close-com` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Close API key via the `api_key` secret; it is sent as the HTTP Basic auth username with
an empty password (`"mode": "basic", "username": "{{ secrets.api_key }}", "password": ""`),
matching legacy's `connsdk.Basic(secret, "")` exactly (`close_com.go:242`). It is never logged.
`base_url` defaults to `https://api.close.com/api/v1` and may be overridden for tests/proxies
(legacy's own `closeBaseURL` validates scheme+host the same way; the engine's base-URL resolution
has no equivalent runtime validation, but every conformance fixture only ever points at an
httptest server, so this is not exercised differently on either side).

## Streams notes

All five streams (`leads`, `contacts`, `opportunities`, `activities`, `users`) are simple list
endpoints under Close's singular, trailing-slashed resource paths (`/lead/`, `/contact/`,
`/opportunity/`, `/activity/`, `/user/`), records at the top-level `data` key. Pagination is
Close's `_skip`/`_limit` offset convention (`pagination.type: offset_limit`, `limit_param: _limit`,
`offset_param: _skip`, `page_size: 100` matching legacy's `closeDefaultPageSize`) — the engine's
`offset_limit` paginator stops on a short page (fewer than `page_size` records), which is
DATA-equivalent to legacy's own `has_more != "true" || len(records)==0` stop rule (`close_com.go:
170-176`) for every real Close response shape: a full page always means more records remain, a
short page always means none do (see Known limits for the one theoretical edge case).

Legacy declares a `CursorFields: ["date_updated"]` on every stream (`streams.go:35` etc.) and an
`InitialState` returning an empty starting cursor (`close_com.go:98-106`), but never actually wires
`date_updated` into any request parameter or client-side filter anywhere in the harvest loop
(`close_com.go:144-180`) — every sync is a full, unfiltered `_skip`/`_limit` walk of the entire
resource regardless of any prior cursor. Because there is no real incremental FILTERING behavior
to port (only a schema-level field that happens to look like a cursor), this bundle declares no
`incremental` block and no `x-cursor-field` on any schema, honestly representing legacy's actual
behavior rather than its unused declared intent. `date_updated` remains a plain schema property.

## Write actions & risks

None. Close is read-only in this connector (legacy's own package doc: "no obviously-safe
reverse-ETL writes... it exposes no reverse-ETL writes, so Capabilities.Write is false");
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **Declared `CursorFields`/`InitialState` are not modeled as an `incremental` block.** See Streams
  notes above — legacy declares the scaffolding for incremental sync (`CursorFields`,
  `InitialState`) but the harvest loop never actually filters by it, so porting only the unused
  declaration (without any request-side behavior) would create a schema-level `x-cursor-field`
  that promises incremental-append sync-mode support the connector cannot deliver. Omitting it is
  the honest representation of legacy's real behavior; `date_updated` remains queryable as a plain
  field for any downstream consumer that wants to filter post-sync.
- **The engine's `offset_limit` paginator stops on a short PAGE, not on legacy's own `has_more`
  flag.** These two stop rules are DATA-equivalent for every real Close response shape (Close's own
  `has_more` flag directly tracks whether the requested page was full) — the one theoretical edge
  case (the true last page happens to contain EXACTLY `page_size` records, with `has_more: false`)
  still terminates correctly on both sides: legacy stops immediately on `has_more: false`, while the
  engine's paginator would issue one further request that returns zero records and stop via the
  same short/empty-page rule — an extra request, never an extra or missing RECORD.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`closePageSize`/`closeMaxPages`, `close_com.go:275-303`). The engine's `offset_limit`
  paginator's `PageSize`/`MaxPages` fields are plain JSON values in `streams.json`, not templated
  against `config.*` — there is no mechanism in this dialect to wire a runtime config value into
  either field. This bundle ships legacy's own default (`page_size: 100`, `max_pages` unbounded) as
  a static value.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance,
  `close_com.go:185-225`) stamps extra fields onto every fixture-mode record — `connector` (a
  static "close-com" marker), `stream` (which stream the record came from), and `previous_cursor`
  (echoing `req.State["cursor"]` when set) — none of which are part of the live record shape. This
  bundle's schemas and fixtures target the live path only; the engine's own conformance/fixture-
  replay harness provides the credential-free test affordance this bundle needs.
