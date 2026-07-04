# Overview

Toggl is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the practical documented
surface of the Toggl Track API v9 (full OpenAPI spec: `https://engineering.toggl.com/assets/files/
api-b2b7afb49ca646faff07d1e19b7b2afd.json`, 282 real operations across every published tag — see
`api_surface.json`). It reads time entries, projects, clients, tags, tasks, and users
(`GET https://api.track.toggl.com/api/v9/...`), and writes create/update/delete/stop mutations for
time entries, projects, clients, tags, and tasks. This bundle is migrated at capability parity from
`internal/connectors/toggl` (the hand-written connector it replaces) for its original 5 read
streams; the legacy package stays registered and unchanged until wave6's registry flip, and never
implemented any of the Pass B write actions or the 2 new streams (`tags`, `tasks`) — those are new
capability, not a parity port, and are called out as such below and in `api_surface.json`.

## Auth setup

Provide a Toggl API token via the `api_token` secret. Toggl authenticates API-token requests as
HTTP Basic auth with the token as the username and the literal string `api_token` as the password
(`connsdk.Basic(token, "api_token")`, `toggl.go:122`) — this bundle reproduces that exact shape via
`base.auth`'s `{"mode":"basic","username":"{{ secrets.api_token }}","password":"api_token"}`, where
`password` is a static literal (no `{{ }}` markers), matching chargebee's identical
token-as-username Basic-auth precedent (`docs/migration/conventions.md`). `base_url` defaults to
`https://api.track.toggl.com/api/v9` and may be overridden for tests/proxies.

## Streams notes

All five streams declare `"projection": "passthrough"` (conventions.md §3/§5 defect class 1):
legacy's `Read` (`toggl.go:94-103`) does `emit(connectors.Record(rec))` on every record read from
the raw API response with no field-building/filtering step at all — `streamSpecs[...].fields` is
Catalog-only decoration (`toggl.go:125-137`, consumed solely by `Catalog()`'s `connectors.Stream`
construction), never applied to the emitted record itself. Default `"schema"` projection mode would
silently drop every real Toggl field not named in each stream's declared schema properties (e.g.
`time_entries`' real `tags`/`tag_ids`/`billable`/`at`/`server_deleted_at`/`duronly`/`created_with`,
`projects`' real `color`/`template`/`auto_estimates`/`is_private`/`billable`/`rate`/`currency`), an
undocumented silent data-shape change relative to legacy's raw passthrough. Each schema still
declares the real Toggl Track API v9 wire-shape properties it knows about (both the current
`workspace_id`/`client_id`/`user_id`-style names and the API's legacy-compat aliases
`wid`/`cid`/`uid`/`pid`/`tid` — Toggl's v9 API emits both on every record) for `x-primary-key`
typing and `records_match_schema` coverage, but passthrough mode means ANY other real field Toggl
adds or returns still survives unfiltered, matching legacy exactly.

`time_entries` reads `GET /me/time_entries` for the authenticated user, optionally filtered by
`start_date`/`end_date` config values sent only when configured (legacy: `toggl.go:82-89`,
conditional `q.Set` calls) — reproduced here via the opt-in `omit_when_absent` query dialect
(conventions.md §3).

`projects`, `clients`, and `workspace_users` are workspace-scoped
(`GET /workspaces/{{ config.workspace_id }}/{projects,clients,users}`), matching legacy's
`workspacePath` helper (`toggl.go:158-166`, `url.PathEscape(id)`) — path interpolation's per-segment
`urlencode` default (conventions.md §3) reproduces the identical escaping. `organization_users` is
organization-scoped (`GET /organizations/{{ config.organization_id }}/users`), matching legacy's
`organizationPath` helper (`toggl.go:167-175`). Neither `workspace_id` nor `organization_id` is
globally `required` in `spec.json` (only the streams that need them fail without one) — an absent
value hard-errors at path-interpolation time with an unresolved-`config.*`-key error, the same
failure classification legacy produces (`"toggl connector requires config workspace_id"` /
`"...organization_id"`) via a differently-worded message, per conventions.md §5's config-validation
parity precedent. None of the five legacy streams expose an incremental cursor field in legacy, so
all five are always full-refresh reads. Neither of the 5 legacy streams is paginated — Toggl's
`/me/time_entries` and workspace/organization list endpoints return their full result set in one
response in legacy's own implementation (`toggl.go` never follows a next-page link for any of
them).

**New Pass B streams** (no legacy precedent — capability expansion, not a parity port):

