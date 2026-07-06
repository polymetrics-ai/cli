# Overview

Reads Shortcut stories, epics, projects, and iterations through the Shortcut REST API.

Readable streams: `stories`, `epics`, `projects`, `iterations`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.shortcut.com/api/rest/v3.

## Auth setup

Connection fields:

- `api_token` (optional, secret, string); Shortcut API token, sent as the Shortcut-Token header.
- `base_url` (optional, string); default `https://api.app.shortcut.com`; format `uri`; Shortcut API
  base URL.
- `page_size` (optional, integer); default `100`.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.app.shortcut.com`, `page_size=100`.

Authentication behavior:

- API key authentication in `Shortcut-Token` using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v3/stories`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `next`; next token from `next`.

- `stories`: GET `/api/v3/stories` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `next`; next token from
  `next`; computed output fields `state`.
- `epics`: GET `/api/v3/epics` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `next`; next token from
  `next`.
- `projects`: GET `/api/v3/projects` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `next`; next token from
  `next`.
- `iterations`: GET `/api/v3/iterations` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `next`; next token from
  `next`.

## Write actions & risks

This connector is read-only. Read behavior: external Shortcut API read of story, epic, project, and
iteration data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=4.
