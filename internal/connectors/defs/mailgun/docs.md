# Overview

Reads Mailgun sending domains, email events, mailing lists, and analytics tags through the Mailgun
v3 REST API.

Readable streams: `domains`, `events`, `mailing_lists`, `tags`.

This connector is read-only; no write actions are declared.

Service API documentation: https://documentation.mailgun.com/en/latest/api_reference.html.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.mailgun.net`; format `uri`; Mailgun API base
  URL. Defaults to the US region host; EU-region accounts must override this to
  https://api.eu.mailgun.net explicitly (see docs.md's Known limits).
- `domain_name` (optional, string); Mailgun sending domain name, substituted into domain-scoped
  resource paths (required for the events and tags streams).
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000).
- `private_key` (required, secret, string); Mailgun account private API key, sent as the password of
  HTTP Basic auth (username is the literal 'api'). Never logged.

Secret fields are redacted in logs and write previews: `private_key`.

Default configuration values: `base_url=https://api.mailgun.net`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `secrets.private_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v3/domains`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `events`, `mailing_lists`, `tags`; offset_limit: `domains`.

- `domains`: GET `/v3/domains` - records path `items`; query `limit`=`{{ config.page_size }}`;
  offset/limit pagination; offset parameter `skip`; limit parameter `limit`; page size 100; computed
  output fields `id`.
- `events`: GET `/v3/{{ config.domain_name }}/events` - records path `items`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path `paging.next`; next
  URLs stay on the configured API host; computed output fields `log_level`, `message_id`.
- `mailing_lists`: GET `/v3/lists/pages` - records path `items`; query `limit`=`{{ config.page_size
  }}`; follows a next-page URL from the response body; URL path `paging.next`; next URLs stay on the
  configured API host.
- `tags`: GET `/v3/{{ config.domain_name }}/tags` - records path `items`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path `paging.next`; next
  URLs stay on the configured API host; computed output fields `first_seen`, `last_seen`, `tag`.

## Write actions & risks

This connector is read-only. Read behavior: external Mailgun API read of sending-domain, event,
mailing-list, and tag data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=9.
