# Overview

Reads Jira issues, projects, and users through the Jira Cloud REST API v3 using HTTP Basic auth
(email + API token). The connector also declares a complete Jira Cloud REST API v3 operation ledger
from the official Atlassian OpenAPI document and exposes connector-owned bounded command metadata.

Readable ETL streams: `issues`, `projects`, `users`.

Operation ledger source: https://developer.atlassian.com/cloud/jira/platform/swagger-v3.v3.json

Source evidence:

- sha256: `8439da27e1b2dd7b013a0ae721b8aeaa7746bc8e2d816fa28aa1a582e8597501`
- md5: `ae49a3d84a12210d4686315cb36442be`
- operations inventoried: `616`
- executable reverse-ETL write actions: `291`
- blocked reverse-ETL shared-foundation gaps: `5`
- implemented bounded direct-read commands: `286`
- blocked binary/direct operation rows: `3`

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

The API-surface ledger includes every official Jira Cloud REST API v3 operation exactly once as an
ETL stream, bounded direct read, reverse-ETL write, or blocked operation-ledger row. Direct reads are
fixed endpoint commands; there is no raw method/path/body passthrough.

## Write actions & risks

Jira write actions are named reverse-ETL actions in `writes.json`; they do not run directly from
inspection or docs commands. They use the existing plan -> preview -> explicit approval -> execute
flow. DELETE actions and other destructive actions declare `confirm: "destructive"`, so execution
also requires the typed `--confirm destructive` challenge printed by the plan output.

Generated write command metadata:

- implemented provider-style write commands with scalar required fields: `283`
- partial/provider-style commands that require structured JSON records through reverse ETL: `8`
- no live write certification is claimed by this bundle.

Representative fixture-backed write shapes are included for safe replay only; no live Jira provider
call was used to create this ledger.

## Known limits

- Official operation inventory is generated from Atlassian's OpenAPI file; if that source changes,
  regenerate the ledger and update counts rather than editing totals by hand.
- Bounded binary downloads are declared as blocked operation rows because the shared command runner
  intentionally lacks an operation-backed binary/file executor in this slice.
- `5` reverse-ETL operations that require raw scalar or binary request bodies are
  recorded as blocked operation rows; implementing them needs a shared raw-body write dialect rather
  than connector-local JSON approximations.
- Provider-style one-off CLI commands cannot express required nested JSON object fields as flags.
  Those actions remain available to reverse ETL record-driven plans, and their command metadata is
  marked partial instead of pretending scalar flags are sufficient.
- Fixture-only and replay evidence does not certify live Jira behavior. Live certification requires
  separate credentials, sandbox policy, and write cleanup approval.
- Operation-ledger blocked row models: admin_reverse_etl=5, binary_read=2, deprecated=28, direct_read=1.
