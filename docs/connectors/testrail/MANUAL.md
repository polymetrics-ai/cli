# pm connectors inspect testrail

```text
NAME
  pm connectors inspect testrail - TestRail connector manual

SYNOPSIS
  pm connectors inspect testrail
  pm connectors inspect testrail --json
  pm credentials add <name> --connector testrail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads TestRail projects, suites, cases, milestones, plans, runs, users, and reference data (case types/fields, priorities, statuses, result fields, templates), and writes approved test-management mutations (projects, milestones, suites, cases, plans, runs, results) through the TestRail v2 API.

ICON
  id: simple-icons-testrail
  asset: icons/simple-icons/testrail.svg
  title: TestRail
  simple_icon_slug: testrail
  simple_icon_hex: 65C179
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=TestRail
  match: exact-name-or-slug
  matched_by: testrail

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  username (required)
  password (secret) (required)

ETL STREAMS
  projects:
    primary key: id
    fields: announcement(string), id(integer), is_completed(boolean), name(string)
  users:
    primary key: id
    fields: email(string), id(integer), is_active(boolean), name(string), role(string), role_id(integer)
  case_types:
    primary key: id
    fields: id(integer), is_default(boolean), name(string)
  case_fields:
    primary key: id
    fields: id(integer), is_active(boolean), label(string), name(string), system_name(string), type_id(integer)
  priorities:
    primary key: id
    fields: id(integer), is_default(boolean), name(string), priority(integer), short_name(string)
  statuses:
    primary key: id
    fields: id(integer), is_final(boolean), is_system(boolean), is_untested(boolean), label(string), name(string)
  result_fields:
    primary key: id
    fields: id(integer), is_active(boolean), label(string), name(string), system_name(string), type_id(integer)
  templates:
    primary key: id, project_id
    fields: id(integer), is_default(boolean), name(string), project_id(string)
  suites:
    primary key: id
    fields: description(string), id(integer), is_completed(boolean), is_master(boolean), name(string), project_id(string), url(string)
  milestones:
    primary key: id
    fields: completed_on(integer), description(string), due_on(integer), id(integer), is_completed(boolean), is_started(boolean), name(string), parent_id(integer), project_id(string), start_on(integer), started_on(integer), url(string)
  cases:
    primary key: id
    cursor: updated_on
    fields: created_by(integer), created_on(integer), estimate(string), id(integer), milestone_id(integer), priority_id(integer), project_id(string), refs(string), section_id(integer), suite_id(integer), template_id(integer), title(string), type_id(integer), updated_by(integer), updated_on(integer)
  plans:
    primary key: id
    fields: assignedto_id(integer), completed_on(integer), created_by(integer), created_on(integer), description(string), failed_count(integer), id(integer), is_completed(boolean), milestone_id(integer), name(string), passed_count(integer), project_id(string), untested_count(integer), url(string)
  runs:
    primary key: id
    fields: assignedto_id(integer), completed_on(integer), created_by(integer), created_on(integer), description(string), failed_count(integer), id(integer), is_completed(boolean), milestone_id(integer), name(string), passed_count(integer), plan_id(integer), project_id(string), suite_id(integer), untested_count(integer), url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  add_project:
    endpoint: POST index.php?/api/v2/add_project
    required fields: name
    risk: creates a new top-level TestRail project; low-risk external mutation, no approval required
  add_milestone:
    endpoint: POST index.php?/api/v2/add_milestone/{{ record.project_id }}
    required fields: project_id, name
    risk: creates a new milestone under the target project; low-risk external mutation, no approval required
  add_suite:
    endpoint: POST index.php?/api/v2/add_suite/{{ record.project_id }}
    required fields: project_id, name
    risk: creates a new test suite under the target project; low-risk external mutation, no approval required
  add_case:
    endpoint: POST index.php?/api/v2/add_case/{{ record.section_id }}
    required fields: section_id, title
    risk: creates a new test case in the target section; low-risk external mutation, no approval required
  update_case:
    endpoint: POST index.php?/api/v2/update_case/{{ record.id }}
    required fields: id
    risk: mutates an existing test case's title, type, priority, milestone, estimate, or references
  add_plan:
    endpoint: POST index.php?/api/v2/add_plan/{{ record.project_id }}
    required fields: project_id, name
    risk: creates a new test plan under the target project; low-risk external mutation, no approval required
  add_run:
    endpoint: POST index.php?/api/v2/add_run/{{ record.project_id }}
    required fields: project_id, name
    risk: creates a new test run under the target project, selecting test cases into it for execution; low-risk external mutation, no approval required
  close_run:
    endpoint: POST index.php?/api/v2/close_run/{{ record.id }}
    required fields: id
    risk: closes and archives an existing test run; no further results can be added to it after closing
  delete_run:
    endpoint: POST index.php?/api/v2/delete_run/{{ record.id }}
    required fields: id
    risk: permanently deletes a test run and all of its tests and results; irreversible
  add_result_for_case:
    endpoint: POST index.php?/api/v2/add_result_for_case/{{ record.run_id }}/{{ record.case_id }}
    required fields: run_id, case_id
    risk: records a new test result (pass/fail/etc.) against a case within a run; appends to result history, does not overwrite prior results

SECURITY
  read risk: external TestRail API read of project, suite, case, milestone, plan, run, and reference data
  write risk: external TestRail API mutation (create/update projects, milestones, suites, cases, plans, runs; close/delete runs; add test results)
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect testrail

  # Inspect as structured JSON
  pm connectors inspect testrail --json

AGENT WORKFLOW
  - Run pm connectors inspect testrail before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
