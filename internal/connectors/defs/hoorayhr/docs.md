# Overview

Reads HoorayHR users, time-off, leave-types, and sick-leave records through the HoorayHR REST API
using session-token authentication.

Readable streams: `users`, `time_off`, `leave_types`, `sick_leaves`.

This connector is read-only; no write actions are declared.

Service API documentation: https://hoorayhr.io.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.hooray.nl`; format `uri`; HoorayHR API base
  URL override for tests or proxies.
- `hoorayhrpassword` (required, secret, string); HoorayHR account password. Used only in the
  /authentication login exchange body; never logged.
- `hoorayhrusername` (required, string); HoorayHR account email address, used as the username in the
  session-token login exchange.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `hoorayhrpassword`.

Default configuration values: `base_url=https://api.hooray.nl`.

Authentication behavior:

- The connector logs in by POSTing `{email, password, strategy: "local"}` to `/authentication` and
  uses the returned `accessToken` as the session token.
- The session token is sent raw in the `Authorization` header, without a `Bearer` prefix.
- The token is fetched once and cached for the lifetime of the run; there is no refresh or expiry
  handling.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users`.

## Streams notes

Default pagination: single request; no pagination.

All streams support full-refresh reads only; no incremental cursor fields are available.

- `users`: GET `/users` - records at response root.
- `time_off`: GET `/time-off` - records at response root.
- `leave_types`: GET `/leave-types` - records at response root.
- `sick_leaves`: GET `/sick-leave` - records at response root.

## Write actions & risks

This connector is read-only. Read behavior: external HoorayHR API read of employee, time-off,
leave-type, and sick-leave data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1.
