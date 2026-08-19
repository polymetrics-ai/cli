# Overview

Asana reads implemented project-management streams through the Asana v1 REST API and executes typed reverse-ETL write actions across tasks, projects, sections, tags, stories, goals, portfolios, teams, users, workspaces, custom fields, exports, templates, OOO entries, and time-tracking entries. This bundle also carries the complete pinned official Asana OpenAPI operation ledger so every documented operation is represented exactly once as executable `covered_by` metadata or as a blocked/planned fixed-target operation row.

Official source inventory:

- Source ID: `asana_openapi_pinned`
- Pinned source: https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml
- OpenAPI: `3.0.0` / info version `1.0`
- Operation count: `249` (`GET=119`, `POST=81`, `PUT=26`, `DELETE=23`)
- Retrieval provenance, SHA-256, byte count, and operation inventory: `sources/asana-operation-source-lock.json` (retrieved 2026-08-19)
- Parent #380 lane counts: `etl_read=111`, `reverse_etl_write=125`, `direct_read_query_search=3`, `file_upload=1`, `cdc_changefeed=8`, `excluded_not_applicable=1`

Readable streams currently executable by the declarative engine: `custom_fields`, `project_statuses`, `projects`, `sections`, `stories`, `tags`, `tasks`, `team_memberships`, `teams`, `users`, `workspace_memberships`, `workspaces`.

Write actions currently executable by the declarative engine: 73 (51 `create`, 18 `update`, 4 `delete`), bound to 73 of the 130 official POST/PUT/DELETE rows. `writes.json` is the authoritative per-action contract; `pm connectors inspect asana` and the generated `docs/connectors/asana/SKILL.md` render every action's endpoint, required record fields, and risk note.

Service API documentation: https://developers.asana.com/reference/rest-api-reference.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Asana personal access token, sent as a Bearer token. Never log or place this value in issue bodies, docs, fixtures, or command arguments.
- `base_url` (optional, string); default `https://app.asana.com/api/1.0`; format `uri`; used by fixture replay and safe test proxies.
- `workspace_id` (optional, string); scopes implemented workspace/project/team/custom-field/member streams where a workspace path or query parameter is required.
- `project_id` (optional, string); scopes the implemented `tasks` stream when set.
- `assignee` (optional, string); scopes the implemented `tasks` stream when set.
- `team_id` (optional, string); scopes the implemented `team_memberships` stream when set.
- `mode` (optional, string); fixture metadata only.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://app.asana.com/api/1.0`.

Authentication behavior: bearer token authentication using `secrets.access_token`.

Connection checks call GET `/users/me`. That identity alias is documented by Asana and already used by this connector, but it is not present in the pinned OpenAPI `paths` map and is therefore intentionally not counted in `api_surface.json`'s 249 official operation ledger.

## Streams notes

Default pagination follows Asana's `next_page.uri` response field with same-host guarding.

Implemented streams remain intentionally bounded and fixture-backed:

- `workspaces`: GET `/workspaces`; records path `data`; first-stream fixture-backed.
- `projects`: GET `/projects`; optional `workspace` query from `workspace_id`.
- `tasks`: GET `/tasks`; optional `workspace`, `project`, and `assignee` query values.
- `users`: GET `/users`; optional `workspace` query.
- `teams`: GET `/workspaces/{{ config.workspace_id }}/teams`.
- `tags`: GET `/tags`; optional `workspace` query.
- `sections`: GET `/projects/{{ fanout.id }}/sections`; project fan-out stamps `project_gid`.
- `stories`: GET `/tasks/{{ fanout.id }}/stories`; task fan-out stamps `task_gid`.
- `custom_fields`: GET `/workspaces/{{ config.workspace_id }}/custom_fields`.
- `project_statuses`: GET `/projects/{{ fanout.id }}/project_statuses`; project fan-out stamps `project_gid`.
- `team_memberships`: GET `/team_memberships`; optional `team` and `workspace` query values.
- `workspace_memberships`: GET `/workspaces/{{ config.workspace_id }}/workspace_memberships`.

