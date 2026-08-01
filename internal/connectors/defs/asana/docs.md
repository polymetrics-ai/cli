# Overview

Asana reads implemented project-management streams through the Asana v1 REST API and safely plans the currently implemented task/project/section/tag reverse-ETL actions. This bundle also carries the complete pinned official Asana OpenAPI operation ledger so every documented operation is represented exactly once as executable `covered_by` metadata or as a blocked/planned fixed-target operation row.

Official source inventory:

- Source ID: `asana_openapi_pinned`
- Pinned source: https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml
- OpenAPI: `3.0.0` / info version `1.0`
- Operation count: `249` (`GET=119`, `POST=81`, `PUT=26`, `DELETE=23`)
- Parent #380 lane counts: `etl_read=109`, `reverse_etl_write=124`, `direct_read_query_search=3`, `binary_file=4`, `cdc_changefeed=8`, `excluded_not_applicable=1`

Readable streams currently executable by the declarative engine: `custom_fields`, `project_statuses`, `projects`, `sections`, `stories`, `tags`, `tasks`, `team_memberships`, `teams`, `users`, `workspace_memberships`, `workspaces`.

Write actions currently executable by the declarative engine: `add_comment`, `create_project`, `create_section`, `create_tag`, `create_task`, `delete_project`, `delete_section`, `delete_tag`, `delete_task`, `update_project`, `update_section`, `update_tag`, `update_task`.

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

All other official read-like operations are represented in `api_surface.json`, `operations.json`, and `cli_surface.json` as blocked/planned fixed-target metadata until a connector-local stream or direct-read command adds a bounded schema, query flag contract, sanitized fixtures, and conformance evidence. Provider search/typeahead operations remain blocked on #2985. Audit/event/job/webhook changefeed surfaces remain blocked on #2986/#2988. Attachment/binary operations remain blocked until max-byte, redaction, and file-transfer fixtures are authored.

## Write actions & risks

Overall write risk: every external mutation must be run through reverse ETL plan -> preview -> explicit approval -> execute. Destructive/admin/delete operations also require typed confirmation (`confirm: "destructive"` / `--confirm destructive`) before execution.

Implemented write actions:

- `create_task`: POST `/tasks`; requires `data.name` and `data.workspace`; requires plan -> preview -> explicit approval -> execute.
- `update_task`: PUT `/tasks/{{ record.gid }}`; redacts `gid`; requires plan -> preview -> explicit approval -> execute.
- `delete_task`: DELETE `/tasks/{{ record.gid }}`; redacts `gid`; idempotent 404; `confirm: "destructive"`; requires plan -> preview -> explicit approval -> execute.
- `create_project`: POST `/projects`; requires `data.name`; requires plan -> preview -> explicit approval -> execute.
- `update_project`: PUT `/projects/{{ record.gid }}`; redacts `gid`; requires plan -> preview -> explicit approval -> execute.
- `delete_project`: DELETE `/projects/{{ record.gid }}`; redacts `gid`; idempotent 404; `confirm: "destructive"`; requires plan -> preview -> explicit approval -> execute.
- `create_section`: POST `/projects/{{ record.project_gid }}/sections`; redacts `project_gid`; requires plan -> preview -> explicit approval -> execute.
- `update_section`: PUT `/sections/{{ record.gid }}`; redacts `gid`; requires plan -> preview -> explicit approval -> execute.
- `delete_section`: DELETE `/sections/{{ record.gid }}`; redacts `gid`; idempotent 404; `confirm: "destructive"`; requires plan -> preview -> explicit approval -> execute.
- `create_tag`: POST `/tags`; requires `data.name`; requires plan -> preview -> explicit approval -> execute.
- `update_tag`: PUT `/tags/{{ record.gid }}`; redacts `gid`; requires plan -> preview -> explicit approval -> execute.
- `delete_tag`: DELETE `/tags/{{ record.gid }}`; redacts `gid`; idempotent 404; `confirm: "destructive"`; requires plan -> preview -> explicit approval -> execute.
- `add_comment`: POST `/tasks/{{ record.task_gid }}/stories`; redacts `task_gid`; requires plan -> preview -> explicit approval -> execute.

The remaining official POST/PUT/DELETE operations are not blanket-excluded. They are represented as blocked/planned operations with source evidence. They become executable only when a future connector-local action supplies a bounded record schema, path/body field redaction, sanitized write fixture, idempotency/destructive notes where applicable, and the existing reverse-ETL approval path.

## Known limits

- Fixture-only status: this connector is not live-certified. `certification.json` declares fixture defaults only; no live Asana credentials or provider calls were requested.
- `api_surface.json` uses `operation_ledger_version: 1`: legacy `excluded` classifiers are intentionally not used. Blocked/planned operation rows are the source of truth for unimplemented operations.
- `/batch` is the only not-applicable official lane row. It is disallowed because it is a generic batch subrequest wrapper and would recreate raw method/path/body passthrough; each underlying Asana operation is represented individually instead.
- Existing executable count remains the current 12 streams + 13 writes. The 224 remaining official rows are planned/blocked metadata, not executable runtime claims.
- Provider search/typeahead execution depends on #2985. CDC/changefeed/audit/webhook truthfulness depends on #2986/#2988. Attachment upload/download/delete execution needs connector-local binary/file contracts and fixtures.
- No generic shell, generic HTTP request/write, raw SQL write, arbitrary GraphQL, unrestricted file, unrestricted binary, or raw passthrough tool is exposed by this connector.
