# Overview

Runs GitLab REST API v4 through 582 source-bound direct reads, four ETL streams,
and 381 source-bound mutations through direct-write and approval-gated reverse-ETL
commands.

Readable streams: `projects`, `groups`, `users`, `issues`.

The source-lock ledger retains 1,752 cited GitLab REST identities. This slice makes 582 direct reads,
four ETL-only stream reads, and 381 mutations executable through one-request direct-write and
approval-gated reverse-ETL command pairs. The remaining 785 source identities stay discoverable with
cited named-foundation outcomes; no source identity is silently omitted.

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

GitLab retains 382 source-bound write actions. 381 have executable direct-write and reverse-ETL command
pairs; one action remains a typed partial outcome because its retained JSON-Schema regex is not accepted
by the current closed engine compiler. `capabilities.write=true` does not promote any other provider
mutation.

Read behavior: external GitLab API reads use the declared direct-read or ETL contract. Every materialized
write requires plan, preview, explicit approval, and execute. Output is not redacted by this connector.

## Known limits

- Batch defaults: read_page_size=50.
- Source-lock denominator: 1,752 cited REST identities from GitLab OpenAPI 3.0.0 `info.version`
  `19.3.0-pre`, retrieved 2026-08-05.
- Source mapping: 1,746 mapped identities and 6 blocked identities with no canonical runtime mapping;
  the latter retain exact citations and named gaps rather than disappearing.
- Runtime-enabled source rows: 967 — 582 direct reads, 4 ETL-only stream reads, and 381 mutations.
- Blocked source rows: 785, each with one or more named foundations. The 1,845 `source_contract`
  gap entries overlap by row and are not a source denominator.
- `execution bundle`, `cli_surface.json`, and the operation-evidence artifact expose the command or
  typed cited outcome for each source identity.
