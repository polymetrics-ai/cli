# Overview

Reads TestRail projects, suites, cases, milestones, plans, runs, users, and reference data (case
types/fields, priorities, statuses, result fields, templates), and writes approved test-management
mutations (projects, milestones, suites, cases, plans, runs, results) through the TestRail v2 API.

Readable streams: `projects`, `users`, `case_types`, `case_fields`, `priorities`, `statuses`,
`result_fields`, `templates`, `suites`, `milestones`, `cases`, `plans`, `runs`.

Write actions: `add_project`, `add_milestone`, `add_suite`, `add_case`, `update_case`, `add_plan`,
`add_run`, `close_run`, `delete_run`, `add_result_for_case`.

Service API documentation:
https://support.testrail.com/hc/en-us/articles/7077039051284-Introduction-to-the-TestRail-API.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://example.testrail.io`; format `uri`; TestRail
  instance base URL override for tests or proxies.
- `password` (required, secret, string); TestRail password or API key, sent as the Basic auth
  password. Never logged.
- `username` (required, string); TestRail username (email), sent as the Basic auth username.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://example.testrail.io`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `index.php?/api/v2/get_projects`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `cases`, `plans`, `runs`; none: `projects`, `users`, `case_types`,
`case_fields`, `priorities`, `statuses`, `result_fields`, `templates`, `suites`, `milestones`.

- `projects`: GET `index.php?/api/v2/get_projects` - records path `.`.
- `users`: GET `index.php?/api/v2/get_users` - records path `.`.
- `case_types`: GET `index.php?/api/v2/get_case_types` - records path `.`.
- `case_fields`: GET `index.php?/api/v2/get_case_fields` - records path `.`.
- `priorities`: GET `index.php?/api/v2/get_priorities` - records path `.`.
- `statuses`: GET `index.php?/api/v2/get_statuses` - records path `.`.
- `result_fields`: GET `index.php?/api/v2/get_result_fields` - records path `.`.
- `templates`: GET `index.php?/api/v2/get_templates/{{ fanout.id }}` - records path `.`; fan-out;
  ids from request `index.php?/api/v2/get_projects`; id-list records path `.`; id field `id`; id
  inserted into the request path; stamps `project_id`.
- `suites`: GET `index.php?/api/v2/get_suites/{{ fanout.id }}` - records path `.`; fan-out; ids from
  request `index.php?/api/v2/get_projects`; id-list records path `.`; id field `id`; id inserted
  into the request path; stamps `project_id`.
- `milestones`: GET `index.php?/api/v2/get_milestones/{{ fanout.id }}` - records path `.`; fan-out;
  ids from request `index.php?/api/v2/get_projects`; id-list records path `.`; id field `id`; id
  inserted into the request path; stamps `project_id`.
- `cases`: GET `index.php?/api/v2/get_cases/{{ fanout.id }}` - records path `cases`; follows a
  next-page URL from the response body; URL path `_links.next`; next URLs stay on the configured API
  host; fan-out; ids from request `index.php?/api/v2/get_projects`; id-list records path `.`; id
  field `id`; id inserted into the request path; stamps `project_id`.
- `plans`: GET `index.php?/api/v2/get_plans/{{ fanout.id }}` - records path `plans`; follows a
  next-page URL from the response body; URL path `_links.next`; next URLs stay on the configured API
  host; fan-out; ids from request `index.php?/api/v2/get_projects`; id-list records path `.`; id
  field `id`; id inserted into the request path; stamps `project_id`.
- `runs`: GET `index.php?/api/v2/get_runs/{{ fanout.id }}` - records path `runs`; follows a
  next-page URL from the response body; URL path `_links.next`; next URLs stay on the configured API
  host; fan-out; ids from request `index.php?/api/v2/get_projects`; id-list records path `.`; id
  field `id`; id inserted into the request path; stamps `project_id`.

## Write actions & risks

Overall write risk: external TestRail API mutation (create/update projects, milestones, suites,
cases, plans, runs; close/delete runs; add test results).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `add_project`: POST `index.php?/api/v2/add_project` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `announcement`, `name`, `show_announcement`, `suite_mode`;
  risk: creates a new top-level TestRail project; low-risk external mutation, no approval required.
- `add_milestone`: POST `index.php?/api/v2/add_milestone/{{ record.project_id }}` - kind `create`;
  body type `json`; path fields `project_id`; required record fields `project_id`, `name`; accepted
  fields `description`, `due_on`, `name`, `parent_id`, `project_id`, `start_on`; risk: creates a new
  milestone under the target project; low-risk external mutation, no approval required.
- `add_suite`: POST `index.php?/api/v2/add_suite/{{ record.project_id }}` - kind `create`; body type
  `json`; path fields `project_id`; required record fields `project_id`, `name`; accepted fields
  `description`, `name`, `project_id`; risk: creates a new test suite under the target project;
  low-risk external mutation, no approval required.
- `add_case`: POST `index.php?/api/v2/add_case/{{ record.section_id }}` - kind `create`; body type
  `json`; path fields `section_id`; required record fields `section_id`, `title`; accepted fields
  `estimate`, `milestone_id`, `priority_id`, `refs`, `section_id`, `template_id`, `title`,
  `type_id`; risk: creates a new test case in the target section; low-risk external mutation, no
  approval required.
- `update_case`: POST `index.php?/api/v2/update_case/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `estimate`, `id`,
  `milestone_id`, `priority_id`, `refs`, `title`, `type_id`; risk: mutates an existing test case's
  title, type, priority, milestone, estimate, or references.
- `add_plan`: POST `index.php?/api/v2/add_plan/{{ record.project_id }}` - kind `create`; body type
  `json`; path fields `project_id`; required record fields `project_id`, `name`; accepted fields
  `description`, `milestone_id`, `name`, `project_id`; risk: creates a new test plan under the
  target project; low-risk external mutation, no approval required.
- `add_run`: POST `index.php?/api/v2/add_run/{{ record.project_id }}` - kind `create`; body type
  `json`; path fields `project_id`; required record fields `project_id`, `name`; accepted fields
  `assignedto_id`, `case_ids`, `description`, `include_all`, `milestone_id`, `name`, `project_id`,
  `suite_id`; risk: creates a new test run under the target project, selecting test cases into it
  for execution; low-risk external mutation, no approval required.
- `close_run`: POST `index.php?/api/v2/close_run/{{ record.id }}` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: closes and archives an
  existing test run; no further results can be added to it after closing.
- `delete_run`: POST `index.php?/api/v2/delete_run/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `400`; risk: permanently deletes a test run and all of its tests and
  results; irreversible.
- `add_result_for_case`: POST `index.php?/api/v2/add_result_for_case/{{ record.run_id }}/{{
  record.case_id }}` - kind `create`; body type `json`; path fields `run_id`, `case_id`; required
  record fields `run_id`, `case_id`; accepted fields `case_id`, `comment`, `defects`, `elapsed`,
  `run_id`, `status_id`, `version`; risk: records a new test result (pass/fail/etc.) against a case
  within a run; appends to result history, does not overwrite prior results.

## Known limits

- API coverage includes 13 stream-backed endpoint group(s), 10 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=13, destructive_admin=14, duplicate_of=14, non_data_endpoint=2, out_of_scope=38,
  requires_elevated_scope=2.
