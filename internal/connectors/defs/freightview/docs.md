# Overview

Reads Freightview shipments, quotes, and tracking events through the Freightview v2.0 REST API using
the client-credentials session-token flow.

Readable streams: `shipments`, `quotes`, `tracking`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.freightview.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string).
- `client_id` (required, secret, string).
- `client_secret` (required, secret, string).
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

- `shipments`: GET connector-managed request path - records path `data`.
- `quotes`: GET connector-managed request path - records path `data`.
- `tracking`: GET connector-managed request path - records path `data`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 3 stream-backed endpoint group(s).
