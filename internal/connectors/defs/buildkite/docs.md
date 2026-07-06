# Overview

Reads and writes Buildkite organizations, pipelines, builds, agents, teams, and clusters through the
Buildkite REST API v2.

Readable streams: `organizations`, `pipelines`, `builds`, `agents`, `teams`, `clusters`.

Write actions: `create_pipeline`, `update_pipeline`, `archive_pipeline`, `unarchive_pipeline`,
`delete_pipeline`, `create_build`, `cancel_build`, `rebuild_build`, `create_annotation`,
`retry_job`, `unblock_job`, `stop_agent`, `pause_agent`, `resume_agent`, `create_team`,
`update_team`, `delete_team`.

Service API documentation: https://buildkite.com/docs/apis/rest-api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Buildkite API access token. Sent as Authorization: Bearer
  <api_key>; never logged.
- `base_url` (optional, string); default `https://api.buildkite.com/v2`; format `uri`; Buildkite API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `organization` (optional, string); Buildkite organization slug. Required for the pipelines,
  builds, and agents streams (organization-scoped); not needed for the top-level organizations
  stream.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only builds created at
  or after this time are read (sent as created_from).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.buildkite.com/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/organizations` with query `per_page`=`1`.

## Streams notes

Default pagination: follows RFC 5988 Link headers with rel=next.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `organizations`: GET `/organizations` - records at response root; query `per_page`=`100`; follows
  RFC 5988 Link headers with rel=next.
- `pipelines`: GET `/organizations/{{ config.organization }}/pipelines` - records at response root;
  query `per_page`=`100`; follows RFC 5988 Link headers with rel=next.
- `builds`: GET `/organizations/{{ config.organization }}/builds` - records at response root; query
  `created_from` from template `{{ incremental.lower_bound }}`, omitted when absent;
  `per_page`=`100`; follows RFC 5988 Link headers with rel=next; incremental cursor `created_at`;
  formatted as `rfc3339`; initial lower bound from `start_date`.
- `agents`: GET `/organizations/{{ config.organization }}/agents` - records at response root; query
  `per_page`=`100`; follows RFC 5988 Link headers with rel=next.
- `teams`: GET `/organizations/{{ config.organization }}/teams` - records at response root; query
  `per_page`=`100`; follows RFC 5988 Link headers with rel=next.
- `clusters`: GET `/organizations/{{ config.organization }}/clusters` - records at response root;
  query `per_page`=`100`; follows RFC 5988 Link headers with rel=next.

## Write actions & risks

Overall write risk: external mutation of pipeline lifecycle, build triggering/cancellation, job
control, agent lifecycle, and team management; create_build/rebuild_build run arbitrary
pipeline-defined commands on real agent capacity.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_pipeline`: POST `/organizations/{{ config.organization }}/pipelines` - kind `create`; body
  type `json`; required record fields `name`, `cluster_id`, `repository`; accepted fields
  `cluster_id`, `configuration`, `default_branch`, `description`, `name`, `repository`, `steps`,
  `tags`, `visibility`; risk: creates a new CI/CD pipeline scoped to a cluster and repository;
  low-risk external mutation, no approval required.
- `update_pipeline`: PATCH `/organizations/{{ config.organization }}/pipelines/{{ record.slug }}` -
  kind `update`; body type `json`; path fields `slug`; required record fields `slug`; accepted
  fields `cluster_id`, `configuration`, `default_branch`, `description`, `name`, `repository`,
  `slug`, `visibility`; risk: mutates an existing pipeline's repository, configuration, or
  visibility; a changed configuration/repository affects every future build.
- `archive_pipeline`: POST `/organizations/{{ config.organization }}/pipelines/{{ record.slug
  }}/archive` - kind `custom`; body type `none`; path fields `slug`; required record fields `slug`;
  accepted fields `slug`; risk: archives a pipeline, hiding it from the default pipeline list and
  blocking new builds until unarchived.
- `unarchive_pipeline`: POST `/organizations/{{ config.organization }}/pipelines/{{ record.slug
  }}/unarchive` - kind `custom`; body type `none`; path fields `slug`; required record fields
  `slug`; accepted fields `slug`; risk: restores a previously archived pipeline to active/buildable
  status.
- `delete_pipeline`: DELETE `/organizations/{{ config.organization }}/pipelines/{{ record.slug }}` -
  kind `delete`; body type `none`; path fields `slug`; required record fields `slug`; accepted
  fields `slug`; missing records treated as success for status `404`; risk: permanently deletes a
  pipeline and its build history; irreversible.
