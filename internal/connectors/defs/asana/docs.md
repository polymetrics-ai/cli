# Overview

Asana reads source-bound operations through the Asana v1 REST API and executes typed one-record direct writes through plan -> preview -> approval -> execute. Saved connections use 12 existing declarative streams for warehouse-backed ETL, and `pm reverse` uses the same write actions for warehouse-table reverse ETL. The pinned OpenAPI ledger accounts for all 249 provider operations in `sources/asana-source-lane-matrix.json`; direct-route implementation does not promote a pageable source operation to an executable ETL stream. `POST /batch` is available only through the closed declared-action selector; it is not a raw HTTP escape hatch.

Official source inventory:

- Source ID: `asana_openapi_pinned`
- Pinned source: https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml
- OpenAPI: `3.0.0` / info version `1.0`
- Operation count: `249` (`GET=119`, `POST=81`, `PUT=26`, `DELETE=23`)
- Declared interactive lanes: `direct_read=119` (107 bounded operation-backed + 12 stream-backed commands with a shared one-provider-request budget), `direct_write=131` across all 130 mutation endpoints (129 one-to-one actions + 2 attachment request variants sharing `POST /attachments`)
- Additional command rows: one implemented `binary_upload` alias and one planned legacy attachment operation alias; neither adds a provider operation

The source matrix is the sole identity denominator. Its explicit seven-lane dispositions are:

| Lane | Source disposition |
| --- | --- |
| Direct read | 119 implemented; 130 not applicable |
| Direct write | 130 implemented; 119 not applicable |
| Binary download | 0 applicable; 249 source-evidenced not applicable |
| Binary upload | 1 implemented attachment operation; 248 not applicable |
| ETL through DuckDB | 64 candidates: 12 implemented stream/schema/API projections and 52 `mapped_unproven` descriptor-only cells; 185 not applicable |
| Reverse ETL from DuckDB | 130 implemented; 119 not applicable |
| Sync transport through DuckDB | 3 implemented task event/hydration/snapshot cells; 246 not applicable |

## Source-operation lane matrix

`TestAsanaSourceLaneMatrixRetainsEveryLockedOperationAndLane` retains the immutable 249-operation source denominator. `TestAsanaSourceLaneArtifactsProjectTheTrackAMatrix` verifies that applicable cells resolve to existing definition artifacts and that an artifact or backlink cannot be removed silently. Command aliases are accounted separately from provider identities.

| Source group or overlay | Current lane/accounting | Actual execution status |
| ---: | --- | --- |
| 107 | operation-backed `direct_read` | implemented bounded operation read, including `GET /memberships/{membership_gid}` |
| 12 | stream-backed `direct_read` plus saved ETL | the interactive command has one aggregate provider-request/page budget; saved full-refresh ETL remains exhaustive |
| 52 ETL overlay cells | descriptor-pagination ETL candidates | `mapped_unproven`: retained in the matrix and gap ledger, with no selected stream/schema/API route and no executable-stream claim |
| 130 | 131 `direct_write` actions plus warehouse-table reverse ETL | every mutation endpoint is executable; `POST /attachments` has two closed request variants and `POST /batch` has one closed declared-action adapter |

The first, second, and fourth rows are the mutually exclusive direct-route partition (107 + 12 + 130 = 249). The ETL overlay intersects the GET rows and therefore does not add a second denominator. The source rows project to 252 CLI command rows because `asana.rest.createAttachmentForObject` has four command representations: two implemented direct writes (`attachments upload-attachment-file` and `attachments create-external-attachment`), one planned legacy operation alias, and one implemented `binary_upload` alias. The planned legacy alias is an accounting/compatibility row, not an unmapped provider operation. The provider documents arbitrary full file contents rather than a finite media allow-list, so the file action carries a closed `provider_unrestricted` policy instead of fabricated MIME types. The binary alias uses the same 100 MB cap, project-root confinement, payload digest, preview, and approval binding. No locked Asana operation supplies a binary response contract, so no `binary_download` command is invented.

Readable streams currently executable by the declarative engine: `custom_fields`, `project_statuses`, `projects`, `sections`, `stories`, `tags`, `tasks`, `team_memberships`, `teams`, `users`, `workspace_memberships`, `workspaces`.

Write actions currently executable by the declarative engine: 131 across all 130 official POST/PUT/DELETE rows. There are 129 one-to-one endpoint actions plus two source-backed attachment request variants for the same `POST /attachments` operation. The action kinds are 63 `create`, 28 `update`, 39 `delete`, and one closed `custom` batch action. `writes.json` is the authoritative per-action contract; `pm connectors inspect asana` and the generated `docs/connectors/asana/SKILL.md` render every action's endpoint, required record fields, and risk note.

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

Implemented streams remain intentionally bounded and fixture-backed. They are the only source-backed ETL stream projections; descriptor pagination alone does not admit the other 52 candidate cells:

- `workspaces`: GET `/workspaces`; records path `data`; first-stream fixture-backed.

