# Overview

GoLogin is a declarative-HTTP, read-only connector that reads browser profiles, folders, tags,
and account information through the GoLogin REST API (`https://api.gologin.com`). This bundle is
a capability-parity migration of the legacy hand-written connector at
`internal/connectors/gologin`; the legacy package stays registered and unchanged until the wave6
registry flip.

## Auth setup

Provide a GoLogin API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged.

## Streams notes

Four streams, matching legacy's `gologinStreamEndpoints` routing table exactly:

- `profiles` — `GET /browser/v2`, records at `profiles`. Page-number pagination (`page` query
  param, 1-based, no page-size query parameter is ever sent — GoLogin's profiles list has no
  server-side page-size control). `page_size: 30` in this bundle's `pagination` block is the
  client-side short-page stop threshold only (matches legacy's `gologinDefaultPageSize`), not a
  request parameter (`size_param` is intentionally the empty string, matching legacy's
  `harvest`, which sends only `?page=N`). Legacy publishes `CursorFields: ["updatedAt"]` for this
  stream but never sends any server-side incremental filter (`InitialState` always starts at an
  empty cursor, and `Read` never references it) — so this bundle declares a bare
  `incremental.cursor_field: "updatedAt"` with no `request_param`, matching legacy's "sorts by
  cursor field for downstream dedup, but always reads every page" behavior exactly.
- `folders` — `GET /folders`, root-array response (`records.path: ""`), single page.
- `user` — `GET /user`, single root object (`records.path: ""`), single page. Legacy publishes
  `CursorFields: ["createdAt"]` with the same no-server-filter behavior as `profiles`, so this
  bundle declares the equivalent bare `incremental.cursor_field: "createdAt"`.
- `tags` — `GET /tags/all`, records at `tags`, single page.

`spec.json` does not declare `page_size`/`max_pages` config keys: GoLogin's pagination spec fields
(`PaginationSpec.PageSize`/`MaxPages`) are plain, non-templated integers in this engine dialect, so
neither value can be wired to a runtime config override — declaring config keys that no template
anywhere in the bundle consumes would be dead config (F6, `docs/migration/conventions.md`). This
mirrors legacy's own `gologinPageSize`/`gologinMaxPages` config-parsing helpers, which exist in
legacy Go code but have no equivalent expression path in this dialect; the fixed literal values
(`page_size: 30`, unbounded `max_pages`) reproduce legacy's actual defaults.

## Write actions & risks

None. GoLogin is read-only in this bundle, matching legacy exactly: `capabilities.write: false`
and no `writes.json` file. Legacy's `Write` method is a stub returning
`connectors.ErrUnsupportedOperation`; the GoLogin API has no reverse-ETL-safe write surface
documented.

## Known limits

- `page_size`/`max_pages` are not config-overridable in this bundle (see Streams notes above) —
  an accepted, documented scope narrowing versus legacy's config-parsing helpers, since the
  underlying values (30 / unbounded) are unchanged and legacy itself only ever varies them via
  config that most callers never set.
- Full GoLogin API surface (proxy management, browser automation triggers, team/workspace admin
  endpoints) is out of scope; see `api_surface.json`. Only the 4 legacy-parity read streams are
  implemented, matching `gologinStreamEndpoints`' exact route table.
