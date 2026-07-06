# Overview

Reads BlueTally IT asset management data (assets, employees, licenses, maintenances, accessories)
through the BlueTally REST API.

Readable streams: `assets`, `employees`, `licenses`, `maintenances`, `accessories`, `components`,
`consumables`, `categories`, `departments`, `depreciations`, `locations`, `manufacturers`,
`products`, `statuses`, `suppliers`, `audits`, `activity`, `tenants`.

This connector is read-only; no write actions are declared.

Service API documentation: https://bluetally.readme.io/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); BlueTally API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://app.bluetallyapp.com`; format `uri`; BlueTally API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.bluetallyapp.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v1/assets` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 50.

Pagination by stream: none: `tenants`; offset_limit: `assets`, `employees`, `licenses`,
`maintenances`, `accessories`, `components`, `consumables`, `categories`, `departments`,
`depreciations`, `locations`, `manufacturers`, `products`, `statuses`, `suppliers`, `audits`,
`activity`.

- `assets`: GET `/api/v1/assets` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50.
- `employees`: GET `/api/v1/employees` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50.
- `licenses`: GET `/api/v1/licenses` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50.
- `maintenances`: GET `/api/v1/maintenances` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 50.
- `accessories`: GET `/api/v1/accessories` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 50.
- `components`: GET `/api/v1/components` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50.
- `consumables`: GET `/api/v1/consumables` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 50.
- `categories`: GET `/api/v1/categories` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50.
- `departments`: GET `/api/v1/departments` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 50.
- `depreciations`: GET `/api/v1/depreciations` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 50.
- `locations`: GET `/api/v1/locations` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50.
- `manufacturers`: GET `/api/v1/manufacturers` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 50.
- `products`: GET `/api/v1/products` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50.
- `statuses`: GET `/api/v1/statuses` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50.
- `suppliers`: GET `/api/v1/suppliers` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50.
- `audits`: GET `/api/v1/audits` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50.
- `activity`: GET `/api/v1/activity` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50.
- `tenants`: GET `/api/v1/tenants` - records path `tenants`.

## Write actions & risks

This connector is read-only. Read behavior: external BlueTally API read of IT asset management data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 18 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=16, out_of_scope=57.
