# Overview

Capsule CRM is a Tier-1 declarative-HTTP migration of `internal/connectors/capsule-crm`
(legacy Go package `capsulecrm`), expanded in Pass B to the full practical Capsule v2 REST
surface. It reads Capsule CRM v2 parties, opportunities, cases (`kases`), tasks, users, tags,
custom field definitions, teams, sales pipelines, pipeline milestones, lost reasons, task
categories, kanban boards, and board stages — and writes party/opportunity/case/task create,
update, and delete actions.

## Auth setup

Provide a Capsule CRM personal access token via the `bearer_token` secret; it is used only
for Bearer auth (`Authorization: Bearer <bearer_token>`) and is never logged.

## Streams notes

The 5 legacy-parity streams (`parties`, `opportunities`, `kases`, `tasks`, `users`) share the
same shape: `GET` against the Capsule v2 list endpoint, records unwrapped from the top-level
JSON key matching the resource name (e.g. `{"parties": [...]}`), primary key `["id"]` (a
numeric Capsule id). Pagination follows Capsule's 1-based `page`/`perPage` convention
(`pagination.type: page_number`, `page_param: page`, `size_param: perPage`, `start_page: 1`),
stopping on a short/empty page exactly like legacy's `connsdk.PageNumberPaginator`.

`streams.json`'s `pagination.page_size` is a static JSON int (`PaginationSpec.PageSize`),
resolved once at bundle load with no config-driven override — the same shape documented in
`auth0`'s and `searxng`'s goldens (`docs/migration/conventions.md`). Legacy's real default is
50 (`capsuleDefaultPageSize`, configurable up to 100 via a `page_size` config value); this
bundle declares `page_size: 50` to match that default exactly (a smaller static value would
silently narrow every live page fetch to a fraction of legacy's request size, multiplying the
number of API calls per sync even though fixture replay would look identical either way). The
required 2-page `parties` conformance fixture (`docs/migration/conventions.md` §4) accordingly
ships a full 50-record page 1 and a short 1-record page 2 to exercise the real stop threshold.
Legacy's `page_size`/`max_pages` config properties are consequently genuinely
dead in this dialect (no template anywhere reads them) and are intentionally NOT declared in
`spec.json` (F6, REVIEW.md: a declared-but-unwireable key is worse than an absent one).

Legacy's stream catalog declares `CursorFields: []string{"updatedAt"}` for every stream, but
`capsulecrm.Read` never actually filters by it — every read is a full pass over the
collection (full-refresh only, no `incremental` request param or client-side filtering
exists in legacy). This bundle mirrors that exactly: each of the 5 legacy-parity schemas
declares `x-cursor-field: updated_at` (preserving the catalog metadata parity) but
intentionally declares **no `incremental` block** in `streams.json` — adding one would
introduce new, behavior-changing server-side or client-side filtering legacy never performed.

Several legacy record mappers flatten nested objects into `_id`/`_name` scalar fields
(`opportunities`' `party`/`milestone`/`value` objects, `kases`' `party` object, `tasks`'
`category`/`party`/`opportunity`/`kase` objects). This bundle reproduces every one of these
via `computed_fields` dotted-path references against the raw record (e.g.
`"party_id": "{{ record.party.id }}"`, `"value_amount": "{{ record.value.amount }}"`) — each
is a bare single `{{ record.<path> }}` reference with no filter stage, so the engine's typed
extraction preserves the nested id's native integer type rather than stringifying it,
matching legacy's raw `item["id"]` (an `int`/`float64` from JSON) assignment exactly.
Capsule's camelCase wire fields (`firstName`, `organisationName`, `createdAt`, `updatedAt`,
etc.) are renamed to legacy's snake_case output field names the same way, via bare
`computed_fields` references.

**Pass B account-configuration streams** (`tags`, `custom_fields`, `teams`, `pipelines`,
`milestones`, `lost_reasons`, `categories`, `boards`, `stages`) are new in this wave: every one
is unpaginated in practice (Capsule returns each as a single small array with no `page`/
`perPage` query support), so each declares `"pagination": {"type": "none"}` and a single-page
fixture. `custom_fields` reads `GET /customfields`, whose response envelope key is
`definitions` (not `customfields`) per Capsule's own documented shape; `entityType`/
`restrictedToType` are renamed to `entity_type`/`restricted_to_type` via `computed_fields`,
matching this bundle's snake_case convention. `pipelines`' `displayOrder` and `stages`'
`displayOrder` are similarly renamed to `display_order`; `boards`' `entityType` to
`entity_type`. `milestones.pipeline_id` and `stages.board_id` are bare
`{{ record.pipeline.id }}`/`{{ record.board.id }}` computed-field dotted-path extractions,
the same nested-object-flatten pattern the legacy-parity streams already use.

