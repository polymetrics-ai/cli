# Overview

Asana reads source-bound operations through the Asana v1 REST API and executes typed one-record direct writes through plan -> preview -> approval -> execute. Saved connections use the same 12 streams for warehouse-backed ETL, and `pm reverse` uses the same write actions for warehouse-table reverse ETL. The pinned OpenAPI ledger accounts for all 249 provider operations: 213 have executable `covered_by` routes and the remaining 36 are explicit non-executable or not-applicable rows with their source citation and current foundation reason.

Official source inventory:

- Source ID: `asana_openapi_pinned`
- Pinned source: https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml
- OpenAPI: `3.0.0` / info version `1.0`
- Operation count: `249` (`GET=119`, `POST=81`, `PUT=26`, `DELETE=23`)
- Declared interactive lanes: `direct_read=118` (106 bounded operation-backed + 12 stream-backed commands with the request-bound gap below), `direct_write=96` across 95 provider endpoints (94 one-to-one actions + 2 attachment request variants sharing `POST /attachments`)
- Remaining ledger rows: one source-bound GET with a named foundation gap, 34 mutations with an exact declared contract/foundation gap, and `/batch` as the sole not-applicable generic-wrapper route

## Source-operation lane matrix

The connector-local `TestEveryLockedOperationHasOneAccountedCommandLane` constructs this matrix from the immutable 249-operation source lock, `api_surface.json`, and `cli_surface.json`. It counts provider operations once; command aliases are accounted separately.

| Locked provider operations | Current lane/accounting | Actual execution status |
| ---: | --- | --- |
| 106 | operation-backed `direct_read` | implemented bounded operation read |
| 12 | stream-backed read plus saved ETL | saved full-refresh ETL is executable; the interactive commands remain declared `direct_read` but do not yet impose a one-provider-request bound |
| 95 | 96 `direct_write` actions plus warehouse-table reverse ETL | all 96 actions pass isolated warehouse plan -> preview -> approval -> one captured provider request -> completed run |
| 1 | planned `direct_read` (`GET /memberships/{membership_gid}`) | blocked on `cli-openapi30-reference-sibling-foundation-r1` |
| 34 | planned `direct_write` | blocked on a named bounded request-schema projection; `createMembership` also carries the sibling-reference gap |
| 1 | unsupported generic `/batch` wrapper | deliberately not exposed because it would recreate raw subrequest passthrough |

The rows total 249. They project to 252 CLI command rows because `asana.rest.createAttachmentForObject` has four command representations: two implemented direct writes (`attachments upload-attachment-file` and `attachments create-external-attachment`), one planned legacy operation alias, and one planned `binary_upload` alias. The file-upload write itself is executable; the binary alias remains planned because the provider does not publish a fixed media allow-list. No locked Asana operation supplies a binary response contract, so no `binary_download` command is invented.

Readable streams currently executable by the declarative engine: `custom_fields`, `project_statuses`, `projects`, `sections`, `stories`, `tags`, `tasks`, `team_memberships`, `teams`, `users`, `workspace_memberships`, `workspaces`.

Write actions currently executable by the declarative engine: 96, comprising 94 actions bound one-to-one to official POST/PUT/DELETE rows plus two source-backed attachment request variants for the same `POST /attachments` operation. The 94 endpoint-bound actions comprise 51 `create`, 20 `update`, and 23 `delete` actions. This includes 19 source-complete DELETEs and the no-body `approve_access_request` and `reject_access_request` POSTs. `writes.json` is the authoritative per-action contract; `pm connectors inspect asana` and the generated `docs/connectors/asana/SKILL.md` render every action's endpoint, required record fields, and risk note.

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

Default pagination follows Asana's `next_page.uri` response field with same-host guarding. Source-bound direct reads use the declared closed pagination contract only: callers navigate with `--page` or `--page-cursor`; raw provider controls such as `offset` and `limit` are not command flags.

Implemented streams remain intentionally bounded and fixture-backed:

- `workspaces`: GET `/workspaces`; records path `data`; first-stream fixture-backed.

The 106 operation-backed direct reads are deliberately distinct from stream-backed reads: each operation read returns one response-capped provider page and reports only the completeness its declared pagination can prove. The 12 stream-backed interactive commands, including `pm asana workspaces list`, currently call `Connector.Read` with a record limit but no `MaxPages` request bound. Pagination can therefore make more than one provider request, and `sections`, `stories`, and `project_statuses` add a fan-out discovery request plus child reads. Until the command route imposes and tests a one-request/page/fan-out bound, those 12 commands do not satisfy the connector's `direct_read` contract even though their saved full-refresh ETL streams are executable. Saved connections use the same declarations for warehouse-backed ETL through `pm etl`; the interactive command itself is not an ETL run.

`event_source_contract.json` is the machine-readable source-evidence projection for the project-scoped `tasks` incremental lane. Its closed schema binds `sync_transport.json`'s exact `asana_event_token_source` executor/conformance reference to the immutable lock and cites `asana.rest.getEvents`, the `resource`/`sync` parameters, first/expired 412 rebootstrap, `data`/`sync`/`has_more`, project-to-task scope, `EventResponse` actions and resource `gid`/type, `asana.rest.getTask` hydration, the project-filtered `asana.rest.getTasks` snapshot, pagination, and auth. It explicitly records `event_total_order=not_documented`. The file is evidence, not a runtime lifecycle: retry, page caps, window coalescing, checkpoint acknowledgement, and request execution remain owned by the registered Asana executor.

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

