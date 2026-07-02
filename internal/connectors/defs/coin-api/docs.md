# Overview

CoinAPI is a read-only market-data source. This bundle migrates `internal/connectors/coin-api` (the
hand-written legacy connector) to a declarative Tier-1 bundle at parity: 3 metadata reference
streams (`symbols`, `exchanges`, `assets` — `GET /v1/<resource>` returning a top-level JSON array)
and 2 symbol-scoped historical streams (`ohlcv_historical_data`, `trades_historical_data` —
`GET /v1/<resource>/<symbol_id>/history`, paginated by `limit` + advancing `time_start` past the
last record's time cursor). The legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a CoinAPI key via the `api_key` secret; it is sent as the `X-CoinAPI-Key` request header
(`streams.json` `base.auth`'s `api_key_header` mode) and never logged, matching legacy's
`connsdk.APIKeyHeader(coinAPIAuthHeader, secret, "")` exactly.

`base_url` optionally overrides the API host (for tests/proxies). When unset, `environment`
(`production` default, or `sandbox`) selects between `https://rest.coinapi.io` and
`https://rest-sandbox.coinapi.io` — legacy's own `coinAPIBaseURL` fallback logic. Only the fixed
production/sandbox pair is representable via the engine's plain `{{ config.base_url }}` +
`"default"` materialization mechanism; the `environment`-conditional derivation itself is left to
operator convention (set `base_url` explicitly for sandbox), documented under Known limits below.

## Streams notes

`symbols`, `exchanges`, and `assets` are full-refresh reference lists: `GET /v1/<resource>` with no
query parameters, records read from the top-level JSON array (`records.path: ""`), no pagination
(`type: none`) — matching legacy's `readMetadata`, which issues exactly one request per stream.

`ohlcv_historical_data` and `trades_historical_data` are symbol-scoped historical series requiring
`symbol_id` (and, for OHLCV, `period`) in config. Both send `limit` (default `100`, materialized
from `spec.json`'s default exactly like legacy's `coinAPIDefaultLimit` fallback) and an optional
`time_end` (`config.end_date`, `omit_when_absent: true` — sent only when configured, matching
legacy's conditional `base.Set("time_end", timeEnd)`). `ohlcv_historical_data` additionally sends
`period_id` (`config.period`, required — legacy hard-errors when absent for this stream; this bundle
relies on the same absent-value-propagates-as-empty-string engine behavior, so an unset `period`
sends an empty `period_id` rather than erroring, a minor parity note — see Known limits).

Pagination is `type: cursor` with `last_record_field` (not `token_path`): the next page's
`time_start` is read from the LAST record's own time-cursor field
(`time_period_start`/`time_exchange`) rather than a separate token in the response envelope — this
is exactly legacy's `readTimeseries` loop (`timeStart = lastCursor`, where `lastCursor` comes from
`stringField(item, endpoint.cursorField)` on the last emitted record of the page). No `stop_path` is
declared: like legacy (`len(records) < limit || lastCursor == "" || lastCursor == timeStart`), the
engine's `lastRecordCursor` paginator stops when a page is empty (`recordCount == 0`) or the last
record has no usable cursor value; the engine does not special-case a *short-but-nonempty* page the
way legacy's `len(records) < limit` check does, so a short final page triggers one additional
request that returns empty before the read terminates — never emits duplicate or incorrect data,
just one harmless extra round trip (documented parity deviation, see Known limits).

Incremental reads use `time_period_start`/`time_exchange` as the cursor field and send `time_start`
(`param_format: rfc3339`, verbatim passthrough — CoinAPI's ISO-8601 wire format needs no
reformatting, matching legacy's raw string `time_start`/`start_date` passthrough) computed from the
persisted state cursor or, on a fresh sync, from `start_date`. `symbol_id` (and `period` for OHLCV)
are stamped onto every emitted record via `computed_fields` referencing `config.*` (never present on
the raw wire record) — matching legacy's explicit `rec["symbol_id"] = symbolID` /
`rec["period_id"] = period` assignments in `readTimeseries`.

## Write actions & risks

None. CoinAPI is a read-only market-data API; `capabilities.write` is `false` and this bundle ships
no `writes.json`, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **`environment`-based sandbox/production base URL selection is not fully modeled.** Legacy derives
  `base_url` from `environment` in code (`coinAPIBaseURL`: `production` → `https://rest.coinapi.io`,
  `sandbox` → `https://rest-sandbox.coinapi.io`, only when no explicit `base_url` override is set).
  The engine's `spec.json` `"default"` materialization only fills a fixed literal, not a value
  derived from another config field, so this bundle declares `environment` (documented, with its own
  enum and default) alongside `base_url` for operator guidance, but `streams.json`'s `base.url`
  templates `{{ config.base_url }}` directly with no default — an operator must set `base_url`
  explicitly (to either CoinAPI host) rather than relying on `environment` alone to select it. This
  is a config-surface narrowing versus legacy's in-code derivation, not a behavior change for any
  input that explicitly sets `base_url`.
- **A short-but-nonempty final page triggers one extra, empty-page request** before pagination
  stops, versus legacy's `len(records) < limit` short-page check which stops immediately on the
  final page itself. This never changes emitted record data (the extra request returns zero
  records) — see Streams notes above.
- **`ohlcv_historical_data`'s `period` config is required for correct behavior but not enforced at
  validate time.** Legacy hard-errors when `period` is unset for this stream
  (`"coin-api config period is required for ohlcv_historical_data"`); this bundle's declarative
  `period_id` query param has no equivalent required-at-read-time check beyond the engine's
  standard unresolved-key error path. `symbol_id` is documented as required in `spec.json`'s
  description for both historical streams for the same reason (the engine's draft-07 dialect has no
  per-stream conditional `required[]`).
- Full CoinAPI API surface (current quotes, orderbooks, exchange rates, indexes, metrics) is out of
  scope for this migration; see `api_surface.json`'s `excluded: {category: out_of_scope, reason:
  "Pass B capability expansion"}` entries. Only the 3 metadata streams and 2 historical streams
  legacy itself implements are covered.
