# Overview

Pipeliner is a wave2 fan-out migration of `internal/connectors/pipeliner` (the hand-written
connector it replaces). It reads Pipeliner CRM accounts, contacts, opportunities, and leads through
the documented REST entity-list API. This bundle is read-only, matching legacy exactly; the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `username` and `password` as secrets; both are sent as HTTP Basic auth credentials
(`Authorization: Basic base64(username:password)`) and are never logged. A `space_id` config value
(the Pipeliner space/account id) is required and is interpolated into every request path
(`/spaces/{space_id}/entities/<Resource>`).

## Streams notes

All 4 streams (`accounts`, `contacts`, `opportunities`, `leads`) share the identical shape, matching
legacy's `streamEndpoints` table: `GET /spaces/{space_id}/entities/<Resource>`, records at the
top-level `data` array, primary key `["id"]`. Pagination is offset+limit
(`pagination.type: offset_limit`, `limit_param: limit`, `offset_param: offset`, `page_size: 100`,
matching legacy's `defaultPageSize`) — the engine stops when a page returns fewer records than
`page_size`, identical to legacy's own `len(records) < pageSize` stop condition.

`updated_at` is a `computed_fields` rename from the raw API's `modified` field — legacy's own
`entityRecord` mapper tries `updated_at`/`modified`/`Modified`/`UpdateDate` in that order, but
legacy's own unit test (`pipeliner_test.go`) fixes the real wire shape as lowercase `modified`, so
this bundle projects from `modified` as the confirmed-real field name (see Known limits for the
unexercised PascalCase fallbacks).

Legacy never sends an incremental filter parameter for any stream (no `request_param`-driven
cursor), so no `incremental` block is declared here either — every stream is full-refresh only,
matching legacy's actual (non-incremental) read behavior exactly. `updated_at` is exposed as a
plain field for downstream dedup/sort use, not as a declared cursor.

## Write actions & risks

None. Pipeliner is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- Legacy's `entityRecord`/`first(item, "id", "Id", "ID")`-style field lookup defensively tries
  PascalCase (`Id`/`Name`/`Status`/`Modified`/`UpdateDate`) and camelCase (`ID`) variants alongside
  lowercase field names, because the exact wire casing was never independently confirmed against a
  live Pipeliner account at authoring time. Legacy's own unit test
  (`internal/connectors/pipeliner/pipeliner_test.go`) is the strongest available ground truth and
  fixes the real API response shape as lowercase (`id`, `name`, `modified`) — this bundle's schema
  projection uses exactly those field names. The engine's `computed_fields` dialect has no
  coalesce/fallback-across-multiple-source-keys mechanism (a single bare `{{ record.<path> }}`
  reference names exactly one source field), so the defensive PascalCase/camelCase fallback paths
  are not reachable here; this only changes behavior if a live Pipeliner deployment actually emits
  PascalCase keys, which legacy's own test evidence contradicts. Documented per the parity-deviation
  ledger as an accepted, unreachable-in-practice narrowing rather than a silent gap.
- Full Pipeliner entity surface (activities, projects, documents, custom entities) and all
  write/mutation endpoints are out of scope for this wave; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
- `page_size`/`max_pages` config overrides from legacy (`intConfig` reading `config.page_size`/
  `config.max_pages`) have no runtime-config-driven equivalent in this engine dialect
  (`PaginationSpec.PageSize`/`MaxPages` are bundle-fixed values, never read from `RuntimeConfig` —
  see `docs/migration/conventions.md`'s searxng precedent for the identical narrowing) — `page_size`
  and `max_pages` are therefore not declared in `spec.json` at all (a declared-but-unwireable key is
  worse than an absent one, per the F6 dead-config rule) rather than accepted but silently ignored.
