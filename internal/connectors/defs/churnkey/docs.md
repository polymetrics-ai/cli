# Overview

Churnkey is a wave2 fan-out Tier-1 declarative migration. It reads Churnkey cancel-flow sessions and
aggregated session counts through Churnkey's read-only Data API
(`https://api.churnkey.co/v1/data`). This bundle targets capability parity with
`internal/connectors/churnkey` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip. Churnkey is read-only in both legacy and this
bundle (no `writes.json` — legacy's own comment: "There are no safe reverse-ETL write actions — the
only mutating endpoints are GDPR deletes").

## Auth setup

Provide a Churnkey Data API key via the `api_key` secret; it is sent as the `x-ck-api-key` header
(`auth: [{"mode":"api_key_header","header":"x-ck-api-key","value":"{{ secrets.api_key }}"}]`),
matching legacy's `connsdk.APIKeyHeader(churnkeyAPIKeyHdr, secret, "")` exactly. The Churnkey
application id is required via the `app_id` config property, sent as a static `x-ck-app` header on
every request (`streams.json`'s `base.headers`), matching legacy's `DefaultHeaders: {churnkeyAppHdr:
app}`. `base_url` defaults to `https://api.churnkey.co/v1/data` (`spec.json`'s `default`), matching
legacy's `churnkeyDefaultBaseURL`.

## Streams notes

`sessions` (`GET /sessions`) paginates with Churnkey's limit/skip offset convention
(`pagination.type: offset_limit`, `limit_param: limit`, `offset_param: skip`); a page shorter than
the declared `page_size` ends the read, matching legacy's `harvest()` loop (`churnkey.go:154-187`)
exactly. Primary key is `["_id"]` (Churnkey's Mongo-style id) and the incremental cursor field is
`["created_at"]`, matching legacy's `churnkeyStreams()` catalog — no `incremental.request_param` is
declared (Churnkey's Data API accepts no server-side updated-at filter; legacy's own `InitialState`
tracks the cursor for resumability only, never uses it to filter a request), so every read is a full
refresh, exactly matching legacy. Each session's nested `customer`/`acceptedOffer` sub-objects are
both hoisted into flat, snake_case columns (`customer_id`, `customer_email`, `customer_plan_id`,
`customer_billing_interval`, `offer_type`) AND preserved whole under their renamed keys (`customer`,
`accepted_offer`), matching legacy's `churnkeySessionRecord` exactly — the hoisted fields use
`computed_fields` reaching into the nested raw JSON (`{{ record.customer.id }}` etc.), and the
whole-object preservation uses a bare `{{ record.acceptedOffer }}` reference (typed extraction
copies the raw object through unchanged; `customer` needs no rename since its raw key already
matches the schema property name, so plain schema projection covers it).

`session_aggregation` (`GET /session-aggregation`) is Churnkey's unpaginated rollup endpoint,
returning every grouped-count row in one response (`records.path: "."` over a top-level array, no
`pagination` block, matching legacy's `paginated: false` + `readSinglePage`). Legacy normalizes 4
camelCase breakdown-dimension keys (`billingInterval`, `planId`, `offerType`, `saveType`) to
snake_case, DROPPING the original camelCase key once normalized (`churnkeyAggregationRecord`'s
normalization loop explicitly `continue`s past those 4 keys rather than re-adding them) — this
bundle reproduces that exact drop-and-rename via `computed_fields` (`"billing_interval": "{{
record.billingInterval }}"` etc.) combined with `schema`-mode (not `passthrough`) projection, since
passthrough would keep BOTH the raw camelCase key and the renamed snake_case key, which legacy never
emits. See Known limits for the one narrowing this implies (legacy's dynamic "carry through any
OTHER breakdown dimension" fallback is not reproduced).

## Write actions & risks

None — Churnkey is exposed as a read-only source connector in both legacy (`churnkey.go`'s `Write`
returns `connectors.ErrUnsupportedOperation`) and this bundle (`metadata.json`'s
`capabilities.write: false`, no `writes.json` file at all). Churnkey's only mutating endpoints are
GDPR customer-data deletes, excluded from `api_surface.json` as `destructive_admin` — not a
reverse-ETL read/write surface for a source connector.

## Known limits

- Full Churnkey Data API surface (any endpoints beyond `/sessions` and `/session-aggregation`, and
  the GDPR delete endpoints) is out of scope for wave2; see `api_surface.json`'s `excluded` entries.
  Only the 2 legacy-parity streams are implemented.
- **`session_aggregation`'s dynamic extra-dimension passthrough is not reproduced.** Legacy's
  `churnkeyAggregationRecord` carries through ANY additional breakdown-dimension key beyond the 8 it
  explicitly enumerates (`count`/`month`/`trial`/`billingInterval`/`planId`/`aborted`/`canceled`/
  `offerType`/`saveType`), a defensive "don't silently drop an unrecognized dimension" fallback.
  `schema`-mode projection only emits schema-declared properties — a genuinely undocumented, unknown
  9th dimension key Churnkey might one day add would be silently dropped here (whereas legacy would
  preserve it verbatim). This is judged ACCEPTABLE scope-narrowing rather than an `ENGINE_GAP`
  blocker: the 8 enumerated dimensions are Churnkey's complete documented `session-aggregation`
  breakdown surface today (matching legacy's own `churnkeyAggregationFields()` catalog, which
  declares the identical closed set), so no CURRENTLY-DOCUMENTED input differs in emitted shape; a
  future Churnkey API addition would need this bundle's schema/computed_fields extended, exactly
  like any other schema-as-projection bundle when an upstream API adds a field. (`passthrough`
  projection was considered and rejected: it would additionally emit the raw camelCase key
  alongside the renamed snake_case key for all 4 normalized dimensions, which legacy's own
  normalization loop explicitly avoids — that would have been a real emitted-shape deviation, not
  just a documented scope narrowing.)
- **`x-ck-app`/`app_id` config-key aliasing dropped to `app_id` only.** Legacy accepts either a
  catalog-canonical `x-ck-app` config key or a friendlier `app_id` alias (`churnkeyApp`, preferring
  `x-ck-app` when both are set). This bundle declares a single `app_id` spec.json property; an
  operator migrating a legacy `x-ck-app`-only config supplies the identical value under `app_id`
  instead. No request/data change once configured — the value is sent as the same static `x-ck-app`
  header either way.
- **`page_size`/`max_pages` config keys dropped.** `streams.json`'s `pagination.page_size` is a
  static JSON int (`PaginationSpec.PageSize`), not a runtime-templated value, so legacy's
  config-overridable `page_size` (default 100, up to 10,000) cannot be reproduced as live
  runtime-configurable; `page_size` is fixed at legacy's real production default, `100`
  (`churnkeyDefaultPageSize`), matching the value an actual deployment's paginator sends. The
  required 2-page conformance fixture (conventions.md §4) is satisfied with a full 100-record
  `page_1` (`limit=100, skip=0`) followed by a genuinely short 1-record `page_2` (`limit=100,
  skip=100`), the same pattern used by other single-paginated-stream bundles (e.g. bamboo-hr's
  `employees`) — a small static `page_size` is not needed to keep the fixture small. Legacy's
  `max_pages` config property remains genuinely dead config in this dialect and is not declared in
  `spec.json` (F6, conventions.md); `page_size` itself is not declared as a live-configurable
  `spec.json` property either, since the dialect has no template hook for it, but the STATIC value
  now matches legacy's real default rather than a fixture-only shortcut.
- `metadata.json` declares no `rate_limit` block: legacy Churnkey enforces no client-side rate
  limiting (no `rate_limit`/throttle field anywhere in `churnkey.go`), so this bundle adds none
  either, matching conventions.md §3's "informational vs. enforced" rate-limit rule.