- `tags` — `GET /workspaces/{{ config.workspace_id }}/tags`, a bare-array workspace-scoped list
  endpoint (`records.path: "."`), matching the exact shape of the pre-existing `clients`/
  `workspace_users` streams; not paginated (Toggl's own OpenAPI spec declares no page/per_page
  parameters on this endpoint).
- `tasks` — `GET /workspaces/{{ config.workspace_id }}/tasks`, a workspace-scoped flat task list
  (deliberately NOT the project-scoped `GET /workspaces/{id}/projects/{project_id}/tasks` shape,
  which would require the engine's `fan_out` dialect to enumerate every project first); records
  live under a `data` envelope key with `page`/`per_page` pagination (`page_number`, `page_size:
  100`, no `max_pages` — the real API paginates this endpoint and this bundle follows it to
  exhaustion, unlike the unpaginated legacy streams).

## Write actions & risks

**New Pass B capability — legacy is entirely read-only** (package doc: "implements a read-only
native Go connector for the Toggl Track API"; legacy's `Write` unconditionally returns
`connectors.ErrUnsupportedOperation`). `capabilities.write` is now `true` and `writes.json` declares
16 actions, every one a dialect-expressible plain JSON-body create/update/delete/stop mutation
matching a real, documented Toggl Track v9 endpoint (`api_surface.json`'s `covered_by.write`
entries):

- **Time entries**: `create_time_entry` (`POST .../time_entries`; `created_with` is a
  Toggl-required field identifying the calling application), `update_time_entry` (`PUT
  .../time_entries/{id}`), `stop_time_entry` (`PATCH .../time_entries/{id}/stop`, a `kind: custom`
  no-body action — Toggl's own dedicated "stop the running timer" endpoint), `delete_time_entry`
  (`DELETE .../time_entries/{id}`, idempotent on 404).
- **Projects**: `create_project`/`update_project`/`delete_project` (`.../projects[/{id}]`).
- **Clients**: `create_client`/`update_client`/`delete_client` (`.../clients[/{id}]`).
- **Tags**: `create_tag`/`update_tag`/`delete_tag` (`.../tags[/{id}]`).
- **Tasks**: `create_task`/`update_task`/`delete_task` (`.../projects/{{ record.project_id }}/
  tasks[/{{ record.id }}]` — task mutations are project-scoped in the real API even though the
  `tasks` stream itself reads the flatter workspace-scoped list endpoint; every task write action
  declares `project_id` in both `path_fields` and `record_schema.required` for this reason).

Every write's `risk` field states its specific blast radius; all are plain external mutations with
no organization-billing, ownership-transfer, or other `destructive_admin`/`requires_elevated_scope`
action modeled (see `api_surface.json` for the full excluded-endpoint accounting).

## Known limits

- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a
  static `fixture: true` marker and a hardcoded `workspace_id: "fixture_workspace"` string onto two
  synthesized records per stream (`toggl.go:176-187`). None of these are part of the LIVE record
  shape (where `workspace_id` is a real integer, not the fixture-mode string literal); this
  bundle's schemas and fixtures target the live path only. The engine's own conformance/
  fixture-replay harness (`internal/connectors/conformance`) provides the credential-free test
  affordance this bundle needs, so no fixture-mode equivalent is needed here.
- **No pagination is modeled for any of the 5 legacy streams or the new `tags` stream**, matching
  legacy exactly (and Toggl's own documented lack of page parameters on `tags`); only the new
  `tasks` stream paginates, since its real API endpoint genuinely does.
- **Toggl's dual-key legacy-compat aliases (`wid`/`cid`/`uid`/`pid`/`tid`) are read but not written.**
  The read-side schemas model both the modern and legacy-alias field names Toggl's API emits (see
  each stream's schema); the new write actions only ever SEND the modern field names
  (`workspace_id`/`client_id`/`user_id`/`project_id`/`task_id` — configured via path or
  `record_schema`, never the aliases) since Toggl's write payloads document only the modern names
  as accepted request fields (`timeentry.Payload`/`project.Payload`/`client.Payload`/
  `tags.payload`/`task.Payload` in the published OpenAPI spec never declare the alias fields on the
  request side, only on responses).
- **`create_time_entry` requires `created_with`.** Toggl's API documents this as a required field
  on entry creation (used to identify the calling application in Toggl's own UI/reporting); this
  bundle declares it `required` in `record_schema` rather than silently defaulting it, so a caller
  omitting it gets a clear validation failure instead of a confusing 400 from the live API.
