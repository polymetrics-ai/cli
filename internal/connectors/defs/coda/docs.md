# Overview

Coda is a read-only source: it reads docs and doc-scoped objects (tables,
pages, formulas, controls) from the Coda REST API v1. This bundle migrates
`internal/connectors/coda` (the hand-written legacy connector) to a
declarative defs bundle at capability parity; the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Coda API token via the `auth_token` secret; it is sent only as
`Authorization: Bearer <auth_token>` and never logged.

## Streams notes

- `docs` (`GET /docs`) — workspace-level list of every doc the token can
  access. No `doc_id` config needed.
- `tables`/`pages`/`formulas`/`controls` (`GET /docs/{{ config.doc_id
  }}/tables|pages|formulas|controls`) — doc-scoped lists; `doc_id` is a
  required config value for these 4 streams specifically (legacy's
  `resolvePath` returns an error when `doc_id` is unset for a doc-scoped
  stream) even though it is not declared in `spec.json`'s top-level
  `required` array (matching legacy's per-stream, not global, requirement).
  Each of these 4 streams stamps `doc_id` onto every emitted record via
  `computed_fields: {"doc_id": "{{ config.doc_id }}"}`, matching legacy's
  `out["doc_id"] = docID` post-map stamp exactly.

All 5 streams share `cursor`/`token_path` pagination (`pageToken` query
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

Coda is read-only in legacy (`Write` always returns
`ErrUnsupportedOperation`); this bundle ships no `writes.json`
(`capabilities.write: false`).

## Known limits

- Legacy also exposed a runtime-configurable `max_pages` (0/all/unlimited or
  a positive integer) config key. `PaginationSpec.MaxPages` is a fixed value
  in this engine's dialect, not template-resolvable at read time, so it is
  not declared in `spec.json` (F6, conventions.md) — matching the same
  documented gap as coassemble/searxng's `max_pages`.
- Only the 5 legacy-parity read streams are implemented; Coda's row-level
  endpoints (`/docs/{docId}/tables/{tableIdOrName}/rows`, read and write)
  were never implemented in legacy either and are out of scope for this
  migration (`api_surface.json`'s `excluded: {category: out_of_scope}`
  entries) — a Pass B capability expansion, not a parity gap.
