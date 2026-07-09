# Overview

Reads monday.com boards, items, users, teams, and tags through the monday.com GraphQL API.
Read-only.

Readable streams: `boards`, `items`, `users`, `teams`, `tags`.

CLI command surface metadata is available for the implemented stream-backed commands:
`pm monday board list`, `pm monday item list`, `pm monday user list`, `pm monday team list`, and
`pm monday tag list`. Fixed direct-read metadata is also available for `pm monday me view` and
`pm monday account view`; these execute bundled GraphQL query documents only. The operation ledger
inventories 367 canonical GraphQL reference operations (87 queries, 280 mutations) as metadata;
other planned direct-read and mutation commands remain blocked/planned metadata only and do not
execute raw GraphQL or writes.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.monday.com/api-reference/docs.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); monday.com OAuth access token, sent verbatim (no Bearer
  prefix) as the Authorization header. Provide either api_token or access_token.
- `api_token` (optional, secret, string); monday.com personal API token, sent verbatim (no Bearer
  prefix) as the Authorization header. Provide either api_token or access_token.
- `api_version` (optional, string); Optional monday.com API-Version header value (e.g. 2024-01).
  Omitted entirely when unset.
- `base_url` (optional, string); default `https://api.monday.com/v2`; format `uri`; monday.com
  GraphQL API base URL override for tests or proxies. Defaults to https://api.monday.com/v2.
- `max_pages` (optional, string); Permissive parse, never errors: an empty value, "all",
  "unlimited", or any non-positive-integer string means unbounded (0, the same as leaving this
  unset); a positive integer string caps the page count at that value. Was previously consumed by
  the hook but undeclared here (a dead-config-surface gap the S3 engine mini-wave carried-minor
  cleanup closed) - see docs.md Known limits.
- `page_size` (optional, string); default `50`; Records per GraphQL page (1-500).

Secret fields are redacted in logs and write previews: `access_token`, `api_token`.

Default configuration values: `base_url=https://api.monday.com/v2`, `page_size=50`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_token` when `{{ secrets.api_token
  }}`.
- API key authentication in `Authorization` using `secrets.access_token` when `{{
  secrets.access_token }}`.
- No authentication: requests fall back to unauthenticated only when both `api_token` and
  `access_token` are absent; unauthenticated requests fail with a 401.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

All stream reads are GraphQL POST requests with pagination state carried in the request body:
`boards`/`users`/`teams`/`tags` use page-number pagination (`limit`/`page`), while `items` uses
cursor pagination (`limit`/`cursor`).

Reads are always full syncs: cursor fields are advertised in the catalog but are not used to
filter or advance reads.

- `boards`: POST connector-managed request path - records path `data.boards`; catalog cursor
  `updated_at`; formatted as `rfc3339`.
- `items`: POST connector-managed request path - records path `data.next_items_page.items`;
  catalog cursor `updated_at`; formatted as `rfc3339`.
- `users`: POST connector-managed request path - records path `data.users`.
- `teams`: POST connector-managed request path - records path `data.teams`.
- `tags`: POST connector-managed request path - records path `data.tags`.

## Write actions & risks

This connector is read-only. Read behavior: external monday.com GraphQL API read of
boards/items/users/teams/tags.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=2.
