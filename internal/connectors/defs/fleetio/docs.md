# Overview

Reads Fleetio fleet management data: vehicles, contacts, fuel entries, issues, and service entries
through the Fleetio REST API.

Readable streams: `vehicles`, `contacts`, `fuel_entries`, `issues`, `service_entries`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.fleetio.com/.

## Auth setup

Connection fields:

- `account_token` (required, secret, string); Fleetio account token. Sent as the Account-Token
  request header on every request.
- `api_key` (required, secret, string); Fleetio API key. Sent as Authorization: Token <api_key>.
- `base_url` (optional, string); default `https://secure.fleetio.com/api/v1`; format `uri`; Fleetio
  API root. Defaults to https://secure.fleetio.com/api/v1. Also usable as a base URL override for
  tests/proxies.
- `page_size` (optional, integer); default `100`; Number of records requested per page (per_page).
  Between 1 and 100.

Secret fields are redacted in logs and write previews: `account_token`, `api_key`.

Default configuration values: `base_url=https://secure.fleetio.com/api/v1`, `page_size=100`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/vehicles`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `start_cursor`; next token from
`next_cursor`.

- `vehicles`: GET `/vehicles` - records path `records`; query `per_page` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `start_cursor`; next
  token from `next_cursor`.
- `contacts`: GET `/contacts` - records path `records`; query `per_page` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `start_cursor`; next
  token from `next_cursor`.
- `fuel_entries`: GET `/fuel_entries` - records path `records`; query `per_page` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `start_cursor`; next
  token from `next_cursor`.
- `issues`: GET `/issues` - records path `records`; query `per_page` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `start_cursor`; next
  token from `next_cursor`.
- `service_entries`: GET `/service_entries` - records path `records`; query `per_page` from template
  `{{ config.page_size }}`, default `100`; cursor pagination; cursor parameter `start_cursor`; next
  token from `next_cursor`.

## Write actions & risks

This connector is read-only. Read behavior: external Fleetio API read of vehicle, contact, fuel
entry, issue, and service entry data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=7.
