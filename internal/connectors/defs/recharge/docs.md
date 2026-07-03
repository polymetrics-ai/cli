# Overview

Recharge is a declarative-HTTP migration. It reads Recharge customers, subscriptions, and orders
through the Recharge REST API (`GET https://api.rechargeapps.com/...`). This bundle targets
capability parity with `internal/connectors/recharge` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip. Read-only: legacy's
`Write` always returns `connectors.ErrUnsupportedOperation`, and this bundle declares
`capabilities.write: false` with no `writes.json` to match.

## Auth setup

Provide a Recharge API access token via the `access_token` secret. It is sent on the
`X-Recharge-Access-Token` header with no prefix (`X-Recharge-Access-Token: <access_token>`),
matching legacy's `connsdk.APIKeyHeader("X-Recharge-Access-Token", token, "")`
(`recharge.go:161`) exactly via the engine's `api_key_header` auth mode. Every request also sends
an `X-Recharge-Version` header sourced from the `api_version` config value, defaulting to
`2021-11` (legacy's own hardcoded fallback, `recharge.go:158-160`) via `spec.json`'s `default`
materialization. `base_url` defaults to `https://api.rechargeapps.com` and may be overridden for
tests/proxies.

## Streams notes

Three streams, all primary-keyed on `id`: `customers`, `subscriptions`, `orders`. Each hits a flat
Recharge list endpoint (`/customers`, `/subscriptions`, `/orders`) whose records live at the
top-level object key matching the resource name (`connsdk.RecordsAt(resp.Body, endpoint.recordsPath)`,
`recharge.go:114`, identical shape for every stream in legacy — `rechargeStreamEndpoints`,
`recharge.go:170-174`). Every raw Recharge field this bundle projects is already named identically
to legacy's own emitted field (`id`/`email`/`status`/`customer_id`/`created_at`/`updated_at` — see
`rechargeCustomerRecord`/`rechargeSubscriptionRecord`/`rechargeOrderRecord`, `recharge.go:185-195`),
so plain schema-mode projection reproduces legacy's `mapRecord` output field-for-field with no
`computed_fields` renames needed.

Pagination is `cursor`/`token_path` (`pagination.type: cursor`, `cursor_param: cursor`,
`token_path: next_cursor`), matching legacy's own cursor-and-`next_cursor`-body-field convention
(`harvest`, `recharge.go:102-133`): the next page's `cursor` query param is read from the response
body's top-level `next_cursor` field, and pagination stops when that field is absent or empty —
identical to legacy's `strings.TrimSpace(next) == ""` stop check. Every request sends `limit=250`
(matches legacy's default `rechargeDefaultPageSize`) via each stream's static `query: {"limit":
"250"}`, and `pagination.page_size: 250` documents the same value at the base level (informational
only for the `cursor`/`token_path` paginator, which never reads `PageSize` itself — only
`page_number`/`offset_limit` paginators do).

None of the three streams expose a server-side incremental filter parameter in legacy (`Read`
never sends a date-filter query param — `harvest` only ever sends `limit`/`cursor`), so this bundle
declares no `incremental` block and no `x-cursor-field` for any stream, matching legacy exactly
(legacy's own `Catalog` declares no `CursorFields` for any of the three streams either —
`recharge.go:176-183`).

## Write actions & risks

None. Legacy's own `Metadata()` declares `Write: false`; `capabilities.write` is `false` and this
bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (`rechargeDefaultPageSize`/`rechargeMaxPageSize`, `recharge.go:203-205`, bounded 1-250). The
  engine's `cursor` paginator's `page_size` (`PaginationSpec.PageSize`) is a bundle-declared
  constant, not templated, and is not even read by the `token_path` cursor variant — the actual
  page size sent on the wire is the static per-stream `query: {"limit": "250"}`, which is likewise
  not config-driven (`stream.Query`'s plain-string entries are static literals here, matching
  legacy's own default value but not its override mechanism). This bundle therefore fixes legacy's
  default (`limit=250`) and does not declare `page_size` in `spec.json` at all (a declared-but-
  unwireable config key is worse than an absent one, per the bitly/searxng/recreation F6
  precedent).
- **`max_pages` is not modeled.** Legacy's hard request-count cap override
  (`rechargeMaxPages`/`maxPagesConfig`, `recharge.go:207-209,247-260`, accepting an integer, `all`,
  or `unlimited`) has no engine-side equivalent wired to a config value — `PaginationSpec.MaxPages`
  is a plain int field, not a template, so it cannot be sourced from `config.max_pages`. Pagination
  is bounded only by the `next_cursor`-absent stop signal, matching legacy's own unbounded default
  (`max_pages=0`) behavior when the config value is left unset (the common case).
- **Fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached when
  `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a fixed
  2-record set with synthetic field values (`recharge.go:135-146`), including an `address_id`
  field this bundle's schema does not model (that field is never part of legacy's live-mode
  `mapRecord` output for any of the three streams — it is a fixture-mode-only extra). The engine's
  own conformance/fixture-replay harness (`internal/connectors/conformance`) provides the
  credential-free test affordance this bundle needs, so no fixture-mode equivalent is needed here.
