# Overview

Reads and writes CircleCI projects, pipelines, workflows, jobs, contexts, schedules, environment
variables, checkout keys, and workflow insights through the CircleCI v2 REST API.

Readable streams: `projects`, `pipelines`, `workflows`, `jobs`, `contexts`, `schedules`,
`checkout_keys`, `environment_variables`, `insights_workflow_summary`.

Write actions: 34 closed contracts. Seven existing schedule, environment-variable, and checkout-key
actions remain alongside 27 source-derived CircleCI v2 contracts for contexts, webhooks, project
settings, pipeline definitions, organization groups, usage exports, OIDC claims, and typed
cancel/approve actions. Run `pm connectors inspect circleci --json` for the authoritative action
schemas, provider paths, and risk descriptions.

Service API documentation: https://circleci.com/docs/api/v2/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); CircleCI personal API token, sent as the Circle-Token
  header. Never logged.
- `base_url` (optional, string); default `https://circleci.com/api/v2`; format `uri`; CircleCI API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `org` (optional, string); CircleCI organization/account segment of the project slug (e.g. acme).
  Required for the projects, pipelines, schedules, checkout_keys, and environment_variables streams,
  and doubles as the second segment of the contexts stream's derived owner-slug.
- `pipeline_id` (optional, string); CircleCI pipeline ID. Required for the workflows stream.
- `repo` (optional, string); CircleCI repository segment of the project slug (e.g. widgets).
  Required for the projects, pipelines, schedules, checkout_keys, environment_variables, and
  insights_workflow_summary streams.
- `vcs_type` (optional, string); CircleCI VCS type segment of the project slug (e.g. gh, bb).
  Required for the projects, pipelines, schedules, checkout_keys, and environment_variables streams,
  and doubles as the first segment of the contexts stream's derived owner-slug.
- `workflow_id` (optional, string); CircleCI workflow ID. Required for the jobs stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://circleci.com/api/v2`.

Authentication behavior:

- API key authentication in `Circle-Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/me`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `pipelines`, `workflows`, `jobs`, `contexts`, `schedules`,
`checkout_keys`, `environment_variables`, `insights_workflow_summary`; none: `projects`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `projects`: GET `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}` - records path
  None; computed output fields `default_branch`, `vcs_url`.
- `pipelines`: GET `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/pipeline` -
  records path `items`; cursor pagination; cursor parameter `page-token`; next token from
  `next_page_token`; incremental cursor `created_at`; formatted as `rfc3339`.
- `workflows`: GET `/pipeline/{{ config.pipeline_id }}/workflow` - records path `items`; cursor
  pagination; cursor parameter `page-token`; next token from `next_page_token`; incremental cursor
  `created_at`; formatted as `rfc3339`.
- `jobs`: GET `/workflow/{{ config.workflow_id }}/job` - records path `items`; cursor pagination;
  cursor parameter `page-token`; next token from `next_page_token`; incremental cursor `started_at`;
  formatted as `rfc3339`.
- `contexts`: GET `/context` - records path `items`; query `owner-slug`=`{{ config.vcs_type }}/{{
  config.org }}`; cursor pagination; cursor parameter `page-token`; next token from
  `next_page_token`.
- `schedules`: GET `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/schedule` -
  records path `items`; cursor pagination; cursor parameter `page-token`; next token from
  `next_page_token`; incremental cursor `updated-at`; formatted as `rfc3339`.
- `checkout_keys`: GET `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo
  }}/checkout-key` - records path `items`; cursor pagination; cursor parameter `page-token`; next
  token from `next_page_token`.
- `environment_variables`: GET `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo
  }}/envvar` - records path `items`; cursor pagination; cursor parameter `page-token`; next token
  from `next_page_token`.
- `insights_workflow_summary`: GET `/insights/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo
  }}/workflows` - records path `items`; cursor pagination; cursor parameter `page-token`; next token
  from `next_page_token`.

## Write actions & risks

Every CircleCI mutation uses reverse ETL plan → preview → approval → execute. Delete actions also
require typed destructive confirmation. The generated terminal surface exposes nine stream commands
and 34 typed action commands; it has no raw HTTP or generic request-body escape hatch.

Webhook create/update actions redact `signing-secret` in write errors and previews. No secret value
is included in fixtures, examples, or this documentation.

## Known limits

- Batch defaults: read_page_size=100.
- The v2 provider-artifact ledger has 111 documented operations: 40 executable bindings (nine
  stream-backed endpoints and 31 write-backed endpoints for 34 actions) and 71 source-cited blocked
  or disallowed rows.
- The current public OpenAPI document is pinned in `sources/circleci-operation-source-lock.json`
  (111 documented operations, retrieved 2026-08-23; SHA-256
  `61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07`). Operations not represented
  by a typed stream or action remain explicitly blocked with provider-artifact provenance.
