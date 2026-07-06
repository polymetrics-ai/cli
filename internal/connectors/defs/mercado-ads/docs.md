# Overview

Reads Mercado Ads brand, display, and product advertisers and daily campaign metrics from the
Mercado Libre Advertising API.

Readable streams: `brand_advertisers`, `display_advertisers`, `product_advertisers`,
`brand_campaigns_metrics`, `display_campaigns_metrics`, `product_campaigns_metrics`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.mercadolibre.com.ar/en_us/advertising.

## Auth setup

Connection fields:

- `base_url` (optional, string).
- `client_id` (required, secret, string).
- `client_refresh_token` (required, secret, string).
- `client_secret` (required, secret, string).
- `end_date` (optional, string); Cannot exceed 90 days from current day for Product Ads.
- `lookback_days` (required, string).
- `mode` (optional, string).
- `start_date` (optional, string); Cannot exceed 90 days from current day for Product Ads, and 90
  days from "End Date" on Brand and Display Ads.

Secret fields are redacted in logs and write previews: `client_id`, `client_refresh_token`,
`client_secret`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `brand_advertisers`: GET connector-managed request path - records path `data`.
- `display_advertisers`: GET connector-managed request path - records path `data`.
- `product_advertisers`: GET connector-managed request path - records path `data`.
- `brand_campaigns_metrics`: GET connector-managed request path - records path `data`; incremental
  cursor `date`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `display_campaigns_metrics`: GET connector-managed request path - records path `data`; incremental
  cursor `date`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `product_campaigns_metrics`: GET connector-managed request path - records path `data`; incremental
  cursor `date`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 6 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `brand_campaigns_metrics`,
  `display_campaigns_metrics`, `product_campaigns_metrics`.
