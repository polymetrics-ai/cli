# Overview

Reads OneSignal account-level applications through the OneSignal REST API.

Readable streams: `apps`.

This connector is read-only; no write actions are declared.

Service API documentation: https://documentation.onesignal.com/reference.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://onesignal.com/api/v1`; format `uri`; OneSignal
  REST API base URL override for tests or proxies.
- `mode` (optional, string).
- `user_auth_key` (required, secret, string); OneSignal user/organization auth key, sent as
  'Authorization: Basic <user_auth_key>'. Authenticates the account-level apps stream. Never logged.

Secret fields are redacted in logs and write previews: `user_auth_key`.

Default configuration values: `base_url=https://onesignal.com/api/v1`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Basic` using `secrets.user_auth_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/apps`.

## Streams notes

Default pagination: single request; no pagination.

- `apps`: GET `/apps` - records at response root.

## Write actions & risks

This connector is read-only. Read behavior: external OneSignal API read of account-level application
metadata.

## Known limits

- Batch defaults: read_page_size=1.
- API coverage includes 1 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
