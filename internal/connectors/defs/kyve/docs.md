# Overview

Reads public KYVE pools, stakers, funders, and Cosmos validators through the KYVE network's public
REST query endpoints. Read-only; no credentials required.

Readable streams: `pools`, `stakers`, `funders`, `validators`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.kyve.network/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.korellia.kyve.network`; format `uri`; KYVE
  Cosmos-style REST query endpoint base URL. No credentials required; defaults to the public KYVE
  Korellia network endpoint.
- `max_pages` (optional, string); default `0`; Maximum pages to read; use 0, all, or unlimited to
  exhaust the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000), sent as the
  pagination.limit query parameter.

Default configuration values: `base_url=https://api.korellia.kyve.network`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/kyve/query/v1beta1/pools` with query `pagination.limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `pagination.key`; next token from
`pagination.next_key`.

- `pools`: GET `/kyve/query/v1beta1/pools` - records path `pools`; query `pagination.limit`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `pagination.key`; next token from
  `pagination.next_key`; computed output fields `id`, `name`, `runtime`.
- `stakers`: GET `/kyve/query/v1beta1/stakers` - records path `stakers`; query
  `pagination.limit`=`{{ config.page_size }}`; cursor pagination; cursor parameter `pagination.key`;
  next token from `pagination.next_key`; computed output fields `address`, `amount`.
- `funders`: GET `/kyve/query/v1beta1/funders` - records path `funders`; query
  `pagination.limit`=`{{ config.page_size }}`; cursor pagination; cursor parameter `pagination.key`;
  next token from `pagination.next_key`; computed output fields `address`, `amount`.
- `validators`: GET `/cosmos/staking/v1beta1/validators` - records path `validators`; query
  `pagination.limit`=`{{ config.page_size }}`; cursor pagination; cursor parameter `pagination.key`;
  next token from `pagination.next_key`; computed output fields `moniker`, `operator_address`,
  `status`.

## Write actions & risks

This connector is read-only. Read behavior: external read of public KYVE network
pool/staker/funder/validator data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
