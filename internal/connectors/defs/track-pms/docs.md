# Overview

Track PMS is a wave2 fan-out declarative-HTTP migration, expanded in Pass B against Track PMS's
published per-domain OpenAPI specs (`https://developer.trackhs.com/reference`, fetched 2026-07-03 —
see `api_surface.json`; the real API is a very large multi-domain surface — PMS reservations/
units/owners, CRM, channel, guest-communications, system — comparable in scope to GitHub's REST
API). It reads reservations, guests, units, owners, CRM contacts, and unit types
(`GET https://api.trackhs.com/...`), and writes create/update mutations for reservations, units,
owners, and contacts. This bundle is migrated at capability parity from
`internal/connectors/track-pms` (the hand-written `trackpms` package it replaces) for its original
4 read streams; the legacy package stays registered and unchanged until wave6's registry flip, and
never implemented any write action or the 2 new streams (`contacts`, `unit_types`) — those are new
capability, not a parity port. **Two significant pre-existing legacy/real-API discrepancies were
discovered during this research** — see Known limits.

## Auth setup

Provide a Track PMS API access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`track_pms.go:146`). `base_url` defaults to `https://api.trackhs.com` and
may be overridden for tests/proxies.

## Streams notes

All four streams (`reservations`, `guests`, `units`, `owners`) are `page_number`-paginated list
endpoints using `limit`/`page` query parameters (legacy's `harvest` function, `track_pms.go:87-119`),
records extracted from a top-level key matching the resource name. Pagination is declared with
`page_size: 100` and `max_pages: 1`, matching legacy's own `defaultPageSize`/`defaultMaxPages`
constants (`track_pms.go:19-21`) exactly — legacy only fetches beyond one page when `max_pages`
is explicitly configured to a larger number, `"all"`, or `"unlimited"` (`track_pms.go:170-183`).
`reservations` declares `arrival_date` as its `x-cursor-field` for manifest-surface parity; legacy
never actually filters or advances reads by it (no `incremental` block on legacy's own reservation
read path), so no `incremental` block is declared here either — matching legacy's real full-refresh
behavior.

**New Pass B streams** (real documented paths — no legacy precedent, so these use the CURRENT
documented shape rather than reproducing any legacy convention):

- `contacts` — `GET /crm/contacts`, the real modern CRM contact resource (see Known limits for why
  this is distinct from legacy's `guests` stream). Paginated `page_number` (`page`/`size` query
  params, `page_size: 100`, no `max_pages` cap — genuinely unbounded, matching the real API's own
  contract). Records live at `_embedded.contacts` (Track PMS's real API wraps every list response
  in a HAL-style `_embedded` envelope — see Known limits). `computed_fields` renames every
  camelCase field (`firstName`, `primaryEmail`, `isVip`, etc.) to this bundle's snake_case schema.
- `unit_types` — `GET /pms/units/types`, the unit-type taxonomy (bed count, occupancy, node/lodging
  assignment) each real unit belongs to. Same `page_number`/`_embedded.units`/`computed_fields`
  shape as `contacts`.

## Write actions & risks

**New Pass B capability — legacy is entirely read-only** (legacy's `Write` unconditionally returns
`connectors.ErrUnsupportedOperation`). `capabilities.write` is now `true` and `writes.json` declares
7 actions against the REAL documented `/pms/...`/`/crm/...` paths (not legacy's unprefixed
`/reservations`/`/units`/`/owners` paths — see Known limits):

- `create_reservation` (`POST /pms/reservations`; required `unitId`/`arrivalDate`/`departureDate`
  — modeled from Track PMS's own `reservation.request.v1` schema).
- `create_unit`/`update_unit` (`.../pms/units[/{{ record.id }}]`).
- `create_owner`/`update_owner` (`.../pms/owners[/{{ record.id }}]`; update uses `PATCH`, matching
  the real API's documented "Update Owner" verb).
- `create_contact`/`update_contact`/`delete_contact` (`.../crm/contacts[/{{ record.id }}]`; update
  uses `PATCH` — the real API also documents an equivalent `PUT` variant, excluded in
  `api_surface.json` as `duplicate_of` since this bundle's per-record write model needs only one
  update verb per resource).

Every write's `record_schema` models a practical field subset (the required fields plus the most
commonly-set optional ones), not every property the real API's request schemas document — Track
PMS's real create/update payloads run to 20-80+ optional properties per resource (unit creation
alone documents dozens of amenity/policy/tax fields); modeling every one is a distinct, larger
follow-on effort, not a Pass B blocker (every field this bundle DOES model is real and verified
against the published OpenAPI spec, never guessed). See `api_surface.json` for the full accounting
of every excluded sub-resource write (payment plans, fees, access codes, owner statements,
fractional ownership, and the many admin-taxonomy configuration endpoints).

## Known limits

- **Legacy's `/reservations`, `/guests`, `/units`, `/owners` paths and bare-top-level-key response
  shape do not match Track PMS's currently documented API.** The real, currently-published Track
  PMS OpenAPI specs (fetched 2026-07-03) document these resources at `/pms/reservations`,
  `/crm/contacts` (not `/guests` at all — no `/guests` path exists anywhere in the current docs),
  `/pms/units`, and `/pms/owners`, and EVERY real list response wraps its records in a HAL-style
  envelope (`_embedded.<resource>`, e.g. `{"_embedded":{"reservations":[...]}, "page":1,
  "page_count":5, ...}`) rather than the bare top-level key (`{"reservations":[...]}`) legacy's
  `harvest` function (`track_pms.go:87-119`, `connsdk.RecordsAt(resp.Body, spec.recordsPath)` with
  `recordsPath` set to the bare resource name) assumes. This bundle reproduces legacy's exact
  path/response-shape assumptions UNCHANGED for the 4 legacy-parity streams (parity-preserving per
  the meta-rule — this is not a Pass B regression, legacy itself has apparently drifted from Track
  PMS's real current API surface, an independent pre-existing issue this research surfaced but did
  not introduce). The 2 NEW Pass B streams (`contacts`, `unit_types`) have no legacy precedent, so
  they correctly use the REAL current paths and the real `_embedded.<resource>` records path from
  the start. Reconciling the 4 legacy streams against Track PMS's real current API is a distinct,
  deliberate parity-deviation decision for a future pass, not a Pass B capability-expansion change.
- **Legacy's Bearer-token auth does not match Track PMS's currently documented authentication.**
  Track PMS's own authentication docs (`https://developer.trackhs.com/docs/authentication`)
  describe a multi-tenant API (`https://{customer}.trackhs.com/api`) authenticated with HTTP Basic
  Auth (an API key/secret pair as username/password) or HMAC request signing — not a Bearer token
  against a single shared `api.trackhs.com` host, which is what legacy (`track_pms.go:146`,
  `connsdk.Bearer(token)`) and this bundle's `base.auth` both implement. This is left unchanged for
  the identical parity-preserving reason as the path/response-shape discrepancy above — flagged
  here as a significant finding for whoever next reconciles this connector's auth against the real
  API, not silently corrected mid-pass.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`pageSize`/`maxPages` helpers, `track_pms.go:167-183`, `page_size` bounded 1-500 via
  `boundedInt`, `max_pages` accepting a literal integer or the sentinels `"all"`/`"unlimited"` for
  unbounded). The engine's `page_number` paginator's `PageSize`/`MaxPages` fields are plain static
  integers in `streams.json` — never templated against a runtime config value (`bundle.go`'s
  `PaginationSpec`; `paginate.go`'s constructor reads them as fixed ints) — so neither can be wired
  to a config override at all. This bundle therefore declares legacy's own DEFAULTS
  (`page_size: 100`, `max_pages: 1`) as fixed pagination values and does not declare `page_size`/
  `max_pages` in `spec.json` (F6, REVIEW.md precedent: a declared-but-unwireable config key is
  worse than an absent one). Because `max_pages: 1` genuinely caps every read at one page (matching
  legacy's own default), this bundle ships single-page fixtures for every stream, following
  searxng's identical `max_pages: 1` + single-page-fixture precedent
  (`internal/connectors/defs/searxng/fixtures`) — proving 2-page pagination termination would
  require the paginator to fetch a page this connector's declared configuration can never actually
  request.
- **Legacy's dual-key field fallbacks (`confirmationNumber`/`arrivalDate`/`full_name`/`unit_name`)
  are not modeled.** Legacy's `reservationRecord`/`personRecord`/`unitRecord` mapping functions
  each accept EITHER a snake_case OR a camelCase/alternate key via a `first(item, ...)` helper
  (`track_pms.go:225-241`) — e.g. `confirmation_number` OR `confirmationNumber`, `name` OR
  `full_name`/`unit_name` — preferring the snake_case key first. The engine's `computed_fields`
  dialect has no coalesce/fallback filter (each output field name resolves against exactly one
  template), so only the snake_case shape legacy's own test suite exercises
  (`track_pms_test.go:23`: `confirmation_number`/`arrival_date`) is modeled via plain schema
  projection. This is a documented scope narrowing, not a data change for any input legacy's own
  tests demonstrate as the real wire shape; if the live Track PMS API ever sends the camelCase
  variant instead, this bundle would silently drop that field where legacy would have populated it
  — flagged here rather than fudged.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps a static `connector: "track-pms"` marker and a `fixture:
  true` flag onto two synthesized records per stream (`track_pms.go:121-135`). Neither is part of
  the LIVE record shape; this bundle's schemas and fixtures target the live path only. The engine's
  own conformance/fixture-replay harness provides the credential-free test affordance this bundle
  needs, so no fixture-mode equivalent is needed here.
