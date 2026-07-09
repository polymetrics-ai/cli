# Overview

Reads GitLab projects, groups, users, and issues through the GitLab REST API v4.

Readable streams: `projects`, `groups`, `users`, `issues`.

The connector declares a command surface in `cli_surface.json`. Implemented commands map to the four
stream-backed reads plus four bounded direct reads (`project view`, `group view`, `user events`, and
`issue view`). Direct reads use a 1 MiB default response cap and the `json_redacted` output policy.
Binary, local workflow, raw API, GraphQL mutation, and reverse-ETL commands remain blocked or deferred
until their dedicated lanes add bounded executors and approval policy.

This connector is read-only for durable ETL/reverse ETL; no write actions are declared.

Service API documentation: https://docs.gitlab.com/ee/api/rest/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); GitLab personal access token or OAuth access token.
  Used only for Bearer auth; never logged.
- `base_url` (optional, string); default `https://gitlab.com/api/v4`; format `uri`; GitLab API base
  URL override, e.g. https://gitlab.example.com/api/v4 for self-managed instances, or for
  tests/proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-100).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound applied as each stream's
  matching since-filter (last_activity_after for projects, created_after for users, updated_after
  for issues; groups has no since-filter upstream).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://gitlab.com/api/v4`, `page_size=50`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/user`.

## Streams notes

Default pagination: follows RFC 5988 Link headers with rel=next.

- `projects`: GET `/projects` - records at response root; query `last_activity_after` from template
  `{{ config.start_date }}`, omitted when absent; `per_page`=`50`; follows RFC 5988 Link headers
  with rel=next.
- `groups`: GET `/groups` - records at response root; query `per_page`=`50`; follows RFC 5988 Link
  headers with rel=next.
- `users`: GET `/users` - records at response root; query `created_after` from template `{{
  config.start_date }}`, omitted when absent; `per_page`=`50`; follows RFC 5988 Link headers with
  rel=next.
- `issues`: GET `/issues` - records at response root; query `per_page`=`50`; `updated_after` from
  template `{{ config.start_date }}`, omitted when absent; follows RFC 5988 Link headers with
  rel=next; computed output fields `author_id`.

## Write actions & risks

This connector is read-only. Read behavior: external GitLab API read of projects, groups, users,
issues, and bounded direct-read detail endpoints.

The operation ledger inventories 1,144 official GitLab OpenAPI REST operations plus one `/users`
compatibility row for the existing stream. Non-enabled operations are blocked by default as direct
read candidates, binary reads, sensitive/admin reverse ETL candidates, destructive actions, or
deprecated endpoints. Sensitive/admin/destructive operations require reverse ETL plan → preview →
approval → execute plus typed confirmation before any future execution can be considered.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 4 stream-backed endpoint group(s), 4 bounded direct-read commands, and a
  complete blocked-by-default operation ledger.
- `cli_surface.json` intentionally does not expose generic raw API commands, unsafe binary downloads,
  generic shell/SQL writes, arbitrary GraphQL mutations, or GitLab write execution.
- GitLab GraphQL is not required for the current REST-backed commands. Future GraphQL support must use
  fixed documents, typed variables, bounded pagination, and no generic mutation escape hatch.
