# Overview

Reads DefiLlama DeFi analytics: protocols, chains, stablecoins, DEX volumes, and fees/revenue from
the public DefiLlama REST API. Read-only; no authentication required.

Readable streams: `protocols`, `chains`, `stablecoins`, `dexs`, `fees`, `options`, `open_interest`,
`pools`, `stablecoin_chains`, `historical_chain_tvl`.

This connector is read-only; no write actions are declared.

Service API documentation: https://defillama.com/docs/api.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.llama.fi`; format `uri`; DefiLlama main API
  base URL override for tests or proxies. Used by protocols, chains, dexs, and fees streams.
- `mode` (optional, string).

Default configuration values: `base_url=https://api.llama.fi`.

Authentication is handled by the connector-specific implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/protocols` with query `limit`=`1`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `stablecoins`, `dexs`, `fees`, `options`, `open_interest`, `pools`,
`stablecoin_chains`, `historical_chain_tvl`; offset_limit: `protocols`, `chains`.

- `protocols`: GET `/protocols` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 1000.
- `chains`: GET `/v2/chains` - records path `.`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 1000.
- `stablecoins`: GET `https://stablecoins.llama.fi/stablecoins` - records path `peggedAssets`; query
  `includePrices`=`true`.
- `dexs`: GET `/overview/dexs` - records path `protocols`; query `excludeTotalDataChart`=`true`;
  `excludeTotalDataChartBreakdown`=`true`.
- `fees`: GET `/overview/fees` - records path `protocols`; query `excludeTotalDataChart`=`true`;
  `excludeTotalDataChartBreakdown`=`true`.
- `options`: GET `/overview/options` - records path `protocols`; query
  `excludeTotalDataChart`=`true`; `excludeTotalDataChartBreakdown`=`true`.
- `open_interest`: GET `/overview/open-interest` - records path `protocols`; query
  `excludeTotalDataChart`=`true`; `excludeTotalDataChartBreakdown`=`true`.
- `pools`: GET `https://yields.llama.fi/pools` - records path `data`.
- `stablecoin_chains`: GET `https://stablecoins.llama.fi/stablecoinchains` - records path `.`.
- `historical_chain_tvl`: GET `/v2/historicalChainTvl` - records path `.`.

## Write actions & risks

This connector is read-only. Read behavior: external DefiLlama API read of public DeFi analytics
data.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 10 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=1, duplicate_of=8, out_of_scope=7, requires_elevated_scope=6.
