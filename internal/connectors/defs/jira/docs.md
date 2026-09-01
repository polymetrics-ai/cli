# Overview

Reads and writes the documented Jira Cloud platform REST API v3 surface using HTTP Basic auth
(email + API token).

Current official operation ledger: 617 documented HTTP operations (276 GET, 134 POST, 118 PUT, 89
DELETE). Implemented rows: 584 = 3 stream-backed reads + 292 bounded direct reads + 286 typed writes
+ 3 binary downloads. Declared `partial` and not executable: 6 typed writes. Blocked/planned rows:
27. Validated rows: 0 (fixture-only; no live provider calls were made).

Readable streams: `issues`, `projects`, `users`.

Service API documentation: https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Jira API token, sent as the Basic auth password. Never
  logged.
- `base_url` (required, string); format `uri`; Jira Cloud site base URL, e.g.
  https://your-company.atlassian.net.
- `email` (required, string); Atlassian account email used as the Basic auth username.

Secret fields are redacted in logs and write previews: `api_token`.

Authentication behavior:

- HTTP Basic authentication using `config.email`, `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/rest/api/3/myself`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `startAt`; limit parameter
`maxResults`; page size 50.

- `issues`: GET `/rest/api/3/search` - records path `issues`; offset/limit pagination; offset
  parameter `startAt`; limit parameter `maxResults`; page size 50; computed output fields
  `assignee`, `created`, `issuetype`, `priority`, `project`, `reporter`, `status`, `summary`,
  `updated`.
- `projects`: GET `/rest/api/3/project/search` - records path `values`; offset/limit pagination;
  offset parameter `startAt`; limit parameter `maxResults`; page size 50.
- `users`: GET `/rest/api/3/users/search` - records path `.`; offset/limit pagination; offset
  parameter `startAt`; limit parameter `maxResults`; page size 50.

## Write actions & risks

The connector declares 292 typed write actions (106 POST creates, 102 PUT updates, 84 DELETE
removals) across issues, projects, users, fields, workflows, dashboards, and instance
administration.

Writes are only available through reverse ETL plan -> preview -> explicit approval -> execute. Every
DELETE action is gated as destructive and additionally requires a typed confirmation. The bundle
does not expose arbitrary request bodies, raw query strings, generic method/path/body, file bytes,
shell commands, or passthrough HTTP tools.

A further 25 documented mutations stay blocked under the `sensitive_reverse_etl` model, and 6
declared write commands are `partial` and refuse execution.

Read behavior: external Jira Cloud API read of issue, project, and user data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 3 stream-backed endpoint group(s); the remaining documented reads are
  exposed as bounded direct reads rather than ETL streams.
- Other documented endpoints are not exposed by this connector where they are blocked in the
  operation ledger as sensitive_reverse_etl=25, direct_read=2.
- Fixture-only evidence: no live Jira credentials, provider calls, provider writes, or
  validation run were used.
