# Overview

Reads Intercom contacts, companies, conversations, admins, and tags through the Intercom REST API.
The connector also carries metadata for the full Intercom REST API 2.14 surface so future CLI parity
lanes can add safe streams, direct reads, binary/export policies, and typed reverse-ETL actions
without exposing raw generic HTTP writes.

Readable streams currently implemented: `contacts`, `companies`, `conversations`, `admins`, `tags`.

Official source used for the surface ledger:
https://raw.githubusercontent.com/intercom/Intercom-OpenAPI/main/descriptions/2.14/api.intercom.io.yaml.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Intercom access token. Used only for Bearer auth; never
  logged.
- `api_version` (optional, string); optional `Intercom-Version` header value. When unset, the header
  is omitted and Intercom uses the workspace default API version.
- `base_url` (optional, string); default `https://api.intercom.io`; format `uri`; Intercom API base
  URL override for tests or proxies.
- `page_size` (optional, string); default `50`; records per page (1-150).

Secret fields are redacted in logs and write previews: `access_token`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/admins`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `starting_after`; next token from
`pages.next.starting_after`.

- `contacts`: GET `/contacts` - records path `data`; query `per_page`=`50`; cursor pagination.
- `companies`: GET `/companies` - records path `data`; query `per_page`=`50`; cursor pagination.
- `conversations`: GET `/conversations` - records path `data`; query `per_page`=`50`; cursor
  pagination.
- `admins`: GET `/admins` - records path `data`; cursor pagination.
- `tags`: GET `/tags` - records path `data`; cursor pagination.

Additional Intercom reads such as articles, tickets, calls, activity logs, exports, and Help Center
resources are tracked in `api_surface.json` as blocked-by-default metadata for later stream,
direct-read, or binary policy slices.

## Write actions & risks

This connector currently declares no executable write actions and has `capabilities.write: false`.

The official Intercom API includes create, update, delete, admin/configuration, export lifecycle,
message, ticket, contact, conversation, and tag mutations. In this bundle those operations are
recorded in `api_surface.json` as blocked-by-default operation metadata, not executable writes.
Future reverse-ETL support must add explicit `writes.json` actions with schemas, path fields, risk
text, redaction, approval copy, and `confirm: destructive` where applicable. Reverse ETL must remain
plan → preview → approval → execute.

## Known limits

- Current executable read coverage remains the five legacy streams: `contacts`, `companies`,
  `conversations`, `admins`, and `tags`.
- `api_surface.json` enumerates all 149 official Intercom REST API 2.14 operations: GET 67, POST 47,
  PUT 16, DELETE 19.
- Non-implemented operations are blocked by default as metadata for follow-up issues #168-#171;
  they are not raw API escape hatches.
- Binary/export and call recording/transcript endpoints require bounded output policies before CLI
  exposure.
- No live or credentialed Intercom checks are required for this metadata slice.
