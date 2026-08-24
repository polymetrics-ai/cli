# Overview

Reads GitLab projects, groups, users, and issues through the GitLab REST API v4.

Readable streams: `projects`, `groups`, `users`, `issues`.

The provider-owned OpenAPI v3 inventory pinned on 2026-08-19 covers 1,755 callable GitLab
operations. Only the four existing stream reads are executable; every other public operation remains
explicitly disabled until it has a bounded connector-local contract.

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

Secret fields are protected in credential storage and never logged: `access_token`.

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

The GitLab provider inventory includes 1,006 writes. `capabilities.write=false` until at least one
GitLab write action is executable; the provider inventory is not an executable capability. This G1
bundle declares no GitLab write action.

Read behavior: external GitLab API read of projects, groups, users, and issues. Any future write must
use plan, preview, explicit approval, and execute. Output is not redacted by this connector.

## Known limits

- Batch defaults: read_page_size=50.
- Provider inventory: 1,755 callable OpenAPI operations (749 reads; 1,006 writes), sourced from
  GitLab OpenAPI 3.0.0 `info.version` `v4`, retrieved 2026-08-19 and pinned in
  `sources/gitlab-operation-source-lock.json`.
- Executable in G1: 4 stream-backed reads (`GET /projects`, `GET /groups`, `GET /users`, and
  `GET /issues`).
- Remaining provider operations: 1,751 are explicitly disabled with an exact reason in
  `api_surface.json`; the batch rejection ledger retains the fixed-vocabulary disposition and recovery
  condition for each one. The four command citations are in `cli_surface.json`.
