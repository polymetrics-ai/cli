# Overview

Aha! is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the full flat-list-endpoint
Aha! surface. It reads Aha! features, products, ideas, releases, initiatives, goals, epics, and users
through the Aha! REST API (`GET https://<account>.aha.io/api/v1/...`). This bundle originally
targeted capability parity with `internal/connectors/aha` (the hand-written connector it migrates,
which is read-only); Pass B's full-surface research (`api_surface.json`) goes beyond that legacy
parity baseline per docs/migration/conventions.md's Pass B scope, though writes remain unimplemented
— see Write actions & risks below for the `ENGINE_GAP` blocking every Aha! mutation. The legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an Aha! API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's treatment of Aha! API keys "exactly like
OAuth bearer tokens" (`aha.go:252`, `connsdk.Bearer(secret)`), and is never logged. `base_url`
defaults to `https://secure.aha.io` (legacy's `ahaDefaultBaseURL`) but Aha! is account-scoped
(`<company>.aha.io`), so almost every real deployment overrides it — see Known limits for legacy's
`url` config alias, which this bundle does not model.

## Streams notes

All eight streams share the identical Aha! page-number envelope: `GET /api/v1/<resource>` returns
`{"<resource>":[...],"pagination":{"total_records":N,"total_pages":N,"current_page":N}}`, records
live at the resource-named top-level key (identical to the resource segment for every stream here,
including the two new Pass B streams below). Pagination is `page_number` (`page`/`per_page`,
`start_page: 1`, static `page_size: 30` matching legacy's `ahaDefaultPageSize`). The engine's
`page_number` paginator stops on a short page (`recordCount < page_size`); legacy instead reads the
`pagination.total_pages`/`current_page` object and falls back to short-page stop only when
`total_pages` is absent. These two stop conditions are equivalent for every dataset except one whose
total record count is an exact multiple of 30, where legacy stops immediately via the `total_pages`
comparison and the engine would issue one additional request returning an empty page before stopping
— no different records are ever emitted either way (same non-data-affecting divergence documented on
this wave's adobe-commerce-magento bundle).

Every stream stamps a static-literal `resource` marker field (`"feature"`/`"product"`/`"idea"`/
`"release"`/`"initiative"`/`"goal"`/`"epic"`/`"user"`) via `computed_fields`, matching legacy's own
hardcoded `"resource": "feature"` (etc.) in each `mapRecord` function for the 6 original streams, and
extending the same convention to the 2 new streams below.

Every Aha! object publishes `updated_at` as `x-cursor-field` (matching legacy's own
`CursorFields: []string{"updated_at"}`), but Aha!'s list endpoints expose no server-side
incremental filter parameter and legacy's own `harvest` never applies one — every read is a full
paginated sweep regardless of any prior sync's cursor. This bundle therefore declares
`incremental.cursor_field` with no `request_param`/`start_config_key`/`client_filtered`: the cursor
field is published (enabling `incremental_append_deduped` sync-mode eligibility for downstream
consumers) without the engine ever computing or sending a filter, matching legacy's true read
behavior exactly.

**New (Pass B) streams**: `epics` (`GET /api/v1/epics`) shares the identical shape, incremental
cursor, and pagination as every pre-existing stream. `users` (`GET /api/v1/users`) has no `updated_at`
timestamp in Aha!'s own user object, so it declares no `incremental` block at all (bare full sync
every time, matching the "no cursor field exists" case conventions.md §8 rule 2 describes, not an
approximation of one).

Other real Aha! resources were researched and NOT added this pass because their real endpoint is
nested under an already-covered parent (comments, to-dos, votes, competitors, teams, release_phases
— each enumerable only per-feature/per-product/per-idea/etc., which would need a `fan_out` hop
layered on an already-fanned-out or already-covered parent stream) or because the flat top-level
list endpoint's exact real path could not be confirmed with sufficient confidence from the
documentation sources available this pass (a to-dos-specific ambiguity — see `api_surface.json`).
See `api_surface.json` for the complete reasoned exclusion list.

## Write actions & risks

None — `ENGINE_GAP`. Every Aha! v1 mutation endpoint (feature/idea/release/initiative/goal/epic/
product create and update) requires the request body to be wrapped in a resource-named JSON envelope
key, e.g. `PUT /api/v1/features/{id}` with body `{"feature": {"name": "New name", ...}}`, not a flat
`{"name": "New name", ...}` body. This dialect's write-body construction (`write.go`'s
`buildJSONBody`) always sends the record's own fields (minus `path_fields`) as the TOP-LEVEL JSON
body — there is no declarative mechanism anywhere in `streams.json`/`writes.json` to wrap that body
in an outer named key. Declaring a `writes.json` action here would silently send a body Aha!'s real
API does not expect (and, per Aha!'s own documented behavior, would likely be rejected or
misinterpreted rather than applied) — exactly the kind of workaround-that-diverges-from-real-behavior
conventions.md §5's meta-rule forbids shipping. `capabilities.write` stays `false`; this bundle ships
no `writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation` (which was
already correct for a different reason — legacy's own package doc: "there is no obvious safe
reverse-ETL write surface" — this Pass B research additionally confirms the dialect could not express
the write shape even if a safe write surface were chosen). If the engine dialect later grows a
`body_wrap`/envelope-key mechanism, every Aha! create/update endpoint documented in
`api_surface.json`'s `ENGINE_GAP`-reasoned exclusions becomes immediately implementable without
further research.

## Known limits

- **The `url` config alias for `base_url` is not modeled.** Legacy accepts either `base_url` or a
  bare `url` config key as an alias (`ahaBaseURL`, `aha.go:268-289`) — either name works
  interchangeably. This bundle declares only `base_url`; an operator previously using the `url` key
  name must rename it. This is a config-surface naming narrowing only — the resolved effective base
  URL behavior once configured is identical.
- **`per_page`/`page_size` and `max_pages` config overrides are not modeled.** Legacy exposes
  `per_page` (aliased from `page_size`, 1-200, default 30) and `max_pages`
  (0/all/unlimited or a positive integer cap) as config-driven overrides (`ahaPageSize`/
  `ahaMaxPages`). The engine's `page_number` paginator has no config-driven page-size or
  request-count-cap knob (mirrors this wave's adobe-commerce-magento precedent and stripe's
  resolved ledger item 3); `per_page`/`max_pages` are therefore not declared in `spec.json`, and
  this bundle sends Aha!'s own default (`per_page=30`) as a static pagination-block value.
- **`total_pages`-based early stop is approximated by short-page stop only** — see Streams notes
  above; the only observable difference is one extra empty-page request when a stream's total
  record count is an exact multiple of 30, never a difference in which records are emitted.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (reached only
  when `config.mode == "fixture"`) stamps a `previous_cursor` field echoing `req.State["cursor"]`
  when set (`aha.go:227-229`), a credential-free conformance-harness affordance with no live-path
  equivalent. This bundle's schemas and fixtures target the live record shape only; the engine's own
  `internal/connectors/conformance` fixture-replay harness provides the credential-free test
  affordance this bundle needs.
- **`users` declares no incremental cursor field.** Unlike every other Aha! object modeled here,
  Aha!'s user resource carries no `updated_at` (or any other) timestamp field in its documented
  shape — there is genuinely no cursor to publish, not an omitted one. `users`'s schema therefore has
  no `x-cursor-field` and its stream declares no `incremental` block at all; every read is a full,
  non-incremental sweep, which is the correct representation per conventions.md §8 rule 2 ("neither
  → no incremental block") rather than a narrowing of existing behavior (this is new Pass B
  coverage with no legacy precedent to diverge from).
