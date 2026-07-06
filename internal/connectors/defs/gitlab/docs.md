# Overview

Reads GitLab projects, groups, users, and issues through the GitLab REST API v4.

Readable streams: `projects`, `groups`, `users`, `issues`.

This connector is read-only; no write actions are declared.

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

This connector is read-only. Read behavior: external GitLab API read of projects, groups, users, and
issues.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=7.
