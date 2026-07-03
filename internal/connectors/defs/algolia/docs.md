# Overview

Algolia reads indices, API keys, and index settings through the Algolia Search REST API
(`https://<application_id>.algolia.net`). This bundle migrates `internal/connectors/algolia` (the
hand-written connector) to a declarative defs bundle. It is a read-only configuration/management API
connector (full-refresh only), matching legacy.

## Auth setup

Provide `api_key` (an Algolia Admin API key) as a secret, sent as the `X-Algolia-API-Key` header, and
`application_id` (a plain config value), sent as the `X-Algolia-Application-Id` header. Neither is
ever logged.

`base_url` is **required** in this bundle (`https://<application_id>.algolia.net` when following
legacy convention) rather than derived automatically from `application_id`, as legacy's
`algoliaBaseURL` does at runtime — the engine's `spec.json` `"default"` mechanism only materializes a
FIXED literal value, not one computed from another config field, so an application-id-derived default
cannot be expressed without inventing ad hoc Go (see `docs/migration/conventions.md`'s spec-defaults
section). This is a documented, honest config-surface narrowing, not a silent behavior change: set
`base_url` to `https://<application_id>.algolia.net` (or your proxy/test override) explicitly.

## Streams notes

All 3 of legacy's `algoliaStreamEndpoints` entries are now migrated:

- `indices` — `GET /1/indexes`, records at `items`, page-number paginated (`page_param: page`,
  `size_param: ""` — Algolia never sends a page-size param on this endpoint, the server default
  applies — `start_page: 0`, `page_size: 100` used only as the engine's client-side short-page stop
  threshold, `max_pages: 100` matching legacy's `algoliaDefaultMaxPages`). Algolia's real indices
  listing is genuinely 0-indexed (legacy's `for page := 0; ...`); this is now expressible thanks to
  the S4 engine mini-wave's `PaginationSpec.StartPage *int` (distinguishes an explicit `start_page:
  0` from an omitted one — see `docs/migration/conventions.md` §3). Legacy's own stop condition is
  `nbPages`-based (`page+1 >= nbPages`); the dialect has no body-driven total-pages stop primitive,
  so the engine's ordinary `page_number` short-page stop (`recordCount < page_size`) is used instead
  — a data-preserving approximation (every real Algolia response returns exactly the server's
  default page size except on the final, necessarily-short page), not a data-dropping one. Several
  raw camelCase fields (`createdAt`, `updatedAt`, `dataSize`, `fileSize`, `lastBuildTimeS`,
  `numberOfPendingTasks`, `pendingTask`) are renamed to snake_case via `computed_fields`, matching
  legacy's `algoliaIndexRecord` mapping exactly.
- `api_keys` — `GET /1/keys`, records at `keys`, unpaginated (matches legacy's `readSingle`). Several
  raw camelCase fields (`createdAt`, `maxHitsPerQuery`, `maxQueriesPerIPPerHour`) are renamed to
  snake_case via `computed_fields`, matching legacy's `algoliaKeyRecord` mapping exactly.
- `index_settings` — `GET /1/indexes/{index_name}/settings`, records at `.` (the response is a flat
  object with no envelope — `connsdk.RecordsAt`'s `"."` path returns a single-object one-element
  record set, matching legacy's `recordsPath: "."`); requires config `index_name`. The configured
  `index_name` is stamped onto every emitted record via `computed_fields` (`{{ config.index_name
  }}`), matching legacy's `item["index_name"] = indexName` injection before mapping.

None of the 3 streams are incremental — Algolia's management API supports full-refresh sync only,
matching legacy (no `CursorFields` declared on any stream).

## Write actions & risks

None. Algolia is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's `Write`
always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`indices`'s `nbPages`-based termination is approximated, not reproduced exactly.** Legacy stops
  when `page+1 >= nbPages` (an explicit total-pages count reported by the API), or on an empty page;
  the engine's `page_number` paginator stops purely on a short page (`recordCount < page_size`, here
  100). This is behaviorally identical for every real Algolia response (the API always returns
  exactly the server's default page size — 100 — except on the final page, which is necessarily
  short), so no records are silently dropped for any input legacy itself would accept — an
  ACCEPTABLE deviation (`docs/migration/conventions.md` §5 meta-rule), not an `ENGINE_GAP`.
- API key creation/deletion and analytics/search endpoints are out of scope for this wave; see
  `api_surface.json`'s `excluded` entries (`requires_elevated_scope`/`destructive_admin` for key
  mutation, `out_of_scope` for Pass B deferrals).
