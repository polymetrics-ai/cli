# Overview

Reads Shippo addresses, parcels, shipments, and transactions through the Shippo REST API.

Readable streams: `addresses`, `parcels`, `shipments`, `transactions`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.goshippo.com/docs/intro.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Shippo API token, sent as an Authorization header
  (Authorization: ShippoToken <api_token>). Never logged.
- `base_url` (optional, string); default `https://api.goshippo.com`; format `uri`; Shippo API base
  URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.goshippo.com`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `ShippoToken` using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/addresses` with query `page`=`1`; `results`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next`; next URLs stay
on the configured API host.

- `addresses`: GET `/addresses` - records path `results`; query `results`=`100`; follows a next-page
  URL from the response body; URL path `next`; next URLs stay on the configured API host; computed
  output fields `id`, `name`, `updated_at`.
- `parcels`: GET `/parcels` - records path `results`; query `results`=`100`; follows a next-page URL
  from the response body; URL path `next`; next URLs stay on the configured API host; computed
  output fields `id`, `name`, `updated_at`.
- `shipments`: GET `/shipments` - records path `results`; query `results`=`100`; follows a next-page
  URL from the response body; URL path `next`; next URLs stay on the configured API host; computed
  output fields `id`, `name`, `updated_at`.
- `transactions`: GET `/transactions` - records path `results`; query `results`=`100`; follows a
  next-page URL from the response body; URL path `next`; next URLs stay on the configured API host;
  computed output fields `id`, `name`, `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Shippo API read of address, parcel, shipment,
and transaction data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
