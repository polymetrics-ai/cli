# Overview

Asana is a pure declarative-HTTP Tier-1 migration of `internal/connectors/asana`, Pass-B
full-surface expanded (legacy is connsdk-HTTP-based: a single Bearer-authenticated requester, plain
JSON responses, no protocol-native SDK, no hook-worthy auth/stream logic). It reads Asana
workspaces, projects, tasks, sections, tags, stories (comments), users, teams, custom field
definitions, project statuses, and team/workspace memberships through the Asana v1 REST API
(`https://app.asana.com/api/1.0/...`), and writes task/project/section/tag create/update/delete
plus task comments. This bundle originally targeted read-only capability parity with legacy's 3
streams (`workspaces`/`projects`/`tasks`); Pass B (`docs/migration/conventions.md` §8,
`api_surface.json`) researched Asana's complete official OpenAPI spec
(https://raw.githubusercontent.com/Asana/openapi/master/defs/asana_oas.yaml, 250 documented
operations counting `/users/me`) and expanded to 13 streams + 13 writes covering the core PM
domain; every write here is a genuinely new capability beyond legacy parity (legacy's own `Write`
always returned `connectors.ErrUnsupportedOperation`), added per the Pass B mandate to implement
every dialect-expressible mutation, not a migrated legacy behavior.

## Auth setup

Provide an Asana personal access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`asana.go:154`). `base_url` defaults to
`https://app.asana.com/api/1.0` and may be overridden for tests/proxies.

## Streams notes

