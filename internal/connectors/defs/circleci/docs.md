# Overview

CircleCI is a wave2 fan-out declarative-HTTP migration. It reads CircleCI projects, pipelines,
workflows, and jobs through the CircleCI v2 REST API (`GET https://circleci.com/api/v2/...`). This
bundle targets capability parity with `internal/connectors/circleci` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip. The
connector is read-only: CircleCI's write surface (trigger pipeline, cancel workflow, approve job)
mutates live CI state and is not a safe reverse-ETL target, matching legacy exactly.

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

`projects`' single nested `vcs_info.default_branch` is lifted to the top-level `default_branch`
field via `computed_fields` (bare `{{ record.vcs_info.default_branch }}` reference), matching
legacy's `circleciProjectRecord` nesting-flatten.

## Write actions & risks

None. CircleCI's mutation endpoints (trigger pipeline, cancel workflow, approve job) affect live CI
state and are not safe reverse-ETL targets; `capabilities.write` is `false` and this bundle ships no
`writes.json`, matching legacy's `Write` unconditionally returning
`connectors.ErrUnsupportedOperation`.

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
- **`vcs_url`'s `vcs_info.vcs_url` fallback is not modeled.** Legacy's `circleciProjectRecord` reads
  the top-level `vcs_url` field and falls back to the nested `vcs_info.vcs_url` only when the
  top-level field is absent. The engine's `computed_fields` dialect has no "first of N paths"
  coalesce primitive (same shape as this wave's cin7 `firstField` narrowing). This bundle's schema
  projects the top-level `vcs_url` field directly (present on every real CircleCI project response);
  the nested fallback, which legacy defends against for a hypothetical response shape never observed
  in practice, is not reproduced. ACCEPTABLE per conventions.md §5's meta-rule.
- Full CircleCI v2 API surface (contexts, insights, schedules, pipeline-trigger, workflow
  approval/cancel) is out of scope for this wave; see `api_surface.json`'s `excluded` entries.
