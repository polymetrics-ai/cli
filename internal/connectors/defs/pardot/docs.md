# Overview

Pardot reads Salesforce Account Engagement (Pardot) prospects, campaigns, and lists through Pardot
API v5 REST endpoints. This bundle migrates `internal/connectors/pardot` (the hand-written legacy
connector) to a declarative Tier-1 defs bundle; the legacy package stays registered and unchanged
until the wave6 registry flip. Legacy is intentionally conservative: it expects a pre-issued
Salesforce OAuth access token and never performs token acquisition/refresh itself — this bundle
preserves that exact scope.

## Auth setup

Provide a pre-issued Salesforce OAuth access token via the `access_token` secret; it is used only
for Bearer auth (`Authorization: Bearer <access_token>`) and is never logged. Token
acquisition/refresh remains out of scope, matching legacy exactly. The required `business_unit_id`
config value is sent as the `Pardot-Business-Unit-Id` header on every request (legacy hard-errors
if either `access_token` or `business_unit_id` is unset; this bundle's `required` array on both
`spec.json` properties reproduces that same hard requirement).

## Streams notes

Three streams: `prospects` (`GET /api/v5/objects/prospects`), `campaigns`
(`GET /api/v5/objects/campaigns`), `lists` (`GET /api/v5/objects/lists`). All share the same shape:
records at `values`, primary key `["id"]` (a Pardot integer id, not a string — schema types it
`"integer"` matching the real wire shape), incremental cursor field `updatedAt`. Every request
sends `limit=200` (matches legacy's default `page_size`) via each stream's static `query:
{"limit": "200"}`. Pagination follows legacy's `harvestNextPageURL`: a `next_url`-type pagination
block reads the absolute next-page URL from the body's `nextPageUrl` field and stops when that
value is empty — legacy's exact stop condition (`strings.TrimSpace(next) == ""`), with the query
params dropped entirely for subsequent requests (`query = nil` in legacy's loop) since the
returned `nextPageUrl` already carries every needed param, identical to the engine's `next_url`
paginator behavior.

## Write actions & risks

None. This connector is read-only, matching legacy (`Capabilities.Write: false`,
`Write` returns `ErrUnsupportedOperation`).

## Known limits

- Only `prospects`, `campaigns`, and `lists` are implemented, matching legacy's exact stream set.
  Opportunities, visitors, visits, emails, forms, custom fields, and users are out of scope for
  this wave; see `api_surface.json`'s `excluded` entries.
- Per the `next_url` pagination sanctioned exception (`docs/migration/conventions.md` §4), each
  stream ships a single-page fixture only: the real next-page URL is the replay server's own
  address, unknown until the harness picks a port at runtime, so a static fixture file cannot embed
  a correct absolute second-page URL. Every fixture's `nextPageUrl` is empty, terminating after one
  page — this satisfies `fixtures_present`/`read_fixture_nonempty` for all three streams.
  `pagination_terminates` exercises whichever stream conformance selects as eligible; a live
  2-page `next_url` parity test is Tier-2/paritytest scope, not available in this wave2 JSON+docs
  pass (no Go was created for this connector).
- Legacy's runtime `limit`/`max_pages` config overrides are not exposed as `spec.json` properties:
  the engine's `next_url` pagination type has no config-templated page-size or page-count field, so
  a declared-but-unwireable `spec.json` property would be dead config (F6). This bundle sends a
  fixed `limit=200` per request (matching legacy's default) and leaves the page count unbounded
  (matching legacy's `max_pages=0`/unlimited default).
