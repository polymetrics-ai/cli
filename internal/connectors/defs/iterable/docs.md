# Overview

Reads Iterable lists, campaigns, and templates through the Iterable REST API. Read-only.

Readable streams: `lists`, `campaigns`, `templates`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.iterable.com/api/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Iterable API key, sent as the Api-Key request header (no
  prefix).
- `base_url` (optional, string); default `https://api.iterable.com/api`; format `uri`; Iterable API
  base URL. Defaults to the standard US endpoint.
- `page_size` (optional, integer); default `100`; Records requested per page (pageSize query param),
  between 1 and 1000.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.iterable.com/api`, `page_size=100`.

Authentication behavior:

- API key authentication in `Api-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/lists` with query `pageSize`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `pageToken`; next token from
`nextPageToken`.

- `lists`: GET `/lists` - records path `lists`; query `pageSize`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `pageToken`; next token from `nextPageToken`.
- `campaigns`: GET `/campaigns` - records path `campaigns`; query `pageSize`=`{{ config.page_size
  }}`; cursor pagination; cursor parameter `pageToken`; next token from `nextPageToken`.
- `templates`: GET `/templates` - records path `templates`; query `pageSize`=`{{ config.page_size
  }}`; cursor pagination; cursor parameter `pageToken`; next token from `nextPageToken`.

## Write actions & risks

This connector is read-only. Read behavior: external Iterable API read of lists, campaigns, and
templates.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=1.
