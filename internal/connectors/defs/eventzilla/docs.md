# Overview

Eventzilla is a wave2 fan-out declarative-HTTP migration. It reads Eventzilla events, categories,
and users through the Eventzilla v2 REST API (`GET https://www.eventzillaapi.net/api/v2/...`).
This bundle migrates 3 of the 5 streams `internal/connectors/eventzilla` (the hand-written
connector) implements; the legacy package stays registered and unchanged until wave6's registry
flip. Eventzilla is read-only: legacy exposes no reverse-ETL writes, so `capabilities.write` is
`false` and this bundle ships no `writes.json`.

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

## Write actions & risks

None. Eventzilla is a read-only source in legacy (`eventzilla.go`'s package doc: "Eventzilla
exposes only full-refresh reads, so the connector is read-only"); `capabilities.write` is `false`
and no `writes.json` is shipped.

## Known limits

- **`attendees` and `tickets` are NOT migrated in this wave — reported as a blocked capability, not
  silently dropped.** Legacy implements both as per-event sub-resource fan-out reads
  (`eventzilla.go`'s `readSubstream`): list every event via `/events`, then issue one additional
  request per discovered event id to `/events/{event_id}/attendees` (or `/tickets`), stamping the
  parent `event_id` onto every child record. This is one of `docs/migration/conventions.md` §1's
  explicitly named Tier-2 `StreamHook` triggers ("sub-resource fan-out reads, e.g. issue → comments
  per issue") — the declarative dialect has no mechanism to fan a list read out over another
  stream's discovered ids, and this wave is JSON-only (no hooks packages permitted). Implementing
  these 2 streams requires a follow-up wave to add `internal/connectors/hooks/eventzilla/hooks.go`
  with a `StreamHook.ReadStream` override (see `api_surface.json`'s `excluded` entries for both
  endpoints). `events`, `categories`, and `users` — the 3 streams with no sub-resource fan-out
  requirement — are fully migrated at parity in this bundle.
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
