# Yahoo Finance Price Connector

## Overview

Reads public Yahoo Finance chart prices as declaration-bound OHLCV records. Read-only.

Readable streams: `prices`.

Service API documentation: https://www.yahoofinanceapi.com/.

## Auth setup

Connection fields:

- `interval` (optional, string); The interval of between prices queried.
- `range` (optional, string); The range of prices to be queried.
- `symbol` (optional, string); Ticker symbol to query through the fixed Yahoo Finance chart route.

Authentication uses declared mode(s): `none`.

## Execution contract

Connection check: `GET /v8/finance/chart/{{ config.symbol }}`
Check query: `interval`=`{{ config.interval }}`; `range`=`{{ config.range }}`.

## Streams notes

- `prices`: `GET /v8/finance/chart/{{ config.symbol }}`; records `chart.result`
  - Query: `interval`=`{{ config.interval }}`; `range`=`{{ config.range }}`.
  - Incremental cursor: `timestamp`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.


## Declared response errors

- `prices`: `message_field`=`description`, `path`=`chart.error`.
