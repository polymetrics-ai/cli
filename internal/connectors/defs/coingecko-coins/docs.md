# Overview

Reads a coin's current metadata/market snapshot and exchange tickers from the CoinGecko REST API
(GET /coins/{id}, GET /coins/{id}/tickers). Read-only; unauthenticated by default, an optional pro
api_key unlocks the pro base URL and higher limits.

Readable streams: `coin`, `tickers`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.coingecko.com/reference/introduction.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Optional CoinGecko pro API key, sent on the x-cg-pro-api-key
  header. Unauthenticated calls against the public base URL are allowed; a key unlocks the pro base
  URL and higher limits. Never logged.
- `base_url` (required, string); format `uri`; CoinGecko API base URL, e.g.
  https://api.coingecko.com/api/v3 (public) or https://pro-api.coingecko.com/api/v3 (pro, when
  api_key is set).
- `coin_id` (required, string); CoinGecko coin id (e.g. bitcoin) to read the metadata/market
  snapshot for.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- API key authentication in `x-cg-pro-api-key` using `secrets.api_key` when `{{ secrets.api_key }}`.
- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/ping`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `coin`; page_number: `tickers`.

- `coin`: GET `/coins/{{ config.coin_id }}` - records path `.`; query `community_data`=`false`;
  `developer_data`=`false`; `localization`=`false`; `tickers`=`false`.
- `tickers`: GET `/coins/{{ config.coin_id }}/tickers` - records path `tickers`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 100; computed
  output fields `market_identifier`.

## Write actions & risks

This connector is read-only. Read behavior: external CoinGecko public API read of a single coin's
metadata/market snapshot.

## Known limits

- Batch defaults: read_page_size=1.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=59, requires_elevated_scope=25.
