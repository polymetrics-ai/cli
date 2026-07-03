# Overview

Capsule CRM is a Tier-1 declarative-HTTP migration of `internal/connectors/capsule-crm`
(legacy Go package `capsulecrm`). It reads Capsule CRM v2 parties, opportunities, cases
(`kases`), tasks, and users through the Capsule v2 REST API. The connector is read-only:
legacy exposes no allow-listed reverse-ETL writes for Capsule CRM.

## Auth setup

Provide a Capsule CRM personal access token via the `bearer_token` secret; it is used only
for Bearer auth (`Authorization: Bearer <bearer_token>`) and is never logged.

## Streams notes

All 5 streams (`parties`, `opportunities`, `kases`, `tasks`, `users`) share the same shape:
`GET` against the Capsule v2 list endpoint, records unwrapped from the top-level JSON key
matching the resource name (e.g. `{"parties": [...]}`), primary key `["id"]` (a numeric
Capsule id). Pagination follows Capsule's 1-based `page`/`perPage` convention
(`pagination.type: page_number`, `page_param: page`, `size_param: perPage`, `start_page: 1`),
stopping on a short/empty page exactly like legacy's `connsdk.PageNumberPaginator`.

`streams.json`'s `pagination.page_size` is a static JSON int (`PaginationSpec.PageSize`),
resolved once at bundle load with no config-driven override — the same shape documented in
`auth0`'s and `searxng`'s goldens (`docs/migration/conventions.md`). Legacy's real default is
50 (`capsuleDefaultPageSize`, configurable up to 100 via a `page_size` config value); this
bundle declares `page_size: 50` to match that default exactly (a smaller static value would
silently narrow every live page fetch to a fraction of legacy's request size, multiplying the
number of API calls per sync even though fixture replay would look identical either way). The
required 2-page `parties` conformance fixture (`docs/migration/conventions.md` §4) accordingly
ships a full 50-record page 1 and a short 1-record page 2 to exercise the real stop threshold.
Legacy's `page_size`/`max_pages` config properties are consequently genuinely
dead in this dialect (no template anywhere reads them) and are intentionally NOT declared in
`spec.json` (F6, REVIEW.md: a declared-but-unwireable key is worse than an absent one).

Legacy's stream catalog declares `CursorFields: []string{"updatedAt"}` for every stream, but
`capsulecrm.Read` never actually filters by it — every read is a full pass over the
collection (full-refresh only, no `incremental` request param or client-side filtering
exists in legacy). This bundle mirrors that exactly: each schema declares
`x-cursor-field: updated_at` (preserving the catalog metadata parity) but intentionally
declares **no `incremental` block** in `streams.json` — adding one would introduce new,
behavior-changing server-side or client-side filtering legacy never performed.

Several legacy record mappers flatten nested objects into `_id`/`_name` scalar fields
(`opportunities`' `party`/`milestone`/`value` objects, `kases`' `party` object, `tasks`'
`category`/`party`/`opportunity`/`kase` objects). This bundle reproduces every one of these
via `computed_fields` dotted-path references against the raw record (e.g.
`"party_id": "{{ record.party.id }}"`, `"value_amount": "{{ record.value.amount }}"`) — each
is a bare single `{{ record.<path> }}` reference with no filter stage, so the engine's typed
extraction preserves the nested id's native integer type rather than stringifying it,
matching legacy's raw `item["id"]` (an `int`/`float64` from JSON) assignment exactly.
Capsule's camelCase wire fields (`firstName`, `organisationName`, `createdAt`, `updatedAt`,
etc.) are renamed to legacy's snake_case output field names the same way, via bare
`computed_fields` references.

## Write actions & risks

None. Capsule CRM is read-only in this bundle, matching legacy's `Capabilities.Write: false`
(legacy's own package comment: "the API has no obviously-safe reverse-ETL surface to
allow-list").

## Known limits

- Only the 5 legacy-parity read streams are implemented; the broader Capsule v2 surface
  (tags, custom fields, users/current, mutating endpoints) is out of scope for this wave —
  see `api_surface.json`'s `excluded` entries.
- No incremental sync is implemented, matching legacy exactly: legacy declares cursor
  fields in its catalog for forward compatibility but never reads them back to filter a
  request or a page of records. This bundle preserves that behavior rather than introducing
  new incremental filtering under the guise of a migration.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate
  limiting for Capsule CRM, so none is added here either (see
  `docs/migration/conventions.md` §3's rate_limit rule).
