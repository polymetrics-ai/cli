# Overview

Reads StockData market quotes, prices, splits, dividends, news, entity stats, entities, and source
metadata through the StockData API.

Readable streams: `tickers`, `eod_prices`, `intraday_prices`, `news`, `quotes`,
`intraday_adjusted_prices`, `splits`, `dividends`, `news_similar`, `news_by_uuid`,
`news_stats_intraday`, `news_stats_aggregation`, `news_stats_trending`, `entity_search`,
`news_sources`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.stockdata.org/documentation.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); StockData API access token, sent as the api_token query
  parameter. Never logged.
- `base_url` (optional, string); default `https://api.stockdata.org/v1`; format `uri`; StockData API
  base URL override for tests or proxies.
- `countries` (optional, string); Optional country filter for entity stats endpoints.
- `date_from` (optional, string).
- `date_to` (optional, string).
- `entity_search` (optional, string); Search term required by the entity_search stream.
- `entity_types` (optional, string); Optional comma-separated entity type filter for the
  entity_search stream.
- `group_by` (optional, string); Optional grouping field for entity stats requests.
- `group_domains` (optional, string); Optional group-domain flag for the news_sources stream.
- `interval` (optional, string); Optional interval for entity stats time-series requests.
- `language` (optional, string); Optional language filter for news and source endpoints.
- `news_uuids` (optional, string); Comma-separated news UUIDs for news_similar and news_by_uuid
  fan-out streams.
- `published_after` (optional, string); Optional lower-bound publication timestamp for news/stat
  endpoints.
- `published_before` (optional, string); Optional upper-bound publication timestamp for news/stat
  endpoints.
- `published_on` (optional, string); Optional publication date filter for news/stat endpoints.
- `sentiment_gte` (optional, string); Optional minimum sentiment filter for news/stat endpoints.
- `sentiment_lte` (optional, string); Optional maximum sentiment filter for news/stat endpoints.
- `symbols` (optional, string); Comma-separated ticker symbols. Required for quote and market data
  streams; optional for news filter streams.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.stockdata.org/v1`.

Authentication behavior:

- API key authentication in query parameter `api_token` using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/news/all` with query `limit`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

Pagination by stream: none: `news_by_uuid`; page_number: `tickers`, `eod_prices`, `intraday_prices`,
`news`, `quotes`, `intraday_adjusted_prices`, `splits`, `dividends`, `news_similar`,
`news_stats_intraday`, `news_stats_aggregation`, `news_stats_trending`, `entity_search`,
`news_sources`.

