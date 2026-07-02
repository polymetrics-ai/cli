# Overview

Tremendous is a wave2 fan-out declarative-HTTP migration. It reads campaigns, orders, rewards, and
funding sources through the Tremendous API v2 (`GET https://testflight.tremendous.com/api/v2/...`).
This bundle is migrated at capability parity from `internal/connectors/tremendous` (the hand-written
connector it replaces); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a Tremendous API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`tremendous.go:146`). `base_url` defaults to
`https://testflight.tremendous.com` — legacy's own default points at Tremendous's sandbox/testflight
host rather than a production host (`tremendous.go:18`), reproduced here verbatim as the spec
default; production callers must override `base_url` explicitly, matching legacy's own behavior.

## Streams notes

All four streams (`campaigns`, `orders`, `rewards`, `funding_sources`) are `page_number`-paginated
list endpoints under `/api/v2/...` using `limit`/`page` query parameters (legacy's `harvest`
function, `tremendous.go:87-119`), records extracted from a top-level key matching the resource
name. Pagination is declared with `page_size: 100` and `max_pages: 1`, matching legacy's own
`defaultPageSize`/`defaultMaxPages` constants (`tremendous.go:19-21`) exactly — legacy only fetches
beyond one page when `max_pages` is explicitly configured to a larger number, `"all"`, or
`"unlimited"` (`tremendous.go:167-180`). None of the four streams expose an incremental cursor
field in legacy, so all four are always full-refresh reads.

## Write actions & risks

None. Legacy's `Write` unconditionally returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`boundedInt`/`configuredMaxPages` helpers, `tremendous.go:167-194`, `page_size` bounded
  1-1000, `max_pages` accepting a literal integer or the sentinels `"all"`/`"unlimited"` for
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
- **Legacy's dual-key field fallbacks (`campaignId`/`paymentStatus`/`createdAt`/`orderId`) are not
  modeled.** Legacy's `namedRecord`/`orderRecord`/`rewardRecord` mapping functions each accept
  EITHER a snake_case OR a camelCase key via a `first(item, ...)` helper
  (`tremendous.go:222-238`) — e.g. `campaign_id` OR `campaignId`, `created_at` OR `createdAt` —
  preferring the snake_case key first. The engine's `computed_fields` dialect has no
  coalesce/fallback filter (each output field name resolves against exactly one template), so only
  the snake_case shape legacy's own test suite exercises (`tremendous_test.go:23`:
  `campaign_id`/`payment_status`/`created_at`) is modeled via plain schema projection. This is a
  documented scope narrowing, not a data change for any input legacy's own tests demonstrate as the
  real wire shape; if the live Tremendous API ever sends the camelCase variant instead, this bundle
  would silently drop that field where legacy would have populated it — flagged here rather than
  fudged.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps a static `connector: "tremendous"` marker and a `fixture:
  true` flag onto two synthesized records per stream (`tremendous.go:121-135`). Neither is part of
  the LIVE record shape; this bundle's schemas and fixtures target the live path only. The engine's
  own conformance/fixture-replay harness provides the credential-free test affordance this bundle
  needs, so no fixture-mode equivalent is needed here.
