# Overview

Shutterstock is a wave2 fan-out declarative-HTTP migration. It reads Shutterstock image, video,
and audio search metadata through the Shutterstock REST API
(`GET https://api.shutterstock.com/v2/{images,videos,audio}/search`). This bundle migrates
`internal/connectors/shutterstock` (the hand-written legacy connector) to a declarative defs
bundle at capability parity; the legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Shutterstock OAuth access token via the `access_token` secret; it is sent as a Bearer
token (`Authorization: Bearer <access_token>`, matching legacy's `connsdk.Bearer(secret)` at
`shutterstock.go:127`) and is never logged. `base_url` defaults to `https://api.shutterstock.com`
(legacy's `shutterstockDefaultBaseURL`) and may be overridden for tests/proxies.

## Streams notes

All three streams (`images`, `videos`, `audio`) share an identical shape: `GET
/v2/<stream>/search`, records at the top-level `data` key, and Shutterstock's own 1-based
page-number pagination (`pagination.type: page_number`, `page_param: page`, `size_param:
per_page`, `start_page: 1`) with a short-page stop threshold, matching legacy's `harvest` loop
(`shutterstock.go:90-116`)'s `len(records) < pageSize` check exactly (the engine's
`PageNumberPaginator` implements the identical stop rule natively). Four optional, config-driven
search filters — `query`, `sort`, `orientation`, `category` — are passed through to every stream
verbatim (`stream.Query`'s `omit_when_absent` object form), matching legacy's `filters` helper
(`shutterstock.go:174-182`) which only sets a query param when the corresponding config value is
non-empty. None of the three streams expose a real server-side incremental filter in legacy (no
date-range/updated-since query parameter is ever sent); `x-cursor-field: updated_at` is declared
purely as catalog/sort-key metadata matching legacy's own `CursorFields` declaration
(`shutterstock.go:163`), and no `incremental` block is declared on any stream, matching legacy's
full-refresh-only read behavior exactly.

## Write actions & risks

None. Shutterstock's legacy connector is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`page_size`/`per_page` is not runtime-configurable.** Legacy exposes `page_size` as a
  config-driven override (`shutterstock.go:217-230`, default 100, max 500). The engine's
  `PaginationSpec.PageSize` (used by the `page_number` paginator for both the `per_page` query
  value and the short-page stop threshold) is a fixed bundle-authored literal with no
  config-templating mechanism — this bundle declares `page_size: 100`, matching legacy's own
  default, and does not expose a `page_size` config property at all (a declared-but-unwireable
  config key is worse than an absent one, per conventions.md F6/`docs/migration/conventions.md`'s
  bitly precedent for the identical limitation). An operator can no longer override the per-page
  request size; the full record set synced is unaffected (only the number of requests it takes
  changes), so this is judged an ACCEPTABLE, documented scope-narrowing rather than an
  `ENGINE_GAP`.
- **Fallback field names are not modeled.** Legacy's `shutterstockRecord` mapper reads
  `description` with a fallback from `description` to `title`, `media_type` with a fallback from
  `media_type` to `asset_type`, and `updated_at` with a fallback from `updated_time` to
  `updated_at` (`shutterstock.go:142-144`). This bundle implements only the PRIMARY field of each
  pair — there is no coalesce/first-non-null filter in this dialect's `computed_fields`
  templating. Legacy's own test suite (`shutterstock_test.go`) only ever exercises the primary
  field names (`description`, `updated_time`); this is judged an ACCEPTABLE, documented
  scope-narrowing rather than an `ENGINE_GAP`, per the `encharge` bundle's identical precedent for
  an unexercised defensive fallback.
- **`max_pages` is not runtime-configurable.** Legacy exposes a `max_pages` config override
  (`0`/`all`/`unlimited` for unbounded, or a positive integer hard cap, `shutterstock.go:231-241`).
  The engine's `PaginationSpec.MaxPages` is a fixed bundle-authored literal, not config-driven;
  this bundle leaves it unset (unbounded), matching legacy's own default.
- **Legacy's fixture-mode-only stamped fields are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) stamps a `fixture: true` marker onto every emitted
  record (`shutterstock.go:151`); this is a credential-free conformance-harness affordance, not
  part of the live record shape, and is intentionally not reproduced — the engine's own
  `internal/connectors/conformance` fixture-replay harness provides the equivalent test
  affordance.