- `tickers`: GET `/data/tickers` - records path `data`; query `date_from` from template `{{
  config.date_from }}`, omitted when absent; `date_to` from template `{{ config.date_to }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1;
  page size 100.
- `eod_prices`: GET `/data/eod` - records path `data`; query `date_from` from template `{{
  config.date_from }}`, omitted when absent; `date_to` from template `{{ config.date_to }}`, omitted
  when absent; `symbols`=`{{ config.symbols }}`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100.
- `intraday_prices`: GET `/data/intraday` - records path `data`; query `date_from` from template `{{
  config.date_from }}`, omitted when absent; `date_to` from template `{{ config.date_to }}`, omitted
  when absent; `symbols`=`{{ config.symbols }}`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100.
- `news`: GET `/news/all` - records path `data`; query `date_from` from template `{{
  config.date_from }}`, omitted when absent; `date_to` from template `{{ config.date_to }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1;
  page size 100.
- `quotes`: GET `/data/quote` - records path `data`; query `date_from` from template `{{
  config.date_from }}`, omitted when absent; `date_to` from template `{{ config.date_to }}`, omitted
  when absent; `symbols`=`{{ config.symbols }}`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `intraday_adjusted_prices`: GET `/data/intraday/adjusted` - records path `data`; query `date_from`
  from template `{{ config.date_from }}`, omitted when absent; `date_to` from template `{{
  config.date_to }}`, omitted when absent; `symbols`=`{{ config.symbols }}`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough
  records.
- `splits`: GET `/data/splits` - records path `data`; query `date_from` from template `{{
  config.date_from }}`, omitted when absent; `date_to` from template `{{ config.date_to }}`, omitted
  when absent; `symbols`=`{{ config.symbols }}`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `dividends`: GET `/data/dividends` - records path `data`; query `date_from` from template `{{
  config.date_from }}`, omitted when absent; `date_to` from template `{{ config.date_to }}`, omitted
  when absent; `symbols`=`{{ config.symbols }}`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `news_similar`: GET `/news/similar/{{ fanout.id }}` - records path `data`; query `countries` from
  template `{{ config.countries }}`, omitted when absent; `group_by` from template `{{
  config.group_by }}`, omitted when absent; `interval` from template `{{ config.interval }}`,
  omitted when absent; `language` from template `{{ config.language }}`, omitted when absent;
  `published_after` from template `{{ config.published_after }}`, omitted when absent;
  `published_before` from template `{{ config.published_before }}`, omitted when absent;
  `published_on` from template `{{ config.published_on }}`, omitted when absent; `sentiment_gte`
  from template `{{ config.sentiment_gte }}`, omitted when absent; `sentiment_lte` from template `{{
  config.sentiment_lte }}`, omitted when absent; `symbols` from template `{{ config.symbols }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `limit`; starts
  at 1; page size 100; fan-out; ids from config field `news_uuids`; id inserted into the request
  path; stamps `uuid`; emits passthrough records.
- `news_by_uuid`: GET `/news/uuid/{{ fanout.id }}` - single-object response; records at response
  root; fan-out; ids from config field `news_uuids`; id inserted into the request path; stamps
  `uuid`; emits passthrough records.
- `news_stats_intraday`: GET `/news/stats/intraday` - records path `data`; query `countries` from
  template `{{ config.countries }}`, omitted when absent; `group_by` from template `{{
  config.group_by }}`, omitted when absent; `interval` from template `{{ config.interval }}`,
  omitted when absent; `language` from template `{{ config.language }}`, omitted when absent;
  `published_after` from template `{{ config.published_after }}`, omitted when absent;
  `published_before` from template `{{ config.published_before }}`, omitted when absent;
  `published_on` from template `{{ config.published_on }}`, omitted when absent; `sentiment_gte`
  from template `{{ config.sentiment_gte }}`, omitted when absent; `sentiment_lte` from template `{{
  config.sentiment_lte }}`, omitted when absent; `symbols` from template `{{ config.symbols }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `limit`; starts
  at 1; page size 100; emits passthrough records.
- `news_stats_aggregation`: GET `/news/stats/aggregation` - records path `data`; query `countries`
  from template `{{ config.countries }}`, omitted when absent; `group_by` from template `{{
  config.group_by }}`, omitted when absent; `interval` from template `{{ config.interval }}`,
  omitted when absent; `language` from template `{{ config.language }}`, omitted when absent;
  `published_after` from template `{{ config.published_after }}`, omitted when absent;
  `published_before` from template `{{ config.published_before }}`, omitted when absent;
  `published_on` from template `{{ config.published_on }}`, omitted when absent; `sentiment_gte`
  from template `{{ config.sentiment_gte }}`, omitted when absent; `sentiment_lte` from template `{{
  config.sentiment_lte }}`, omitted when absent; `symbols` from template `{{ config.symbols }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `limit`; starts
  at 1; page size 100; emits passthrough records.
- `news_stats_trending`: GET `/news/stats/trending` - records path `data`; query `countries` from
  template `{{ config.countries }}`, omitted when absent; `group_by` from template `{{
  config.group_by }}`, omitted when absent; `interval` from template `{{ config.interval }}`,
  omitted when absent; `language` from template `{{ config.language }}`, omitted when absent;
  `published_after` from template `{{ config.published_after }}`, omitted when absent;
  `published_before` from template `{{ config.published_before }}`, omitted when absent;
  `published_on` from template `{{ config.published_on }}`, omitted when absent; `sentiment_gte`
  from template `{{ config.sentiment_gte }}`, omitted when absent; `sentiment_lte` from template `{{
  config.sentiment_lte }}`, omitted when absent; `symbols` from template `{{ config.symbols }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `limit`; starts
  at 1; page size 100; emits passthrough records.
- `entity_search`: GET `/entity/search` - records path `data`; query `search`=`{{
  config.entity_search }}`; `types` from template `{{ config.entity_types }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `news_sources`: GET `/news/sources` - records path `data`; query `group_domains` from template `{{
  config.group_domains }}`, omitted when absent; `language` from template `{{ config.language }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `limit`; starts
  at 1; page size 100; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external StockData API read of ticker, quote, price,
split, dividend, news, entity, stats, and source data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 15 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=2.
