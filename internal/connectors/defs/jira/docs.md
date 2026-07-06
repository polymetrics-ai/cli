# Overview

Reads Jira issues, projects, and users through the Jira Cloud REST API v3 using HTTP Basic auth
(email + API token). Read-only.

Readable streams: `issues`, `projects`, `users`.

This connector is read-only; no write actions are declared.

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

This connector is read-only. Read behavior: external Jira Cloud API read of issue, project, and user
data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=1, duplicate_of=3, non_data_endpoint=1, out_of_scope=3,
  requires_elevated_scope=3.
