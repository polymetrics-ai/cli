# Overview

TickTick reads TickTick projects and project-scoped tasks, and writes task create/complete/delete
actions, through the TickTick Open API (`https://api.ticktick.com/open/v1/...`). This bundle
originated as a capability-parity migration from `internal/connectors/ticktick` (the hand-written
connector it migrates; the legacy package stays registered and unchanged until wave6's registry
flip) and was then expanded to TickTick's full documented Open API v1 surface (Pass B). TickTick's
own published OpenAPI 3.0 document for this exact product (servers: `https://ticktick.com`, base
path `/open/v1` — matching this bundle's `base_url` exactly) declares exactly 6 operations across
4 paths; every one is accounted for by a stream, a write action, or a documented `duplicate_of`
exclusion in `api_surface.json`.

## Auth setup

TickTick issues an OAuth access token. Legacy accepts it under any of 3 aliased secret keys, in a
first-non-empty-wins fallback chain (`firstSecret(cfg, "bearer_token", "client_access_token",
"access_token")`, `ticktick.go:109,178-190`): `bearer_token` first, `client_access_token` second,
`access_token` third. This bundle reproduces the exact same precedence as 3 ordered `bearer`
auth candidates, each gated by a `when` clause on its own secret's truthiness — `base.auth`'s
first-match-wins `selectAuth` evaluation (conventions.md §3, "Dual-auth ordering is
load-bearing") reproduces legacy's fallback chain exactly: `bearer_token` wins if set (regardless
of the other two), otherwise `client_access_token` wins if set, otherwise `access_token` is used.
Whichever one resolves is sent as `Authorization: Bearer <token>`; none is ever logged.
`base_url` defaults to `https://api.ticktick.com/open/v1`, matching legacy's `defaultBaseURL`
fallback.

## Streams notes

`projects` (`GET /project`, records at the JSON response root `.`) has no pagination — TickTick's
project-list endpoint returns every project in one call, matching legacy's single unpaginated
`Do` request (`ticktick.go:81`). `tasks` (`GET /project/{project_id}/data`, records at the
`tasks` envelope key) requires `project_id` (the TickTick project id to scope reads to), matching
legacy's own hard requirement (`ticktick.go:125-129`: "ticktick tasks stream requires config
project_id") — the engine's path-interpolation hard-errors identically when `project_id` is
unset, with no special-casing needed. Both streams declare primary key `["id"]`.

Both streams declare `"projection": "passthrough"` (conventions.md §8 rule 1): legacy's `Read`
emitted records verbatim with no field-building step, and the schemas below now declare the full
documented field set of TickTick's own `Project`/`Task` OpenAPI schemas (not just legacy's
narrower catalog list) — `projects` adds `closed`/`groupId`/`viewMode`/`permission`/`kind`;
`tasks` adds `isAllDay`/`completedTime`/`desc`/`dueDate`/`priority`/`reminders`/`repeatFlag`/
`sortOrder`/`startDate`/`timeZone`. Passthrough mode means any other real field TickTick returns
(beyond even this fuller declared set) still survives unfiltered. Two documented boolean-shaped
fields (`projects.closed`, `tasks.isAllDay`) are typed `string` (not `boolean`) because TickTick's
own OpenAPI document declares them as string-valued `"true"`/`"false"`, not JSON booleans — this is
the real documented wire shape, not a widening workaround. `tasks.priority`/`sortOrder`/`status`
and `projects.sortOrder` are typed `integer`, matching the OpenAPI document's `int32`/`int64`
fields exactly (no string-ification).

The single-project detail endpoint (`GET /project/{projectId}`) is not modeled as its own stream:
its response is a strict subset of the same `Project` schema returned by the `projects` list
stream, so it would duplicate already-covered data (`api_surface.json`, `duplicate_of`).

## Write actions & risks

- **`create_task`** (`POST /task`) creates a new task in the caller's TickTick account — in the
  given `projectId`, or the default Inbox if omitted. Low-risk external mutation; no approval
  required.
- **`complete_task`** (`POST /project/{projectId}/task/{id}/complete`) marks an existing task
  completed. A completed task is removed from active task lists/reminders for every collaborator
  on the project.
- **`delete_task`** (`DELETE /project/{projectId}/task/{id}`) permanently removes a task;
  irreversible via the API (idempotent: a 404 on an already-deleted task counts as written, not
  failed).

No project-mutation write (create/update/delete project) or task-update write is implemented:
TickTick's own published Open API document declares no such operations at all (see
`api_surface.json`'s scope note) — only unofficial, uncorroborated community references describe
them, and this migration does not fabricate a write shape against a real user's TickTick account
without a corroborated source.

## Known limits

- **`tasks` has no incremental/cursor support**, matching legacy exactly — TickTick's
  `project/{id}/data` endpoint returns the full project snapshot (tasks + columns) on every call
  with no server-side filter parameter; this bundle declares no `incremental` block for either
  stream, an honest 1:1 match of the API's own behavior, not a scope narrowing.
- **Only one `project_id` can be read per sync.** TickTick's Open API has no "all tasks across all
  projects" endpoint; a caller wanting every project's tasks must run one sync per `project_id`.
- **No task-update write.** TickTick's official Open API document exposes create/complete/delete
  for tasks but no update-in-place operation; a caller wanting to change an existing task's fields
  must currently do so outside this connector (e.g. in the TickTick app/website).
- **No project-mutation writes** (create/update/delete). See `api_surface.json`'s `out_of_scope`
  exclusions — these operations are absent from TickTick's own published Open API document.
- **Kanban `columns` are not a standalone stream.** Column data for a project is already returned
  inline on the `tasks` stream's `/project/{id}/data` response (`ProjectData.columns`); the Open
  API exposes no endpoint to list columns independently of a project.
- **Tags, habits, and focus/pomodoro stats are out of scope.** These are exposed only by TickTick's
  unofficial internal web API (V2), never by the public Open API (V1) this bundle authenticates
  against (`api_surface.json`, `requires_elevated_scope`).
