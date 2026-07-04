# Overview

Teamwork is a declarative-HTTP migration of `internal/connectors/teamwork` (the hand-written
legacy connector this bundle migrates; the legacy package stays registered and unchanged until
wave6's registry flip). It reads Teamwork projects, people, companies, tags, time entries,
tasklists, milestones, and tasks, and writes approved project-management mutations through the
Teamwork V1 REST API (`https://api.teamwork.com/<resource>.json`).

This is a Pass B full-surface expansion: the wave2 migration covered only the single legacy-parity
`projects` stream; every other stream and every write action here is new coverage researched
against Teamwork's published Postman collection and API reference (`api_surface.json`).

## Auth setup

Provide a Teamwork username (email) via the `username` config value and an API token via the
`password` secret; both are required. They are sent as HTTP Basic auth
(`Authorization: Basic base64(username:password)`), matching legacy's
`connsdk.Basic(username, password)` (`teamwork.go:110`). `password` is never logged. `base_url`
defaults to `https://api.teamwork.com` and may be overridden for tests/proxies.

Legacy's own `Check` is a pure config-presence validation with no network call
(`teamwork.go:34-48`); this bundle's declarative `check` (`GET /projects.json`) performs a real,
richer network round-trip instead — an unavoidable, structural consequence of the engine's
declarative `check` dialect, not a per-connector authoring choice.

## Streams notes

Pagination follows a 1-based page-number convention (`pagination.type: page_number`,
`page_param: page`, `size_param: pageSize`, `page_size: 100`), matching legacy's hand-rolled loop
and shared by every stream via `base.pagination`.

**Global streams**:
- `projects` (legacy-parity): `GET /projects.json`, records at `projects`, primary key `["id"]`.
  `created_at` computed-field-renamed from the raw `created-on`.
- `people`: `GET /people.json`, records at `people`, primary key `["id"]`. `first_name`/`last_name`
  computed-field-renamed from the raw `first-name`/`last-name`; `email-address`/`user-name`/
  `company-id` need no rename (schema declares the hyphenated property names directly, matching
  Teamwork's own wire field names verbatim).
- `companies`: `GET /companies.json`, records at `companies`, primary key `["id"]`. No renames
  needed.
- `tags`: `GET /tags.json`, records at `tags`, primary key `["id"]`. No renames needed.
- `time_entries`: `GET /time_entries.json`, records at `time-entries`, primary key `["id"]`,
  `x-cursor-field: created_at`. `created_at`/`person_id`/`project_id`/`todo_item_id`
  computed-field-renamed from the raw hyphenated `created-at`/`person-id`/`project-id`/
  `todo-item-id`.
- `tasks`: `GET /tasks.json` (Teamwork's "all tasks across all projects" endpoint — a genuinely
  global, non-project-scoped list), records at `todo-items`, primary key `["id"]`,
  `x-cursor-field: created_at`. `created_at` computed-field-renamed from the raw `created-on`.
  Deliberately NOT modeled as a per-tasklist fan-out (`GET /tasklists/{id}/tasks.json`) — the
  global endpoint already returns the identical record set without needing a fan-out at all.

**Project-scoped streams (fanned out via the engine's `fan_out` dialect)** — each fans out over
every project id discovered via a preliminary, fully-paginated `GET /projects.json` request
(`fan_out.ids_from.request`), stamping the fanned-out project id onto every emitted record's
`project_id` field (`fan_out.stamp_field`, ALWAYS a string per the engine's stamp contract):
- `tasklists`: `GET /projects/{project_id}/tasklists.json`, records at `todo-lists`, primary key
  `["id"]`.
- `milestones`: `GET /projects/{project_id}/milestones.json`, records at `milestones`, primary key
  `["id"]`, `x-cursor-field: created_at`. `created_at` computed-field-renamed from the raw
  `created-on`.

## Write actions & risks

Every Teamwork V1 write body requires a resource-name WRAPPER key (`{"project": {...}}`,
`{"todo-list": {...}}`, `{"todo-item": {...}}`, `{"milestone": {...}}`, `{"company": {...}}`,
`{"time-entry": {...}}`) — this is modeled by declaring the wrapper key itself as a REQUIRED nested
object field on each action's `record_schema`. The engine's `body_type: json` write path sends
record fields verbatim as the top-level JSON body (`buildJSONBody`, minus `path_fields`); since the
record's one non-path-field top-level key IS the wrapper name, the body produced is byte-for-byte
the wrapped shape the legacy V1 API requires — no engine change was needed, this is a genuinely
sanctioned Tier-1 pattern (bitly's `create_qr_code.destination` already proves nested-object record
fields pass through the write body untouched).

- `create_project` (create, `POST /projects.json`, body `{"project": {name, description, ...}}`):
  creates a new project; low-risk, no approval required.
- `update_project` (update, `PUT /projects/{id}.json`, body `{"project": {...}}`): mutates an
  existing project's name/description.
- `create_tasklist` (create, `POST /projects/{project_id}/tasklists.json`, body
  `{"todo-list": {name, description}}`): creates a new tasklist under a project; low-risk, no
  approval required.
- `create_task` (create, `POST /tasklists/{tasklist_id}/tasks.json`, body
  `{"todo-item": {content, description, priority}}`): creates a new task in a tasklist; low-risk,
  no approval required.
- `update_task` (update, `PUT /tasks/{id}.json`, body `{"todo-item": {...}}`): mutates an existing
  task's content/description/priority.
- `complete_task` (update, `PUT /tasks/{id}/complete.json`, no body): marks a task complete; a
  visible, notifiable state change for every task follower.
- `create_milestone` (create, `POST /projects/{project_id}/milestones.json`, body
  `{"milestone": {title, description, deadline}}`): creates a new milestone under a project;
  low-risk, no approval required.
- `create_company` (create, `POST /companies.json`, body `{"company": {name, ...}}`): creates a new
  company record; low-risk, no approval required.
- `create_time_entry` (create, `POST /projects/{project_id}/time_entries.json`, body
  `{"time-entry": {description, date, hours, minutes, person-id, ...}}`): logs a new time entry
  against a project; contributes to billable-hours totals and any linked invoice.

`metadata.json` now declares `capabilities.write: true`.

## Known limits

- User-account provisioning/administration (`POST /people.json`, role/permission management,
  billing/invoicing) is excluded as `requires_elevated_scope` — these require Teamwork
  administrator-tier credentials beyond ordinary project read/write access.
- Destructive deletes beyond the covered writes (project/tasklist/task/milestone/company delete,
  and every sub-object delete) are excluded as `destructive_admin` — see `api_surface.json`.
- Binary/multipart file-attachment endpoints are excluded (`binary_payload`) — the engine's
  `body_type` dialect (`json`/`form`/`none`) has no multipart shape.
- Kanban boards, calendar events, notebooks, message boards/comments, bookmark links, and
  webhook-subscription management are excluded as `out_of_scope` — each is a separate
  sub-application object domain distinct from core project/task/time data; Pass B
  breadth-vs-cost triage.
- **`page_size` is not runtime-configurable** (carried over from the wave2 golden): legacy exposes
  a config-driven `page_size` override (`teamwork.go:124-130`), but the engine's `page_number`
  paginator constructor reads `PaginationSpec.PageSize` as a static bundle-level integer, not a
  config-templated field. This bundle hardcodes `page_size: 100`, legacy's own default. `page_size`
  is not declared in `spec.json` at all (F6, REVIEW.md).
- All fixtures (`fixtures/streams/**`, `fixtures/writes/**`, `fixtures/check.json`) represent
  Teamwork's real wire shape, including hyphenated field names (`created-on`, `first-name`,
  `person-id`, etc.) before their `computed_fields` renames, and the resource-name-wrapped write
  request bodies described above.
- `project_id` is stamped as a STRING on `tasklists`/`milestones` (the engine's `stamp_field`
  contract), matching every other id field in this bundle's schemas (already string-typed to match
  Teamwork's own wire convention, so no type mismatch arises here unlike a numeric-id API).
