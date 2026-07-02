# Overview

Aha! is a wave2 fan-out declarative-HTTP migration. It reads Aha! features, products, ideas,
releases, initiatives, and goals through the Aha! REST API (`GET https://<account>.aha.io/api/v1/...`).
This bundle targets capability parity with `internal/connectors/aha` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an Aha! API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's treatment of Aha! API keys "exactly like
OAuth bearer tokens" (`aha.go:252`, `connsdk.Bearer(secret)`), and is never logged. `base_url`
defaults to `https://secure.aha.io` (legacy's `ahaDefaultBaseURL`) but Aha! is account-scoped
(`<company>.aha.io`), so almost every real deployment overrides it — see Known limits for legacy's
`url` config alias, which this bundle does not model.

## Streams notes

All six streams share the identical Aha! page-number envelope: `GET /api/v1/<resource>` returns
`{"<resource>":[...],"pagination":{"total_records":N,"total_pages":N,"current_page":N}}`, records
live at the resource-named top-level key (identical to the resource segment for every stream here).
Pagination is `page_number` (`page`/`per_page`, `start_page: 1`, static `page_size: 30` matching
legacy's `ahaDefaultPageSize`). The engine's `page_number` paginator stops on a short page
(`recordCount < page_size`); legacy instead reads the `pagination.total_pages`/`current_page`
object and falls back to short-page stop only when `total_pages` is absent. These two stop
conditions are equivalent for every dataset except one whose total record count is an exact
multiple of 30, where legacy stops immediately via the `total_pages` comparison and the engine
would issue one additional request returning an empty page before stopping — no different records
are ever emitted either way (same non-data-affecting divergence documented on this wave's
adobe-commerce-magento bundle).

Every stream stamps a static-literal `resource` marker field (`"feature"`/`"product"`/`"idea"`/
`"release"`/`"initiative"`/`"goal"`) via `computed_fields`, matching legacy's own hardcoded
`"resource": "feature"` (etc.) in each `mapRecord` function.

Every Aha! object publishes `updated_at` as `x-cursor-field` (matching legacy's own
`CursorFields: []string{"updated_at"}`), but Aha!'s list endpoints expose no server-side
incremental filter parameter and legacy's own `harvest` never applies one — every read is a full
paginated sweep regardless of any prior sync's cursor. This bundle therefore declares
`incremental.cursor_field` with no `request_param`/`start_config_key`/`client_filtered`: the cursor
field is published (enabling `incremental_append_deduped` sync-mode eligibility for downstream
consumers) without the engine ever computing or sending a filter, matching legacy's true read
behavior exactly.

## Write actions & risks

None. Aha! has no obvious safe reverse-ETL write surface (legacy's own package doc: "there is no
obvious safe reverse-ETL write surface"); `capabilities.write` is `false` and this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **The `url` config alias for `base_url` is not modeled.** Legacy accepts either `base_url` or a
  bare `url` config key as an alias (`ahaBaseURL`, `aha.go:268-289`) — either name works
  interchangeably. This bundle declares only `base_url`; an operator previously using the `url` key
  name must rename it. This is a config-surface naming narrowing only — the resolved effective base
  URL behavior once configured is identical.
- **`per_page`/`page_size` and `max_pages` config overrides are not modeled.** Legacy exposes
  `per_page` (aliased from `page_size`, 1-200, default 30) and `max_pages`
  (0/all/unlimited or a positive integer cap) as config-driven overrides (`ahaPageSize`/
  `ahaMaxPages`). The engine's `page_number` paginator has no config-driven page-size or
  request-count-cap knob (mirrors this wave's adobe-commerce-magento precedent and stripe's
  resolved ledger item 3); `per_page`/`max_pages` are therefore not declared in `spec.json`, and
  this bundle sends Aha!'s own default (`per_page=30`) as a static pagination-block value.
- **`total_pages`-based early stop is approximated by short-page stop only** — see Streams notes
  above; the only observable difference is one extra empty-page request when a stream's total
  record count is an exact multiple of 30, never a difference in which records are emitted.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (reached only
  when `config.mode == "fixture"`) stamps a `previous_cursor` field echoing `req.State["cursor"]`
  when set (`aha.go:227-229`), a credential-free conformance-harness affordance with no live-path
  equivalent. This bundle's schemas and fixtures target the live record shape only; the engine's own
  `internal/connectors/conformance` fixture-replay harness provides the credential-free test
  affordance this bundle needs.
