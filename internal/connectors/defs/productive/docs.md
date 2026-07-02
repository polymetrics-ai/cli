# Overview

Productive is a wave2 fan-out declarative-HTTP migration. It reads Productive projects, people,
companies, and tasks through the Productive JSON:API-style REST API
(`GET https://api.productive.io/api/v2/<resource>`). This bundle targets capability parity with
`internal/connectors/productive` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Productive API token via the `api_key` secret; it is sent as the `X-Auth-Token` header
(matching legacy's `connsdk.APIKeyHeader("X-Auth-Token", token, "")`, an unprefixed header value),
and is never logged. Productive additionally requires an `organization_id` config value on every
request (legacy: `secret()`/config lookup that hard-errors when absent) — this bundle declares it as
a required `spec.json` property and sends it verbatim as the `X-Organization-Id` header on every
request (not conditionally omitted like Stripe's optional `account_id`, since legacy treats it as
mandatory, not optional). `base_url` defaults to `https://api.productive.io/api/v2`.

## Streams notes

All four streams share the identical Productive JSON:API envelope: `GET /api/v2/<resource>` returns
`{"data":[{"id","type","attributes":{...}}],"meta":{"current_page","total_pages"}}`, records live at
the top-level `data` array. `id` and `type` are top-level JSON:API fields that survive schema
projection directly; `name`, `created_at`, and `updated_at` live under `attributes` and are lifted to
top-level output fields via `computed_fields` (`{{ record.attributes.name }}` etc.), matching
legacy's `mapRecord` which spreads every `attributes` key onto the emitted record.

Pagination is `page_number` (`page`/`per_page`, `start_page: 1`, static `page_size: 100` matching
legacy's `defaultPageSize`). The engine's `page_number` paginator stops on a short page
(`recordCount < page_size`); legacy instead reads `meta.total_pages` and stops when
`page >= total_pages`, falling back to short-page-only behavior only via that same total_pages
check never being non-zero. These two stop conditions are equivalent for every dataset except one
whose total record count is an exact multiple of 100, where legacy stops immediately via the
`total_pages` comparison and the engine would issue one additional request returning an empty page
before stopping — no different records are ever emitted either way (same non-data-affecting
divergence documented on this wave's aha/adobe-commerce-magento bundles).

Every Productive object publishes `updated_at` as `x-cursor-field` (matching legacy's own
`CursorFields: []string{"updated_at"}`), but Productive's list endpoints expose no server-side
incremental filter parameter that legacy ever wires, and legacy's own `Read` never applies one —
every read is a full paginated sweep regardless of any prior sync's cursor. This bundle therefore
declares `incremental.cursor_field` with no `request_param`/`start_config_key`/`client_filtered`: the
cursor field is published (enabling `incremental_append_deduped` sync-mode eligibility for downstream
consumers) without the engine ever computing or sending a filter, matching legacy's true read
behavior exactly.

## Write actions & risks

None. Legacy `productive` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- **`name` fallback to `title` is not modeled.** Legacy's `mapRecord` sets
  `rec["name"] = first(item, "name", "title")` — if a record's `attributes.name` is absent, it falls
  back to `attributes.title`. This bundle's `computed_fields` template
  (`{{ record.attributes.name }}`) has no coalesce/fallback mechanism in the current dialect (no
  `first`/`coalesce` filter exists), so only `attributes.name` is read; a resource whose only naming
  field is `title` (none of Productive's projects/people/companies/tasks resources use `title` in
  practice — only `name`) would emit a null `name` here where legacy would emit the title. Documented
  scope narrowing, not expected to affect real data for these 4 resource types.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-200,
  default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides. The engine's `page_number` paginator has no config-driven page-size or request-count-cap
  knob (mirrors this wave's aha precedent and stripe's resolved ledger item 3); neither is declared in
  `spec.json`, and this bundle sends Productive's own default (`per_page=100`) as a static
  pagination-block value.
- **`total_pages`-based early stop is approximated by short-page stop only** — see Streams notes
  above; the only observable difference is one extra empty-page request when a stream's total record
  count is an exact multiple of 100, never a difference in which records are emitted.
- **Legacy's `raw` passthrough field is not modeled.** Legacy's `mapRecord` stashes the entire raw
  JSON:API item under a `raw` key on every emitted record; this bundle's schema projection keeps only
  the declared parity fields (`id`, `type`, `name`, `created_at`, `updated_at`), matching the "schema
  is a field-for-field projection of legacy's own `mapRecord` output" convention — `raw` was an
  internal escape hatch, never itself consumed by legacy's own `streams()` field catalog.
