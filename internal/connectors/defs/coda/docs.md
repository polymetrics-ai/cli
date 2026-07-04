# Overview

Coda reads docs and doc-scoped objects (tables, rows, columns, pages,
formulas, controls) from the Coda REST API v1, and writes table rows and
doc pages back. This bundle migrated `internal/connectors/coda` (the
hand-written, read-only legacy connector) to a declarative defs bundle at
capability parity, then (Pass B) expanded it to the full documented public
Coda API v1 surface: 7 streams and 8 write actions, up from the original 5
read-only streams. The legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide a Coda API token via the `auth_token` secret; it is sent only as
`Authorization: Bearer <auth_token>` and never logged.

## Streams notes

- `docs` (`GET /docs`) — workspace-level list of every doc the token can
  access. No `doc_id` config needed.
- `tables`/`pages`/`formulas`/`controls` (`GET /docs/{{ config.doc_id
  }}/tables|pages|formulas|controls`) — doc-scoped lists; `doc_id` is a
  required config value for these streams specifically (legacy's
  `resolvePath` returns an error when `doc_id` is unset for a doc-scoped
  stream) even though it is not declared in `spec.json`'s top-level
  `required` array (matching legacy's per-stream, not global, requirement).
  Each stamps `doc_id` onto every emitted record via `computed_fields:
  {"doc_id": "{{ config.doc_id }}"}`, matching legacy's `out["doc_id"] =
  docID` post-map stamp exactly.
- `columns`/`rows` (`GET /docs/{{ config.doc_id
  }}/tables/{{ fanout.id }}/columns|rows`) — Pass B additions. Coda has no
  single "all rows across all tables" endpoint; rows and columns are always
  scoped to one table. Both streams use the `fan_out` dialect
  (conventions.md §3) over the doc's own table list (`ids_from.request`
  re-reads `/docs/{{ config.doc_id }}/tables`, `records_path: items`,
  `id_field: id`) so every table in the doc is walked automatically —
  `into.path_var: table_id` substitutes each table id into the stream's own
  `path`, and `stamp_field: table_id` stamps it onto every emitted row/
  column record (alongside the same `doc_id` computed-field stamp every
  other doc-scoped stream carries). `rows` requests `valueFormat:
  simpleWithArrays` so multi-select/array cell values survive as JSON
  arrays rather than Coda's comma-joined `simple` string default; `values`
  is schema-typed as a passthrough `object` (a hash of column-id → cell
  value) since Coda's own cell-value union (`Value`: scalar or array of
  scalars, keyed by an arbitrary, per-doc column id) has no fixed field set
  to enumerate.

All 7 streams share `cursor`/`token_path` pagination (`pageToken` query
param, `nextPageToken` response-body path), matching legacy's
`connsdk.CursorPaginator{CursorParam: "pageToken", TokenPath:
"nextPageToken"}`. `limit` is sent via the optional-query dialect
(`{"template": "{{ config.page_size }}", "default": "25"}`) so
`config.page_size` genuinely drives the request page size at runtime,
falling back to legacy's own default (25) when unset — unlike coassemble/
searxng's `page_number.page_size`, `stream.Query` params ARE templated
against config, so this knob is fully wireable here.

Coda's list endpoints expose no incremental cursor; every stream here is
full refresh only, matching legacy's `CursorFields`-empty catalog.

## Write actions & risks

Pass B capability expansion — legacy shipped no writes at all
(`capabilities.write` flips `false` → `true` in this bundle).

- `upsert_rows` (`POST /docs/{{ config.doc_id }}/tables/{{ record.table_id
  }}/rows`) — inserts new rows, or upserts existing ones by `keyColumns`,
  in one request (`rows: [{cells: [{column, value}, ...]}, ...]`). Async:
  Coda always answers `202` and applies the change within seconds. No
  approval required (bounded, additive).
- `update_row` (`PUT .../rows/{{ record.row_id }}`) — overwrites named cell
  values on one existing row (`row: {cells: [...]}`); unincluded cells are
  left unchanged. No approval required.
- `delete_row` / `delete_rows` (`DELETE .../rows/{{ record.row_id }}` /
  `DELETE .../rows` with a JSON `rowIds` body) — irreversible; approval
  required for both.
- `push_button` (`POST .../rows/{{ record.row_id
  }}/buttons/{{ record.column_id }}`) — pushes a button column on a row.
  Coda's own docs warn the underlying button can perform ANY action the
  doc's formulas define, including writes to other tables and Pack actions
  entirely outside this connector's declared surface; approval required.
- `create_page` (`POST /docs/{{ config.doc_id }}/pages`) — creates a new
  page (optionally a subpage via `parentPageId`); requires Doc Maker access
  in the workspace upstream. No approval required.
- `update_page` (`PUT .../pages/{{ record.page_id }}`) — renames/re-icons/
  hides an existing page. No approval required.
- `delete_page` (`DELETE .../pages/{{ record.page_id }}`) — irreversibly
  removes a page and its subpages/content; approval required.

All row/page write actions return `202` (queued for async processing);
this connector does not poll `/mutationStatus/{requestId}` for completion
(see Known limits).

## Known limits

- Legacy also exposed a runtime-configurable `max_pages` (0/all/unlimited or
  a positive integer) config key. `PaginationSpec.MaxPages` is a fixed value
  in this engine's dialect, not template-resolvable at read time, so it is
  not declared in `spec.json` (F6, conventions.md) — matching the same
  documented gap as coassemble/searxng's `max_pages`.
- Write actions do not poll `/mutationStatus/{requestId}`: Coda's row/page
  mutation endpoints all return `202` immediately and apply the change
  asynchronously (generally within seconds); this connector reports the
  write as accepted, not as confirmed-applied. Polling for completion would
  require a second declarative request per write with no dialect support
  for a post-write wait/poll loop; every other async-202 write-capable
  bundle in this migration (e.g. airtable) makes the identical choice.
- `getRow`/`getTable`/`getColumn`/`getFormula`/`getControl`/`getDoc`
  single-detail endpoints are not modeled as separate streams
  (`api_surface.json`'s `duplicate_of` entries): each returns a strict
  subset of its list stream's own per-record shape, so no record data is
  lost by reading the list endpoint instead.
- Doc lifecycle (create/update/delete a doc), sharing/permissions/ACL,
  publishing, custom domains, and doc-authored automation triggers are all
  out of scope (`api_surface.json`'s `requires_elevated_scope`/
  `destructive_admin`/`out_of_scope` entries) — these are workspace-admin
  or opaque-blast-radius actions, not bounded per-row/per-page content
  mutations.
- Page content export (`POST/GET .../pages/{pageId}/export{,/​{requestId}}`)
  is excluded as `binary_payload`: it returns a downloadable HTML/Markdown
  blob via an async job, not a declarative JSON read/write.
