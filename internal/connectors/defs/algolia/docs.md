# Overview

Algolia reads API keys and index settings through the Algolia Search REST API
(`https://<application_id>.algolia.net`). This bundle migrates `internal/connectors/algolia` (the
hand-written connector) to a declarative defs bundle. It is a read-only configuration/management API
connector (full-refresh only), matching legacy. **The `indices` stream is NOT migrated in this wave**
— see "Known limits" below for the specific engine gap blocking it.

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

Two streams are migrated, matching 2 of legacy's 3 `algoliaStreamEndpoints` entries:

- `api_keys` — `GET /1/keys`, records at `keys`, unpaginated (matches legacy's `readSingle`). Several
  raw camelCase fields (`createdAt`, `maxHitsPerQuery`, `maxQueriesPerIPPerHour`) are renamed to
  snake_case via `computed_fields`, matching legacy's `algoliaKeyRecord` mapping exactly.
- `index_settings` — `GET /1/indexes/{index_name}/settings`, records at `.` (the response is a flat
  object with no envelope — `connsdk.RecordsAt`'s `"."` path returns a single-object one-element
  record set, matching legacy's `recordsPath: "."`); requires config `index_name`. The configured
  `index_name` is stamped onto every emitted record via `computed_fields` (`{{ config.index_name
  }}`), matching legacy's `item["index_name"] = indexName` injection before mapping.

Neither stream is incremental — Algolia's management API supports full-refresh sync only, matching
legacy (no `CursorFields` declared on any stream).

**`indices` is NOT implemented — see Known limits.**

## Write actions & risks

None. Algolia is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's `Write`
always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Blocked: `indices` stream (`ENGINE_GAP`).** Legacy's `harvestIndices` pages Algolia's indices
  listing starting at page 0 (`for page := 0; ...`), which is Algolia's real 0-based page-number
  convention for this endpoint. The engine's `page_number` pagination type delegates directly to
  `connsdk.PageNumberPaginator`, whose `Start()` method is `p.page = p.StartPage; if p.page == 0 {
  p.page = 1 }` — this unconditionally coerces an explicit `start_page: 0` to `1`, because a Go `int`
  cannot distinguish "the zero value because start_page was never set" from "explicitly configured to
  0". Every live read of `indices` would therefore silently begin at Algolia's real SECOND page
  (`page=1`), permanently skipping the first page's records — this is an accepted-input EMITTED-DATA
  change (real records silently dropped), not a cosmetic/request-count deviation, so it fails the
  parity-deviation meta-rule (`docs/migration/conventions.md` §5) and cannot be shipped as a
  documented-acceptable deviation. No Tier-1 workaround exists (`PaginationSpec` fields are plain JSON,
  never templated, so there is no config-driven escape hatch either), and this wave's hard rules forbid
  writing a Tier-2 hook to patch around it. Confirmed by direct execution against the real engine
  (`engine.Read` against a live `httptest.Server`) during authoring, not merely inferred from reading
  `connsdk.PageNumberPaginator`'s source. This is an `ENGINE_GAP` blocker for a follow-up
  engine-dialect increment (`PaginationSpec` needs a way to distinguish "start_page unset" from
  "start_page explicitly 0", e.g. a pointer/explicit-presence field or a documented sentinel), not a
  per-connector patch. Once closed, `indices` should follow the exact shape legacy uses:
  `pagination.type: page_number`, `page_param: page`, `size_param: ""` (Algolia never sends a
  page-size param — the server default applies), `start_page: 0`; several raw camelCase fields
  (`createdAt`, `dataSize`, `fileSize`, `lastBuildTimeS`, `numberOfPendingTasks`, `pendingTask`,
  `updatedAt`) need `computed_fields` renames to the snake_case schema (see legacy's
  `algoliaIndexRecord`); `nbPages`-based termination has no dialect-native stop condition either (a
  `page_size`-threshold short-page stop is the closest approximation and was verified separately to be
  an acceptable, data-preserving deviation on its own — see the removed `indices` stream's prior
  authoring notes in this bundle's git history) — but that approximation is moot until the 0-based
  start itself is fixed.
- API key creation/deletion and analytics endpoints are out of scope for this wave; see
  `api_surface.json`'s `excluded` entries (`requires_elevated_scope`/`destructive_admin` for key
  mutation, `out_of_scope` for the blocked `indices` endpoint and Pass B deferrals).
