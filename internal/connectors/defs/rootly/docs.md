# Overview

Reads Rootly incidents, services, and users through fixed JSON:API routes. Read-only.

Readable streams: `incidents`, `services`, `users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.rootly.com/api-reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Rootly bearer token.
- `start_date` (required, string); retained initial ETL lower-bound configuration.

Secret fields are redacted in logs and write previews: `api_key`.

The runtime uses only the fixed `https://api.rootly.com` origin and declared bearer authentication. It does not accept caller-provided origins or fixture modes.

Connection checks read one incidents page.

## Streams notes

Default pagination follows the provider-declared `links.next` URL with a 100-page maximum.

- `incidents`: GET `/v1/incidents`; JSON:API attributes are projected to `title` and `status`.
- `services`: GET `/v1/services`; JSON:API attributes are projected to `title` and `status`.
- `users`: GET `/v1/users`; JSON:API attributes are projected to `title` and `status`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 3 stream-backed endpoint groups.
