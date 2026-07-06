# Overview

Reads Metabase cards, dashboards, collections, databases, and users through the Metabase REST API
using session-token authentication.

Readable streams: `cards`, `dashboards`, `collections`, `databases`, `users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.metabase.com/docs/latest/api-documentation.

## Auth setup

Connection fields:

- `base_url` (optional, string).
- `instance_api_url` (required, string); URL to your metabase instance API.
- `mode` (optional, string).
- `password` (optional, secret, string).
- `session_token` (optional, secret, string); To generate your session token, you need to run the
  following command: ``` curl -X POST \ -H "Content-Type: application/json" \ -d '{"username":
  "person@metabase.com", "password": "fakepassword"}' \ http://localhost:3000/api/session ``` Then
  copy the value of the `id` field returned by a successful call to that API. Note that by default,
  sessions are good for 14 days and needs to be regenerated.
- `username` (required, string).

Secret fields are redacted in logs and write previews: `password`, `session_token`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

- `cards`: GET connector-managed request path - records path `data`.
- `dashboards`: GET connector-managed request path - records path `data`.
- `collections`: GET connector-managed request path - records path `data`.
- `databases`: GET connector-managed request path - records path `data`.
- `users`: GET connector-managed request path - records path `data`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
