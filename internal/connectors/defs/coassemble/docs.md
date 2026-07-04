# Overview

Coassemble is a headless-API source with course-lifecycle and identity/
translation writes. This bundle migrated `internal/connectors/coassemble`
(the hand-written, read-only legacy connector) to a declarative defs bundle
at capability parity, then (Pass B) expanded it to the full documented
Coassemble headless API surface: 9 streams (up from 3) and 16 write actions
(up from none), covering Courses, Tracking, Identities, Collections, and
Translations. The legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide `user_id` and `user_token` secrets (both required). They are never
sent as separate header/query values; instead they are composed into a
single custom `Authorization` header matching Coassemble's documented
`COASSEMBLE-V1-SHA256 UserId=<user_id>, UserToken=<user_token>` scheme, via
the engine's `api_key_header` auth mode with a multi-reference `value`
template (`"COASSEMBLE-V1-SHA256 UserId={{ secrets.user_id }}, UserToken={{
secrets.user_token }}"`, `prefix: ""`). Neither secret is ever logged. (The
current public API reference additionally documents a newer
workspace-level `Authorization: COASSEMBLE:<workspaceId>:<apiKey>` scheme;
this bundle keeps legacy's original per-user scheme unchanged, since it is
the one actually implemented/working — parity is preserved rather than
switched to an alternate, unverified auth shape.)

## Streams notes

