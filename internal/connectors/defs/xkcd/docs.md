# Overview

Reads public XKCD comic metadata from the JSON API. Read-only.

Readable streams: `latest`, `comic`.

This connector is read-only; no write actions are declared.

Service API documentation: https://xkcd.com/json.html.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; XKCD API base URL. There is no default: leaving
  `base_url` unset is a configuration error, so it must always be supplied explicitly.
- `comic_number` (optional, string); Specific comic number to read for the 'comic' stream (required
  for that stream only). Interpolated into the request path as a single, urlencoded path segment; a
  value containing a path-traversal or separator character is rejected, including percent-encoded
  `..` segments (e.g. `%2e%2e`).

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value.

Connection checks call GET `info.0.json`.

## Streams notes

Default pagination: single request; no pagination.

- `latest`: GET `info.0.json` - single-object response; records path `.`; emits passthrough records.
- `comic`: GET `{{ config.comic_number }}/info.0.json` - single-object response; records path `.`;
  emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: public XKCD comic metadata read, no credentials
involved.

## Known limits

- API coverage includes 2 stream-backed endpoint group(s).
