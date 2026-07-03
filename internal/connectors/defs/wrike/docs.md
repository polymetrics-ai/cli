# Overview

Wrike is a wave2 fan-out declarative-HTTP migration. It reads tasks, folders, and contacts
through the Wrike REST API v4 (`GET {{ config.base_url }}/...`). This bundle is migrated from
`internal/connectors/wrike` (the hand-written connector it replaces); the legacy package stays
registered and unchanged until wave6's registry flip. Read-only (`capabilities.write` is `false`,
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a Wrike API access token via the `access_token` secret; it is sent as a Bearer token on
every request (`mode: bearer`), matching legacy's `connsdk.Bearer(token)`. `base_url` defaults to
`https://www.wrike.com/api/v4` (legacy's `defaultBaseURL`) and may be overridden for test proxies.

## Streams notes

All 3 streams (`tasks`, `folders`, `contacts`) share the identical envelope (records at the
top-level `data` array) and `page_number` pagination (`page`/`pageSize` query params, matching
legacy's `PageNumberPaginator{PageParam: "page", SizeParam: "pageSize", StartPage: 1}`). `page_size`
defaults to 100 (legacy's `defaultPageSize`); legacy bounds it to a max of 1000 (`maxPageSize`) and
`max_pages` defaults to 1 (legacy's `readMaxPages` default) when unset.

`tasks` (`GET /tasks`) emits `id`/`title`/`updatedDate`, matching legacy's field set exactly
(Wrike's own camelCase field name is preserved, not renamed — legacy itself emits `updatedDate`
verbatim, no `computed_fields` rename is needed). `folders` (`GET /folders`) emits the identical
shape. `contacts` (`GET /contacts`) emits `id`/`firstName`/`lastName` — legacy declares no cursor
field for this stream (`cursorFields: []`), so `contacts`' schema declares no `x-cursor-field`
either, matching that exactly. Primary key is `id` for every stream; `tasks`/`folders` declare
`updatedDate` as the incremental cursor field for manifest-surface parity, matching legacy's
`cursorFields`, though neither legacy nor this bundle actually issues a server-side incremental
filter — legacy's `Read` performs a full stream read every time regardless of any prior cursor.
All 3 streams declare `"projection": "passthrough"`: legacy's `Read` hands `connsdk.Harvest` a
callback that does `emit(connectors.Record(rec))` verbatim for every stream — no field-built
`connectors.Record{...}` mapping anywhere in `wrike.go` — so schema-mode projection would silently
drop any Wrike field beyond the three currently declared; passthrough reproduces legacy's actual
raw-emission behavior exactly, and the schema remains a documentation surface of the known shape.

## Write actions & risks

None. Legacy `wrike.go`'s `Write` returns `connectors.ErrUnsupportedOperation` unconditionally;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` config-driven overrides are not modeled.** Legacy reads
  `config["page_size"]` (bounded 1-1000) and `config["max_pages"]` (default 1) at request time via
  `boundedInt`/`readMaxPages`. The engine's `page_number` paginator reads `PaginationSpec.PageSize`
  from the static `streams.json` `base.pagination` block only — there is no per-request
  config-driven override mechanism for either value in the current dialect. `page_size`/`max_pages`
  remain declared in `spec.json` as documentation of legacy's accepted config surface, but neither
  is wired into any template in this bundle.
- **No incremental filter is modeled**, matching legacy: `updatedDate` is declared as
  `x-cursor-field` on `tasks`/`folders` for manifest parity, but Wrike's `/tasks` and `/folders`
  endpoints (as legacy calls them) accept no time-range query parameter — both connectors always
  perform a full stream read on every sync. `contacts` has no cursor field at all, matching
  legacy's empty `cursorFields` for that stream.
- The full Wrike API surface (task/folder/contact mutation, comments, timelogs, attachments) is
  out of scope for this wave; see `api_surface.json`'s `excluded` entries.
