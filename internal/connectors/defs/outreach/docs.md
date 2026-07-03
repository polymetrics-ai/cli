# Outreach

## Overview

Reads Outreach prospects, accounts, sequences, and mailings through the Outreach API v2
(`https://api.outreach.io/api/v2`), a JSON:API-shaped REST API. Read-only: there is no approved
reverse-ETL write surface for Outreach in pm, matching the legacy `internal/connectors/outreach`
package. This bundle is capability-parity migrated from that legacy connector; the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an Outreach OAuth2 **access token** (`secrets.access_token`); it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) on every request and is never logged. Legacy performs its
own `client_id`/`client_secret`/`refresh_token` OAuth2 refresh-token-grant exchange
(`internal/connectors/outreach/outreach.go`'s `refreshTokenAuth`) to obtain this access token; that
exchange is performed **outside** this connector — only the resolved, already-issued access token is
consumed here. The engine's declarative auth dialect has no `oauth2_refresh_token` mode (only
`oauth2_client_credentials`, a different OAuth2 grant, is declarative), so modeling the refresh-grant
exchange itself would require a Tier-2 hook; narrowing the config surface to a pre-obtained
`access_token` avoids that entirely while leaving every read behavior (streams, pagination, record
shape) byte-identical to legacy. This mirrors `defs/linkedin-ads`'s identical
`access_token`-only narrowing for the same reason — see its `docs.md` Known limits for the same
pattern applied to a different vendor.

## Streams notes

All four streams (`prospects`, `accounts`, `sequences`, `mailings`) share the identical shape
legacy's own generic `jsonAPIRecord` mapper produces: `GET` against the Outreach list endpoint,
records at the JSON:API `data` array, primary key `["id"]`, incremental cursor field `updated_at`.
Every request sends `page[size]=100` (matches legacy's `defaultPageSize`) via each stream's static
`query: {"page[size]": "100"}`. Pagination follows Outreach's own JSON:API `links.next` **absolute**
URL convention (`pagination.type: next_url`, `next_url_path: "links.next"`) — exactly legacy's own
`links.next` follow-until-empty loop (`outreach.go`'s `harvest`), with no `page[size]` requery
needed on subsequent pages since the next URL already carries every needed query parameter, and no
`MaxPages` cap either (legacy's own `maxPages` config existed but this bundle does not need an
equivalent — see Known limits). The engine's `next_url` paginator's same-host SSRF guard
(THREAT-MODEL §3) passes cleanly since Outreach's `links.next` values are always same-origin
absolute URLs pointing back at `api.outreach.io`.

Incremental reads send `filter[updatedAt]` as an RFC3339 value (`param_format` defaults to
`rfc3339`, matching legacy's `lowerBound` which sends the raw RFC3339 string verbatim, never
Unix-seconds), computed either from the sync's persisted cursor or, on a fresh sync, from the
`start_date` config value — identical to legacy's `lowerBound`/`filter[updatedAt]` logic. The
engine sets `filter[updatedAt]` on the request automatically whenever the incremental lower bound
resolves (`streams.json`'s `incremental.request_param`); it is intentionally not also declared in
each stream's static `query` block (matching stripe's `created[gte]` pattern — see
`docs/migration/conventions.md` §3's `{{ incremental.lower_bound }}` note).

Every record's fields are derived via `computed_fields` from the raw JSON:API envelope, mirroring
legacy's `jsonAPIRecord` exactly: `id`/`type` copied from the JSON:API envelope (not
`attributes`), `email`/`name`/`created_at`/`updated_at` read from `attributes.email` /
`attributes.name` / `attributes.createdAt` / `attributes.updatedAt`.

## Write actions & risks

None. Outreach is read-only in pm (`capabilities.write: false`), matching legacy exactly (legacy's
own `Write` always returns `connectors.ErrUnsupportedOperation`).

## Known limits

- **`name` fallback to `displayName` is not modeled**: legacy's generic mapper
  (`outreach.go`'s `first(attrs, "name", "displayName")`) falls back to `attributes.displayName`
  when `attributes.name` is absent, for any of the four resource types. The engine's
  `computed_fields` dialect has no coalesce/fallback-chain filter (only a single bare
  `{{ record.<path> }}` reference, a filter chain, or a mixed literal template — see
  `docs/migration/conventions.md` §3's typed-extraction note), so only the primary `name` field is
  read here; `displayName` is never consulted. Outreach's real prospect/account/sequence/mailing
  resources set `name` directly in the overwhelming common case (the `displayName` fallback exists
  in legacy defensively, not for a documented common shape), so this is a narrow, documented scope
  narrowing — never a data change for any record that sets `name`. ACCEPTABLE per
  `docs/migration/conventions.md` §5's parity-deviation meta-rule (never changes emitted data for
  any input legacy's *primary* path would accept).
- **`max_pages` config surface is not carried over**: legacy accepted a `max_pages` config value
  (0/`all`/`unlimited` = unbounded, or a positive integer cap) as a hard request-count safety valve
  independent of the `links.next` stop signal. The engine's `next_url` paginator has no
  `MaxPages`-equivalent knob wired to a config value (`PaginationSpec.MaxPages` is read only by the
  page-count-driven paginators — `page_number`/`offset_limit`/`cursor` — never by `next_url`, per
  `paginate.go`'s `newPaginator`); pagination is bounded only by the short/empty-`links.next` stop
  signal, matching Outreach's own real termination behavior (an unbounded sync still terminates
  correctly, it just has no independent request-count safety cap). `page_size` is likewise the only
  page-shape knob retained in `spec.json`, for config-surface documentation; it is not wired into
  the pagination block itself (the `next_url` paginator reads no page-size field at all) — the
  static `page[size]: "100"` in each stream's `query` is the sole place page size is actually
  requested, matching stripe's `limit_param`/`page_size` dead-config precedent
  (`docs/migration/conventions.md` §5 item 3).
- **Fixture pagination**: every stream's `fixtures/streams/<name>/page_1.json` is a single-page
  fixture (the sanctioned `next_url` exception, `docs/migration/conventions.md` §4) — a `next_url`
  stream's next-page URL is the replay server's own runtime-assigned address, which cannot be
  embedded in a static fixture file. `pagination_terminates` exercises this bundle's first stream
  (`prospects`) and passes because the engine's `next_url` paginator stops correctly on the fixture's
  empty `links.next` value, proving the real stop condition without needing a second page.
- Full Outreach API surface (calls, tasks, opportunities, users, sequence states, webhooks, custom
  objects) is out of scope for this migration; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 4
  legacy-parity read streams are implemented.
