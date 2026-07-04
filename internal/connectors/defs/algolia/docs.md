# Overview

Algolia reads indices, API keys, index settings, dictionary settings/languages, security (vault)
sources, and API request logs, and writes index settings/API keys, through the Algolia Search REST
API (`https://<application_id>.algolia.net`). This bundle originally migrated
`internal/connectors/algolia` (the hand-written connector, read-only, 3 streams) to a declarative
defs bundle; this Pass B pass expands it to the full documented Algolia Search API surface
(`api-clients-automation`'s `specs/search/spec.yml`) — 7 read streams and 2 write actions — while
keeping every original legacy-parity stream and behavior unchanged.

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

None of the original 3 legacy streams are incremental — Algolia's management API supports
full-refresh sync only, matching legacy (no `CursorFields` declared on any stream).

**New in Pass B** (full-surface expansion — `api_surface.json` covers the complete Algolia Search
API OpenAPI spec):

- `vault_sources` — `GET /1/security/sources`, records at `.` (a bare top-level JSON array, unlike
  every other stream in this bundle — `connsdk.RecordsAt`'s `"."` path also returns an array's
  elements verbatim when the root node is already an array). Lists the IP address ranges allowed to
  access the application (`security`/`admin` ACL). Unpaginated; Algolia's own docs don't paginate
  this endpoint either.
- `dictionary_settings` — `GET /1/dictionaries/*/settings` (the literal `*` segment is part of
  Algolia's real path, not a template placeholder — it is a fixed wildcard route, not a per-dictionary
  resource path). Records at `.` (single flat object). Since the endpoint has no natural id of its
  own, `computed_fields` stamps a fixed literal `"id": "dictionary_settings"` (a static-literal
  computed field, matching the dialect's `searxng`-precedent marker-stamp pattern — see
  `docs/migration/conventions.md` §3) so the schema can declare `x-primary-key`.
- `dictionary_languages` — `GET /1/dictionaries/*/languages`, whose response is a JSON OBJECT keyed
  by language code (`{"en": {...}, "fr": {...}}`), not an array — modeled with `records.keyed_object:
  true` / `key_field: language` (the S4 engine mini-wave's keyed-object flatten, see
  `docs/migration/conventions.md` §3), stamping each language code onto its own record's `language`
  field. Each value carries nested `plurals`/`stopwords`/`compounds` sub-objects (each an
  optional `{"nbCustomEntries": N}` or `null`); these pass through schema projection unrenamed since
  their raw field names already match the schema exactly (no `computed_fields` needed).
- `logs` — `GET /1/logs`, records at `logs`, `offset_limit` paginated (`offset_param: offset`,
  `limit_param: length`, `page_size: 10` matching Algolia's own default `length`). Every field in
  Algolia's log-entry response is already snake_case (`answer_code`, `query_body`, `query_headers`,
  `processing_time_ms`, etc.) except the entry's own unique identifier, which Algolia calls `sha1`;
  `computed_fields` stamps that onto `id` (`x-primary-key`) via a bare `{{ record.sha1 }}` reference
  (typed extraction — the dialect's bare-single-reference rule, see `docs/migration/conventions.md`
  §3 — though `sha1` is already a string so typing is moot here). Logs are held for 7 days by
  Algolia and require the `logs` ACL on the configured `api_key`.

## Write actions & risks

Two write actions were added in this Pass B pass (`capabilities.write` flipped to `true`):

- `update_index_settings` — `PUT /1/indexes/{index_name}/settings`. Overwrites the named index's
  search settings (ranking, faceting, searchable attributes, `hitsPerPage`, etc.); settings omitted
  from the submitted record are left unchanged by Algolia itself (a partial-update semantic on
  Algolia's side, not this connector's), but any field that IS included replaces its current value
  immediately for live search traffic. `path_fields: ["index_name"]` — the record's `index_name`
  selects the target index and is not itself sent in the body.
- `create_api_key` — `POST /1/keys`. Creates a brand-new live Algolia API key with the requested
  `acl`/`indexes`/rate-limit scope. This is a genuinely new standing credential, not a mutation of an
  existing resource — see Known limits below for why key rotation/deletion is deliberately NOT
  covered alongside it.

Both actions require the `admin`/`editSettings` ACL on the configured `api_key` and are marked
`"approval": "required"` in `metadata.json.risk` — see each action's own `risk` string in
`writes.json` for the specific blast radius.

## Known limits

- **`indices`'s `nbPages`-based termination is approximated, not reproduced exactly.** Legacy stops
  when `page+1 >= nbPages` (an explicit total-pages count reported by the API), or on an empty page;
  the engine's `page_number` paginator stops purely on a short page (`recordCount < page_size`, here
  100). This is behaviorally identical for every real Algolia response (the API always returns
  exactly the server's default page size — 100 — except on the final page, which is necessarily
  short), so no records are silently dropped for any input legacy itself would accept — an
  ACCEPTABLE deviation (`docs/migration/conventions.md` §5 meta-rule), not an `ENGINE_GAP`.
- **`ENGINE_GAP`: synonyms, rules, and custom-dictionary-entry LISTING is not migrated.** Algolia's
  only way to list all synonyms/rules/dictionary-entries is a POST-with-body "search" endpoint
  (`searchSynonyms`/`searchRules`/`searchDictionaryEntries`, an empty query returns everything,
  page/hitsPerPage paginated) — the engine's declarative stream dialect has no wired mechanism to
  send a stream-level request body: `streams.json`'s `StreamSpec.Body` field is accepted by the
  bundle loader's JSON schema but never read anywhere in the read-path engine (`read.go`'s
  declarative loop always calls `rt.Requester.Do(...)` with a nil body, regardless of whether
  `stream.Body` is set). Implementing these 3 streams would require either an engine change (wiring
  `stream.Body` through the read path) or a Tier-2 `StreamHook`; both are out of scope for this Pass
  B pass. See `api_surface.json`'s 3 matching `excluded` entries for the exact endpoints.
- Record/object CRUD, search-query execution, index browse/export, batch operations, index
  copy/move, and per-task-id status polling are out of scope — these are data-plane (search
  results/records) or batch-shaped surfaces, not the config/metadata-plane resources this connector
  targets. See `api_surface.json`'s `out_of_scope` entries.
- Multi-cluster (userID mapping) endpoints are excluded as `deprecated` — Algolia's own OpenAPI spec
  marks every one of them `deprecated: true`.
- API key rotation/deletion, security-source append/delete, dictionary-settings mutation, and
  index-task-status/copy-move endpoints are excluded as `requires_elevated_scope`/`destructive_admin`
  — see `api_surface.json` for the specific risk reasoning per endpoint (mirroring the original
  wave2 exclusions for key creation/deletion, now joined by the newly-enumerated admin surface).
