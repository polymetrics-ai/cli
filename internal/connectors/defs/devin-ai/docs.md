# Overview

Devin AI is a wave2 migration of `internal/connectors/devin-ai` (the
hand-written legacy connector this bundle migrates; the legacy package stays
registered and unchanged until wave6's registry flip). It reads Devin
sessions, session insights, session messages, playbooks, and secret metadata
through the Devin v3 REST API's organization-scoped list endpoints. It is
read-only: Devin has no reverse-ETL write surface legacy modeled, so
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Auth setup

Provide `api_token` as a secret: a Devin service-user API key (`cog_...`),
sent as `Authorization: Bearer <api_token>` — identical to legacy's
`connsdk.Bearer(secret)` construction. Provide `org_id` (required, non-secret
config): every stream reads `/v3/organizations/{org_id}/...`, matching
legacy's `devinPath(orgID, resource)` path construction.

`base_url` defaults to `https://api.devin.ai` (materialized via `spec.json`'s
`"default"`, matching legacy's `devinDefaultBaseURL` in-code fallback); an
override must be an absolute `http`/`https` URL with a host — the engine's
path-traversal and same-origin guards apply exactly as documented in
conventions.md §3, matching legacy's own `devinBaseURL` scheme/host
validation intent.

## Streams notes

All 5 streams share Devin's uniform list-endpoint envelope: `GET` against an
org-scoped resource, records extracted from the response's top-level `items`
array, cursor pagination via `pagination.type: cursor` with
`token_path: end_cursor` and `stop_path: has_next_page` (the next page's
`after` query param is read from the response body's `end_cursor`, but
pagination stops as soon as `has_next_page` is falsy regardless of whether
`end_cursor` is still populated) — this reproduces legacy's own stop rule
(`hasNext != "true" || strings.TrimSpace(endCursor) == ""` —
`devinai.go:202` — either condition alone is sufficient to stop).

`sessions`, `sessions_insights`, and `session_messages` are session-derived
and declare `incremental.cursor_field: created_at` /
`incremental.request_param: created_after` / `incremental.start_config_key:
start_date`, matching legacy's `incrementalLowerBound` (state cursor, falling
back to `start_date` config, empty for a full sync) forwarded as the
`created_after` query param. `playbooks` and `secrets` are full-refresh
metadata streams with no `incremental` block, matching legacy exactly (no
`created_after` sent for either).

Every stream's `first` query param is wired from `config.page_size` via the
optional-query dialect (`{"template": "{{ config.page_size }}", "default":
"100"}`), defaulting to 100 when unset — matching legacy's
`devinPageSize` (default 100, max 200).

## Write actions & risks

None. Legacy `devin-ai` is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `metadata.json` declares
`capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Devin v3 API surface (session creation, attachments, knowledge base
  management) is out of scope for wave2; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries. Only the 5 legacy-parity read streams are
  implemented.
- **Legacy's `mode: fixture` credential-free affordance is NOT part of this
  bundle.** Legacy's `readFixture`/`fixtureMode` (`devinai.go:212-252`) emit
  synthetic records without any network call when `config.mode ==
  "fixture"` — this is a legacy-only testing convenience, not part of the
  live record shape; this bundle's own `fixtures/` directory (recorded-shape,
  sanitized) is the wave2 substitute used by `conformance`'s dynamic
  (fixture-replay) checks, matching the wave1-pilot convention.
- **`max_pages` is not part of this bundle's config surface (documented
  scope narrowing, matching searxng's F6 precedent).** Legacy accepts a
  `max_pages` config value, but the engine's `PaginationSpec.MaxPages` is a
  static integer field with no template support (conventions.md §3's
  pagination table) — there is no way to wire a runtime `config.max_pages`
  value into it, so declaring the property would be dead config a caller
  could set with no effect (F6, conventions.md). `spec.json` therefore does
  not declare `max_pages`; a caller wanting to bound total pages read can
  still do so externally (e.g. capping the sync scheduler), and no emitted
  record shape differs for any accepted input.
- All fixtures (`fixtures/streams/**`, `fixtures/check.json`) represent
  Devin's real wire shape, including `has_next_page`/`end_cursor` exactly as
  the API returns them (a JSON boolean and a `null` cursor on the final
  page, not stringified or omitted).
