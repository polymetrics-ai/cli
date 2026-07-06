# Overview

Reads ShipStation orders, shipments, products, and customers through the ShipStation REST API.

Readable streams: `orders`, `shipments`, `products`, `customers`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.shipstation.com/docs/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); ShipStation API key, used as the HTTP Basic auth username.
  Never logged.
- `api_secret` (required, secret, string); ShipStation API secret, used as the HTTP Basic auth
  password. Never logged.
- `base_url` (optional, string); default `https://ssapi.shipstation.com`; format `uri`; ShipStation
  API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`, `api_secret`.

Default configuration values: `base_url=https://ssapi.shipstation.com`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`, `secrets.api_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/orders` with query `page`=`1`; `pageSize`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `pageSize`; starts
at 1; page size 100.

- `orders`: GET `/orders` - records path `orders`; page-number pagination; page parameter `page`;
  size parameter `pageSize`; starts at 1; page size 100; computed output fields `id`, `modified_at`,
  `order_number`, `status`.
- `shipments`: GET `/shipments` - records path `shipments`; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1; page size 100; computed output fields `id`,
  `modified_at`, `order_number`, `status`.
- `products`: GET `/products` - records path `products`; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1; page size 100; computed output fields `id`,
  `modified_at`, `name`.
- `customers`: GET `/customers` - records path `customers`; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1; page size 100; computed output fields `id`,
  `modified_at`, `name`.

## Write actions & risks

This connector is read-only. Read behavior: external ShipStation API read of order, shipment,
product, and customer data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
