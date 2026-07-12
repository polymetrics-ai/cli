# Overview

Reads GitLab projects, groups, users, and issues through the GitLab REST API v4.

Readable streams: `projects`, `groups`, `users`, `issues`.

The connector declares a full operation-parity command surface in `cli_surface.json`, modeled after
the GitHub parity scaffold. Implemented commands cover all non-deprecated official GitLab REST
operations plus the existing `/users` compatibility stream row. The four durable ETL streams and four
bounded direct reads are runnable today; operation-backed read/binary/HEAD commands remain
feature-gated until an operation executor exists. Generated write commands are typed reverse-ETL write
actions and still require plan → preview → approval → execute before any live mutation.

This connector now declares typed write actions for non-deprecated GitLab mutation endpoints, but no
write runs from a plain command invocation and no raw generic API command is exposed.

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

Read behavior: external GitLab API read of projects, groups, users, issues, and bounded direct-read
detail endpoints. Operation-backed direct/binary/HEAD read commands are present for parity metadata
but remain feature-gated by the command runner until a bounded operation executor is implemented.

The operation ledger inventories 1,144 official GitLab OpenAPI REST operations plus one `/users`
compatibility row for the existing stream. All non-deprecated official operations are now covered by a
stream, bounded direct read, operation-backed command, or typed reverse-ETL write action. The three
deprecated operations remain blocked-by-default operation-ledger rows. Sensitive/admin/destructive
write actions require reverse ETL plan → preview → approval → execute plus typed confirmation before
execution.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 1,142 covered endpoint rows: 4 stream-backed endpoint groups, 4 runnable
  bounded direct-read commands, 497 operation-backed read/binary/HEAD commands, and 637 typed
  reverse-ETL write actions. Three deprecated endpoints remain blocked.
- Operation-backed direct reads and binary/HEAD commands are parity metadata until the operation
  executor exists; this avoids unsafe binary downloads or raw dispatch. Leaf help such as
  `pm gitlab repo branches check --help` still renders from metadata without opening credentials or
  project state.
- `cli_surface.json` intentionally does not expose generic raw API commands, generic shell/SQL writes,
  or arbitrary GraphQL mutations.
- GitLab GraphQL is not required for the current REST-backed commands. Future GraphQL support must use
  fixed documents, typed variables, bounded pagination, and no generic mutation escape hatch.
