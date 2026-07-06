# Overview

Reads PrestaShop customers, orders, products, addresses, and carts through the PrestaShop Webservice
REST API.

Readable streams: `customers`, `orders`, `products`, `addresses`, `carts`.

This connector is read-only; no write actions are declared.

Service API documentation: https://devdocs.prestashop-project.org/9/webservice/.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); Your PrestaShop access key. See <a
  href="https://devdocs.prestashop.com/1.7/webservice/tutorials/creating-access/#create-an-access-key">
  the docs </a> for info on how to obtain this.
- `base_url` (optional, string).
- `mode` (optional, string).
- `start_date` (required, string); The Start date in the format YYYY-MM-DD.
- `url` (required, string); Shop URL without trailing slash.

Secret fields are redacted in logs and write previews: `access_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `customers`: GET connector-managed request path - records path `data`; incremental cursor
  `date_upd`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `orders`: GET connector-managed request path - records path `data`; incremental cursor `date_upd`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `products`: GET connector-managed request path - records path `data`; incremental cursor
  `date_upd`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `addresses`: GET connector-managed request path - records path `data`; incremental cursor
  `date_upd`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `carts`: GET connector-managed request path - records path `data`; incremental cursor `date_upd`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `customers`, `orders`, `products`, `addresses`,
  `carts`.
