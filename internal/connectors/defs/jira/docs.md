# Overview

Reads and writes the documented Jira Cloud platform REST API v3 surface using HTTP Basic auth
(email + API token).

`api_surface.json` owns Jira's source-backed endpoint ledger, while `cli_surface.json` owns
per-command availability. The bundle includes stream-backed reads, bounded direct and binary reads,
typed reverse-ETL actions, and explicitly blocked or partial operations; use runtime help or
`cli_surface.json` to determine which command paths are currently executable. Certification is
fixture-only; no live provider calls were made.

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

The connector declares typed write actions across issues, projects, users, fields, workflows,
dashboards, and instance administration. `writes.json` owns those declarations; `cli_surface.json`
records whether each action has an executable command path.

Writes are only available through reverse ETL plan -> preview -> explicit approval -> execute. Every
DELETE action is gated as destructive and additionally requires a typed confirmation. The bundle
does not expose arbitrary request bodies, raw query strings, generic method/path/body, file bytes,
shell commands, or passthrough HTTP tools.

Some documented mutations remain blocked under the `sensitive_reverse_etl` model, and some declared
write commands are `partial` and refuse execution; `api_surface.json` and `cli_surface.json` hold
the current per-operation state.

Read behavior: external Jira Cloud API read of issue, project, and user data.

## Known limits

- Batch defaults: read_page_size=50.
- `api_surface.json` is the authoritative per-endpoint coverage and blocked-operation ledger.
- Fixture-only evidence: no live Jira credentials, provider calls, provider writes, or
  certification run were used.
