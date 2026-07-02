# Overview

Track PMS is a wave2 fan-out declarative-HTTP migration. It reads reservations, guests, units, and
owners through the Track PMS REST API (`GET https://api.trackhs.com/...`). This bundle is migrated
at capability parity from `internal/connectors/track-pms` (the hand-written `trackpms` package it
replaces); the legacy package stays registered and unchanged until wave6's registry flip.

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

## Write actions & risks

None. Legacy's `Write` unconditionally returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

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
