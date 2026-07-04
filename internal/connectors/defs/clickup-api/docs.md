# Overview

ClickUp-api reads ClickUp workspaces (teams), spaces, folders, lists, tasks, goals, space tags,
and webhooks, and writes task/folder/list/space/webhook lifecycle mutations, task comments, tags,
custom field values, and goal creation, through the ClickUp v2 REST API
(`https://api.clickup.com/api/v2`). It began as a wave2 fan-out migration of
`internal/connectors/clickup-api` (the hand-written connector it migrates; the legacy package
stays registered and unchanged until wave6's registry flip) and was expanded to 19 write actions
and 3 new read streams in a Pass B pass, researched against `developer.clickup.com`'s full
OpenAPI-backed API Reference (~140 documented endpoints as of 2026-07-04).
`capabilities.write` is now `true`.

## Auth setup

Provide a ClickUp personal API token via the `api_token` secret; it is sent RAW in the
`Authorization` header with no `Bearer ` prefix (`mode: api_key_header`, `header: Authorization`,
`prefix: ""`), matching legacy's `connsdk.APIKeyHeader("Authorization", token, "")` exactly (a
ClickUp personal token convention, not OAuth Bearer). `base_url` defaults to
`https://api.clickup.com/api/v2` and may be overridden for tests/proxies.

## Streams notes

Streams are declared `tasks`/`teams`/`spaces`/`folders`/`lists` in this bundle's `streams.json`
(reordered from legacy's own catalog order `teams, spaces, folders, lists, tasks` purely so the
one genuinely-paginated stream, `tasks`, is exercised by conformance's `pagination_terminates`
dynamic check, which always picks the bundle's first declared stream — catalog position carries
no legacy-parity meaning in this dialect, unlike stream NAMES, which are unchanged).

- **`teams`** (`GET /team`): unparameterised, matching legacy's default stream.
- **`spaces`** (`GET /team/{{ config.team_id }}/space`): requires config `team_id` (an absent
  `team_id` hard-errors on both sides — legacy: `"clickup-api stream spaces requires config
  team_id"`; engine: an unresolved `config.team_id` path-template key — same failure
  classification, different literal text, per conventions.md §5's config-validation-parity
  precedent).
