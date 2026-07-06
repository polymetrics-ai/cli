# Overview

Reads Bugsnag organizations, projects, collaborators, errors, events, and releases through the
Bugsnag Data Access API.

Readable streams: `organizations`, `projects`, `collaborators`, `errors`, `events`, `releases`.

This connector is read-only; no write actions are declared.

Service API documentation: https://bugsnagapiv2.docs.apiary.io/.

## Auth setup

Connection fields:

- `auth_token` (required, secret, string); Bugsnag personal auth token, sent as 'Authorization:
  token <auth_token>'. Never logged.
- `base_url` (optional, string); default `https://api.bugsnag.com`; format `uri`; Bugsnag Data
  Access API base URL override for tests or proxies.
- `organization_id` (optional, string); Comma-separated list of Bugsnag organization ids to fan out
  over for the projects/collaborators streams (one GET per id).
- `page_size` (optional, string); default `100`; Records per page (1-100, per_page).
- `project_id` (optional, string); Comma-separated list of Bugsnag project ids to fan out over for
  the errors/events/releases streams (one GET per id).

Secret fields are redacted in logs and write previews: `auth_token`.

Default configuration values: `base_url=https://api.bugsnag.com`, `page_size=100`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `token` using `secrets.auth_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/user/organizations` with query `per_page`=`1`.

## Streams notes

Default pagination: follows RFC 5988 Link headers with rel=next.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `organizations`: GET `/user/organizations` - records at response root; query `per_page`=`{{
  config.page_size }}`; follows RFC 5988 Link headers with rel=next.
- `projects`: GET `/organizations/{{ fanout.id }}/projects` - records at response root; query
  `per_page`=`{{ config.page_size }}`; follows RFC 5988 Link headers with rel=next; fan-out; ids
  from config field `organization_id`; id inserted into the request path; stamps `organization_id`.
- `collaborators`: GET `/organizations/{{ fanout.id }}/collaborators` - records at response root;
  query `per_page`=`{{ config.page_size }}`; follows RFC 5988 Link headers with rel=next; fan-out;
  ids from config field `organization_id`; id inserted into the request path.
- `errors`: GET `/projects/{{ fanout.id }}/errors` - records at response root; query `per_page`=`{{
  config.page_size }}`; follows RFC 5988 Link headers with rel=next; incremental cursor `last_seen`;
  formatted as `rfc3339`; computed output fields `events_count`; fan-out; ids from config field
  `project_id`; id inserted into the request path; stamps `project_id`.
- `events`: GET `/projects/{{ fanout.id }}/events` - records at response root; query `per_page`=`{{
  config.page_size }}`; follows RFC 5988 Link headers with rel=next; incremental cursor
  `received_at`; formatted as `rfc3339`; fan-out; ids from config field `project_id`; id inserted
  into the request path; stamps `project_id`.
- `releases`: GET `/projects/{{ fanout.id }}/releases` - records at response root; query
  `per_page`=`{{ config.page_size }}`; follows RFC 5988 Link headers with rel=next; incremental
  cursor `release_time`; formatted as `rfc3339`; fan-out; ids from config field `project_id`; id
  inserted into the request path; stamps `project_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Bugsnag API read of organization, project,
collaborator, and error/event/release data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=4.
