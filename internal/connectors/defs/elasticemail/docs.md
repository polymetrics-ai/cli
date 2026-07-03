# Overview

Elastic Email is a wave2 fan-out declarative-HTTP migration. It reads contacts, campaigns, lists,
segments, and templates through the Elastic Email v4 REST API (`GET
{{ config.base_url }}/...`). This bundle is migrated from `internal/connectors/elasticemail` (the
hand-written connector it replaces); the legacy package stays registered and unchanged until
wave6's registry flip. Read-only (`capabilities.write` is `false`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide an Elastic Email API key via the `api_key` secret; it is sent as the
`X-ElasticEmail-ApiKey` header with no prefix (`mode: api_key_header`, empty `prefix`), matching
legacy's `connsdk.APIKeyHeader("X-ElasticEmail-ApiKey", secret, "")`. `base_url` defaults to
`https://api.elasticemail.com/v4` and may be overridden for test proxies.

## Streams notes

All 5 streams share the identical shape: `GET`, records at the response root (`records.path: ""`
— every Elastic Email v4 list endpoint returns a top-level JSON array, matching
`connsdk.RecordsAt(resp.Body, "")`'s root-array selection and legacy's own `RecordsAt(resp.Body,
"")` call), and `offset_limit` pagination (`limit`/`offset` query params). Primary keys match each
stream's natural identifier rather than a synthetic id, exactly as legacy's own catalog declares:
`Email` for `contacts`, `ListName` for `lists`, `Name` for `campaigns`/`segments`/`templates`.
`contacts` declares a bare `incremental.cursor_field: "DateUpdated"` (no `request_param`) and its
schema's `x-cursor-field` names the same field, matching legacy's `Catalog()` `CursorFields:
[]string{"DateUpdated"}` hint for the `contacts` stream — legacy's own `Read` never filters on it
(Elastic Email v4 list endpoints have no request-side time-range filter parameter its own
connector code ever sends), so per `docs/migration/conventions.md`'s incremental-declaration truth
table (bare `cursor_field` iff legacy publishes `CursorFields`; `request_param` iff legacy sends a
server-side filter) this is a bare declaration only, not a `request_param`. `campaigns`/`lists`/
`segments`/`templates` declare no `CursorFields` in legacy and correspondingly have no
`incremental` block or `x-cursor-field`.

`contacts`' schema includes `Activity`/`Consent`/`CustomFields` (nested objects legacy's
`contactRecord` mapper emits) even though legacy's separate `contactFields()` catalog-description
function omits them — the schema is a projection of what `mapRecord` actually emits (the
authoritative behavior per `docs/migration/conventions.md`'s schema-as-projection rule), not of
the narrower `Fields` catalog list, which is descriptive metadata only and does not gate what
`Read` returns.

## Write actions & risks

None. Legacy `elasticemail.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** Legacy accepts optional
  `page_size` (1-1000, default 100) and `max_pages` (default unlimited, `all`/`unlimited`/`0`
  synonyms) config keys read at request time (`elasticEmailPageSize`/`elasticEmailMaxPages`). The
  engine's `PaginationSpec.PageSize`/`MaxPages` fields are plain fixed JSON integers baked into
  `streams.json` — there is no templating/config-driven override mechanism for them.
  `base.pagination.page_size` is set to legacy's real production default, `100`
  (`elasticEmailDefaultPageSize`) — this is the actual value a live deployment's paginator sends;
  it is not a fixture convenience. The `contacts` stream declares a stream-level `pagination`
  override (`page_size: 2`) so its required 2-page conformance fixture
  (`fixtures/streams/contacts/{page_1,page_2}.json`, §4 of `docs/migration/conventions.md`) can
  stay small and readable; since stream-level `pagination` replaces the base spec wholesale, this
  is an intentional, ledgered per-stream deviation from legacy's uniform 100-record page size —
  `contacts` reads in smaller, more numerous pages than legacy would, everywhere else identical.
  `campaigns`/`lists`/`segments`/`templates` are unaffected and use legacy's true 100-record page
  size end-to-end, matching their single-page fixtures' `limit=100` request/response. No
  `max_pages` cap is declared (unbounded, matching legacy's own default). Neither key is declared
  in `spec.json` (F6, `docs/migration/conventions.md`: dead, unwireable config is worse than
  absent config). This never changes which records are emitted for an in-range request — only
  request cadence.
- **`base_url` scheme/host validation is enforced by legacy in Go** with dedicated error messages
  (`elasticEmailBaseURL`); the engine has no equivalent declarative URL-shape validator, so a
  malformed `base_url` here surfaces as a generic request-construction/connection error rather
  than legacy's specific `"config base_url must use http or https"`/`"must include a host"`
  messages. This never changes behavior for any valid `base_url`.
- The full Elastic Email v4 API surface (contact/campaign/template mutation, delivery statistics,
  event webhooks) is out of scope for this wave; see `api_surface.json`'s `excluded` entries.