All other official read-like operations are represented in `api_surface.json`, `operations.json`, and `cli_surface.json` as blocked/planned fixed-target metadata until a connector-local stream or direct-read command adds a bounded schema, query flag contract, sanitized fixtures, and conformance evidence. Provider search/typeahead operations remain blocked on #2985. Audit/event/job/webhook changefeed surfaces remain blocked on #2986/#2988. Attachment metadata reads remain blocked until connector-local JSON schema, redaction, and sanitized fixture evidence are authored; attachment uploads remain blocked until file-upload input contracts and multipart fixture evidence exist.

## Write actions & risks

Overall write risk: every external mutation must be run through reverse ETL plan -> preview -> explicit approval -> execute. Destructive/admin/delete operations also require typed confirmation (`confirm: "destructive"` / `--confirm destructive`) before execution.

Implemented write actions: 73 named actions in `writes.json`, which is their authoritative contract (endpoint, bounded record schema, required/accepted fields, redacted path fields, idempotency and confirmation notes). Read them with `pm connectors inspect asana` or in the generated `docs/connectors/asana/SKILL.md`; this file does not restate per-action fields.

By resource family: tasks and subtasks (create/update/delete, duplicate, instantiate from template, set parent, add dependencies/dependents/project/tag/followers), projects (create for team/workspace, update, delete, duplicate, save as template, briefs, statuses, custom-field settings, members, followers, portfolio settings), sections (create, insert, update, delete, add task), tags (create, create for workspace, update, delete), stories (`add_comment`, goal stories, update), goals and goal relationships (create/update, metrics, supporting relationships, followers, custom-field settings), portfolios (create/update, add item/members/custom-field setting, duplicate), custom fields and enum options, teams and team membership, users and workspace membership, workspaces, status updates, rule triggers, exports, OOO entries, and time-tracking entries.

Every action routes through reverse ETL plan -> preview -> explicit approval -> execute. Exactly four are destructive (`delete_task`, `delete_project`, `delete_section`, `delete_tag`): they carry the legacy `confirm: "destructive"` declaration normalized by the shared typed gate, treat 404 as success, and redact their path fields. No other DELETE or admin/elevated operation is bound to an action. The shared gate makes the 36 `destructive_action` rows technically bindable, but they stay unbound pending connector-local typed action schemas, canonical command mappings, and fixtures; `reverse_etl_execute_test.go`'s `TestDestructiveOperationsStayBlocked` fails if that count moves without that authoring work.

The remaining official POST/PUT/DELETE operations are not blanket-excluded. They are represented as blocked/planned operations with source evidence. They become executable only when a future connector-local action supplies a bounded record schema, path/body field redaction, sanitized write fixture, idempotency/destructive notes where applicable, and the existing reverse-ETL approval path.

## Known limits

- Fixture-only status: this connector is not live-certified. `certification.json` declares fixture defaults only; no live Asana credentials or provider calls were requested.
- `api_surface.json` uses `operation_ledger_version: 1`: legacy `excluded` classifiers are intentionally not used. Blocked/planned operation rows are the source of truth for unimplemented operations.
- `/batch` is the only not-applicable official lane row. It is disallowed because it is a generic batch subrequest wrapper and would recreate raw method/path/body passthrough; each underlying Asana operation is represented individually instead.
- Executable surfaces are 12 streams + 73 writes (85 `covered_by` rows). The 164 remaining official rows are planned/blocked metadata, not executable runtime claims.
- Every promoted write's record schema is derived from the pinned OpenAPI source above, never inferred from response shapes. Envelope and resource levels are closed with `additionalProperties: false`; deeply nested provider-defined regions (for example `custom_fields` on `create_task`) stay `type: object` with `additionalProperties: true`, which is the bundle's bounded-but-not-exhaustive convention.
- Provider search/typeahead execution depends on #2985. CDC/changefeed/audit/webhook truthfulness depends on #2986/#2988. Attachment metadata read/upload/delete execution needs connector-local JSON/file-upload contracts and fixtures.
- No generic shell, generic HTTP request/write, raw SQL write, arbitrary GraphQL, unrestricted file, unrestricted binary, or raw passthrough tool is exposed by this connector.
