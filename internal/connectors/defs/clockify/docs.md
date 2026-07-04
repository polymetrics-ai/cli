# Overview

Clockify is a wave2 fan-out declarative-HTTP migration, expanded to the full documented Clockify
API v1 surface in Pass B. It reads Clockify workspaces, clients, projects, tags, users, the current
user, custom fields, user groups, holidays, expense categories, time-off policies, per-project
tasks, and per-user time entries; it writes clients, projects, tags, and tasks (create/update/
delete) through the Clockify REST API v1 (`https://api.clockify.me/api/v1/...`). This bundle
targets capability parity with `internal/connectors/clockify` (the hand-written connector it
migrates) for its original 5 streams; the legacy package stays registered and unchanged until
wave6's registry flip. The Pass B streams/writes (`current_user`, `custom_fields`, `user_groups`,
`holidays`, `expense_categories`, `time_off_policies`, `tasks`, `time_entries`, and every
`writes.json` action) are new coverage beyond legacy's own scope — legacy never implemented them —
so there is no parity constraint on their record shape; schemas are derived directly from
Clockify's published OpenAPI spec (`https://docs.clockify.me/openapi.json`).

## Auth setup

Provide a Clockify API key via the `api_key` secret; it is sent as the `X-Api-Key` header
(`api_key_header` auth mode), matching legacy's `connsdk.APIKeyHeader("X-Api-Key", secret, "")`
(`clockify.go:202`). It is never logged. `base_url` defaults to `https://api.clockify.me/api` and
may be overridden for tests/proxies (legacy's own `clockifyBaseURL` validates scheme+host the same
way; the engine's base-URL resolution has no equivalent runtime validation, but every
conformance fixture only ever points at an httptest server, so this is not exercised differently
on either side).

## Streams notes

`workspaces` reads the top-level, unscoped `/v1/workspaces` endpoint. `clients`, `projects`,
`tags`, and `users` are scoped under `/v1/workspaces/{{ config.workspace_id }}/<resource>` — an
absent `workspace_id` hard-errors on both sides (legacy: `"clockify connector requires config
workspace_id for this stream"`; engine: an unresolved `config.workspace_id` path-template key —
same failure classification, different literal text, per conventions.md §5's precedent for
config-validation parity).

All five list endpoints return a bare top-level JSON array (no envelope), so every stream declares
`records.path: "."` (the dotted-path root selector). Pagination is 1-indexed page-number
(`pagination.type: page_number`, `page_param: page`, `size_param: page-size`) with `page_size: 50`
matching legacy's `clockifyDefaultPageSize`; a page returning fewer than 50 records is the last
page, matching `connsdk.PageNumberPaginator`'s exact stop rule legacy itself uses
(`clockify.go:136-141`).

None of Clockify's five list endpoints expose an incremental cursor field (legacy's own
`clockifyStreams` comment: "Clockify list endpoints do not expose an updated-at cursor field, so
these streams are full-refresh (no cursor)") — this bundle declares no `incremental` block for any
stream, matching legacy exactly. None of the Pass B streams below expose one either (Clockify's own
API has no `updatedAt` query filter on any of these list endpoints).

**Pass B streams**: `current_user` (`GET /v1/user`, single-object, `pagination.type: none`) and
`custom_fields`/`user_groups` (`pagination.type: none` — Clockify does not paginate these two
endpoints at all) read directly. `holidays`, `expense_categories` (records at the `categories` key,
not root), and `time_off_policies` are ordinary workspace-scoped list streams using the shared base
pagination. `tasks` and `time_entries` are genuine sub-resource fan-outs (`stream.fan_out`,
conventions.md §3): `tasks` lists every project id from `GET
/v1/workspaces/{{ config.workspace_id }}/projects` (reusing the `projects` stream's own request
shape as the id source, `id_field: id`), then reads `GET .../projects/{{ fanout.id }}/tasks` once
per project, stamping the project id onto `projectId` (a no-op stamp — Clockify's own task response
already carries `projectId` — included for defense-in-depth per the dialect's stated pattern).
`time_entries` mirrors this exactly over `users`: lists every user id from `GET
/v1/workspaces/{{ config.workspace_id }}/users`, then reads `GET
.../user/{{ fanout.id }}/time-entries` once per user, stamping `userId`. Neither fan-out passes a
`start`/`end` date-range query filter (both optional per Clockify's docs) — every time entry ever
recorded for that user is read on every sync; a future increment could add a `stream.Query`
opt-in-optional `start`/`end` pair templated against `config.*` if operators need date-bounded
syncs, but no such config property is wired today (declaring one with no template consuming it
would itself be dead config, per conventions.md's `default_type_mismatch`-adjacent "don't declare
what nothing wires" rule).

## Write actions & risks

`create_client`/`update_client`/`delete_client`, `create_project`/`update_project`/`delete_project`,
`create_tag`/`update_tag`/`delete_tag`, and `create_task`/`update_task`/`delete_task` are new Pass B
writes (legacy never implemented any Clockify write path — legacy's `Write` always returns
`connectors.ErrUnsupportedOperation`, and this bundle now supersedes that for these 4 resources).
Every action is a live external mutation against the real Clockify workspace; `risk` on each action
requires approval. `update_task`/`delete_task` both require `path_fields: ["projectId", "id"]`
since Clockify's task URLs are nested under their owning project
(`/projects/{projectId}/tasks/{taskId}`) — a record missing either field fails write validation
before any request is issued. `capabilities.write` is now `true`.

Time entries, custom fields, user groups, holidays, expense categories, and time-off policies have
no write action in this bundle: time-entry writes accept a polymorphic body (plain/lump-sum/
lump-sum-service variants select different required fields on the SAME endpoint) with no single
dialect-expressible `record_schema` shape decided yet (see `api_surface.json`'s exclusion reasons);
the remaining resources (custom fields, user groups, holidays, expense categories, time-off
policies) are workspace-configuration objects with no demonstrated write demand — see
`api_surface.json` for the itemized reason on every excluded endpoint.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`clockifyPageSize`/`clockifyMaxPages`, `clockify.go:252-280`). The engine's
  `page_number` paginator's `PageSize`/`MaxPages` fields are plain JSON values in `streams.json`,
  not templated against `config.*` — there is no mechanism in this dialect to wire a runtime
  config value into either field. This bundle ships legacy's own default (`page_size: 50`,
  `max_pages` unbounded) as a static value; an operator can no longer override the page size or
  cap request count per sync. This mirrors the identical, already-accepted limitation documented
  for `bitly`'s `next_url` paginator and other wave1 goldens (conventions.md's fixture-rules
  section references this pattern).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance,
  `clockify.go:154-185`) stamps a broader, cross-stream synthetic record shape (e.g. every fixture
  record carries `workspaceId`, `clientId`, `duration`, etc. regardless of stream) that does not
  match any single stream's real live-API record shape. This bundle's schemas and fixtures target
  the live per-stream record shape only; the engine's own conformance/fixture-replay harness
  provides the credential-free test affordance this bundle needs, so no fixture-mode equivalent is
  needed here.
- **`api_url` alternate config-key name is not modeled.** Legacy's `clockifyBaseURL` also accepts
  `config.api_url` as a fallback name for the base URL override (`clockify.go:233-235`); this
  bundle declares only `base_url` (spec.json's single, canonical property name) since the engine's
  `spec.json` "default" materialization mechanism has no analogous "try this key, then that key"
  fallback chain. An operator relying on the `api_url` config key name specifically would need to
  rename it to `base_url`.
