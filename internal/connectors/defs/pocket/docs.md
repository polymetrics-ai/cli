# Overview

Reads saved Pocket items through the v3 retrieve API.

Readable streams: `items`.

This connector is read-only; no write actions are declared.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); The user's Pocket access token.
- `base_url` (optional, string).
- `consumer_key` (required, secret, string); Your application's Consumer Key.
- `content_type` (optional, string); Select the content type of the items to retrieve.
- `detail_type` (optional, string); Select the granularity of the information about each item.
- `domain` (optional, string); Only return items from a particular `domain`.
- `favorite` (optional, string); Retrieve only favorited items.
- `mode` (optional, string).
- `search` (optional, string); Only return items whose title or url contain the `search` string.
- `since` (optional, string); Only return items modified since the given timestamp.
- `sort` (optional, string); Sort retrieved items by the given criteria.
- `state` (optional, string); Select the state of the items to retrieve.
- `tag` (optional, string); Return only items tagged with this tag name. Use _untagged_ for
  retrieving only untagged items.

Secret fields are redacted in logs and write previews: `access_token`, `consumer_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `items`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 1 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `items`.
