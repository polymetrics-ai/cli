# Overview

Reads SafetyCulture audits, templates, and users through the SafetyCulture API. Read-only.

Readable streams: `audits`, `templates`, `users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.safetyculture.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string).
- `base_url` (optional, string).
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

- `audits`: GET connector-managed request path - records path `data`.
- `templates`: GET connector-managed request path - records path `data`.
- `users`: GET connector-managed request path - records path `data`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 3 stream-backed endpoint group(s).
