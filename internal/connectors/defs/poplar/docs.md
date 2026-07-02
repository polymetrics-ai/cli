# Overview

Poplar is a wave2-fan-out declarative-HTTP migration. It reads Poplar campaigns and orders through
read-only REST list endpoints (`GET https://api.heypoplar.com/v1/...`). This bundle targets
capability parity with `internal/connectors/poplar` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip. Legacy's own package doc
notes Poplar's public docs vary by account features, so the connector intentionally limits itself
to common list endpoints and never performs write operations.

## Auth setup

Provide a Poplar API token via the `api_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_token>`) and is never logged, matching legacy's `connsdk.Bearer(token)`
(`poplar.go:181`). `base_url` defaults to `https://api.heypoplar.com/v1` and may be overridden for
tests/proxies.

## Streams notes

Both streams (`campaigns`, `orders`) share the same shape: `GET` against the Poplar list endpoint,
records at `data`, primary key `["id"]`. `orders` models legacy's `CursorFields: []string{"created_at"}`
catalog declaration as an `incremental.cursor_field` (no `request_param`/`client_filtered`, so this
adds no request-shape or filtering behavior beyond what legacy's own `Read` performs — legacy never
filters by date server- or client-side, it only declares the cursor field for catalog purposes).

Pagination follows Poplar's own `meta.next_page` body-path convention (`poplar.go:114`'s
`connsdk.StringAt(resp.Body, "meta.next_page")`): declared as `pagination.type: cursor` with
`token_path: meta.next_page` — Poplar's `next_page` value is itself the next page NUMBER (not an
opaque token), sent back as the `page` query param on the following request
(`cursor_param: "page"`), matching legacy's `page, err = strconv.Atoi(next)` exactly. `page=1` /
`per_page=100` are declared as a static per-stream `query` (matching legacy's `defaultPageSize`),
re-sent on every page request per the `cursor` paginator's normal merge behavior.

## Write actions & risks

None. Legacy's `Write` unconditionally returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (default
  100, clamped 1-500) and `max_pages` (default 3, or `"all"`/`"unlimited"` for unbounded) as
  config-driven overrides (`poplar.go:199-222`). The engine's `cursor` paginator has no
  config-driven page-size or max-pages knob (`PaginationSpec.PageSize`/`MaxPages` are static ints
  set once in `streams.json`, not template-resolvable), so this bundle sends legacy's own default
  page size (`per_page=100`) as a static per-stream query literal and declares no `max_pages` cap
  (unbounded, matching legacy's zero-value/`"unlimited"` shape) rather than legacy's default cap of
  3 pages — `spec.json` does not declare `page_size`/`max_pages` at all (F6, REVIEW.md: a
  declared-but-unwireable config key is worse than an absent one).
- **The short-page stop signal is not modeled.** Legacy stops when `next_page` is empty OR the
  current page returned fewer records than `page_size` (`poplar.go:118`'s
  `strings.TrimSpace(next) == "" || len(records) < pageSize`) — a defensive belt-and-braces check.
  The engine's `cursor`+`token_path` paginator stops only on an absent/falsy token
  (`meta.next_page`), matching legacy's PRIMARY stop condition exactly; it does not independently
  check page fullness. This never diverges for Poplar's real API (its own `next_page: null`/absent
  is the authoritative last-page signal, confirmed by legacy's own `poplar_test.go` fixture where
  both conditions agree on the same page) — only a hypothetical malformed response (a full page
  with a null `next_page`, or vice versa) would exercise the difference, and Poplar's documented
  behavior gives no evidence this occurs.
- **The `createdAt`/`campaign_id` fallback fields are not modeled.** Legacy's `mapRecord` functions
  defensively fall back to a camelCase `createdAt` if `created_at` is absent
  (`first(item, "created_at", "createdAt")`), and `orders` falls back to `campaign_id` if `name` is
  absent (`first(item, "name", "campaign_id")`) — the engine's schema-as-projection dialect matches
  by exact key name only and has no coalesce/fallback-chain filter. Poplar's real, documented field
  names are snake_case (`created_at`); no test evidence (fixture or live) shows the camelCase
  variant or a missing `name` ever occurring in practice, so this bundle declares the primary field
  name only. If Poplar's API is later observed to emit either fallback shape, this is a
  Pass-B/capability-expansion `computed_fields` addition, not a Tier-2 hook (the dialect has no
  multi-source coalesce; a fallback that never fires on any observed real payload is not treated as
  a proven-necessary `ENGINE_GAP`).
- **`docs.poplar.studio` was unreachable during this migration** (HTTP 403 on fetch); per
  conventions.md, legacy code (and its test fixtures) is ground truth over docs in this situation,
  so every stream/field/pagination shape above is derived from `poplar.go`/`poplar_test.go` only.
