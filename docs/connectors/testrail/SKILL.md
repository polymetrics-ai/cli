---
name: pm-testrail
description: TestRail connector knowledge and safe action guide.
---

# pm-testrail

## Purpose

Reads TestRail projects, suites, cases, milestones, plans, runs, users, and reference data (case types/fields, priorities, statuses, result fields, templates), and writes approved test-management mutations (projects, milestones, suites, cases, plans, runs, results) through the TestRail v2 API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- username
- password (secret)

## ETL Streams

- projects:
  - primary key: id
  - fields: announcement(), id(), is_completed(), name()
- users:
  - primary key: id
  - fields: email(), id(), is_active(), name(), role(), role_id()
- case_types:
  - primary key: id
  - fields: id(), is_default(), name()
- case_fields:
  - primary key: id
  - fields: id(), is_active(), label(), name(), system_name(), type_id()
- priorities:
  - primary key: id
  - fields: id(), is_default(), name(), priority(), short_name()
- statuses:
  - primary key: id
  - fields: id(), is_final(), is_system(), is_untested(), label(), name()
- result_fields:
  - primary key: id
  - fields: id(), is_active(), label(), name(), system_name(), type_id()
- templates:
  - primary key: id, project_id
  - fields: id(), is_default(), name(), project_id()
- suites:
  - primary key: id
  - fields: description(), id(), is_completed(), is_master(), name(), project_id(), url()
- milestones:
  - primary key: id
  - fields: completed_on(), description(), due_on(), id(), is_completed(), is_started(), name(), parent_id(), project_id(), start_on(), started_on(), url()
- cases:
  - primary key: id
  - cursor: updated_on
  - fields: created_by(), created_on(), estimate(), id(), milestone_id(), priority_id(), project_id(), refs(), section_id(), suite_id(), template_id(), title(), type_id(), updated_by(), updated_on()
- plans:
  - primary key: id
  - fields: assignedto_id(), completed_on(), created_by(), created_on(), description(), failed_count(), id(), is_completed(), milestone_id(), name(), passed_count(), project_id(), untested_count(), url()
- runs:
  - primary key: id
  - fields: assignedto_id(), completed_on(), created_by(), created_on(), description(), failed_count(), id(), is_completed(), milestone_id(), name(), passed_count(), plan_id(), project_id(), suite_id(), untested_count(), url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- add_project:
  - endpoint: POST index.php?/api/v2/add_project
  - risk: creates a new top-level TestRail project; low-risk external mutation, no approval required
- add_milestone:
  - endpoint: POST index.php?/api/v2/add_milestone/{{ record.project_id }}
  - required fields: project_id
  - risk: creates a new milestone under the target project; low-risk external mutation, no approval required
- add_suite:
  - endpoint: POST index.php?/api/v2/add_suite/{{ record.project_id }}
  - required fields: project_id
  - risk: creates a new test suite under the target project; low-risk external mutation, no approval required
- add_case:
  - endpoint: POST index.php?/api/v2/add_case/{{ record.section_id }}
  - required fields: section_id
  - risk: creates a new test case in the target section; low-risk external mutation, no approval required
- update_case:
  - endpoint: POST index.php?/api/v2/update_case/{{ record.id }}
  - required fields: id
  - risk: mutates an existing test case's title, type, priority, milestone, estimate, or references
- add_plan:
  - endpoint: POST index.php?/api/v2/add_plan/{{ record.project_id }}
  - required fields: project_id
  - risk: creates a new test plan under the target project; low-risk external mutation, no approval required
- add_run:
  - endpoint: POST index.php?/api/v2/add_run/{{ record.project_id }}
  - required fields: project_id
  - risk: creates a new test run under the target project, selecting test cases into it for execution; low-risk external mutation, no approval required
- close_run:
  - endpoint: POST index.php?/api/v2/close_run/{{ record.id }}
  - required fields: id
  - risk: closes and archives an existing test run; no further results can be added to it after closing
- delete_run:
  - endpoint: POST index.php?/api/v2/delete_run/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a test run and all of its tests and results; irreversible
- add_result_for_case:
  - endpoint: POST index.php?/api/v2/add_result_for_case/{{ record.run_id }}/{{ record.case_id }}
  - required fields: run_id, case_id
  - risk: records a new test result (pass/fail/etc.) against a case within a run; appends to result history, does not overwrite prior results

## Security

- read risk: external TestRail API read of project, suite, case, milestone, plan, run, and reference data
- write risk: external TestRail API mutation (create/update projects, milestones, suites, cases, plans, runs; close/delete runs; add test results)
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect testrail
```

### Inspect as structured JSON

```bash
pm connectors inspect testrail --json
```

## Agent Rules

- Run pm connectors inspect testrail before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
