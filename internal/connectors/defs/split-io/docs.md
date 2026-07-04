# Overview

Split.io is a wave2 fan-out migration from `internal/connectors/split-io` (the legacy
hand-written connector this bundle replaces at capability parity). It reads Split.io workspaces,
environments, feature flags (splits), and segments through the Split Admin API. Read-only; the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Split.io Admin API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)`.

## Streams notes

`workspaces` needs no path parameter. `environments`, `splits`, and `segments` each require
`config.workspace_id` to substitute the `{workspace_id}` path segment (`path:
"/internal/api/v2/environments/ws/{{ config.workspace_id }}"`, etc.) — legacy's
`resolveResource` hard-errors with `"stream requires config workspace_id"` when unset; this bundle
reproduces the identical requirement by declaring `workspace_id` as a plain (non-required-at-spec-
level, since `workspaces` doesn't need it) config property referenced directly in the three
path-scoped streams' `path` templates — an absent `config.workspace_id` is a hard interpolation
error on those three streams' reads exactly like legacy's explicit check, just surfaced through
the engine's own path-interpolation error rather than a bespoke message.

All 4 streams share the identical pagination shape: `offset_limit` (`limit_param: limit`,
`offset_param: offset`, `page_size: 100`), records at `objects` — matches legacy's
`connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam: "offset", PageSize: pageSize}` reading
`endpoint.recordsPath` (`"objects"` for every stream) with `defaultPageSize = 100`.
`page_size`/`max_pages` were legacy config knobs with no equivalent config-driven mechanism in
this dialect (`PaginationSpec.PageSize`/`MaxPages` are static `streams.json` fields, not
`{{ }}`-templated) — see Known limits.

Legacy performs no incremental/state-cursor filtering during `Read` (no persisted cursor is read
or sent as a request filter) even though `splits`/`segments` declare `updatedAt` as a
`CursorFields` catalog hint; this bundle matches that exactly — `schemas/splits.json` and
`schemas/segments.json` declare `x-cursor-field: updatedAt` (descriptive, matching legacy's
catalog metadata) but no stream declares an `incremental` block, since legacy never actually
applies one during a read.

### Pass B additions

Three new read streams, added against the real, live "api-settings" OpenAPI spec (fetched directly
from the embedded page-state JSON `docs.split.io` renders on every reference page — see
`api_surface.json`'s `scope` note):

- **`groups`** — `GET /internal/api/v2/groups`, records at `objects`. The real API's list response
  includes `nextMarker`/`previousMarker` cursor fields, but no request-side cursor query parameter
  is documented anywhere in the spec for this endpoint (unlike `users`, below); rather than guess
  an unconfirmed parameter name, this stream declares `pagination: {"type": "none"}` — a single
  request reads whatever the account's default page size returns.
  See Known limits.
- **`traffic_types`** — `GET /internal/api/v2/trafficTypes/ws/{{ config.workspace_id }}`. The real
  API returns a BARE ARRAY for this endpoint (`[{id, name, displayAttributeId}, ...]`), unlike
  `workspaces`/`splits`/`segments`'s `{objects: [...], offset, limit, totalCount}` envelope — so
  this stream declares `records.path: ""` (root-is-the-array) and `pagination: {"type": "none"}`
  (the real API documents no pagination parameters for this endpoint at all).
- **`users`** — `GET /internal/api/v2/users`, records at `data`. Cursor-paginated using the real
  API's own `after`/`nextMarker` convention: `pagination: {"type": "cursor", "cursor_param":
  "after", "token_path": "nextMarker"}` — the identical `cursor`+`token_path` paginator type every
  other cursor-paginated bundle uses, just with Split's own non-generic parameter/field names.

## Write actions & risks

Six write actions, all added against the real, live OpenAPI spec. Every one is a feature-flag or
segment-targeting lifecycle mutation — legacy shipped none of these (its `Write` returned
`connectors.ErrUnsupportedOperation` unconditionally), so there is no parity baseline to diverge
from; `docs/migration/conventions.md` §5's parity-deviation ledger has no new entries for this
bundle.

- **`kill_feature_flag_in_environment`** (`PUT .../splits/ws/{{ record.workspace_id }}/{{
  record.feature_flag_name }}/environments/{{ record.environment_id }}/kill`) — forces every SDK
  evaluating this flag in the given environment onto its off/default treatment. Path-only, no
  request body (`body_type: none`); the real API's PUT variant takes none. High-impact production
  traffic-shaping mutation, approval required.
- **`restore_feature_flag_in_environment`** (`.../restore`) — reverts a killed flag back to its
  configured targeting rules; same path-only shape, same risk class.
- **`archive_feature_flag`** (`PUT .../splits/ws/{{ record.workspace_id }}/{{
  record.feature_flag_name }}/archive`) — archives a feature flag account-wide. Real body is
  `{title?, comment?}`, both optional (a change-log annotation, not flag configuration); modeled as
  a flat JSON body with no required fields beyond the two path components.
- **`unarchive_feature_flag`** (`.../unarchive`) — restores an archived flag to active use; same
  optional `{title?, comment?}` body shape.
- **`add_segment_keys_in_environment`** (`PUT .../segments/{{ record.environment_id }}/{{
  record.segment_name }}/uploadKeys`) — adds member keys to a segment in an environment via the
  real API's documented JSON body shape (`{"keys": [...], "comment": "..."}`, confirmed from the
  reference page's rendered example; the OpenAPI page-state JSON did not itself carry a body
  schema for this endpoint). Changes which end-users match segment-based targeting rules for
  every feature flag using that segment; approval required.
- **`remove_segment_keys_from_environment`** (`.../removeKeys`) — removes member keys from a
  segment in an environment; identical `{"keys": [...], "comment": "..."}` body shape.

Every archive/kill/restore action's path resolves `workspace_id`/`feature_flag_name`/
`environment_id`/`segment_name` from the WRITE RECORD itself (`path_fields`), not from
`config.workspace_id` (the read-side config property) — a write record must carry its own
identifying fields since a single write batch may target flags/segments across different
workspaces or environments.

## Known limits

- **`page_size`/`max_pages` are not exposed as config properties.** Legacy accepts
  `config.page_size` (bounded 1-1000, default 100) and `config.max_pages` (default unbounded) at
  read time. The engine's `PaginationSpec.PageSize`/`MaxPages` fields are static values baked into
  `streams.json`'s pagination block, with no `{{ }}` templating support from `config.*` — there is
  no per-run override mechanism at all (matching searxng's `page_size`/`max_pages` precedent, F6
  REVIEW.md). This bundle hard-codes `page_size: 100` (legacy's own default) and declares no
  `max_pages` (unbounded, matching legacy's own default when the config value is unset/`all`/
  `unlimited`). A caller that previously overrode either value per-run loses that capability;
  every default-config caller sees byte-identical behavior.
- **The wave2 `environments` and `segments` streams do not match the real API's documented
  response shapes exactly, and this bundle does not correct them.** This is a pre-existing wave2
  discrepancy discovered during this Pass B research pass, not introduced by it — both streams are
  left untouched at their already-migrated parity shape per the meta-rule against altering
  already-migrated behavior:
  - `environments`' real response (`GET /internal/api/v2/environments/ws/{workspace_id}`) is a BARE
    ARRAY (`[{id, name}, ...]`), not the `{objects: [...], offset, limit, totalCount}` envelope
    this bundle's `records.path: "objects"` + `offset_limit` pagination expects. Legacy's own
    `streamEndpoints["environments"]` assumed the SAME uniform `objects`-envelope shape as every
    other legacy stream, so this bundle faithfully reproduces legacy's (incorrect, per this pass's
    live-API research) assumption rather than silently correcting it outside this task's scope.
  - `segments`' real per-item object identifies a segment by `name`, not `id` (segment objects in
    the real API have no `id` field at all) — this bundle's `schemas/segments.json` declares
    `x-primary-key: ["id"]`, matching legacy's assumption, not the real API's actual identifying
    field.
  - The NEW `groups`/`traffic_types`/`users` streams added in this pass were authored directly
    against the live spec and do not inherit either discrepancy.
- **`groups`' real cursor-pagination parameter name is unconfirmed.** The response envelope
  documents `nextMarker`/`previousMarker`, but the endpoint's OpenAPI parameter list names only
  `limit` — no `after`/`before`/`marker` request parameter is documented for this specific
  endpoint (unlike `users`, which explicitly documents `before`/`after`). Declaring an unconfirmed
  parameter name risks silently sending a no-op query param and reading only the first page
  forever with no error; `pagination: {"type": "none"}` is the honest representation of what could
  be confirmed from this pass's research, not a corner cut.
- Full endpoint-by-endpoint accounting, including every excluded write and the reasoning for each
  (change requests, attributes, large segments, flag sets, rule-based segments, identities, API
  keys, and account-structure CRUD), is in `api_surface.json`.
