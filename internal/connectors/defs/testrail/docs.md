# Overview

TestRail is a declarative-HTTP migration of `internal/connectors/testrail` (the hand-written
legacy connector this bundle migrates; the legacy package stays registered and unchanged until
wave6's registry flip). It reads TestRail projects, suites, cases, milestones, plans, runs, users,
and reference data (case types/fields, priorities, statuses, result fields, templates), and writes
approved test-management mutations through the TestRail v2 REST API
(`<base_url>/index.php?/api/v2/<method>`, TestRail's PHP-front-controller URL convention).

This is a Pass B full-surface expansion: the wave2 migration covered only the single legacy-parity
`projects` stream; every other stream and every write action here is new coverage researched
against TestRail's published API reference (`api_surface.json`).

## Auth setup

Provide a TestRail username (email) via the `username` config value and a password or API key via
the `password` secret; both are required. They are sent as HTTP Basic auth
(`Authorization: Basic base64(username:password)`), matching legacy's
`connsdk.Basic(username, password)` (`testrail.go:101`). `password` is never logged. `base_url`
defaults to `https://example.testrail.io` (legacy's own placeholder default) and should be
overridden to the operator's real TestRail instance URL.

## Streams notes

TestRail's front controller URL convention embeds the real API path as a raw (unencoded, no `=`)
query string segment rather than a normal path segment — every `stream.path` here is a STATIC
LITERAL (no `{{ }}` templating outside `fanout.id`), so it passes through `InterpolatePath`
unmodified and `url.Parse` splits it into `Path: /index.php` + `RawQuery: /api/v2/<method>...`
exactly as legacy's own `connsdk.Requester` does.

**Global (non-project-scoped) streams** — one request each, no pagination, records at the response
body's array root (`records.path: "."`):
- `projects` (legacy-parity): `GET get_projects`, primary key `["id"]`.
- `users`: `GET get_users`, primary key `["id"]`.
- `case_types`, `case_fields`, `priorities`, `statuses`, `result_fields`: TestRail reference/
  vocabulary lists, each primary-keyed `["id"]`.

**Project-scoped streams (fanned out via the engine's `fan_out` dialect)** — each fans out over
every project id discovered via a preliminary, fully-paginated `GET get_projects` request
(`fan_out.ids_from.request`), running the identical per-project sub-sequence once per id and
stamping the fanned-out project id onto every emitted record's `project_id` field
(`fan_out.stamp_field`). The stamped `project_id` is ALWAYS a string (the fan_out dialect's id
values are string-typed regardless of the source field's own JSON type), so every stream schema
here declares `project_id` as `["string", "null"]` even though TestRail's own raw wire `project_id`
field (where present) is a JSON integer — this is the engine's documented `stamp_field` contract
(see breezy-hr's `position_id` for the same pattern), not a TestRail-specific approximation.
- `templates`: `GET get_templates/{project_id}`, no pagination, primary key `["id", "project_id"]`
  (a template id is scoped per-project, not globally unique, so the composite key is required to
  avoid cross-project primary-key collisions across the fan-out).
- `suites`: `GET get_suites/{project_id}`, no pagination, primary key `["id"]`.
- `milestones`: `GET get_milestones/{project_id}`, no pagination, primary key `["id"]`.
- `cases`: `GET get_cases/{project_id}`, `next_url` pagination (`_links.next`, TestRail's real
  `offset`/`limit`+`_links.next|prev` envelope), records at `cases`, primary key `["id"]`,
  `x-cursor-field: updated_on` (schema-only — no `incremental` block is declared, since a
  fan_out-driven stream's per-id sub-sequence has no natural single "the connection's lower bound"
  request param to gate against across every project, and legacy itself had no incremental
  filtering to preserve parity with).
- `plans`: `GET get_plans/{project_id}`, `next_url` pagination, records at `plans`, primary key
  `["id"]`.
- `runs`: `GET get_runs/{project_id}`, `next_url` pagination, records at `runs`, primary key
  `["id"]`.

Per §4's sanctioned `next_url` exception, `cases`/`plans`/`runs` ship a fixture that proves the
fan-out's first page (the `get_projects` id-listing request) and each project's own first response
page, without attempting to fabricate a second same-project page (the replay server's own address
is unknown at fixture-authoring time for a true `next_url` continuation) — `pagination_terminates`
exercises the `projects` stream (unpaginated) instead, per the same sanctioned pattern bitly/
calendly use.

**Not migrated — nested/multi-level path parameters** (`api_surface.json`): `get_sections` (needs
project_id AND suite_id), `get_tests` (needs run_id, itself only discoverable via a per-project
runs listing), and `get_results`/`get_results_for_run`/`get_results_for_case` (need a test_id or
run_id one or two levels deeper than a single project fan-out reaches) are excluded as
`out_of_scope` with an `ENGINE_GAP`-flavored reason: the engine's `fan_out` dialect resolves exactly
ONE id level per stream declaration (`FanOutSpec` is a single field on `Stream`, not itself
recursive/nestable), and these resources need a second (or third) nesting level. This is a genuine
dialect limitation, not a per-connector workaround — see `api_surface.json`'s per-endpoint
`reason` text for the exact nesting each excluded endpoint would need.

## Write actions & risks

- `add_project` (create, `POST add_project`): creates a new top-level project; low-risk, no
  approval required.
- `add_milestone` (create, `POST add_milestone/{project_id}`): creates a new milestone under a
  project; low-risk, no approval required.
- `add_suite` (create, `POST add_suite/{project_id}`): creates a new test suite under a project;
  low-risk, no approval required.
- `add_case` (create, `POST add_case/{section_id}`): creates a new test case in a section;
  low-risk, no approval required.
- `update_case` (update, `POST update_case/{id}`): mutates an existing case's title/type/priority/
  milestone/estimate/refs.
- `add_plan` (create, `POST add_plan/{project_id}`): creates a new test plan under a project;
  low-risk, no approval required.
- `add_run` (create, `POST add_run/{project_id}`): creates a new test run, optionally selecting
  specific case_ids into it; low-risk, no approval required.
- `close_run` (update, `POST close_run/{id}`): closes and archives a run; no further results can be
  added afterward.
- `delete_run` (delete, `POST delete_run/{id}`, `missing_ok_status: [400]` — TestRail's PHP API
  answers an invalid/already-deleted id with HTTP 400, not a REST-conventional 404): permanently
  deletes a run and every test/result nested under it; irreversible.
- `add_result_for_case` (create, `POST add_result_for_case/{run_id}/{case_id}`): records a new test
  result against a case within a run; appends to result history rather than overwriting.

Destructive project/suite/section/case/milestone/plan/run/config/group/dataset/variable deletes
beyond `delete_run`, and admin-tier group/permission management, are excluded — see
`api_surface.json`'s `destructive_admin`/`requires_elevated_scope` entries.

`metadata.json` now declares `capabilities.write: true`.

## Known limits

- Nested/multi-level resources (`sections`, `tests`, `results`) are not migrated — see the
  Streams notes section above; this is an `ENGINE_GAP` (the `fan_out` dialect is single-level),
  not a per-connector scope-narrowing choice.
- Binary/multipart endpoints (attachments) and report-generation endpoints are excluded
  (`binary_payload`/`non_data_endpoint`) — the engine's `body_type` dialect (`json`/`form`/`none`)
  has no multipart shape, and a generated report is a PDF/HTML artifact, not a JSON record.
- BDD templates, shared steps, configuration groups, and datasets/variables are excluded as
  `out_of_scope` — each is a niche, opt-in TestRail feature (some Enterprise-tier-gated) not
  detectable at authoring time; Pass B breadth-vs-cost triage.
- All fixtures (`fixtures/streams/**`, `fixtures/writes/**`, `fixtures/check.json`) record
  TestRail's PHP-front-controller request shape as the replay harness sees it after Go's own
  `url.Parse` split (`path: "/index.php"`, `query: {"/api/v2/<method>...": ""}`), matching the
  wave2 `projects` stream's existing fixture convention.
- `project_id` is stamped as a STRING on every fanned-out stream's records (the engine's
  `stamp_field` contract), not TestRail's own numeric wire type — see Streams notes above.