- `courses` (`GET /api/v1/headless/courses`) — paginated with `page_number`
  pagination (`page`/`length` params, 1-based, page size 20 matching
  legacy's `coassembleDefaultPageSize`); stops on a short/empty page exactly
  like legacy's `harvest` loop. Primary key `["id"]`.
- `screen_types` (`GET /api/v1/headless/screen/types`) — legacy serves this
  as a single un-paginated array (`endpoint.paginated: false`); this stream
  overrides the base pagination with `{"type": "none"}` so no `page`/`length`
  params are ever sent, matching legacy exactly. No stable primary key is
  published (legacy's own catalog entry carries no `PrimaryKey`), so only
  `full_refresh_append` applies.
- `trackings` (`GET /api/v1/headless/trackings`) — same `page_number`
  pagination as `courses`. No stable primary key is published, matching
  legacy's catalog entry.
- `collections`/`clients`/`users` (Pass B additions, `GET
  /api/v1/headless/collections|clients|users`) — the API reference documents
  these with a 0-indexed `page` default, unlike `courses`/`trackings`'s
  1-indexed convention inherited from legacy; each stream overrides the base
  pagination with its own `page_number` block (`start_page: 0, page_size:
  100`, matching the endpoint's own documented `length` default) rather than
  inheriting the base's 1-indexed courses/trackings convention. Primary keys:
  `collections` → `["id"]`; `clients` → `["clientIdentifier"]`; `users` →
  `["identifier"]`.
- `user_trackings` (Pass B addition, `GET
  /api/v1/headless/user/trackings?identifier=...`) — Coassemble has no
  single "all user trackings" endpoint; it is always scoped to one learner
  identifier. Uses the `fan_out` dialect (conventions.md §3) over this
  bundle's own `users` stream (`ids_from.request` re-reads `GET
  /api/v1/headless/users`, `records_path: ""`, `id_field: identifier`);
  `into.query_param: identifier` sends each user's identifier as the query
  parameter, and `stamp_field: identifier` re-stamps it onto the emitted
  record (the response body already carries its own `identifier` field, so
  this is a harmless re-stamp of the same value, not new data). The response
  is a single object per user (totals + a nested `trackings` array), so
  `records.path: ""` yields exactly one record per fanned-out user, matching
  `connsdk.RecordsAt`'s "bare object = one whole-object record" rule.
- `collection_trackings` (Pass B addition, `GET
  /api/v1/headless/collection/trackings?id=...`) — same `fan_out` shape,
  scoped over this bundle's own `collections` stream (`id_field: id`);
  `into.query_param: id`, `stamp_field: collection_id`. Unlike
  `user_trackings`, the response here IS a list of per-learner tracking rows
  for that collection, so this stream can genuinely paginate per-collection.
- `translations` (Pass B addition, `GET
  /api/v1/headless/translations/{{ fanout.id }}`) — fanned out over this
  bundle's own `courses` stream (`id_field: id`) via `into.path_var:
  course_id`; the underlying endpoint takes no `page`/`length` params at all
  (`pagination: {"type": "none"}` at both the id-listing request — since the
  fan_out id-listing request reuses the SURROUNDING stream's own effective
  pagination spec, not `courses`'s base-level 1-based pagination — and the
  per-course translations request itself). Composite primary key `["id",
  "language"]` (one row per source-course × translated-language pair).

None of Coassemble's headless list endpoints expose an incremental cursor,
so every stream here is full refresh only (no `incremental` block anywhere
in `streams.json`), matching legacy's `CursorFields`-empty catalog exactly.

## Write actions & risks

Pass B capability expansion — legacy shipped no writes at all
(`capabilities.write` flips `false` → `true` in this bundle).

- `publish_course` / `duplicate_course` / `restore_course` — course
  lifecycle actions with documented, bounded effects. No approval required.
- `delete_course` — soft-deletes a course (Coassemble retains it for
  `restore_course`); approval required regardless, since the connector has
  no way to express "recoverable within some window" as a lesser risk tier.
- `delete_tracking` — permanently erases one learner's progress record for
  a course (`DELETE /api/v1/headless/tracking`, body `{id, identifier}`);
  irreversible, approval required.
- `create_collection` — creates a new collection of courses. No approval
  required.
- `delete_collection` / `restore_collection` — soft-delete/restore pair,
  same shape as courses; `delete_collection` requires approval.
- `update_client` / `delete_client` — `update_client` overwrites a client's
  `metadata` bag (no approval required); `delete_client` irreversibly
  removes a multi-tenant sub-account (approval required).
- `update_user` / `delete_user` — `update_user` overwrites a learner's
  `name`/`avatar`/`metadata`/`clientIdentifier` (no approval required).
  `delete_user` removes a learner identity; see Known limits for why the
  real endpoint's optional `action`/`reallocateTo`/`clientIdentifier` query
  params are NOT exposed by this action (approval required regardless).
- `translate_course` / `set_default_translation` / `sync_translation` /
  `delete_translation` — course-translation lifecycle. `translate_course`
  (kicks off machine translation into a new language) and
  `set_default_translation`/`sync_translation` require no approval;
  `delete_translation` (removes a language variant) requires approval.

## Known limits

- Legacy exposed runtime-configurable `page_size` (1-100, default 20) and
  `max_pages` (0/all/unlimited or a positive integer) config keys. Neither
  `PaginationSpec.PageSize` nor `PaginationSpec.MaxPages` in this engine's
  dialect is template-resolvable at read time — both are fixed values baked
  into `streams.json`'s pagination blocks at bundle-authoring time (the same
  documented gap as searxng's `page_size`/`max_pages`, conventions.md §"read-
  only, no-auth variant"). This bundle bakes in legacy's real *default*
  behavior (`page_size: 20`, unbounded pages) rather than declaring
  `config.page_size`/`config.max_pages` keys that no template anywhere could
  ever consume (F6, conventions.md).
- `courses`/`trackings` fixtures ship a full 20-record first page (matching
  the real page-size threshold that triggers Coassemble's page-increment
  pagination to continue) plus a short second page, per conventions.md §4's
  2-page-fixture-when-paginated rule; `screen_types` and every Pass B
  addition ship a single natural page (the bundle's 2-page proof obligation
  is already satisfied by `courses`, and `pagination_terminates`'
  exact-fixture-consumption assertion only ever exercises the bundle's
  FIRST declared stream, `courses`).
- `screen_trackings` (`GET /api/v1/headless/screen/trackings`) is NOT
  implemented: it requires a "screen id" the API reference never defines a
  list/lookup endpoint for (distinct from the `screen_types` lookup enum
  this connector already exposes, which is a fixed set of UI-block type
  names, not per-instance screen ids) — there is no verifiable source to
  fan out over, so guessing one would risk silently reading nothing or the
  wrong resource (`api_surface.json`'s `out_of_scope` entry).
- `delete_user`'s real endpoint accepts optional `action` (reallocate/
  delete/ignore), `reallocateTo`, and `clientIdentifier` query parameters
  controlling what happens to the deleted learner's course progress.
  This action does not expose them: the write-action path/query dialect has
  no way to send an optional field only when the caller supplies it (no
  `omit_when_absent`-style tolerance exists for `WriteAction.Path`, unlike
  `stream.Query`), and Coassemble's own docs do not fully specify the three
  actions' exact semantics beyond "choose what to do with any courses
  associated with this identifier" — silently defaulting an ambiguous,
  irreversible per-learner-data-retention choice would be worse than
  declaring the narrower, always-safe action out of scope for those extra
  params.
- Deprecated/binary/content-authoring endpoints (`course/generate`,
  `course/url`, `course/{action}`, `course/scorm/{id}`, `course/revert`,
  `course/content/{id}`) are excluded (`api_surface.json`'s `deprecated`/
  `binary_payload`/`out_of_scope` entries) rather than implemented.
