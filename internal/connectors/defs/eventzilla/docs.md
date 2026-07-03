# Overview

Eventzilla is a fan-out declarative-HTTP migration. It reads Eventzilla events, categories,
users, attendees, and ticket types through the Eventzilla v2 REST API (`GET
https://www.eventzillaapi.net/api/v2/...`). This bundle now covers all 5 streams
`internal/connectors/eventzilla` (the hand-written connector) implements; the legacy package stays
registered and unchanged until wave6's registry flip. `attendees` and `tickets` — previously
blocked (`ENGINE_GAP`; see `docs/migration/status.json`'s `partial[]` entry) — are now expressed
via the engine's `fan_out` dialect (S4 engine mini-wave item 2). Eventzilla is read-only: legacy
exposes no reverse-ETL writes, so `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Auth setup

Provide the Eventzilla API key via the `api_key` secret; it is sent as the `x-api-key` request
header (`{"mode":"api_key_header","header":"x-api-key","value":"{{ secrets.api_key }}"}`), matching
legacy's `connsdk.APIKeyHeader(eventzillaAPIKeyHeader, secret, "")` exactly (`eventzilla.go`'s
`requester`, `eventzillaAPIKeyHeader = "x-api-key"`). `base_url` defaults to
`https://www.eventzillaapi.net/api/v2` and may be overridden for tests/proxies.

## Streams notes

`events`, `categories`, and `users` are top-level list endpoints (`GET /events`, `/categories`,
`/users`); records live at the resource-named JSON key. Eventzilla returns `{"<field>":[...]}`
with no total/has_more marker, so pagination stops on a short page
(`pagination.type: offset_limit`, `limit_param: limit`, `offset_param: offset`), matching legacy's
`harvest` function exactly (`len(records) < pageSize` stop condition). `page_size` is declared as
a fixed literal (`100`) in `base.pagination`, matching legacy's own default
(`eventzillaDefaultPageSize = 100` in `eventzilla.go`, also the config-bounds cap) — the engine's
`offset_limit` paginator has no config-driven page-size override mechanism (see Known limits), so
this bundle bakes in legacy's default rather than legacy's config-driven override range.

`categories`' primary key is `["category"]` (the category name string itself), matching legacy's
own `PrimaryKey: []string{"category"}` — Eventzilla's category list has no separate id field.

`attendees` and `tickets` are per-event sub-resource fan-out reads, matching legacy's
`readSubstream` (`eventzilla.go:141-160`) exactly: `fan_out.ids_from.request` issues a paginated
`GET /events` sequence (reusing the `attendees`/`tickets` stream's own effective pagination — the
base `offset_limit` block, since neither stream declares its own override), extracting `id_field:
id` off every discovered event record; `into.path_var: "event_id"` threads each discovered event id
into `/events/{{ fanout.id }}/attendees` (or `/tickets`); `stamp_field: "event_id"` writes it onto
every emitted child record after projection — reproducing legacy's `stampParent` (which only fills
`event_id` when the raw record itself omits it; the real Eventzilla API always sends `event_id` on
every attendee/ticket record, so the fan-out stamp's unconditional overwrite lands on the exact
same value legacy's raw record already carried — never a data change for any real API response).
Both streams' `event_id` schema property is typed `["integer", "string"]`: the raw wire value is a
JSON integer, but the fan_out engine's stamp always writes the fan-out id as a Go string (S4 engine
mini-wave item 2, `read.go`'s `fc.id`) — see Known limits for why this is a documented, ACCEPTABLE
parity deviation rather than a silent narrowing.

## Write actions & risks

None. Eventzilla is a read-only source in legacy (`eventzilla.go`'s package doc: "Eventzilla
exposes only full-refresh reads, so the connector is read-only"); `capabilities.write` is `false`
and no `writes.json` is shipped.

## Known limits

- **`attendees`/`tickets`'s `event_id` field is typed `["integer", "string"]`, not a bare
  `integer`.** Legacy's raw Eventzilla API always sends `event_id` as a JSON integer on both
  streams; the engine's `fan_out.stamp_field` mechanism always overwrites that field with the
  fan-out id as a Go **string** (`read.go`'s `fc.id`), applied unconditionally after projection —
  there is no dialect option to stamp a typed (non-string) value. This never changes the emitted
  VALUE for any real Eventzilla response (the stamped string is the same event id legacy's own
  `stampParent` fallback would have used, and the real API always already carries the matching
  integer on the raw record, which the stamp simply overwrites with the string form of the same
  id) — only the JSON **type** of the field differs from legacy's plain-integer schema. ACCEPTABLE
  parity deviation (§5): documented here rather than silently narrowed.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (default
  100, capped at 100) and `max_pages` (0/all/unlimited = unbounded) as config-driven overrides
  (`eventzillaPageSize`/`eventzillaMaxPages` in `eventzilla.go`). The engine's `offset_limit`
  paginator has no config-driven page-size or max-pages knob (`PaginationSpec.PageSize` is a fixed
  literal read once at bundle-load time, not a per-request template), so this bundle declares a
  fixed `page_size: 100` in `base.pagination` (legacy's own default/cap value,
  `eventzillaDefaultPageSize`) and does not declare `page_size`/`max_pages` in `spec.json` at all (a
  declared-but-unwireable config key is worse than an absent one, per conventions.md F6
  precedent). Pagination is bounded only by Eventzilla's own short-page stop signal, matching
  Eventzilla's real termination behavior.
- **Legacy's fixture-mode-only synthetic fields are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) stamps a broad synthetic superset of fields
  shared across all 5 streams' mappers (including `attendees`/`tickets`-only fields), which is not
  the live wire shape any single migrated stream actually returns. This bundle's schemas and
  fixtures target the LIVE record shape only (`eventzilla.go`'s `harvest`/`mapRecord` functions),
  per the bitly-pilot precedent (`docs/migration/conventions.md`'s worked example): the engine's
  own fixture-replay conformance harness supersedes the need for an in-connector fixture-mode
  branch.
