# Overview

Reads Rootly incidents, services, and users through the Rootly API. Read-only.

Readable streams: `incidents`, `services`, `users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://rootly.com/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string).
- `base_url` (optional, string).
- `mode` (optional, string).
- `start_date` (required, string).

Secret fields are redacted in logs and write previews: `api_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

- `incidents`: GET connector-managed request path - records path `data`.
- `services`: GET connector-managed request path - records path `data`.
- `users`: GET connector-managed request path - records path `data`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 3 stream-backed endpoint group(s).
