# Overview

StockData reads market data and news through the stockdata.org REST API
(`https://api.stockdata.org/v1`). This bundle migrates `internal/connectors/stockdata` while
preserving the legacy stream record mappings for `tickers`, `eod_prices`, `intraday_prices`, and
`news`.

Pass B reviewed the live HTML documentation on 2026-07-04. The current docs expose 16 GET
endpoints and no documented POST/PUT/PATCH/DELETE endpoints. This bundle covers every documented
object/list GET endpoint, plus the legacy-only `GET /v1/data/tickers` stream that the current docs
page no longer lists. Scalar reference-list endpoints for entity types and industries are excluded
as non-data/reference endpoints in `api_surface.json`.

## Auth setup

Provide a StockData API access token via the `api_token` secret. It is sent as the `api_token`
query parameter on every request (`mode: api_key_query`), matching legacy's
`connsdk.APIKeyQuery("api_token", token)` exactly; it is never logged.

Market data streams that need symbols use the `symbols` config key. The legacy streams pass
`date_from`/`date_to` through whenever configured, matching legacy's shared `baseQuery`. News and stats streams accept
optional filters such as `language`, `published_after`, `published_before`, `published_on`,
`countries`, `interval`, `group_by`, `sentiment_gte`, and `sentiment_lte`. UUID-scoped news streams
use comma-separated `news_uuids` fan-out. The `entity_search` stream requires the `entity_search`
config value because StockData documents `search` as the lookup term for that endpoint.

## Streams notes

Legacy streams keep their legacy schema-mode projections:
`tickers` emits `symbol`, `name`, and `exchange`; `eod_prices` and `intraday_prices` emit
`ticker`, `date`, and `close`; `news` emits `title`, `url`, and `published_at`.

New Pass B streams use `projection: "passthrough"` so StockData response fields outside the core
schema are not dropped. They cover quotes, adjusted intraday prices, splits, dividends, similar
news, news by UUID, news entity stats, entity search, and news sources. `news_by_uuid` is a
single-object stream with `pagination.type: none`; list-style endpoints inherit the base
`page`/`limit` page-number pagination.

`news_similar` and `news_by_uuid` use `fan_out.ids_from.config_key: news_uuids`. If no UUIDs are
configured, those streams emit no records. With UUIDs configured, the engine runs one request per
UUID and stamps the UUID onto emitted records.

## Write actions & risks

None. StockData is read-only in legacy (`Capabilities.Write` is `false`), and the current
documentation exposes no POST/PUT/PATCH/DELETE endpoints. No `writes.json` is shipped for this
bundle.

## Known limits

- **Dead runtime pagination config is removed.** Legacy accepts `page_size` and `max_pages`, but
  the declarative paginator uses bundle-authored values. This bundle keeps StockData's legacy
  default `limit=100` and does not declare unwired config keys.
- **Legacy fixture-mode-only fields are not modeled.** Legacy stamps `connector` and `fixture` on
  synthetic fixture records. Declarative fixture replay replaces that test-only path.
- **Entity type/industry reference lists are excluded.** These endpoints return static scalar
  reference arrays, not account/market object records; the current records dialect also does not
  fan out scalar arrays into records.
- **No rate-limit block is declared.** Legacy enforces no client-side rate limiting for StockData,
  so none is added here.