The sole non-executable source GET is `asana.rest.getMembership` (`GET /memberships/{membership_gid}`): it records `missing_foundation=cli-openapi30-reference-sibling-foundation-r1` with the pinned source citation. Other non-executable provider operations remain visible in `api_surface.json`, `operations.json`, and `cli_surface.json` with their exact request-contract, encoding, CDC, or binary-input foundation; no generic request or raw provider escape hatch is exposed.

## Write actions & risks

Overall write risk: every interactive external mutation is a one-record direct write that must run through plan -> preview -> explicit approval -> execute. Bulk reverse ETL remains the separate `pm reverse` warehouse-table path. Destructive/admin/delete operations also require typed confirmation (`confirm: "destructive"` / `--confirm destructive`) before execution.

Implemented write actions: 96 named actions in `writes.json`, which is their authoritative contract (endpoint, bounded record schema, required/accepted fields, redacted path fields, idempotency and confirmation notes). Read them with `pm connectors inspect asana` or in the generated `docs/connectors/asana/SKILL.md`; this file does not restate per-action fields.

By resource family: tasks and subtasks (create/update/delete, duplicate, instantiate from template, set parent, add dependencies/dependents/project/tag/followers), projects (create for team/workspace, update, delete, duplicate, save as template, briefs, statuses, custom-field settings, members, followers, portfolio settings), sections (create, insert, update, delete, add task), tags (create, create for workspace, update, delete), stories (`add_comment`, goal stories, update), goals and goal relationships (create/update, metrics, supporting relationships, followers, custom-field settings), portfolios (create/update, add item/members/custom-field setting, duplicate), custom fields and enum options, teams and team membership, users and workspace membership, workspaces, status updates, rule triggers, exports, OOO entries, and time-tracking entries.

Every interactive action routes through the one-record direct-write plan -> preview -> explicit approval -> execute lifecycle; bulk warehouse-table execution remains available through `pm reverse`. All 23 implemented DELETEs require typed `--confirm destructive`, treat 404 as success, and redact their path fields. The source-complete DELETE set is covered by the same action path as the earlier task/project/section/tag deletes; only 16 destructive rows remain non-executable because their current source request contract lacks the required bounded action foundation.

The remaining official POST/PUT/DELETE operations are not blanket-excluded. Their pinned source entries name the exact missing foundation. In particular, 69 implemented actions retain source-partial declarations only where the real provider request requires `cli-request-schema-foundation-r1` (65 operations) or `source-path-parameter-alias-foundation-r1` (4 operations); their declared typed subset remains executable and is never presented as complete provider coverage.

## Known limits

- Static source availability is computed from pinned declarations: no credential or live-provider account is required to determine whether a declared operation has its complete shared foundation.
- `api_surface.json` uses `operation_ledger_version: 1`: non-executable operation rows are the source of truth for the remaining foundation gaps.
- `/batch` is the only not-applicable official lane row. It is disallowed because it is a generic batch subrequest wrapper and would recreate raw method/path/body passthrough; each underlying Asana operation is represented individually instead.
- Provider-route coverage is 12 streams + 106 bounded source-bound operation reads + 95 mutation endpoints (213 `covered_by` rows). Interactive presentation declares 118 direct reads and 96 direct writes because the attachment operation has two source-backed request variants; the 12 stream-backed interactive reads retain the request-bound defect described above. The remaining 36 official rows are explicit non-executable or not-applicable metadata, not executable runtime claims.
- Full-refresh overwrite/append remains available for all 12 saved ETL streams. Provider-token incremental append/upsert/dedupe is admitted only for project-scoped `tasks`, using Asana Events tokens, complete `has_more` windows, durable checkpoint acknowledgement, stable `gid` coalescing/hydration, and `deleted` tombstones. The other 11 stream scopes and ordered history/change capture remain unproven and are not admitted as incremental modes.
- Saved bulk multipart upload currently resolves a warehouse row's `file_path` relative to the app-owned `.polymetrics` runtime directory, while the one-record direct-write route resolves it relative to the public project root. The all-action reverse-ETL proof exercises the current bounded behavior; changing the shared bulk runtime root requires a deliberate App-level policy correction.
- Every promoted write's record schema is derived from the pinned OpenAPI source above, never inferred from response shapes. Envelope and resource levels are closed with `additionalProperties: false`; deeply nested provider-defined regions (for example `custom_fields` on `create_task`) stay `type: object` with `additionalProperties: true`, which is the bundle's bounded-but-not-exhaustive convention.
- Provider search/typeahead, CDC/changefeed/audit/webhook, and attachment binary routes remain visible only with their actual named request/response foundation in the source ledger; source unavailability and missing runtime foundation are distinct states.
- No generic shell, generic HTTP request/write, raw SQL write, arbitrary GraphQL, unrestricted file, unrestricted binary, or raw passthrough tool is exposed by this connector.
