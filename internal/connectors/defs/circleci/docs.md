# Overview

CircleCI is a Pass B full-surface declarative-HTTP migration. It reads and writes CircleCI
projects, pipelines, workflows, jobs, contexts, schedules, checkout keys, environment variables,
and per-workflow insight summaries through the CircleCI v2 REST API
(`https://circleci.com/api/v2/...`), verified against the real OpenAPI 3.0.3 spec fetched from
`https://circleci.com/api/v2/openapi.json` on 2026-07-04. This bundle originally targeted capability
parity with `internal/connectors/circleci` (the hand-written connector it migrates, read-only); the
legacy package stays registered and unchanged until wave6's registry flip. CircleCI's
live-CI-state mutation surface (trigger pipeline, cancel workflow, approve job, rerun workflow)
remains permanently out of scope — it mutates live CI state and is not a safe reverse-ETL
target — but this bundle now covers CircleCI's safe project-CONFIGURATION mutation surface
(schedules, environment variables, checkout keys) that legacy never implemented.

## Auth setup

Provide a CircleCI personal API token via the `api_key` secret; it is sent as the `Circle-Token`
header (`{"mode": "api_key_header", "header": "Circle-Token", "value": "{{ secrets.api_key }}"}`)
and is never logged, matching legacy's `connsdk.APIKeyHeader(circleciTokenHeader, secret, "")`.
`base_url` defaults to `https://circleci.com/api/v2` and may be overridden for tests/proxies.

## Streams notes

