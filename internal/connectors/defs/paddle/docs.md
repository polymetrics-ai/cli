# Overview

Paddle is a wave2 fan-out declarative-HTTP migration. It reads Paddle transactions, customers,
subscriptions, and products through the Paddle REST API (`GET https://api.paddle.com/...`). This
bundle targets capability parity with `internal/connectors/paddle` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.
Read-only: legacy's `Write` always returns `connectors.ErrUnsupportedOperation`, and this bundle
declares `capabilities.write: false` with no `writes.json` to match.

## Auth setup

Provide a Paddle API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`paddle.go:218`). `base_url` defaults to `https://api.paddle.com` and may
be overridden for tests/proxies (legacy's own `baseURL` helper validates scheme+host the same way;
the engine's base-URL resolution has no equivalent runtime validation, but every fixture/conformance
run only ever points at an httptest server, so this is not exercised differently on either side).

## Streams notes

Four streams, all primary-keyed on `id`: `transactions`, `customers`, `subscriptions`, `products`.
Each is a flat list endpoint whose records live at the top-level `data` key
(`connsdk.RecordsAt(resp.Body, "data")`, `paddle.go:167`). Pagination follows Paddle's own
cursor convention: `per_page=100` (legacy's `paddleDefaultPageSize`) is sent as a static per-stream
`query` literal (matching stripe's `limit=100` static-query precedent, `docs/migration/
conventions.md`), and the next page's cursor is read from `meta.pagination.next`
(`pagination.type: cursor`, `token_path: meta.pagination.next`) and sent back as the `after` query
param (`cursor_param: after`) — an empty/absent `next` value stops pagination, matching legacy's own
`strings.TrimSpace(next) == ""` stop check (`paddle.go:180`) exactly; no `stop_path` is declared
since legacy has no separate boolean stop signal beyond the cursor token itself.

None of the four streams expose a server-side incremental filter parameter in legacy (`Read` never
sends a created/updated-after query param — `harvestCursor` only ever sends `per_page`/`after`), so
this bundle declares no `incremental` block for any stream, matching legacy exactly. Each schema
still declares `x-cursor-field: created_at`, mirroring legacy's own `Catalog` `CursorFields`
declaration (`paddle.go:56-84`) for informational/dedup-mode purposes only — per
`docs/migration/conventions.md` §2, `incremental_append` sync modes are gated on the presence of an
`incremental` block, not on `x-cursor-field` alone, so declaring the field without the block adds no
incremental capability legacy itself doesn't have (see zendesk-support's identical precedent).

## Write actions & risks

None. Legacy's own package doc states it is "a read-only native Paddle HTTP API connector";
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (`paddleDefaultPageSize`/`paddleMaxPageSize`, `paddle.go:258-271`) on every stream. The engine's
  `cursor` paginator (token_path variant) never reads `PaginationSpec.PageSize` (only
  `page_number`/`offset_limit` consume it), so a `{{ config.page_size }}` template on the `per_page`
  query param has no wiring mechanism analogous to legacy's override. This bundle therefore sends
  Paddle's own default (`per_page=100`) as a static per-stream query literal and does not declare
  `page_size` in `spec.json` at all (a declared-but-unwireable config key is worse than an absent
  one, per the bitly/searxng F6 precedent).
- **`max_pages` is not modeled.** Legacy's hard request-count cap override (`paddle.go:273-286`) has
  no engine-side equivalent for the `cursor` paginator (no `MaxPages`-style knob wired to a config
  value for this pagination type); pagination is bounded only by the empty-next-cursor stop signal,
  matching Paddle's own real termination behavior.
- **Fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached when
  `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps an extra
  `previous_cursor` field (echoing `req.State["cursor"]` when a prior cursor happens to be set) onto
  every fixture-mode record (`paddle.go:195-197`). This is not part of the LIVE record shape; this
  bundle's schemas target the live path only. The engine's own conformance/fixture-replay harness
  provides the credential-free test affordance this bundle needs, so no fixture-mode equivalent is
  needed here.
