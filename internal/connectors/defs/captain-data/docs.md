# Overview

Reads Captain Data workspace, workflows, jobs, and job results, and writes a launch_workflow action
to trigger a new workflow run, through the Captain Data v3 REST API.

Readable streams: `workspace`, `workflows`, `jobs`, `job_results`.

Write actions: `launch_workflow`.

Service API documentation: https://docs.captaindata.com/api-documentation.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Captain Data API key. Sent as the X-API-Key header; never
  logged.
- `base_url` (optional, string); default `https://api.captaindata.co/v3`; format `uri`; Captain Data
  API base URL override for tests or proxies.
- `job_uid` (optional, string); Parent job uid; required to read the job_results stream (scopes GET
  /jobs/{job_uid}/results).
- `mode` (optional, string).
- `project_uid` (required, string); Captain Data project id.
- `workflow_uid` (optional, string); Parent workflow uid; required to read the jobs stream (scopes
  GET /workflows/{workflow_uid}/jobs).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.captaindata.co/v3`.

Authentication behavior:

- API key authentication in `X-API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/workspace`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `job_results`; none: `workspace`, `workflows`, `jobs`.

- `workspace`: GET `/workspace` - records at response root.
- `workflows`: GET `/workflows` - records at response root.
- `jobs`: GET `/workflows/{{ config.workflow_uid }}/jobs` - records at response root.
- `job_results`: GET `/jobs/{{ config.job_uid }}/results` - records path `results`; cursor
  pagination; cursor parameter `cursor`; next token from `paging.next`; stop flag
  `paging.have_next_page`.

## Write actions & risks

Overall write risk: external mutation; launches a live Captain Data workflow run, consuming account
credits and potentially performing external side effects depending on the workflow's own configured
steps; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `launch_workflow`: POST `/workflows/{{ record.workflow_uid }}/launch` - kind `create`; body type
  `json`; path fields `workflow_uid`; required record fields `workflow_uid`; accepted fields
  `accounts`, `accounts_rotation_enabled`, `output_column`, `parameters`, `workflow_uid`; risk:
  external mutation; launches a live Captain Data workflow run (a new job), consuming account
  credits and potentially performing external actions (scraping, enrichment, outreach) depending on
  the workflow's own configured steps; approval required.

## Known limits

- API coverage includes 4 stream-backed endpoint group(s), 1 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=1, out_of_scope=6, requires_elevated_scope=2.