All 3 streams (`workspaces`, `projects`, `tasks`) share the same shape: `GET` against the Asana
list endpoint, records at `data`, primary key `["gid"]`. Each stream sends its own
`opt_fields` value (matching legacy's per-endpoint `optFields`) and a static `limit=100`
(legacy's `asanaDefaultPageSize`/`asanaMaxPageSize`, both 100). The `next_url` paginator has no
config-driven page-size override mechanism, so the bundle keeps legacy's default request size and
does not declare a dead `page_size` config property.

`projects` and `tasks` optionally scope to a workspace via the `workspace` query param
(`{{ config.workspace_id }}`, `omit_when_absent: true` — left off entirely when unset, matching
legacy's `endpoint.resource != "workspaces"` guard: the `workspaces` stream never sends a
`workspace` param on itself). `tasks` additionally scopes to `project`/`assignee` via
`{{ config.project_id }}`/`{{ config.assignee }}`, both `omit_when_absent: true`, matching
legacy's `asanaQuery`'s `endpoint.resource == "tasks"` guards exactly.

Pagination follows Asana's `next_page.uri` convention (`pagination.type: next_url`,
`next_url_path: "next_page.uri"`) exactly like legacy's `harvest` loop, which follows
`resp.next_page.uri` verbatim as the next request path (with no query) until it is empty
(`asana.go:119-127`). None of the original 3 streams declare an `incremental` block — legacy's
Asana v1 endpoints expose no server-side "modified since" filter this connector's catalog wires up
(no `CursorFields` declared anywhere in legacy's `asanaStreams()`), so every read is full refresh,
matching legacy exactly; the same is true of every Pass B stream added below (Asana's list
endpoints generally have no `modified_since`-style filter parameter at all).

**Pass B additions** — 9 new streams:

- `users` — `GET /users`, optionally scoped by `workspace_id`, primary key `["gid"]`.
- `teams` — `GET /workspaces/{workspace_id}/teams` — **requires `workspace_id`** (Asana has no
  cross-workspace list-all-teams endpoint; teams are always listed within a workspace path
  segment), primary key `["gid"]`.
- `tags` — `GET /tags`, optionally scoped by `workspace_id`, primary key `["gid"]`.
- `sections` — `GET /projects/{project_gid}/sections`, a `fan_out` over every project gid
  discovered from a preliminary fully-paginated `GET /projects` listing
  (`fan_out.ids_from.request`), stamping the source project onto each record's `project_gid`
  field; primary key `["gid"]`.
- `stories` — `GET /tasks/{task_gid}/stories` (task comments/activity), a `fan_out` over every task
  gid discovered from a preliminary fully-paginated `GET /tasks` listing, stamping the source task
  onto `task_gid`; primary key `["gid"]`. Note: fanning out over EVERY task in the account (rather
  than a caller-scoped subset) can be a large number of requests for a big Asana instance — no
  different in kind from `sections`' fan-out over every project, but worth calling out since a
  typical account has many more tasks than projects.
- `custom_fields` — `GET /workspaces/{workspace_id}/custom_fields` — **requires `workspace_id`**
  (same reasoning as `teams`), primary key `["gid"]`.
- `project_statuses` — `GET /projects/{project_gid}/project_statuses`, a `fan_out` over every
  project gid (same preliminary-listing mechanism as `sections`), stamping `project_gid`; primary
  key `["gid"]`.
- `team_memberships` — `GET /team_memberships`, optionally scoped by `team_id`/`workspace_id`
  (Asana's own endpoint requires EITHER `team` alone OR `workspace`+`user` together — a
  cross-field constraint this dialect's per-param `omit_when_absent` cannot enforce; an operator
  who sets only `workspace_id` without a per-user scope will see Asana's own 400 at read time, an
  honest surfacing of the API's own requirement rather than a silent partial read), primary key
  `["gid"]`.
- `workspace_memberships` — `GET /workspaces/{workspace_id}/workspace_memberships` — **requires
  `workspace_id`**, primary key `["gid"]`.

## Write actions & risks

`capabilities.write` is now `true` (Pass B; legacy always returned
`connectors.ErrUnsupportedOperation`). Every Asana write body follows the API's `{"data": {...}}`
envelope convention — the write dialect's default JSON body (every record field except
`path_fields`) sends the record verbatim, so a caller-supplied record's own top-level shape must
already be `{"<path_fields...>": ..., "data": {...fields...}}`; this is the same established
pattern already used elsewhere in this repo for envelope-wrapped write APIs (see e.g. teamtailor's
`create_job`), not an Asana-specific special case.

- `create_task` / `update_task` / `delete_task` (`POST /tasks`, `PUT`/`DELETE /tasks/{gid}`).
  `delete_task` declares `missing_ok_status: [404]` (idempotent delete) and `confirm: "destructive"`.
- `create_project` / `update_project` / `delete_project` (`POST /projects`,
  `PUT`/`DELETE /projects/{gid}`). Same idempotent-delete/destructive-confirm shape as tasks.
- `create_section` / `update_section` / `delete_section` (`POST /projects/{project_gid}/sections`,
  `PUT`/`DELETE /sections/{gid}`). `create_section`'s path carries `project_gid` (a
  `path_fields`-excluded record field, not part of the section's own `data` body) since Asana
  creates a section AS a sub-resource of a project.
- `create_tag` / `update_tag` / `delete_tag` (`POST /tags`, `PUT`/`DELETE /tags/{gid}`).
- `add_comment` (`POST /tasks/{task_gid}/stories`) — posts a task comment (a `story` of
  `resource_subtype: comment_added`); `task_gid` is a `path_fields`-excluded record field, the
  comment text itself lives in `data.text`.

Every write's `risk` field states whether it is destructive/irreversible (the 4 delete actions,
each also carrying `confirm: "destructive"`) or a lower-risk additive mutation (create/update
actions, no `confirm` set) — approval is required for all of them regardless.

**Not implemented — sub-resource/relationship-mutation actions already equivalent to a covered
field on `update_task`/`update_project`** (documented in `api_surface.json` as `duplicate_of`, not
silently dropped): `addTag`/`removeTag`, `addProject`/`removeProject`, `setParent`,
`addCustomFieldSetting`/`removeCustomFieldSetting`, and the section-reordering/task-move actions —
each is a narrower single-field mutation the already-covered `update_task`/`update_project`
actions' own record body already expresses (e.g. setting `data.tags`/`data.parent` directly),
so implementing them as separate actions would be redundant surface, not new capability.

## Known limits

- **`page_size` is not runtime-configurable.** Same limitation as bitly's `next_url`-paginated
  `bitlinks` stream (see `docs/migration/conventions.md` and bitly's own `docs.md`): the engine's
  `next_url` paginator has no analogous page-size knob, so `page_size`/`max_pages` (legacy's
  config-driven overrides, `asanaPageSize`/`asanaMaxPages`) are not wired into the live request —
  Asana's own default (`limit=100`) is sent unconditionally, and `spec.json` does not declare a
  dead `page_size` property.
- **Fixtures ship one page per stream (sanctioned `next_url` exception, conventions.md §4).** A
  `next_url` stream's next-page URL is the replay server's own address, unknown ahead of time to a
  static fixture file — a genuine harness limitation, not a fixture-authoring shortcut. All 3
  streams paginate identically in legacy; `pagination_terminates` exercises whichever stream
  `conformance` selects as its first eligible stream from this single-page shape (an empty
  `next_page` value stops pagination immediately, proving termination on an already-short page).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) synthesizes `gid`/`name`/`resource_type`/`created_at`/
  `modified_at`/`completed` values directly in Go and never talks to a real Asana API
  (`asana.go:132-143`); this bundle's own `conformance`/fixture-replay harness
  (`internal/connectors/conformance`) provides the equivalent credential-free test affordance, so
  no fixture-mode config branch is modeled here — matching SPEC's instruction to target the live
  record shape only.
- **`sections`/`stories`/`project_statuses` fan out over EVERY project/task in the account, not a
  caller-scoped subset.** Asana's API has no "list sections/stories/statuses across all projects
  in one call" endpoint — each is genuinely a per-parent sub-resource. `fan_out.ids_from.request`
  runs a full, separately-paginated `GET /projects` (or `/tasks`) listing first (ignoring
  `workspace_id ` scoping, since the fan-out id-listing request is independent of the parent
  stream's own query wiring) and then issues one sub-request per discovered id — this is the
  correct, honest expression of Asana's actual resource shape, not an approximation, but it does
  mean these 3 streams' total request count scales with the size of the whole account, not just
  the records ultimately of interest.
- **`api_surface.json` scopes out several entire adjacent Asana product sub-domains as future
  capability-expansion passes, not migration gaps**: Goals (12 endpoints, OKR/objective tracking),
  Portfolios (13 endpoints, project-of-projects rollups), Time Tracking/Timesheet
  Approval/Budgets/Rates/Allocations (resource-planning financial add-ons, ~25 endpoints total),
  Webhooks (push-delivery, architecturally orthogonal to this connector's pull-based model),
  Project/Task Templates, and admin-only configuration (Roles, Audit Log, Organization Exports,
  Access Requests). Each is coherent and sizable enough to warrant its own dedicated pass rather
  than a handful of endpoints folded in here; see `api_surface.json`'s per-endpoint `reason` fields
  for the specific justification behind every individual exclusion.
