# Overview

Reads Stigg products, plans, customers, and subscriptions through the Stigg GraphQL-over-HTTP API.
Read-only.

Readable streams: `products`, `plans`, `customers`, `subscriptions`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.stigg.io/reference/api-overview.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Stigg server API key, sent as a Bearer token on every
  GraphQL POST request. Never logged.
- `base_url` (optional, string); default `https://api.stigg.io`; format `uri`; Stigg
  GraphQL-over-HTTP base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.stigg.io`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults. A `base_url` override
must parse as a valid http(s) URL with a non-empty host; malformed values fail closed.

Connection checks call POST `/graphql`.

## Streams notes

Reads are GraphQL POSTs with the query sent in the request body. Every read fetches the full
result set in a single request; streams are not incremental and there is no pagination.

If the GraphQL response contains a non-empty `errors` array, the stream yields zero records
rather than failing hard.

- `products`: POST `/graphql` - records path `data.products`.
- `plans`: POST `/graphql` - records path `data.plans`.
- `customers`: POST `/graphql` - records path `data.customers`.
- `subscriptions`: POST `/graphql` - records path `data.subscriptions`.

## Write actions & risks

This connector is read-only. Read behavior: external Stigg GraphQL API read of
product/plan/customer/subscription entitlement metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