## Write actions & risks

Pass B adds full create/update/delete coverage for the four core CRM entities, following
Capsule's documented resource-envelope convention: every write body is wrapped under a
top-level key matching the resource (`{"party": {...}}`, `{"opportunity": {...}}`,
`{"kase": {...}}`, `{"task": {...}}`) — the record itself must carry that wrapper key, since
the engine's write dialect sends record fields verbatim as the JSON body with no
nested-wrapper construction primitive (the same pattern documented in the `teamwork` golden's
`create_project` action).

- `create_party` / `update_party` / `delete_party` — creates, updates, or irreversibly deletes
  a Capsule contact (person or organisation). Capsule requires `firstName`+`lastName` (person)
  or `name` (organisation) for create; update accepts any partial `party` object; delete is a
  hard, unrecoverable removal including the contact's activity history.
- `create_opportunity` / `update_opportunity` / `delete_opportunity` — creates, updates, or
  deletes a sales opportunity. Create requires `name`, a `party.id`, and a `milestone.id`
  (Capsule's own required-fields rule); update commonly moves `milestone.id` (pipeline stage)
  or sets `closedOn`/`lostReason` to close/lose the deal.
  Update requires an `owner` and/or `team` on the account per Capsule's documented rule, but
  since those values are typically inherited from the existing record rather than supplied on
  every partial update, this bundle does not hard-require them in `record_schema` — an update
  omitting both when neither is already set on the account default will surface as a live-API
  422, not a local validation failure (the dialect's draft-07 subset cannot express "required
  unless already present server-side").
- `create_kase` / `update_kase` / `delete_kase` — creates, updates, or deletes a Capsule case
  (Capsule's own current product naming is "Project", but the API's `/kases` path and `kase`
  envelope key are unchanged to avoid a breaking change — see Capsule's own Case API docs).
  Create requires `name` and a `party.id`; update commonly sets `status: CLOSED` with a
  `closedOn` date.
- `create_task` / `update_task` / `delete_task` — creates, updates, or deletes a task/reminder.
  Create requires `description`; update commonly sets `status` to `PENDING`/`COMPLETED`.

Every write action is `risk: external mutation; approval required` per this connector's
`metadata.json.risk.write`; the three `delete_*` actions are irreversible.

## Known limits

- Detail/show-by-id endpoints (`/parties/{id}`, `/opportunities/{id}`, `/kases/{id}`,
  `/tasks/{id}`, `/users/{id}`, and the same shape for every account-configuration resource)
  are excluded as `duplicate_of` their list stream — the dialect has no per-id detail-stream
  primitive distinct from a list read; see `api_surface.json`.
- `/parties/search`, `/opportunities/search`, `/kases/search`, and the `/deleted` tombstone
  feeds are out of scope: these are query-time/feed operations, not fixed catalog streams.
- Account-configuration write endpoints (creating/updating/deleting tags, custom fields,
  pipelines, milestones, lost reasons, categories, boards, stages, track definitions, activity
  types, titles, and goals) are excluded as `requires_elevated_scope`/`destructive_admin` —
  these mutate shared account configuration affecting every record on them, a materially
  different risk class from the per-record CRM writes this connector exposes.
- Entries (activity-log/history), tracks, webhooks, and the ad hoc `/filters/query` endpoint
  are out of scope: entries have no single stable `records.path` shape across the many entity
  types they attach to, and webhooks/tracks are configuration/integration surfaces rather than
  CRM record data.
- Attachment upload/download (`POST/GET .../attachments`) is excluded as `binary_payload`: the
  write dialect sends a single JSON/form body, not multipart, and attachment download returns
  a raw file, not a JSON record.
- No incremental sync is implemented for any stream, matching legacy exactly: legacy declares
  cursor fields in its catalog for forward compatibility but never reads them back to filter a
  request or a page of records. This bundle preserves that behavior rather than introducing
  new incremental filtering under the guise of a migration.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate
  limiting for Capsule CRM, so none is added here either (see
  `docs/migration/conventions.md` §3's rate_limit rule).
