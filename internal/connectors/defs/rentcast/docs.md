# Overview

Reads RentCast properties, sale listings, rental listings, market data, and value/rental estimates
through the RentCast REST API. Read-only.

Readable streams: `properties`, `sale_listings`, `rental_listings`, `markets`, `value_estimates`,
`rental_estimates`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.rentcast.io/.

## Auth setup

Connection fields:

- `address` (optional, string).
- `api_key` (required, secret, string); RentCast API key, sent as the X-Api-Key header. Never
  logged.
- `base_url` (optional, string); default `https://api.rentcast.io/v1`; format `uri`; RentCast API
  base URL override for tests or proxies.
- `city` (optional, string); Optional city filter applied to every stream's request.
- `property_type` (optional, string); Optional property type filter applied to every stream's
  request (RentCast's propertyType query parameter).
- `state` (optional, string); Optional state filter applied to every stream's request.
- `zip_code` (optional, string); Optional ZIP code filter applied to every stream's request
  (RentCast's zipCode query parameter).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.rentcast.io/v1`.

Authentication behavior:

- API key authentication in `X-Api-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/properties` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `properties`: GET `/properties` - records at response root; query `address` from template `{{
  config.address }}`, omitted when absent; `city` from template `{{ config.city }}`, omitted when
  absent; `propertyType` from template `{{ config.property_type }}`, omitted when absent; `state`
  from template `{{ config.state }}`, omitted when absent; `zipCode` from template `{{
  config.zip_code }}`, omitted when absent; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100; incremental cursor `last_seen_date`; formatted as
  `rfc3339`; computed output fields `address`, `last_seen_date`, `property_type`, `zip_code`.
- `sale_listings`: GET `/listings/sale` - records at response root; query `address` from template
  `{{ config.address }}`, omitted when absent; `city` from template `{{ config.city }}`, omitted
  when absent; `propertyType` from template `{{ config.property_type }}`, omitted when absent;
  `state` from template `{{ config.state }}`, omitted when absent; `zipCode` from template `{{
  config.zip_code }}`, omitted when absent; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100; incremental cursor `last_seen_date`; formatted as
  `rfc3339`; computed output fields `address`, `last_seen_date`, `property_type`.
- `rental_listings`: GET `/listings/rental/long-term` - records at response root; query `address`
  from template `{{ config.address }}`, omitted when absent; `city` from template `{{ config.city
  }}`, omitted when absent; `propertyType` from template `{{ config.property_type }}`, omitted when
  absent; `state` from template `{{ config.state }}`, omitted when absent; `zipCode` from template
  `{{ config.zip_code }}`, omitted when absent; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100; incremental cursor `last_seen_date`; formatted as
  `rfc3339`; computed output fields `address`, `last_seen_date`, `property_type`.
- `markets`: GET `/markets` - records at response root; query `address` from template `{{
  config.address }}`, omitted when absent; `city` from template `{{ config.city }}`, omitted when
  absent; `propertyType` from template `{{ config.property_type }}`, omitted when absent; `state`
  from template `{{ config.state }}`, omitted when absent; `zipCode` from template `{{
  config.zip_code }}`, omitted when absent; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100; computed output fields `zip_code`.
- `value_estimates`: GET `/avm/value` - records at response root; query `address` from template `{{
  config.address }}`, omitted when absent; `city` from template `{{ config.city }}`, omitted when
  absent; `propertyType` from template `{{ config.property_type }}`, omitted when absent; `state`
  from template `{{ config.state }}`, omitted when absent; `zipCode` from template `{{
  config.zip_code }}`, omitted when absent; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100; computed output fields `address`, `id`.
- `rental_estimates`: GET `/avm/rent/long-term` - records at response root; query `address` from
  template `{{ config.address }}`, omitted when absent; `city` from template `{{ config.city }}`,
  omitted when absent; `propertyType` from template `{{ config.property_type }}`, omitted when
  absent; `state` from template `{{ config.state }}`, omitted when absent; `zipCode` from template
  `{{ config.zip_code }}`, omitted when absent; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100; computed output fields `address`, `id`.

## Write actions & risks

This connector is read-only. Read behavior: external RentCast API read of property, listing, market,
and valuation data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
