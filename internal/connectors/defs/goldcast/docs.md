# Overview

Reads Goldcast organizations, events, agenda items, discussion groups, and tracks through the
Goldcast customapi REST API.

Readable streams: `organizations`, `events`, `agenda_items`, `discussion_groups`, `tracks`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.goldcast.io/api-docs.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); Goldcast API access key, sent as Authorization: Token
  <access_key>. Never logged.
- `base_url` (optional, string); default `https://customapi.goldcast.io`; format `uri`; Goldcast
  customapi base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `access_key`.

Default configuration values: `base_url=https://customapi.goldcast.io`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token` using `secrets.access_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/event/`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `events`, `agenda_items`, `discussion_groups`, `tracks`; none:
`organizations`.

- `organizations`: GET `/core/organization/` - records path `.`.
- `events`: GET `/event/` - records path `results`; follows a next-page URL from the response body;
  URL path `next`; next URLs stay on the configured API host.
- `agenda_items`: GET `/event/agenda-item/` - records path `results`; follows a next-page URL from
  the response body; URL path `next`; next URLs stay on the configured API host.
- `discussion_groups`: GET `/event/discussion-groups/` - records path `results`; follows a next-page
  URL from the response body; URL path `next`; next URLs stay on the configured API host.
- `tracks`: GET `/event/tracks/` - records path `results`; follows a next-page URL from the response
  body; URL path `next`; next URLs stay on the configured API host.

## Write actions & risks

This connector is read-only. Read behavior: external Goldcast API read of organization, event, and
event-scoped data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
