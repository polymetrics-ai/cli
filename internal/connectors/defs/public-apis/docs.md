# Overview

Public APIs is a Tier-1 declarative-HTTP migration. It reads public API directory entries from
the api.publicapis.org directory API (`GET https://api.publicapis.org/entries?limit=<n>&offset=<n>`).
This bundle is migrated from `internal/connectors/public-apis` (the hand-written connector); the
legacy package stays registered and unchanged until wave6's registry flip. The upstream API is
read-only, credential-free, and public.

## Auth setup

No credentials are required: `api.publicapis.org` is a fully public, unauthenticated directory
API, matching legacy's own no-auth `connsdk.Requester` construction (`public_apis.go:163-169`).
`base_url` defaults to `https://api.publicapis.org` and may be overridden for tests/proxies.

## Streams notes

`entries` reads `GET /entries`, paginated via `offset_limit` (`limit`/`offset` query params,
`page_size: 100` — legacy's own `defaultPageSize`); the engine stops on a short/empty page,
matching legacy's own `len(records) < pageSize` stop rule (`public_apis.go:130`). Records are
extracted from the `entries` array in the response body. The raw API returns
PascalCase field names (`API`, `Description`, `Category`, `Auth`, `HTTPS`, `Cors`, `Link`);
`computed_fields` renames each to the schema's snake_case/lower-case field names and additionally
stamps `id` from the raw `API` field, matching legacy's `mapEntry`'s `id = api` convention
(`public_apis.go:178-190`) since the raw API has no dedicated id field. The same computed fields
also coalesce legacy's defensive lowercase-key fallbacks (`api`, `description`, etc.). Primary key
is `id`.

Legacy's own short-page stop rule ALSO independently stops when the running offset plus this
page's record count reaches the response body's top-level `count` field
(`public_apis.go:130`: `count > 0 && offset+len(records) >= count`). The engine's `offset_limit`
paginator implements only the short-page stop signal. This is the identical, already-documented
`jamf-pro` `totalCount` deviation shape (`docs/migration/conventions.md` ledger item 13): it can
cause, at most, one harmless extra request on the rare page where a full-size page happens to
exactly exhaust `count` (the following request then returns an empty/short page and stops
normally) — it never omits, duplicates, or reorders any record for any input legacy itself would
accept.

## Write actions & risks

None. Public APIs is a read-only directory API with no mutation endpoints; `capabilities.write`
is `false` and this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`categories` stream is NOT migrated (`ENGINE_GAP`).** Legacy's `readCategories` reads
  `GET /categories`, whose real response is a bare JSON array of scalar strings (or, per legacy's
  defensive fallback, an object with a `categories` key wrapping the same array) — legacy then
  fans that array out one synthetic record (`{"id": category, "category": category}`) per string
  (`public_apis.go:137-161`). Neither `connsdk.RecordsAt` (keeps only array elements that decode
  as a JSON object; a bare string element is silently dropped, yielding zero records) nor
  `records.keyed_object` (explodes a JSON object's VALUES, which must themselves be objects) can
  express a scalar-array-to-record fan-out. This is the same gap class as the already-documented
  `ip2whois` `nameservers` deviation (`docs/migration/conventions.md` ledger item 12) — emitting a
  single joined-string record instead would change record CARDINALITY versus legacy's genuine
  1-record-per-category stream, an accepted-input emitted-DATA change, not cosmetic. Out of scope
  for this migration; see `api_surface.json`'s excluded entry.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-500,
  default 100) and `max_pages` (0/all/unlimited default) as config-driven overrides
  (`pageSize`/`maxPages`, `public_apis.go:263-285`). The engine's `offset_limit` paginator's
  `page_size` is a fixed value baked into `streams.json`'s `base.pagination` block at
  bundle-author time (`PaginationSpec.PageSize` is a plain int, never `config.*`-templated), and
  `MaxPages` is likewise a fixed bundle-time int — matching the identical, already-documented
  bluetally/searxng precedent. This bundle bakes in legacy's own default (`page_size: 100`);
  `max_pages` is left unbounded (0/omitted), matching legacy's own default (empty `max_pages`
  config = unlimited). Neither is declared in `spec.json` (F6, REVIEW.md: a declared-but-unwireable
  config key is worse than an absent one).
