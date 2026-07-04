# Overview

Salesflare is a Pass B full-surface-expansion declarative-HTTP migration. It reads Salesflare
accounts, contacts, opportunities, users, tags, tasks, workflows, groups, stages, pipelines,
persons, currencies, custom-field types, and email data sources, and writes CRM lifecycle
mutations, through the Salesflare REST API. This bundle targets capability parity with
`internal/connectors/salesflare` (the hand-written connector it migrates, `Capabilities{Write:
false}`) as a **superset**: legacy's original 3 read streams (`accounts`, `contacts`,
`opportunities`) are preserved unchanged, and 11 new read streams plus 22 new write actions are
added against the full documented Salesflare API v1.0.0 surface (47 paths, live OpenAPI/Swagger 2.0
spec fetched from `https://api.salesflare.com/openapi.json`, 2026-07-03). The legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Salesflare API key via the `api_key` secret. It is sent as `Authorization: Bearer
<api_key>` via `base.auth`'s `mode: bearer` — identical to legacy's `connsdk.Bearer(token)`. Never
logged. `base_url` defaults to `https://api.salesflare.com`, matching legacy's own in-code default.

## Streams notes

**Legacy-parity streams** (`accounts`, `contacts`, `opportunities`) are unchanged from the prior
wave. All 3 share the same shape: `GET` against the Salesflare list endpoint, records at `data`
(`records.path: "data"`), primary key `["id"]`. `projection: passthrough` is used on every stream
in this bundle (legacy-parity and new alike) because legacy's own `Read` re-emits each decoded
record verbatim (`emit(connectors.Record(rec))` at `salesflare.go:126`) with no field filtering at
all; schema projection alone (which would silently drop any undeclared field) would be a
data-parity regression. The declared `schemas/*.json` properties are still a realistic, honest
field set (for `records_match_schema`'s type-checking of the fields that ARE declared) but do not
gate which fields survive.

Pagination for `accounts`/`contacts`/`opportunities` is `cursor` with `token_path:
pagination.next_page` and `cursor_param: page`: legacy's `readPages` reads `pagination.next_page`
and treats a non-empty value as the next page's literal `page` query param value; a `null`/absent
`next_page` stops pagination. `limit=100` is a fixed literal on these 3 streams, matching legacy's
`defaultPageSize = 100`; `max_pages` is a fixed `100` in `streams.json`'s `base.pagination.max_pages`
(legacy's `defaultMaxPages = 100`). Neither is a runtime config override — see Known limits.

**New streams added this pass** (all `GET`, `records.path: "data"`, `projection: passthrough` for
the same field-preservation reasoning as the legacy-parity streams):

- `users` (`/users`), `tags` (`/tags`), `tasks` (`/tasks`), `workflows` (`/workflows`) —
  `offset_limit` pagination (`limit`/`offset` query params, per the real OpenAPI spec's documented
  parameters for these 4 endpoints — a different pagination convention than the legacy-parity
  streams' `pagination.next_page` cursor shape).
- `groups` (`/groups`), `stages` (`/stages`), `pipelines` (`/pipelines`), `persons` (`/persons`),
  `currencies` (`/currencies`), `custom_field_types` (`/customfields/types`), `email_data_sources`
  (`/datasources/email`) — unpaginated (`pagination: none`); the real OpenAPI spec documents no
  pagination query parameters for any of these 7 endpoints (small, low-cardinality reference/
  configuration collections).

None of the 14 streams declare an `incremental` block, matching legacy's `Catalog` (no
`CursorFields`) and the real OpenAPI spec's lack of a documented updated-since filter parameter on
any of these list endpoints.

## Write actions & risks

`capabilities.write` is now `true` (22 actions added; legacy shipped none — legacy's `Write` always
returns `connectors.ErrUnsupportedOperation`):

- **Accounts**: `create_account` (low-risk, no approval), `update_account` (approval required),
  `delete_account` (destructive/irreversible, approval required).
- **Contacts**: `create_contact` (low-risk, no approval), `update_contact` (approval required),
  `delete_contact` (destructive/irreversible, approval required).
- **Opportunities**: `create_opportunity` (low-risk, no approval), `update_opportunity` (may change
  stage/close state, approval required), `delete_opportunity` (destructive/irreversible, approval
  required).
- **Tags**: `create_tag` (low-risk, no approval), `update_tag` (approval required), `delete_tag`
  (destructive — removes the tag from every record it's applied to, approval required).
- **Tasks**: `create_task` (low-risk, no approval), `update_task` (may mark complete, approval
  required), `delete_task` (destructive/irreversible, approval required).
- **Meetings**: `create_meeting` (low-risk, no approval), `update_meeting` (approval required),
  `delete_meeting` (destructive/irreversible, approval required).
- **Calls**: `create_call` (logs a call activity against an account; low-risk, no approval).
- **Internal notes (messages)**: `create_internal_note` (low-risk, no approval),
  `update_internal_note` (approval required), `delete_internal_note` (destructive/irreversible,
  approval required).

All create/update/delete triples for accounts/contacts/opportunities/tags/tasks/meetings and the
internal-note triple follow the same `body_type: json`, `path_fields: ["id"]` (or the
resource-specific id field name, e.g. `meeting_id`/`message_id`) shape; every `delete_*` action
declares `delete.missing_ok_status: [404]` (idempotent delete, matching the engine's delete
semantics).

## Known limits

- **Account-to-contact/user membership mutations are excluded.** `POST`/`PUT
  /accounts/{account_id}/contacts` and `POST`/`PUT /accounts/{account_id}/users` are list-membership
  add/replace operations, not a single-record create/update/delete on a syncable resource; the
  dialect's write shapes (`body_type` over a fixed field set, `path_fields`) do not model
  list-membership mutations.
- **Per-item-class custom-field CRUD is excluded.** `GET`/`POST`/`PUT`/`DELETE
  /customfields/{itemClass}[/...]` requires a caller-supplied `itemClass` path segment
  (`accounts`/`contacts`/`opportunities`/etc.) chosen per-call, not a fixed resource this dialect's
  static path templating can express as one stream/write.
- **Per-account/per-tag sub-resource reads are excluded as Pass B breadth-vs-cost triage.**
  `GET /accounts/{account_id}/feed`, `GET /accounts/{account_id}/messages`, and
  `GET /tags/{tag_id}/usage` would each require a fan_out read issuing one request per account/tag
  id; not implemented this pass.
- **Settings/configuration singletons are excluded.** `GET`/`PUT /settings/ai` and
  `GET /filterfields/{entity}` are account-wide configuration, not syncable data records.
  `POST /message/{message_id}/feedback` (and its duplicate `/messages/{message_id}/feedback`) is
  AI-message quality feedback, not a CRM data mutation.
- **Workflow authoring/audience-control mutations are excluded.** `POST /workflows`,
  `PUT /workflows/{id}`, and `PUT /workflows/{id}/audience/{record_id}` are marketing-automation
  builder-tool operations (full trigger/step-graph authoring, runtime audience control) with no
  documented flat request schema suited to a plain CRM record create/update.
- **Detail-by-id GETs that duplicate an already-covered list stream are excluded** (accounts,
  contacts, opportunities, groups, stages, tags, users, workflows) — see `api_surface.json`'s
  `duplicate_of` entries.
- **`page_size`/`max_pages` are not runtime config overrides for the legacy-parity streams, and
  neither is declared in `spec.json`.** Legacy accepts both as config-driven overrides (`page_size`
  unbounded, `max_pages` also accepting `all`/`unlimited` sentinels for "no cap"). The engine's
  declarative pagination has no per-read config-driven override mechanism for either
  `PaginationSpec.PageSize` or `PaginationSpec.MaxPages` — both are fixed values baked into
  `streams.json`. This bundle bakes in `max_pages: 100` (legacy's own default) and a fixed
  `limit: 100` query literal for `accounts`/`contacts`/`opportunities`; a caller wanting a different
  page size or cap (or "unlimited") cannot express it here. Documented config-surface narrowing,
  never a data-parity change for any read actually exercised at the default.
- **Legacy's URL/path-shaped `next_page` fallback branches are not modeled.** Legacy's `readPages`
  treats a `next_page` value starting with `http://`, `https://`, or `/` as a full next-page
  URL/path override rather than a literal `page` param value. Salesflare's real API (and legacy's
  own `salesflare_test.go` fixture) only ever returns a bare integer page number at
  `pagination.next_page`, so this branch is unreachable dead code in practice.
- Legacy declares no incremental/cursor-field behavior for any of its 3 streams (no `CursorFields`
  in its `Catalog()`); this bundle matches that for the legacy-parity streams, and extends the same
  no-incremental-filter behavior to every new stream (the real OpenAPI spec documents no
  updated-since filter parameter on any endpoint covered here).
