# Overview

Reads Bunny subscription-billing data (accounts, contacts, invoices, payments, subscriptions) from
the per-tenant Bunny GraphQL API.

Readable streams: `accounts`, `contacts`, `invoices`, `payments`, `subscriptions`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.bunny.net/reference/bunnynet-api-overview.

## Auth setup

Connection fields:

- `apikey` (required, secret, string).
- `base_url` (optional, string).
- `mode` (optional, string).
- `start_date` (optional, string).
- `subdomain` (required, string); The subdomain specific to your Bunny account or service.

Secret fields are redacted in logs and write previews: `apikey`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `accounts`: GET connector-managed request path - records path `data`; incremental cursor
  `updatedAt`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `contacts`: GET connector-managed request path - records path `data`; incremental cursor
  `updatedAt`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `invoices`: GET connector-managed request path - records path `data`; incremental cursor
  `updatedAt`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `payments`: GET connector-managed request path - records path `data`; incremental cursor
  `updatedAt`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `subscriptions`: GET connector-managed request path - records path `data`; incremental cursor
  `updatedAt`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only. Read behavior: external Bunny, Inc.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `accounts`, `contacts`, `invoices`, `payments`,
  `subscriptions`.
