# Overview

Todoist reads projects, sections, tasks, comments, labels, and project collaborators, and writes
project/section/task/comment/label create, update, and delete actions (plus task close/reopen),
through the Todoist REST API v2 (`https://api.todoist.com/rest/v2/...`). This bundle originated as
a capability-parity migration from `internal/connectors/todoist` (the hand-written connector it
replaces; the legacy package stays registered and unchanged until wave6's registry flip) and was
then expanded to Todoist's full documented REST v2 surface (Pass B): 28 operations across 13 paths
per a community-maintained OpenAPI 3.1 document for this exact API (matching this bundle's
`base_url` exactly), covering every GET/write shape Todoist's REST v2 API exposes.

## Auth setup

Provide a Todoist personal API token via the `token` secret; it is sent as a Bearer token
(`Authorization: Bearer <token>`) and is never logged, matching legacy's `connsdk.Bearer(token)`
(`todoist.go:118`). Legacy's `firstSecret(cfg, "token", "bearer_token")` accepts either the
`token` OR `bearer_token` secret, `token` taking precedence when both are configured
(`todoist.go:114,178-190`) — this bundle reproduces that exact precedence with a two-candidate
`base.auth` list (`when`-gated on each secret's presence in declaration order), matching
conventions.md §3's dual-auth-ordering rule. `base_url` defaults to
`https://api.todoist.com/rest/v2` and may be overridden for tests/proxies.

## Streams notes

`projects`, `sections`, and `tasks` are simple, non-paginated list endpoints (`GET /projects`,
`/sections`, `/tasks`); Todoist's REST v2 API returns the full list in one response for personal
task lists (legacy never paginates any of these three, `todoist.go:126-131`), so no `pagination`
block is declared and records are extracted from the response root (`records.path: "."`), matching
legacy's `connsdk.RecordsAt(resp.Body, ".")` (`todoist.go:90`). None of the four streams expose an
incremental cursor field in legacy, so all four are always full-refresh reads.

All four streams declare `"projection": "passthrough"`. Legacy's `Read` emits the raw API record
verbatim (`emit(connectors.Record(rec))`, `todoist.go:94-99`) with no field-building/filtering —
`streamSpecs[stream].fields` is consumed only by `Catalog` (`todoist.go:133-141`), never by `Read`.
Every real Todoist field beyond each stream's narrow catalog schema (e.g. `order`,
`comment_count`, `is_archived`, `url`, `parent_id` on `projects`; `labels`, `priority`,
`assignee_id`, `duration` on `tasks`; `attachment` on `comments`) survives to the emitted record
exactly as legacy would emit it. Declaring the default `"schema"` projection mode here would
silently narrow every emitted record to the catalog schema's properties — a silent, undocumented
parity deviation from legacy's verbatim passthrough — so `passthrough` is required, matching
conventions.md's projection rule (§3) and the passthrough-decision rule (§5 defect class 1:
legacy's raw `emit(record)` with no `mapRecord` field-building is the mechanical signal to use
`passthrough`).

`comments` optionally scopes to one project or task via `project_id`/`task_id` query parameters,
sent only when configured (legacy: `todoist.go:78-85`, `strings.TrimSpace` then conditionally
`q.Set`). This bundle reproduces the identical optional behavior via the opt-in `omit_when_absent`
query dialect (conventions.md §3) — both params are left off the request entirely when their
config keys are unset, exactly like legacy's conditional `q.Set` calls.

**`labels`** (`GET /labels`) is a new Pass B stream for Todoist's personal-label resource: a small,
unpaginated list, same shape as the other 4 original streams.

**`collaborators`** (`GET /projects/{project_id}/collaborators`) is a new Pass B stream using the
engine's `fan_out` dialect (conventions.md §3): the id-listing preliminary request re-reads the
`projects` list endpoint (`ids_from.request`, `records_path: ""` since Todoist's `/projects`
response is a bare top-level array), then issues one collaborators request per resolved project id
(`into.path_var: "project_id"`), stamping `project_id` onto every emitted collaborator record
(`stamp_field`) so a caller can tell which project each collaborator belongs to. Primary key is
`["id", "project_id"]` (composite) since the same person can collaborate on more than one project.

## Write actions & risks

- **`create_project`/`update_project`/`delete_project`** (`POST /projects`, `POST
  /projects/{id}`, `DELETE /projects/{id}`) create, rename/re-style, or permanently delete a
  project. Deleting a project also deletes every section/task/comment inside it — irreversible.
- **`create_section`/`update_section`/`delete_section`** (`POST /sections`, `POST
  /sections/{id}`, `DELETE /sections/{id}`) create, rename, or permanently delete a section.
  Deleting a section also deletes every task inside it — irreversible.
- **`create_task`/`update_task`/`delete_task`** (`POST /tasks`, `POST /tasks/{id}`, `DELETE
  /tasks/{id}`) create, mutate, or permanently delete a task. Delete is irreversible.
- **`close_task`/`reopen_task`** (`POST /tasks/{id}/close`, `POST /tasks/{id}/reopen`) mark a task
  completed (a recurring task instead advances to its next occurrence) or return a completed task
  to the active list.
- **`create_comment`/`update_comment`/`delete_comment`** (`POST /comments`, `POST
  /comments/{id}`, `DELETE /comments/{id}`) post, edit, or permanently delete a comment on a task
  or project.
- **`create_label`/`update_label`/`delete_label`** (`POST /labels`, `POST /labels/{id}`, `DELETE
  /labels/{id}`) create, rename/re-style, or permanently delete a personal label. Deleting or
  renaming a label affects every task already tagged with it.

Every write action uses `body_type: json` — Todoist's real REST v2 API accepts a flat JSON request
body for every one of these mutations (confirmed by the API's own documented example request
bodies), so no engine dialect gap blocks any write here (contrast e.g. Timely, whose API requires a
nested envelope body the engine cannot express).

Every `delete_*` action declares `missing_ok_status: [404]` (idempotent delete): re-deleting an
already-deleted resource counts as written, not failed, matching the engine's standard idempotent-
delete semantics (conventions.md §3).

## Known limits

- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a
  static `fixture: true` marker and synthesizes `content`/`name`/`id` fields onto two fixture
  records per stream (`todoist.go:149-160`). None of these are part of the LIVE record shape; this
  bundle's schemas and fixtures target the live path only. The engine's own conformance/
  fixture-replay harness (`internal/connectors/conformance`) provides the credential-free test
  affordance this bundle needs, so no fixture-mode equivalent is needed here.
- **No pagination is modeled for any read stream**, matching legacy's original 4 streams — none of
  the Todoist REST v2 list endpoints this bundle calls are paginated (Todoist's REST v2 API returns
  the full list in one response for every resource here), so no `pagination` block is declared
  anywhere and every stream ships single-page fixtures.
- **No update-collaborators/remove-collaborator write.** Todoist's REST v2 API exposes no
  mutation endpoint for project collaborators at all (collaborator management is Sync-API/UI-only);
  `collaborators` is read-only, matching the real documented surface.