- **`folders`**/**`lists`** (`GET /space/{{ config.space_id }}/folder` and `.../list`): both
  require config `space_id` (same absent-key parity pattern as `spaces`/`team_id`). Both stamp a
  `space_id` computed field from the raw API's nested `{{ record.space.id }}` reference (ClickUp
  nests a partial space object on folder/list records), matching legacy's own
  `nestedID(item["space"])` helper.
- **`tasks`** (`GET /team/{{ config.team_id }}/task`): requires config `team_id`. Computed fields
  extract every nested-object id ClickUp returns (`status` from `{{ record.status.status }}`,
  `creator_id` from `{{ record.creator.id }}`, `list_id`/`folder_id`/`space_id` from their
  respective nested `.id` references) — bare single-reference `computed_fields` entries copy the
  RAW typed value (so `creator_id` stays a JSON integer, matching ClickUp's real wire shape and
  legacy's own `nestedID` passthrough), matching legacy's `clickupTaskRecord`/`statusName`/
  `nestedID` exactly.

Pagination (`tasks` only; every other stream is a single unpaginated read, matching legacy's
`fetchOnce` for those 4 streams): `pagination.type: page_number`, `page_param: page`, **no
`size_param`** (legacy never sends a page-size query parameter — ClickUp's task list endpoint
returns a fixed ~100 items per page server-side, confirmed by ClickUp's own docs: "Responses are
limited to 100 tasks per page"). `streams.json` pins `page_size: 100`, matching that real
server-side page size, as the client-side short-page-stop threshold (no query param is ever sent
for it — `size_param` is absent — so this value only ever decides when the paginator stops, never
what is requested). The required first-stream 2-page fixture proof (conventions.md §4) follows
chargify's/clarif-ai's identical precedent: page 1 returns exactly 100 tasks (a full page, so the
paginator advances) and page 2 returns 1 (a short page, so it terminates).

Legacy's real stop signal is `last_page == "true" || len(records) == 0`; the engine's
`page_number` paginator has no `stop_path`-equivalent field (that is exclusive to the `cursor`
pagination type in this dialect), so this bundle relies purely on the short-page-stop heuristic —
DATA-identical to legacy in the overwhelmingly common case, diverging only in emitted REQUEST
COUNT (never emitted RECORD data) for a final page landing on an exact multiple of the declared
threshold.

**`start_page` is declared `0`, matching ClickUp's real 0-indexed first page.** Legacy's own loop
is `for page := 0; page < clickupMaxPages; page++` — ClickUp's task list endpoint is genuinely
0-indexed (`page=0` is the real first page). This was originally blocked by an `ENGINE_GAP`
(`PaginationSpec.StartPage` was a plain Go `int`, and `engine/paginate.go`'s `newPaginator`
unconditionally coerced an explicit `"start_page": 0` to `1`, since the zero value could not be
distinguished from "never set" — the identical gap class documented in `algolia`'s and `datadog`'s
`docs.md`). That gap is now closed (S4 engine mini-wave item 1): `PaginationSpec.StartPage` is a
`*int`, so an explicit `start_page: 0` is honored verbatim rather than coerced, and this bundle now
reads `tasks` starting from ClickUp's real first page — full read parity for every stream in this
bundle (`teams`, `spaces`, `folders`, `lists`, `tasks`).

`tasks` publishes `date_updated` as its cursor field, matching legacy catalog metadata, but does
not declare a server-side request parameter or client-side filter. ClickUp's real task list
endpoint exposes no updated-since filter parameter, and legacy's own `Read`/`harvestPaged` emits
every record returned by the paged full scan.

`archived`/`include_closed_tasks` config values are forwarded verbatim as the literal query value
when present (`omit_when_absent`/`default` dialect), matching legacy's own `fetchOnce`
(`archived=true` sent only when configured true, otherwise omitted entirely — ClickUp's own docs
confirm `archived` defaults server-side to `false`, so omission and an explicit `archived=false` are
DATA-equivalent) and `harvestPaged` (`archived` ALWAYS sent, `true` or `false`; `include_closed`
sent only when true). See Known limits for the narrowed accepted-value vocabulary this requires.

**New in this pass** (researched against `developer.clickup.com`'s per-endpoint reference pages,
2026-07-04):

- **`goals`** (`GET /team/{{ config.team_id }}/goal`): unparameterised beyond `team_id`; primary
  key `["id"]`; no incremental cursor (the documented response has no updated-since filter, and
  goals are typically edited in place, not append-only). Records at the response's `goals` key
  (the same response also carries a sibling `folders` key for Goal Folders, not modeled here).
- **`space_tags`** (`GET /space/{{ config.space_id }}/tag`): ClickUp Tags have no `id` field at
  all — `name` IS the tag's identity within a Space (confirmed via the live endpoint's response
  shape: `name`/`tag_fg`/`tag_bg` only) — so `x-primary-key` is `["name"]`, not `["id"]`. A
  `computed_fields` entry stamps the Space id (`{{ config.space_id }}`) onto every record so a
  destination can distinguish same-named tags across different Spaces.
- **`webhooks`** (`GET /team/{{ config.team_id }}/webhook`): returns webhooks created by the
  authenticated user/token, per ClickUp's own docs; primary key `["id"]`, no incremental cursor
  (no updated-since filter documented).

## Write actions & risks

19 write actions, added in this Pass B pass, researched against `developer.clickup.com`'s
per-endpoint reference pages (each cited path/method/body was independently confirmed, 2026-07-04):

- **`create_task`** (`POST /list/{{ config.list_id }}/task`) / **`update_task`**
  (`PUT /task/{{ record.id }}`) / **`delete_task`** (`DELETE /task/{{ record.id }}`,
  `missing_ok_status: [404]`): standard task lifecycle. `create_task` requires config `list_id`
  (tasks are created within a specific List); `update_task`/`delete_task` are pure
  path-parameterized on the task's own `id`. `delete_task`/`update_task` require approval;
  `create_task` is low-risk (additive).
- **`create_task_comment`** (`POST /task/{{ record.task_id }}/comment`,
  `body_fields: ["comment_text", "notify_all", "assignee", "group_assignee"]`): adds a comment;
  low-risk.
- **`add_tag_to_task`** (`POST /task/{{ record.task_id }}/tag/{{ record.tag_name }}`) /
  **`remove_tag_from_task`** (`DELETE /task/{{ record.task_id }}/tag/{{ record.tag_name }}`,
  `missing_ok_status: [404]`): both pure path-parameterized (no body at all); the tag must already
  exist at the Space level (see `space_tags`). Low-risk — `remove_tag_from_task` un-attaches a tag
  from one task only and does not delete the tag definition itself (ClickUp's own docs: "This does
  not delete the Tag from the Space").
- **`set_custom_field_value`** (`POST /task/{{ record.task_id }}/field/{{ record.field_id }}`,
  `body_fields: ["value", "value_options"]`): ClickUp's accepted `value` shape is a discriminated
  union keyed by the Custom Field's own type (text/number/date/dropdown/label/people/task-
  relationship/manual-progress/location/button — see the confirmed shapes in the research this
  action is grounded on); this bundle accepts any JSON value for `value` and forwards it verbatim
  rather than re-implementing that 9-way discriminated union in schema form, since the record
  author is expected to already know the target field's type. **Approval required** since an
  incorrectly-typed value can silently fail (`400` if the field isn't enabled for the task's
  `custom_item_id`) or write a value ClickUp accepts but the operator did not intend.
- **`create_goal`** (`POST /team/{{ config.team_id }}/goal`): low-risk (additive).
- **`create_folder`** (`POST /space/{{ config.space_id }}/folder`) / **`update_folder`**
  (`PUT /folder/{{ record.id }}`, rename only) / **`delete_folder`**
  (`DELETE /folder/{{ record.id }}`, `missing_ok_status: [404]`): standard Folder lifecycle.
  `delete_folder` cascades to every List/task inside it — approval required, as is `update_folder`;
  `create_folder` is low-risk.
- **`create_list`** (`POST /folder/{{ config.folder_id }}/list`) / **`update_list`**
  (`PUT /list/{{ record.id }}`) / **`delete_list`** (`DELETE /list/{{ record.id }}`,
  `missing_ok_status: [404]`): standard List lifecycle, requires config `folder_id` for creation
  (Lists are created within a specific Folder — the folder-scoped `createlist` endpoint, not the
  Space-scoped folderless-list-creation variant). `delete_list` cascades to every task inside it —
  approval required, as is `update_list`; `create_list` is low-risk.
- **`create_space`** (`POST /team/{{ config.team_id }}/space`) / **`update_space`**
  (`PUT /space/{{ record.id }}`) / **`delete_space`** (`DELETE /space/{{ record.id }}`,
  `missing_ok_status: [404]`): standard Space lifecycle. ClickUp's own docs mark every
  `update_space` body field required (name/color/private/admin_can_manage/multiple_assignees/
  features) rather than a true partial-update PATCH — this bundle's `record_schema` only requires
  `id`+`name` (looser than the live API, never stricter, matching the stripe-precedent
  `minProperties`-style permissiveness in conventions.md §5) since the engine has no way to
  pre-populate the other required fields from a prior read automatically; an operator sending a
  partial record risks the live API resetting unspecified feature toggles to their defaults —
  documented in Known limits. `delete_space` cascades to every Folder/List/task inside it —
  approval required, as is `update_space`; `create_space` is low-risk.
- **`create_webhook`** (`POST /team/{{ config.team_id }}/webhook`) / **`update_webhook`**
  (`PUT /webhook/{{ record.id }}`) / **`delete_webhook`** (`DELETE /webhook/{{ record.id }}`,
  `missing_ok_status: [404]`): registers/repoints/removes an outbound event-delivery URL of the
  caller's choosing (bitly's identical `create_webhook`/`update_webhook` risk precedent) —
  approval required for all three.

## Known limits

- **`include_archived`/`include_closed_tasks` accept only the literal strings `"true"`/`"false"`
  (or absent), narrower than legacy's `boolConfig` helper (which also accepts `"1"`/`"yes"` as
  truthy synonyms).** The engine's query-templating dialect has no boolean-normalization filter —
  a `{{ config.include_archived }}` reference forwards whatever string the operator set verbatim
  as the query value, with no equivalent of legacy's `strings.EqualFold`-based truthy-alias
  matching. Declaring the accepted vocabulary as `"true"`/`"false"` only (documented here rather
  than silently accepting-and-mismapping `"1"`/`"yes"`) keeps every ACCEPTED input's emitted
  records identical to legacy; an operator who previously relied on `"1"`/`"yes"` must switch to
  the literal `"true"` string. This is a documented config-surface narrowing, not a data-shape
  regression for any `"true"`/`"false"`/absent input.
- **`tasks` pagination has no `last_page`-boolean stop signal wired.** The engine's `page_number`
  paginator type supports only a short-page stop threshold (no `stop_path`, which is exclusive to
  the `cursor` pagination type in this dialect) — see Streams notes above for why this is
  data-identical to legacy in every case that matters (an over-threshold "extra" request against an
  already-exhausted endpoint returns zero records and stops immediately either way).
- **`update_space`'s required-fields gap**: see Write actions & risks above — the live API treats
  every `update_space` body field as required, but this bundle's `record_schema` only requires
  `id`+`name` to avoid forcing every caller to always re-supply every ClickApp toggle. An operator
  should always populate the full desired feature set on every `update_space` call, not rely on
  partial-update semantics ClickUp's own API does not actually provide.
- Chat, Docs, Views, Guests (Enterprise-plan-gated), User Groups, Templates, legacy+new Time
  Tracking, Attachments (multipart/form-data, not expressible in this dialect's json/form/none
  body types), Checklists, Dependencies/Task Links, and Custom Field/Custom Task Type discovery
  remain out of scope this pass — each is a distinct product area from ClickUp's core
  project-management data model (or, for Attachments, a real dialect limitation) rather than an
  oversight; see `api_surface.json`'s per-endpoint `excluded` entries for the specific category
  and reason (never a blanket "Pass B" bucket).