`projects`/`pipelines` are scoped to a CircleCI project; `workflows` to a pipeline; `jobs` to a
workflow — matching legacy's per-stream `resolvePath` requirement (`errMissingProjectSlug`/
`errMissingPipelineID`/`errMissingWorkflowID`). `projects` returns a single object (not an
`items[]` list); `records.path: ""` handles this the same way every other single-object stream in
this repo does (`connsdk.RecordsAt` returns a root JSON object as a one-element record set).
`pipelines`/`workflows`/`jobs` share CircleCI's `{"items":[...],"next_page_token":...}` list
envelope; pagination is `cursor` with `token_path: next_page_token` (no `stop_path` needed — a
JSON `null` `next_page_token`, CircleCI's real end-of-list value, stringifies to `""` via
`connsdk.StringAt`, which the `token_path` cursor paginator already treats as "stop", identical to
legacy's own `strings.TrimSpace(next) == ""` check). `pipelines`/`workflows` publish
`x-cursor-field: created_at` and `jobs` publishes `started_at`, matching legacy's
`CursorFields`, but none of the 3 send a server-side incremental filter (CircleCI's v2 API exposes
none, and legacy's own `harvest` never applies one) — the cursor field is published for downstream
sync-mode eligibility without ever being requested, the same shape as this wave's aha bundle.

`projects`' nested `vcs_info.default_branch` is lifted to the top-level `default_branch` field via
`computed_fields` (bare `{{ record.vcs_info.default_branch }}` reference), matching legacy's
`circleciProjectRecord` nesting-flatten. `vcs_url` also matches legacy's fallback behavior: the
bundle emits the top-level `vcs_url` when present and falls back to `vcs_info.vcs_url` via
`{{ coalesce record.vcs_url record.vcs_info.vcs_url }}`.

`contexts`/`schedules`/`checkout_keys`/`environment_variables`/`insights_workflow_summary` are new
Pass B streams sharing the same `{"items":[...],"next_page_token":...}` cursor-pagination envelope
as `pipelines`/`workflows`/`jobs`. `contexts` is org-scoped, not project-scoped: it derives its
required `owner-slug` query parameter from `{{ config.vcs_type }}/{{ config.org }}` (the same two
config segments already used to build the project slug) since this bundle has no separate org-only
config key. `schedules`/`checkout_keys`/`environment_variables`/`insights_workflow_summary` are
project-scoped exactly like `projects`/`pipelines`. `schedules` publishes `x-cursor-field:
updated-at` (CircleCI's own hyphenated field name — schema projection is exact key-match, so the
property is declared as `"updated-at"`, not renamed); the other 4 new streams are full-refresh only.
`checkout_keys`' primary key is `fingerprint` (there is no `id` field on this resource);
`environment_variables`' primary key is `name`; `insights_workflow_summary`'s primary key is `name`
(one row per workflow name per project in the current aggregation window — CircleCI's own
`project_id`/`window_start`/`window_end` fields are preserved as opaque columns, not decomposed
further).

## Write actions & risks

`capabilities.write` is `true`. Seven actions cover CircleCI's project-CONFIGURATION mutation
surface — deliberately excluding every live-CI-state mutation (trigger/cancel/approve/rerun),
which remains out of scope exactly as legacy left it:
`create_schedule`/`update_schedule`/`delete_schedule` (a scheduled-pipeline trigger's timetable and
pipeline parameters), `create_environment_variable`/`delete_environment_variable` (project
environment variables — CircleCI's API only supports create-or-overwrite and delete, never a
partial update of an existing variable's value), and `create_checkout_key`/`delete_checkout_key`
(project deploy/checkout SSH keys). All three delete-kind actions declare `delete.missing_ok_status:
[404]` (idempotent delete). `delete_environment_variable`/`delete_checkout_key` path-interpolate
`{{ record.name }}`/`{{ record.fingerprint }}` respectively — a checkout-key fingerprint's `:`
characters pass through the engine's default per-segment `urlencode` filter unescaped (Go's
`url.QueryEscape`, which the engine wraps for path-segment encoding, does not escape `:`;
confirmed empirically against the write-replay harness). This exceeds legacy's own read-only scope
(`Write` unconditionally returned `ErrUnsupportedOperation`) — a deliberate Pass B capability
expansion targeting only configuration writes, never anything that starts, stops, or approves a CI
run.

## Known limits

- **`project_slug` is decomposed into three separate config keys (`vcs_type`/`org`/`repo`),
  replacing legacy's single opaque `project_slug` string (e.g. `gh/acme/widgets`).** The engine's
  `InterpolatePath` applies the `urlencode` filter BY DEFAULT to each individual `{{ }}` reference's
  resolved value before path substitution (percent-encoding every character including `/`) — a
  single reference resolving to a slash-containing value like `gh/acme/widgets` would therefore be
  encoded as one opaque `gh%2Facme%2Fwidgets` segment, corrupting the path CircleCI expects
  (confirmed empirically: `InterpolatePath("project/{{ config.project_slug }}/pipeline", ...)`
  yields `project/gh%2Facme%2Fwidgets/pipeline`). There is no filter in the dialect that emits a raw,
  unencoded multi-segment value from a single reference (`const:` replaces the value with a fixed
  literal, not a passthrough; no "raw"/"identity" filter exists). Decomposing into 3 config keys —
  each a single path segment, each urlencoded independently and correctly — is the same pattern
  already used by this repo's `github` bundle (`config.owner`/`config.repo`, two segments) and
  avoids the gap entirely without narrowing any read behavior: the resolved path is byte-identical
  to legacy's for the exact same effective vcs-type/org/repo triple, just supplied as 3 config
  values instead of 1 pre-joined string. This is a config-surface naming change only (ACCEPTABLE,
  conventions.md §5 meta-rule — never changes emitted record DATA for any legacy-accepted input); a
  caller previously supplying `project_slug: "gh/acme/widgets"` must instead supply `vcs_type: "gh"`,
  `org: "acme"`, `repo: "widgets"`.
- **Webhook management (`/webhook*`), pipeline-definitions/triggers, and project/org-level OIDC
  claims are out of scope**: each requires a project UUID (`scope-id`/`project_id`/`projectID`)
  that CircleCI's API exposes no slug-to-UUID lookup endpoint for; this bundle's config only ever
  carries the human-readable `vcs_type`/`org`/`repo` project-slug triple, never the underlying
  UUID, so these endpoints have no declarative resolution path. See `api_surface.json`'s
  `requires_elevated_scope`-categorized entries.
- **Org/account-tier administration is out of scope**: organizations, groups, URL Orb allow-lists,
  usage-export jobs, OTel exporters, and OPA policy/decision-audit endpoints all operate above this
  connector's per-project scope; see `api_surface.json`.
- Full CircleCI v2 API surface breakdown (every documented endpoint's covered/excluded disposition
  and reason) is recorded in `api_surface.json`.
