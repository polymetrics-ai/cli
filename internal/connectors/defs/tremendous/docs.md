# Overview

Tremendous is a wave2 fan-out declarative-HTTP migration, expanded in Pass B against Tremendous's
published OpenAPI v2 reference (`https://developers.tremendous.com/reference`, fetched 2026-07-03
— see `api_surface.json`). It reads campaigns, orders, rewards, funding sources, products, invoices,
and members through the Tremendous API v2 (`GET https://testflight.tremendous.com/api/v2/...`), and
writes order/reward lifecycle actions plus invoice/member/webhook administration. This bundle is
migrated at capability parity from `internal/connectors/tremendous` (the hand-written connector it
replaces) for its original 4 read streams; the legacy package stays registered and unchanged until
wave6's registry flip, and never implemented any write action or the 3 new streams (`products`,
`invoices`, `members`) — those are new capability, not a parity port.

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

**New Pass B streams** (no legacy precedent — capability expansion, not a parity port):

- `products` — `GET /api/v2/products`, records at `products`; not paginated (Tremendous's own
  OpenAPI spec declares no page/limit parameters on this endpoint — it returns the full product
  catalog in one response).
- `invoices` — `GET /api/v2/invoices`, records at `invoices`; Tremendous's real documented
  pagination for this endpoint is `offset`/`limit` (`offset_limit` pagination, `page_size: 100`, no
  `max_pages` cap), which this NEW stream declares correctly since there is no legacy behavior to
  preserve — see Known limits for why this differs from the 4 legacy streams' `page`/`limit` shape.
  `x-cursor-field: created_at` is declared on the schema for manifest-surface completeness
  (Tremendous documents a `created_at[gte]`/`created_at[lte]` server-side filter), but no
  `incremental` block is wired per conventions.md §8 rule 2 (no legacy precedent to preserve, and
  wiring the bracket-syntax filter param is a distinct follow-on).
- `members` — `GET /api/v2/members`, records at `members`; not paginated (Tremendous's own OpenAPI
  spec declares no parameters at all on this endpoint).

## Write actions & risks

**New Pass B capability — legacy is entirely read-only** (legacy's `Write` unconditionally returns
`connectors.ErrUnsupportedOperation`). `capabilities.write` is now `true` and `writes.json` declares
11 actions, every one matching a real, documented Tremendous v2 endpoint
(`api_surface.json`'s `covered_by.write` entries):

- **Orders**: `create_order` (`POST /orders` — spends funding-source balance to issue a reward;
  modeled as the single-reward order shape, Tremendous's `SingleRewardOrder` request variant, since
  that is the common case; the multi-reward batch-order variant is not modeled), `approve_order`/
  `reject_order` (`POST /order_approvals/{id}/{approve,reject}`, `kind: custom` no-body actions —
  Tremendous's order-approvals workflow).
- **Rewards**: `cancel_reward` (`POST /rewards/{id}/cancel`), `resend_reward`
  (`POST /rewards/{id}/resend`, optional `updated_email`/`updated_phone`), `generate_reward_link`
  (`POST /rewards/{id}/generate_link`).
- **Invoices**: `create_invoice`/`delete_invoice` (`.../invoices[/{id}]`).
- **Members**: `create_member` (`POST /members`; invites a new organization user).
- **Webhooks**: `create_webhook`/`delete_webhook` (`.../webhooks[/{id}]`; Tremendous supports
  exactly one webhook per organization, so `create_webhook` also serves as "replace the existing
  webhook URL").

Every write's `risk` field states its specific blast radius; `create_order`/`approve_order`/
`reject_order`/`cancel_reward`/`create_member` move money or grant organization access and are
flagged as approval-required in `metadata.json.risk.approval`. No `destructive_admin` or
`requires_elevated_scope` action (fraud-rule configuration, API-key issuance, sub-organization
creation) is modeled — see `api_surface.json` for the full excluded-endpoint accounting.

## Known limits

- **The 4 legacy streams' `page`/`limit` pagination parameters do not match Tremendous's currently
  documented pagination shape.** Tremendous's published OpenAPI spec for `campaigns`/`orders`/
  `rewards`/`funding_sources` declares `offset`/`limit` query parameters (`offset_limit`
  pagination), not `page`/`limit` — legacy (`tremendous.go:97`) sends `page`/`limit` regardless,
  which is not a parameter shape Tremendous's current docs recognize for these endpoints. This
  bundle reproduces legacy's exact `page`/`limit` request shape for these 4 streams unchanged
  (parity-preserving per the meta-rule: this bundle must not silently change what legacy sends),
  rather than "fixing" it to the documented `offset`/`limit` shape, which would be an accepted-input
  behavior change outside this pass's mandate. The 3 NEW Pass B streams have no such legacy
  precedent, so `invoices` correctly declares the real `offset_limit` shape from the start (see
  Streams notes); `products`/`members` are unpaginated in Tremendous's own docs, so this
  discrepancy does not apply to them. A future pass reconciling the 4 legacy streams' pagination
  against Tremendous's real API is a distinct, deliberate parity-deviation decision, not a Pass B
  capability-expansion change.
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
- **Legacy's dual-key field fallbacks are modeled.** Legacy's `namedRecord`/`orderRecord`/
  `rewardRecord` mapping functions accept either snake_case or camelCase keys via `first(item, ...)`
  (`tremendous.go:222-238`) — e.g. `campaign_id`/`campaignId`, `payment_status`/`paymentStatus`,
  `created_at`/`createdAt`, `order_id`/`orderId` — preferring the snake_case key first. The bundle
  expresses that with `computed_fields` `coalesce` entries on the four legacy-parity streams.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps a static `connector: "tremendous"` marker and a `fixture:
  true` flag onto two synthesized records per stream (`tremendous.go:121-135`). Neither is part of
  the LIVE record shape; this bundle's schemas and fixtures target the live path only. The engine's
  own conformance/fixture-replay harness provides the credential-free test affordance this bundle
  needs, so no fixture-mode equivalent is needed here.
