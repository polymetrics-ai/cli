# Overview

Reads FastBill customers, invoices, products, recurring invoices, and revenues through the FastBill
JSON API.

Readable streams: `customers`, `invoices`, `products`, `recurring_invoices`, `revenues`.

This connector is read-only; no write actions are declared.

Service API documentation: https://apidocs.fastbill.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Fastbill API key.
- `base_url` (optional, string).
- `mode` (optional, string).
- `username` (required, string); Username for Fastbill account.

Secret fields are redacted in logs and write previews: `api_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

- `customers`: GET connector-managed request path - records path `data`.
- `invoices`: GET connector-managed request path - records path `data`.
- `products`: GET connector-managed request path - records path `data`.
- `recurring_invoices`: GET connector-managed request path - records path `data`.
- `revenues`: GET connector-managed request path - records path `data`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
