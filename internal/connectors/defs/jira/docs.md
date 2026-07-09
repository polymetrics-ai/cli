# Overview

Reads Jira issues, projects, users, and bounded Jira Cloud REST resources through the Jira Cloud
REST API v3 using HTTP Basic auth (email + API token). The full-surface ledger inventories all 620
official OpenAPI operations from `swagger-v3.v3.json`: 276 GET, 135 POST, 119 PUT, and 90 DELETE.

Readable streams: `issues`, `projects`, `users`.

Safe JSON GET endpoints that are not stream-backed are exposed as typed fixed-endpoint direct reads
with `json_redacted` output policy and the engine's direct-read max-byte limit. Mutating endpoints
are modeled as named reverse-ETL actions; they still require plan, preview, approval, and execute
before dispatch. Admin, destructive, sensitive, file upload, and binary download shapes carry
blocked-by-default policy metadata where a specialized executor is not enabled.

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

Jira declares typed reverse-ETL actions for OpenAPI mutating endpoints that can be represented by
bounded record schemas. These actions are not raw API escape hatches: paths are fixed to Jira REST
operations, record fields are schema-derived, and destructive/admin/sensitive actions carry risk text
and confirmation metadata.

Write execution must follow plan, preview, approval, and execute. Sensitive values must come from
environment variables, files, stdin, or the credential manager; never pass secrets in prompt text or
examples.

Read behavior: external Jira Cloud API read of issue, project, user, and administrative metadata.
Write behavior: Jira REST mutations may create, update, transition, archive, delete, or administer
Jira resources and can notify users or trigger automation.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 3 stream-backed endpoint groups plus operation-ledger coverage for all 620
  official Jira Cloud REST API v3 operations.
- JSON GET direct reads are bounded and redacted; collection reads that need resumable sync beyond
  the existing issue/project/user streams remain direct-read commands until promoted to streams.
- POST body read-query operations are classified as `rest_query` and blocked in `api_surface.json`
  until fixed body-variable direct-read execution is wired.
- Binary download and file upload operations are inventoried with max-byte policy but remain blocked
  unless a bounded local file-output/input executor is enabled.
