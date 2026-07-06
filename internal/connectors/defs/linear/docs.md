# Overview

Reads Linear issues, teams, projects, and users through the Linear GraphQL API. Read-only.

Readable streams: `issues`, `teams`, `projects`, `users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.linear.app/docs.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Linear OAuth access token, sent as a Bearer
  Authorization header. Provide either api_key or access_token; access_token takes priority when
  both are set.
- `api_key` (optional, secret, string); Linear personal API key, sent as a bare Authorization header
  (no Bearer prefix). Provide either api_key or access_token.
- `auth_type` (optional, string); default ; allowed values `oauth`, `oauth2.0`; Optional. Not needed
  when access_token is set (access_token always uses Bearer regardless of this value).
- `base_url` (optional, string); default `https://api.linear.app/graphql`; format `uri`; Full Linear
  GraphQL endpoint URL override for tests or proxies. Defaults to https://api.linear.app/graphql.
- `max_pages` (optional, string); Optional hard cap on the number of pages read per stream. Empty,
  "all", or "unlimited" means unbounded (the default). Hooks-consumed; see docs.md Known limits.
- `page_size` (optional, string); default `50`; GraphQL connection page size (1-250, Linear's own
  cap). Hooks-consumed; see docs.md Known limits.

Secret fields are redacted in logs and write previews: `access_token`, `api_key`.

Default configuration values: `auth_type=`, `base_url=https://api.linear.app/graphql`,
`page_size=50`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- Bearer token authentication using `secrets.api_key` when `{{ config.auth_type in ['oauth',
  'oauth2.0'] }}`.
- API key authentication in `Authorization` using `secrets.api_key` when `{{ secrets.api_key }}`.
- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `issues`: POST connector-managed request path - records path `data.issues.nodes`; incremental
  cursor `updated_at`; formatted as `rfc3339`.
- `teams`: POST connector-managed request path - records path `data.teams.nodes`; incremental cursor
  `updated_at`; formatted as `rfc3339`.
- `projects`: POST connector-managed request path - records path `data.projects.nodes`; incremental
  cursor `updated_at`; formatted as `rfc3339`.
- `users`: POST connector-managed request path - records path `data.users.nodes`; incremental cursor
  `updated_at`; formatted as `rfc3339`.

## Write actions & risks

This connector is read-only. Read behavior: external Linear GraphQL API read of
issues/teams/projects/users.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=2.
