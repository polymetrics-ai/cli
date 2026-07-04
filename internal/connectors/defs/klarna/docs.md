# Overview

Klarna reads Klarna settlement payouts and transactions through the Klarna Settlements API. This
bundle migrates `internal/connectors/klarna` (the hand-written connector) to a declarative defs
bundle; the legacy package stays registered and unchanged until wave6's registry flip. Pass B adds
the remaining JSON Settlements list/detail endpoints without changing legacy stream record data. The
Klarna Settlements API is read-only for reverse-ETL purposes, so `capabilities.write` is `false` and
this bundle ships no `writes.json`.

## Auth setup

Provide the Klarna merchant UID via the `username` secret and the API shared secret via the
`password` secret; both flow only into HTTP Basic auth (`Authorization: Basic
base64(username:password)`) and are never logged, matching legacy's `connsdk.Basic(username,
password)` wiring. Legacy reads `username` from `Secrets` first, falling back to `Config` — this
bundle models `username` as `x-secret` and sources it from `secrets.username` exclusively, matching
legacy's actual precedence order (Secrets checked first) rather than its rarely-used Config fallback
(see Known limits).

`base_url` is **required** in this bundle (see Known limits for why the region/playground derivation
is not reproduced).

## Streams notes

`payouts` (`GET /settlements/v1/payouts`, records at the `payouts` array) emits legacy's exact field
set via `computed_fields`: `settlement_amount` is hoisted out of the nested `totals` object
(`{{ record.totals.settlement_amount }}`, a bare single-reference so the engine copies the raw typed
integer value, matching legacy's `klarnaPayoutRecord`'s manual `totals["settlement_amount"]` hoist).
`transactions` (`GET /settlements/v1/transactions`, records at `transactions`) is a flat field-for-field
projection with no `computed_fields` needed — the schema property names match the raw API field
names exactly. `payout_summary` reads the identical `payouts` endpoint as `payouts` but re-shapes each
record to only the settlement totals, keyed by `payout_reference` (matching legacy's
`klarnaPayoutSummaryRecord`), via `computed_fields` hoisting all four `totals.*` fields.

Pass B adds `payout_details` (`GET /settlements/v1/payouts/{payment_reference}`), driven by the
comma-separated `config.payment_references` fan-out list, and `payout_summaries`
(`GET /settlements/v1/payouts/summary`), driven by `config.summary_start_date` and
`config.summary_end_date` with optional `config.summary_currency_code`. These are new streams with
their own schemas, so they do not alter the three legacy stream record shapes.

Pagination is `offset_limit` (`limit_param: size`, `offset_param: offset`, `page_size: 100`, matching
legacy's default `klarnaDefaultPageSize`/`OffsetPaginator` with Klarna's own `size`/`offset`
query-param naming); the engine stops once a page returns fewer than `page_size` records. No
`max_pages` cap is declared, matching legacy's default `klarnaMaxPages` behavior (unlimited). No
stream declares an `incremental` block or `x-cursor-field`: legacy's `klarnaStreams()` catalog
publishes no `CursorFields` for any stream (the Settlements API supports full refresh only), so this
bundle matches that exactly.

## Write actions & risks

None. The Klarna Settlements API is read-only for pm reverse-ETL purposes; `capabilities.write` is
`false` and no `writes.json` is shipped, matching legacy's `Write` stub
(`connectors.ErrUnsupportedOperation`).

## Known limits

- **Region/playground-derived `base_url` is not reproduced; `base_url` is required instead.**
  Legacy derives the API host from a `region` config value (`eu`/`na`/`oc`, defaulting to `eu`) plus
  a `playground` boolean that rewrites the host to its `*.playground.klarna.com` counterpart
  (`klarnaBaseURL`/`klarnaRegionHosts`), only falling back to an explicit `base_url` override when
  one is set. The engine's `streams.json` `base.url` is a single template with no per-value
  conditional/lookup-table mechanism (`spec.json`'s `"default"` materialization is a single fixed
  literal, not a derived function of another config value — see
  `docs/migration/conventions.md`'s `"default"` materialization section, which explicitly calls out
  this exact shape as needing either a required `base_url` or a future computed-base-URL dialect
  extension). This bundle takes the documented, honest path: `base_url` is `required`, with its
  description enumerating the 6 real hosts (3 regions × production/playground) the caller must
  choose from explicitly. This is a config-surface narrowing from legacy's region-shorthand
  convenience, not a data/behavior deviation — every request Klarna itself would accept still
  succeeds once the caller supplies the correct literal host.
- **`username`'s dual Config/Secrets lookup is narrowed to Secrets-only.** Legacy checks
  `cfg.Secrets["username"]` first and only falls back to `cfg.Config["username"]` if that's empty.
  Since Secrets is checked first (i.e., is the precedent path when both single-value dialects are
  compared), this bundle sources `username` from `secrets.username` only. A caller who previously
  supplied `username` only via `Config` (the fallback path) must instead supply it as a secret; this
  is a config-shape narrowing, not a behavior change for the common (Secrets-configured) case.
- **`page_size`/`max_pages` runtime overrides are not modeled.** Legacy accepts `config.page_size`
  (1-500, default 100) and `config.max_pages` (0/`all`/`unlimited` = unbounded). The engine's
  `offset_limit` paginator reads `page_size`/`max_pages` only from fixed `streams.json` pagination
  fields, not from runtime config, so `spec.json` intentionally does not declare these dead config
  properties. The bundle matches legacy defaults: `page_size: 100` and unbounded pages.
- The four report endpoints (`/reports/payout-with-transactions`, `/reports/payout`,
  `/reports/payouts-summary-with-transactions`, and `/reports/payouts-summary`) return CSV or PDF
  payloads. They are accounted for in `api_surface.json` as `binary_payload` exclusions because the
  declarative stream engine consumes JSON records and this pass may not add a CSV/PDF hook package.
