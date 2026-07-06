# Overview

Reads Fastly services, the current user, the current customer (account), and datacenters through the
Fastly REST API. Read-only.

Readable streams: `services`, `current_user`, `current_customer`, `datacenters`, `service_details`,
`users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.fastly.com/reference/api/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.fastly.com`; format `uri`; Fastly API base URL
  override, e.g. for a test server. Defaults to https://api.fastly.com.
- `fastly_api_token` (optional, secret, string); Fastly API token, sent as the Fastly-Key request
  header on every call.

Secret fields are redacted in logs and write previews: `fastly_api_token`.

Default configuration values: `base_url=https://api.fastly.com`.

Authentication behavior:

- API key authentication in `Fastly-Key` using `secrets.fastly_api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/current_user`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `current_user`, `current_customer`, `service_details`, `users`;
page_number: `services`, `datacenters`.

- `services`: GET `/service` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100.
- `current_user`: GET `/current_user` - single-object response; records path `.`.
- `current_customer`: GET `/current_customer` - single-object response; records path `.`.
- `datacenters`: GET `/datacenters` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 2.
- `service_details`: GET `/service/{{ fanout.id }}/details` - single-object response; records path
  `.`; fan-out; ids from request `/service`; id-list records path `.`; id field `id`; id inserted
  into the request path; stamps `service_id`.
- `users`: GET `/customer/{{ fanout.id }}/users` - records path `.`; fan-out; ids from request
  `/current_customer`; id-list records path `.`; id field `id`; id inserted into the request path;
  stamps `customer_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Fastly API read of service/account
configuration metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=17, duplicate_of=4, non_data_endpoint=4, out_of_scope=14,
  requires_elevated_scope=9.
