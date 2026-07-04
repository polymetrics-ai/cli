# Overview

Churnkey is a Pass B full-surface Tier-1 declarative connector. It reads Churnkey cancel-flow
sessions and aggregated session counts through Churnkey's Data API (`https://api.churnkey.co/v1/data`),
and writes usage events / customer attribute updates / billing-contact assignments through Churnkey's
Event Tracking API (`https://api.churnkey.co/v1/api/events/*`) — both APIs share one host
(`api.churnkey.co`) but live at different path prefixes, so `base_url` is now the bare host and every
stream/write path carries its own `/v1/data` or `/v1/api` prefix. This bundle targets capability
parity with `internal/connectors/churnkey` (the hand-written connector it migrates) for reads, and
extends beyond it for writes (legacy never implemented any write path; its own comment asserted "There
are no safe reverse-ETL write actions — the only mutating endpoints are GDPR deletes", which undersold
the documented Event Tracking API); the legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Churnkey Data API key via the `api_key` secret; it is sent as the `x-ck-api-key` header
(`auth: [{"mode":"api_key_header","header":"x-ck-api-key","value":"{{ secrets.api_key }}"}]`),
matching legacy's `connsdk.APIKeyHeader(churnkeyAPIKeyHdr, secret, "")` exactly — the same header
authenticates both the Data API and the Event Tracking API (Churnkey's docs confirm both surfaces use
the identical `x-ck-api-key`/`x-ck-app` header pair). The Churnkey application id is required via the
`app_id` config property, sent as a static `x-ck-app` header on every request (`streams.json`'s
`base.headers`), matching legacy's `DefaultHeaders: {churnkeyAppHdr: app}`. `base_url` defaults to
`https://api.churnkey.co` (`spec.json`'s `default`) — narrowed from legacy's `churnkeyDefaultBaseURL`
(`https://api.churnkey.co/v1/data`) to the bare host now that writes.json needs to reach the sibling
`/v1/api` prefix on the same host; every stream/action path was updated to carry the full `/v1/data/...`
or `/v1/api/...` prefix itself (the engine joins `base_url` + `path` as a plain string concatenation
with no `../`-escape allowed, so a shared bare-host base is the only way to address two sibling path
prefixes from one bundle — see conventions.md's `resolveURL`/`InterpolatePath` reference).

## Streams notes

`sessions` (`GET /v1/data/sessions`) paginates with Churnkey's limit/skip offset convention
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

`session_aggregation` (`GET /v1/data/session-aggregation`) is Churnkey's unpaginated rollup endpoint,
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

Three single-record write actions cover the full documented Event Tracking API, all POSTing a plain
JSON object body (`body_type: "json"`, the default) to `https://api.churnkey.co/v1/api/events/*`:

- **`create_event`** (`POST /v1/api/events/new`) — records a usage/billing event
  (`event`/`customerId` required; `uid`, `eventDate` for backfilling, `eventData` key-value metrics,
  and a nested `user` object for B2B products are optional) against a Churnkey customer. This is the
  primary signal Churnkey's cancel-flow targeting and save-offer eligibility logic consumes.
- **`update_customer`** (`POST /v1/api/events/customer-update`) — overwrites a customer's tracked
  `customerData` attributes (and/or nested `user.data`), identified by `customerId` or `uid` (at
  least one must be supplied by the caller; the dialect's draft-07 subset cannot express Churnkey's
  documented "at least one of `uid`/`customerId`" OR-constraint as a hard requirement — see Known
  limits).
- **`set_billing_users`** (`POST /v1/api/events/customer-update/set-users`) — overwrites the list of
  billing contacts (`users[]`, each with `userId` + `data.email`/`data.name`/`data.billingAdmin`
  required, `data.phone` optional) who receive Payment Recovery emails for a customer.

Every write is an **external mutation requiring approval** (`metadata.json`'s `risk.write`): each
directly influences which cancel offers a real customer sees, or who receives billing-recovery email,
in the connected Churnkey account.

Legacy implemented none of these — its own doc comment ("no safe reverse-ETL write actions") predates
this Pass B research into the documented Event Tracking API, which is a genuinely separate,
purpose-built ingestion surface (distinct from the GDPR deletes legacy was referring to).

## Known limits

- **Bulk-batch write variants are not modeled.** `POST /v1/api/events/bulk` (up to 100 events per
  call) and `POST /v1/api/events/customer-update/set-users/bulk` both accept a JSON ARRAY request
  body; the engine's write dialect (`engine/write.go`'s `executeWriteRecord`) issues exactly one
  request per record with a JSON-OBJECT body built from that record's own fields — there is no
  array-body write primitive. `create_event`/`set_billing_users` already cover the identical
  per-record shape one record at a time; a caller migrating a bulk-array workflow issues N single
  writes instead of 1 batched write (same eventual data, more requests). See `api_surface.json`'s
  `duplicate_of` exclusions.
- **`update_customer`'s "at least one of `uid`/`customerId`" OR-requirement is not enforced.**
  Churnkey's docs require at least one caller-supplied identifier field; the engine's draft-07 schema
  subset has no `anyOf`/`oneOf` (the same limitation stripe's `create_customer` documents, ledger item
  1 in conventions.md §5) — `record_schema` declares both fields optional rather than modeling the
  OR. Strictly more permissive than legacy's real API contract (a record with neither field set would
  be accepted here and rejected by Churnkey itself with a 4xx), never silently wrong-shaped.
  ACCEPTABLE per conventions.md §5's meta-rule.
- **GDPR data-subject-request endpoints (`POST /v1/data/dsr/access`, `POST /v1/data/dsr/delete`) are
  out of scope**, not reverse-ETL write actions — see `api_surface.json`'s `non_data_endpoint`/
  `destructive_admin` exclusions.
- **The hosted-cancel-flow "Customer Data Endpoint" is not a Churnkey-hosted API at all** and is out
  of scope by construction: it is a webhook Churnkey itself calls INTO the customer's own backend
  (HMAC-signed, customer-operated URL) to fetch extra attributes during a live cancel-flow session —
  there is nothing on Churnkey's side to read or write.
- Full Churnkey Data API/Event Tracking API surface beyond the 2 read streams and 3 write actions
  above is covered; see `api_surface.json`'s `excluded` entries for the remaining, deliberately
  out-of-scope endpoints.
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
