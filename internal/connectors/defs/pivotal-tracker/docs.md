# Overview

Pivotal Tracker is a read-only declarative migration of `internal/connectors/pivotal-tracker`
(legacy Go package `pivotaltracker`). It reads Pivotal Tracker projects, stories, iterations, and
epics through API v5. This bundle is capability-parity with legacy; legacy stays registered and
unchanged until wave6's registry flip. The connector's `docs_url` in the wave2 bundle manifest was
recorded as "manual intervention needed"; Pivotal Tracker's REST v5 API is nonetheless fully
documented at the URL below and legacy's own Go source is a complete, authoritative ground truth for
every endpoint/field this bundle implements, so this migration proceeded without an external-docs
blocker.

## Auth setup

Provide a Pivotal Tracker API token via the `api_token` secret. It is sent as the `X-TrackerToken`
header on every request (`auth.mode: api_key_header`), matching legacy's
`connsdk.APIKeyHeader("X-TrackerToken", token, "")` exactly (empty prefix — the raw token value, no
`Bearer `/other scheme prefix). Never logged.

## Streams notes

`projects` is un-scoped (`GET /projects`, no project id needed). `stories`, `iterations`, and
`epics` are project-scoped: their `path` templates `{{ config.project_id }}` into
`/projects/{{ config.project_id }}/<resource>`, matching legacy's `streamEndpoint.path`'s
`projected: true` branch, which errors when `project_id` is unset
(`pivotal-tracker stream requires config project_id`) — the engine's path interpolation reproduces
that hard-error-on-missing-required-config behavior automatically since `project_id` is a declared
(if not globally `required`) spec property referenced only by these three streams' `path` templates.

All 4 streams share the identical `offset_limit` pagination shape (`limit`/`offset` query params,
`page_size: 100`), stopping on a short page exactly like legacy's `len(records) < pageSize` check.

Every stream's record shape is renamed via `computed_fields` to reproduce legacy's exact
per-endpoint `mapRecord` functions field-for-field (all four are bare single `{{ record.<path> }}`
references, so the engine's typed extraction preserves each field's native JSON type — matching
legacy's raw `item["..."]` any-typed pass-through exactly, not a stringified copy):

- `projects`: `id`←`id`, `name`←`name`, `state`←`kind`, `updated_at`←`updated_at`.
- `stories`: `id`←`id`, `name`←`name`, `state`←`current_state`, `updated_at`←`updated_at`.
- `iterations`: `id`←`number`, `name`←`kind`, `state`←`team_strength`, `updated_at`←`finish` — this
  is legacy's own (unusual) field mapping verbatim: iterations have no natural `id`/`name`/`state` of
  their own, so legacy repurposes `number` as the identity and stamps Pivotal Tracker's iteration
  `team_strength` (a decimal number, hence `schemas/iterations.json`'s `state` type union including
  `number`) into the generic `state` slot. Preserved exactly, not "fixed", to stay at parity.
- `epics`: `id`←`id`, `name`←`name`, `state`←`label`, `updated_at`←`updated_at`.

No stream declares an `incremental` block: legacy's `CursorFields: []string{"updated_at"}` on the
`stories` stream is Catalog-only metadata — `Read` never filters or advances by any state cursor for
any of the four streams, every read is a full stream read every time. Declaring an `incremental`
block here (even an inert one with no `request_param`) would misrepresent capability this connector
does not have by exposing `incremental_append`/`incremental_append_deduped` sync modes at derivation
time, so `x-cursor-field` is intentionally absent from every schema too.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly.

## Known limits

- `page_size` (1-500 clamp) and `max_pages` (default 3, or 0/all/unlimited for unbounded) config
  knobs have no bundle-level equivalent; `page_size: 100` is a fixed value in `streams.json`'s base
  pagination block and no `MaxPages` cap is declared, relying solely on the offset paginator's
  short-page stop signal. This is a WIDER default than legacy's own `defaultMaxPages: 3` (legacy caps
  every real read at 3 pages by default unless overridden) — never a stricter read, so no
  legacy-accepted input's emitted record set is truncated relative to what this bundle now returns;
  out of scope for wave2 fan-out (Pass B).
- `project_id` is declared as an optional (not `required[]`) spec property because `projects` does
  not need it; `stories`/`iterations`/`epics` still hard-error on a missing value via ordinary path
  interpolation (an undeclared-or-required-but-missing config key is always a hard error per the
  engine's header/path resolution rules), matching legacy's own explicit check.
- The 2-page conformance fixture lives on the `projects` stream (100-record synthetic first page,
  1-record second page); `stories`/`iterations`/`epics` ship single-page fixtures only, since they
  share the identical `offset_limit` pagination shape already proven by `projects`'s fixture pair.
  Their fixture request paths use the literal string `synthetic-conformance-value` for
  `project_id` — the exact value `conformance`'s dynamic replay harness materializes for every
  non-secret spec property.
