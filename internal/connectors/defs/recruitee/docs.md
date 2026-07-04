# Overview

Recruitee is a wave2 fan-out declarative-HTTP migration. It reads Recruitee job offers,
candidates, departments, sources, and tags through the core ATS REST API
(`GET https://api.recruitee.com/c/{company_id}/...`). This bundle targets capability parity with
`internal/connectors/recruitee` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip. Read-only: legacy's `Write` always returns
`connectors.ErrUnsupportedOperation`, and this bundle declares `capabilities.write: false` with no
`writes.json` to match.

## Auth setup

Provide a Recruitee personal API token via the `api_key` secret. It is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(key)` (`recruitee.go:149`)
exactly. Every request path is scoped under the required `company_id` config value
(`/c/{{ config.company_id }}/...`), matching legacy's own allow-listed
`"c/"+companyID+"/"+endpoint.path` construction (`recruitee.go:100`) rather than accepting an
arbitrary request path; the engine's `InterpolatePath` rejects `..`/traversal segments in the
resolved path exactly like legacy's own `cleanSegment` validation (`recruitee.go:236-241`), and
urlencodes the segment by default. `base_url` defaults to `https://api.recruitee.com` and may be
overridden for tests/proxies.

## Streams notes

Five streams, all primary-keyed on `id`: `offers`, `candidates`, `departments`, `sources`, `tags`.
Each is a flat list endpoint whose records live at a top-level key matching the stream name
(`connsdk.RecordsAt(resp.Body, endpoint.recordsPath)`, `recruitee.go:104`, where
`recordsPath == path` for every stream in legacy). Pagination is 1-based page-number
(`pagination.type: page_number`, `page_param: page`, `size_param: limit`, `start_page: 1`,
`page_size: 100` matching legacy's `recruiteeDefaultPageSize`) — the engine's
`PageNumberPaginator` stops on a short page (fewer records than the page size), matching legacy's
own `len(records) < pageSize` stop rule (`recruitee.go:116`) exactly; there is no other stop
condition on either side.

`offers`, `departments`, `sources`, and `tags` project their raw fields verbatim (`id`/`title`/
`status`/`created_at`/`updated_at` for offers; `id`/`name` for the other three), matching legacy's
`offerRecord`/`namedRecord` (`recruitee.go:180-188`) field-for-field via bare `computed_fields`
references, so the engine's typed extraction preserves `id`'s real wire type (Recruitee's REST API
returns numeric IDs as JSON integers). `candidates` similarly projects `id`/`name`/`email`/
`created_at`/`updated_at`, matching `candidateRecord`'s primary field set.

None of the five streams expose a server-side incremental filter parameter in legacy (`Read` never
sends a date-filter query param — `harvest` only ever sends `page`/`limit`), so this bundle declares
no `incremental` block for any stream, matching legacy exactly. `offers` and `candidates` still
declare `x-cursor-field: updated_at`, mirroring legacy's own `Catalog` `CursorFields` declaration
(`recruitee.go:172-173`) for informational/dedup-mode purposes only — per `docs/migration/
conventions.md` §2, `incremental_append` sync modes are gated on the presence of an `incremental`
block, not on `x-cursor-field` alone.

## Write actions & risks

None. Legacy's own `Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write`
is `false` and this bundle ships no `writes.json`.

## Known limits

- **Full-surface Pass B is quarantined.** The current Recruitee docs page
  (`https://apidocs.recruitee.com/`) was scraped on 2026-07-04 and exposes 948 documented
  method/path actions: 384 GET, 181 POST, 297 PATCH, 8 PUT, and 78 DELETE. The page is HTML-only
  for this purpose; no machine-readable OpenAPI/response schema was discoverable, and the surface
  spans ATS data plus admin, OAuth, billing, payment, SSO, security, attachment, and lifecycle
  controls. Because legacy only grounds the five list streams above and no write surface,
  `docs/migration/quarantine.json` records `recruitee` as `SCHEMA_AMBIGUOUS`. `api_surface.json`
  still enumerates all scraped endpoints and covers the five legacy streams, but excludes the
  remaining endpoints with closed categories instead of inventing stream schemas or write bodies.
- **`name`/`title`-fallback is not modeled.** Legacy's `candidateRecord` maps `name` from
  `first(item, "name", "full_name")` (a defensive OR: use `name` if present, else fall back to
  `full_name`), and `namedRecord` (departments/sources/tags) maps `name` from
  `first(item, "name", "title")` (`recruitee.go:183-197`). The engine's `computed_fields` dialect has
  no coalesce/fallback-between-two-paths primitive (only a single dotted-path reference or a filter
  chain over ONE reference); this bundle therefore projects only the primary field (`record.name`
  for all four streams), matching Recruitee's real, documented API field name in every case (the
  `full_name`/`title` branch is a defensive fallback for a shape not observed in Recruitee's current
  REST API — candidates and named resources consistently emit `name`). This narrows legacy's
  defensive-only fallback behavior, never its accepted-input behavior for the real API's actual
  wire shape.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes config-driven overrides
  (`recruiteeDefaultPageSize`/`recruiteeMaxPageSize`, `recruitee.go:243-265`). The engine's
  `page_number` paginator's `PageSize` is a bundle-declared constant (`streams.json`'s
  `base.pagination.page_size: 100`), with no per-request config-driven override mechanism. This
  bundle therefore fixes Recruitee's own default (`limit=100`) and does not declare `page_size`/
  `max_pages` in `spec.json` at all (a declared-but-unwireable config key is worse than an absent
  one, per the bitly/searxng/pagerduty F6 precedent).
- **Fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached when
  `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a fixed
  2-record set with synthetic `email`/`status` values regardless of stream (`recruitee.go:123-134`);
  this is a test-only affordance, not part of the live record shape. The engine's own
  conformance/fixture-replay harness provides the credential-free test affordance this bundle needs,
  so no fixture-mode equivalent is needed here.