- `create_build`: POST `/organizations/{{ config.organization }}/pipelines/{{ record.pipeline_slug
  }}/builds` - kind `create`; body type `json`; path fields `pipeline_slug`; required record fields
  `pipeline_slug`, `commit`, `branch`; accepted fields `author`, `branch`, `clean_checkout`,
  `commit`, `env`, `ignore_pipeline_branch_filters`, `message`, `meta_data`, `pipeline_slug`; risk:
  immediately triggers a new CI/CD build on the target pipeline/branch; consumes agent capacity and
  may run arbitrary pipeline-defined commands.
- `cancel_build`: PUT `/organizations/{{ config.organization }}/pipelines/{{ record.pipeline_slug
  }}/builds/{{ record.number }}/cancel` - kind `custom`; body type `none`; path fields
  `pipeline_slug`, `number`; required record fields `pipeline_slug`, `number`; accepted fields
  `number`, `pipeline_slug`; risk: cancels a running or scheduled build; any in-progress jobs are
  terminated immediately.
- `rebuild_build`: PUT `/organizations/{{ config.organization }}/pipelines/{{ record.pipeline_slug
  }}/builds/{{ record.number }}/rebuild` - kind `custom`; body type `none`; path fields
  `pipeline_slug`, `number`; required record fields `pipeline_slug`, `number`; accepted fields
  `number`, `pipeline_slug`; risk: triggers a full re-run of a completed build on new agent
  capacity; may run arbitrary pipeline-defined commands again.
- `create_annotation`: POST `/organizations/{{ config.organization }}/pipelines/{{
  record.pipeline_slug }}/builds/{{ record.build_number }}/annotations` - kind `create`; body type
  `json`; path fields `pipeline_slug`, `build_number`; required record fields `pipeline_slug`,
  `build_number`, `body`; accepted fields `append`, `body`, `build_number`, `context`,
  `pipeline_slug`, `style`; risk: posts a visible HTML/Markdown annotation onto a build's detail
  page; low-risk external mutation, no approval required.
- `retry_job`: PUT `/organizations/{{ config.organization }}/jobs/{{ record.job_id }}/retry` - kind
  `custom`; body type `none`; path fields `job_id`; required record fields `job_id`; accepted fields
  `job_id`; risk: re-runs a single failed/finished job on new agent capacity, without re-running the
  rest of the build.
- `unblock_job`: PUT `/organizations/{{ config.organization }}/jobs/{{ record.job_id }}/unblock` -
  kind `custom`; body type `json`; path fields `job_id`; body fields `fields`, `unblocker`; required
  record fields `job_id`; accepted fields `fields`, `job_id`, `unblocker`; risk: releases a manual
  'block' pipeline step, allowing the build to continue past it immediately.
- `stop_agent`: PUT `/organizations/{{ config.organization }}/agents/{{ record.id }}/stop` - kind
  `custom`; body type `json`; path fields `id`; body fields `force`; required record fields `id`;
  accepted fields `force`, `id`; risk: stops an agent; force=true cancels any job it is currently
  processing.
- `pause_agent`: PUT `/organizations/{{ config.organization }}/agents/{{ record.id }}/pause` - kind
  `custom`; body type `json`; path fields `id`; body fields `note`, `timeout_in_minutes`; required
  record fields `id`; accepted fields `id`, `note`, `timeout_in_minutes`; risk: pauses an agent so
  it stops picking up new jobs until resumed or the timeout elapses.
- `resume_agent`: PUT `/organizations/{{ config.organization }}/agents/{{ record.id }}/resume` -
  kind `custom`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: resumes a previously paused agent so it can pick up new jobs again.
- `create_team`: POST `/organizations/{{ config.organization }}/teams` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `default_member_role`, `description`,
  `is_default_team`, `name`, `privacy`; risk: creates a new team; low-risk external mutation, no
  approval required.
- `update_team`: PATCH `/organizations/{{ config.organization }}/teams/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `default_member_role`, `description`, `id`, `name`, `privacy`; risk: mutates an existing team's
  name, privacy, or default permissions; a privacy change from visible to secret hides membership
  immediately.
- `delete_team`: DELETE `/organizations/{{ config.organization }}/teams/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; risk: permanently deletes a team and its
  pipeline/member associations; irreversible.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s), 17 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=3, duplicate_of=14, non_data_endpoint=3, out_of_scope=46,
  requires_elevated_scope=9.
