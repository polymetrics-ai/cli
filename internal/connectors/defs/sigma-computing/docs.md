# Overview

Reads Sigma workbooks, datasets, teams, and members through the Sigma REST API.

Readable streams: `workbooks`, `datasets`, `teams`, `members`.

This connector is read-only; no write actions are declared.

Service API documentation: https://help.sigmacomputing.com/reference/api-overview.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Sigma Computing OAuth access token, sent as a Bearer
  token.
- `base_url` (optional, string); default `https://api.sigmacomputing.com`; format `uri`; Sigma
  Computing API base URL.
- `page_size` (optional, integer); default `100`.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.sigmacomputing.com`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/workbooks`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `nextPage`.

- `workbooks`: GET `/v2/workbooks` - records path `entries`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `page`; next token from
  `nextPage`; computed output fields `name`, `updated_at`.
- `datasets`: GET `/v2/datasets` - records path `entries`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `page`; next token from
  `nextPage`; computed output fields `name`, `updated_at`.
- `teams`: GET `/v2/teams` - records path `entries`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `page`; next token from
  `nextPage`; computed output fields `name`, `updated_at`.
- `members`: GET `/v2/members` - records path `entries`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `page`; next token from
  `nextPage`; computed output fields `name`, `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Sigma Computing API read of workbook, dataset,
team, and member data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