The 107 operation-backed direct reads are deliberately distinct from stream-backed reads: each operation read returns one response-capped provider page and reports only the completeness its declared pagination can prove. The 12 stream-backed interactive commands, including `pm asana workspaces list`, apply one aggregate provider-send and page budget across pagination, retries, redirects, discovery, and fan-out. Saved connections leave that interactive budget unset so warehouse-backed full-refresh ETL remains exhaustive. An interactive stream command is still a bounded `direct_read`; it is not an ETL run.

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

All 119 locked source GET operations are executable: 107 through bounded operation-backed commands and 12 through bounded stream-backed commands. `asana.rest.getMembership` (`GET /memberships/{membership_gid}`) is implemented because its request path/query contract is complete; the retained OpenAPI 3.0 response-schema sibling diagnostic is response-only authoring evidence and does not block the bounded JSON read. No generic request or raw provider escape hatch is exposed.

## Write actions & risks

Overall write risk: every interactive external mutation is a one-record direct write that must run through plan -> preview -> explicit approval -> execute. Bulk reverse ETL remains the separate `pm reverse` warehouse-table path. Destructive/admin/delete operations also require typed confirmation (`confirm: "destructive"` / `--confirm destructive`) before execution.

Implemented write actions: 131 named actions in `writes.json`, which is their authoritative contract (endpoint, bounded record schema, required/accepted fields, redacted path fields, idempotency and confirmation notes). Read them with `pm connectors inspect asana` or in the generated `docs/connectors/asana/SKILL.md`; this file does not restate per-action fields.

By resource family: tasks and subtasks (create/update/delete, duplicate, instantiate from template, set parent, add dependencies/dependents/project/tag/followers), projects (create for team/workspace, update, delete, duplicate, save as template, briefs, statuses, custom-field settings, members, followers, portfolio settings), sections (create, insert, update, delete, add task), tags (create, create for workspace, update, delete), stories (`add_comment`, goal stories, update), goals and goal relationships (create/update, metrics, supporting relationships, followers, custom-field settings), portfolios (create/update, add item/members/custom-field setting, duplicate), custom fields and enum options, teams and team membership, users and workspace membership, workspaces, status updates, rule triggers, exports, OOO entries, and time-tracking entries.

Every interactive action routes through the one-record direct-write plan -> preview -> explicit approval -> execute lifecycle; bulk warehouse-table execution remains available through `pm reverse`. All provider DELETE routes require typed `--confirm destructive`, treat 404 as success, and redact their path fields. POST-based remove/admin actions use their source-projected request schema and the same approval boundary.

Sixty-nine actions retain historical source-partial disposition metadata where their earlier projection required `cli-request-schema-foundation-r1` (65 operations) or `source-path-parameter-alias-foundation-r1` (4 operations). The strict source-first projector now recreates their final closed request/CLI behavior from the schema-v3 lock; the dispositions remain audit evidence and do not block execution.

## Known limits

- Static source availability is computed from pinned declarations: no credential or live-provider account is required to determine whether a declared operation has its complete shared foundation.
- `api_surface.json` uses `operation_ledger_version: 1`: non-executable operation rows are the source of truth for the remaining foundation gaps.
- `/batch` is executable only as the closed `create_batch_request` action. A record supplies 1..10 named, explicitly allow-listed existing actions; the engine derives every subrequest method/path/body, rejects unsupported/nested/query-bearing actions before provider I/O, preserves preview/approval, and fails closed on partial or malformed provider results. Callers never supply raw HTTP fields.
- Provider-route coverage is 12 bounded stream reads + 107 bounded operation reads + all 130 mutation endpoints. Interactive presentation declares 119 direct reads and 131 direct writes; the attachment operation has two source-backed request variants. Those direct routes account for all 249 locked provider rows, but they do not alter the separate 64-cell ETL denominator or make mapped-unproven rows executable streams.
- Full-refresh overwrite/append remains available for the 12 selected saved ETL streams. The other 52 descriptor-pagination candidates remain `mapped_unproven` because no exact source-backed scope/fanout and stream/schema/API projection has been selected. Provider-token incremental append/upsert/dedupe is admitted only for project-scoped `tasks`, using Asana Events tokens, complete `has_more` windows, durable checkpoint acknowledgement, stable `gid` coalescing/hydration, and `deleted` tombstones. The other 11 selected-stream scopes and ordered history/change capture remain unproven and are not admitted as incremental modes.
- Saved bulk multipart upload currently resolves a warehouse row's `file_path` relative to the app-owned `.polymetrics` runtime directory, while the one-record direct-write route resolves it relative to the public project root. The all-action reverse-ETL proof exercises the current bounded behavior; changing the shared bulk runtime root requires a deliberate App-level policy correction.
- Every promoted write's record schema is derived from the pinned OpenAPI source above, never inferred from response shapes. Envelope and resource levels are closed with `additionalProperties: false`; deeply nested provider-defined regions (for example `custom_fields` on `create_task`) stay `type: object` with `additionalProperties: true`, which is the bundle's bounded-but-not-exhaustive convention.
- Provider search/typeahead and audit/webhook operations are bounded direct commands, not evidence that every stream supports incremental ETL. The source-backed Asana attachment action is the sole binary-upload route; no binary-download route is claimed.
- No generic shell, generic HTTP request/write, raw SQL write, arbitrary GraphQL, caller-selected file route/media contract, or raw passthrough tool is exposed by this connector.
