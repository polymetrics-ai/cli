# Overview

Coassemble is a headless-API read-only source. This bundle migrates
`internal/connectors/coassemble` (the hand-written legacy connector) to a
declarative defs bundle at capability parity: it reads courses, screen types,
and learner tracking records through the Coassemble headless REST API. The
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `user_id` and `user_token` secrets (both required). They are never
sent as separate header/query values; instead they are composed into a
single custom `Authorization` header matching Coassemble's documented
`COASSEMBLE-V1-SHA256 UserId=<user_id>, UserToken=<user_token>` scheme, via
the engine's `api_key_header` auth mode with a multi-reference `value`
template (`"COASSEMBLE-V1-SHA256 UserId={{ secrets.user_id }}, UserToken={{
secrets.user_token }}"`, `prefix: ""`). Neither secret is ever logged.

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

None of Coassemble's headless list endpoints expose an incremental cursor,
so every stream here is full refresh only (no `incremental` block anywhere
in `streams.json`), matching legacy's `CursorFields`-empty catalog exactly.

## Write actions & risks

Coassemble is read-only. Legacy's `Write` always returns
`ErrUnsupportedOperation`; this bundle ships no `writes.json`
(`capabilities.write: false`).

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
  ever consume (F6, conventions.md). A future capability-expansion pass could
  close this if the dialect grows config-templated pagination fields.
- `courses`/`trackings` fixtures ship a full 20-record first page (matching
  the real page-size threshold that triggers Coassemble's page-increment
  pagination to continue) plus a short second page, per conventions.md §4's
  2-page-fixture-when-paginated rule; `screen_types` ships a single
  unpaginated page.
