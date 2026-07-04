# Overview

Awin Advertiser is a declarative-HTTP bundle migrated from
`internal/connectors/awin-advertiser` (package `awinadvertiser`; the hand-written legacy connector,
which stays registered and unchanged until wave6's registry flip). It reads Awin advertiser
commission transactions, publisher-aggregated performance reports, publisher-level performance
reports, creative-level performance reports, and publisher relationships, and creates advertiser
promotion/voucher offers, through the Awin Advertiser REST API.

**Pass B full-surface expansion**: this bundle was re-researched against Awin's real published
advertiser API docs (`help.awin.com/apidocs/for-advertisers`, `developer.awin.com/apidocs`,
`help.awin.com/advertisers/docs`) rather than the prior wave2 `api_surface.json`'s placeholder
`creatives`/`vouchers`/`programmeinfo` entries, which do not exist as advertiser-facing endpoints
in Awin's real API (see `api_surface.json`'s `scope` field for the full correction). Two new read
streams (`publisher_performance`, `creative_performance`) and one write action (`create_offer`)
were added; `capabilities.write` is now `true`.

## Auth setup

Provide an Awin API OAuth2 bearer token via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. `advertiserId` (a plain, non-secret config
value — the numeric Awin advertiser account ID) is substituted into every stream's path
(`/advertisers/{{ config.advertiserId }}/...`), matching legacy's `fmt.Sprintf("advertisers/%s/...",
advertiserID)`. Unlike legacy, this bundle does not validate that `advertiserId` is all-digits at
read time (legacy's `awinAdvertiserID` rejects non-numeric values before ever making a request) —
the engine's `spec.json` has no numeric-string-pattern validation primitive; a non-numeric value
here simply produces a 404/error from Awin's own API instead of a pre-flight local validation
error. This is a documented, non-data-changing deviation (no valid numeric `advertiserId` behaves
any differently) — see the parity-deviation ledger note below.

## Streams notes

`publishers` (`GET /advertisers/{id}/publishers/`) and `campaign_performance` (`GET
/advertisers/{id}/reports/aggregated/publisher`) are fully declarative and complete: both return a
bare top-level JSON array (`records.path: "."`), paginated by Awin's `page`/`pageSize` params
(`pagination.type: page_number`, 1-indexed — Awin's own convention starts at page 1, so this is a
clean fit for the engine's `page_number` paginator, unlike Auth0's 0-indexed API; see the
`auth0`/`docs.md` for that unrelated gap).

`streams.json`'s `base.pagination.page_size` is a fixed `100`, matching legacy's
`awinDefaultPageSize`/`awinMaxPageSize` (legacy defaults to 100 and caps `page_size` at 100).
Legacy additionally accepts a `page_size` config override in the 1-100 range (`awinPageSize`), but
the engine's `page_number` paginator reads `PageSize` as a plain, non-templated value straight from
`streams.json` (`bundle.go`'s `PaginationSpec`, `paginate.go`'s `newPaginator`) — there is no
config-driven override mechanism for this field in the current dialect (unlike `stream.Query`
params, which do support `{{ config.* }}`). This bundle narrows the config surface to legacy's own
default/max (100 records per request) rather than exposing a `page_size` override; see Known
limits.

`transactions` (`GET /advertisers/{id}/transactions/`) additionally sends Awin's `startDate`/
`endDate`/`dateType`/`timezone` date-window query params, matching legacy's `awinDateWindow`:

- `dateType: "transaction"` and `timezone: "UTC"` are static per-stream literals (legacy hardcodes
  both identically).
- `startDate` is sent only when an incremental lower bound resolves (state cursor, or the
  `start_date` config on a fresh sync) via the `stream.Query` optional-query dialect
  (`{"template": "{{ incremental.lower_bound }}", "omit_when_absent": true}`) — omitted entirely on
  a genuinely-fresh full sync with no `start_date` configured, matching legacy (`lower == ""` skips
  `base.Set("startDate", ...)` entirely).
- `endDate` is a **static sentinel literal** (`2099-12-31T23:59:59`), not legacy's `time.Now()`. See
  Known limits — this is a documented `ENGINE_GAP`-motivated substitution, argued ACCEPTABLE per the
  parity-deviation meta-rule (never changes emitted data for any legacy-accepted read): a real
  transaction's `transactionDate` can never be later than "now" at read time, so an upper bound far
  in the future returns the exact same result set as legacy's `time.Now()` upper bound for any
  transaction that actually exists.
- `transactions` also accepts two new, purely-optional passthrough filters matching Awin's own
  documented `status` (pending/approved/declined/deleted) and `publisherId` (single ID or
  comma-separated list) query parameters: `config.transaction_status` and `config.publisher_id`,
  both `omit_when_absent` (unset by default — no filter, matching legacy's behavior exactly, since
  legacy never sent either param).

**`publisher_performance`** (`GET /advertisers/{id}/reports/publisher`, new in Pass B) is a
single-page report (`pagination.type: none` — Awin does not paginate this endpoint) aggregating
transaction counts/values/commission by publisher over a date window. `startDate` is optional
(Awin's own documented default when omitted) via `config.report_start_date`
(`omit_when_absent: true`) in Awin's date-only `YYYY-MM-DD` format — a DIFFERENT format than
`transactions`' timestamp-shaped `start_date`, since this endpoint documents date-only bounds.
`endDate` is the same static far-future-sentinel pattern as `transactions` (date-only:
`2099-12-31`), for the same provable-equivalence reason.

**`creative_performance`** (`GET /advertisers/{id}/reports/creative`, new in Pass B) is the
creative-level analogue of `publisher_performance`, also single-page. Awin documents BOTH
`startDate` and `region` as REQUIRED for this endpoint (no all-regions/no-date-bound option) —
unlike every other stream in this bundle, an absent `config.report_start_date`/`config.report_region`
does not omit the parameter but instead falls back to a fixed literal default (`2020-01-01` /
`GB`, via the `default` query-param dialect, not `omit_when_absent`) so the request Awin actually
receives always satisfies its own required-parameter contract. An operator who needs data outside
the `GB` region, or before `2020-01-01`, must set `report_region`/`report_start_date` explicitly.
Primary key is the composite `["creativeId", "publisherId"]`, matching the report's real grain (one
row per creative-publisher pair).

## Write actions & risks

**`create_offer`** (`POST /promotion/advertiser/{id}`, new in Pass B) creates a new promotion or
voucher offer in the advertiser's MyOffers system, immediately visible to publishers. Required
fields: `title`, `description`, `terms`, `type` (`"promotion"` or `"voucher"`), `url`, `startDate`,
`endDate` (both `YYYY-MM-DD`), `appliesToAllRegions` (boolean), `promotionCategories` (integer
array). Conditionally required per Awin's own docs (not separately enforced by this bundle's
`record_schema`, which has no `if`/`then` conditional-required primitive — see Known limits):
`voucherCode` when `type` is `"voucher"`, `regions` (integer array) when `appliesToAllRegions` is
`false`. Optional: `startTime`/`endTime` (`HH:MM:SS`, default `00:00:00`/`23:59:59` UTC),
`timeZone` (default UTC), `campaign`. This is an external, publisher-visible mutation; `risk`
documents that approval is required before use. No `path_fields` are declared (the record itself
carries no `id` — Awin returns no offer id in its create response — so the full record maps
directly to the request body via the default `body_type: json` behavior).

## Known limits

- **`ENGINE_GAP` (documented deviation, not a blocker — see parity-deviation ledger)**: the engine's
  interpolation dialect (`internal/connectors/engine/interpolate.go`) has no "current time" /
  `now()` reference anywhere (`config.*`, `secrets.*`, `record.*`, `cursor`, and
  `incremental.lower_bound` are the only resolvable references) — there is no way to express
  legacy's `endDate: time.Now().UTC().Format(...)` as a genuinely dynamic value. This bundle
  substitutes a static far-future sentinel (`2099-12-31T23:59:59`) instead, which is provably
  equivalent for any transaction that exists (see Streams notes) rather than a silent behavior
  change.
- **Date-format config-surface narrowing**: legacy's `normalizeAwinDate` accepts either an RFC3339
  timestamp or a bare `YYYY-MM-DD` date for `start_date` and normalizes either to Awin's exact
  `YYYY-MM-DDTHH:MM:SS` wire shape (no timezone suffix) before sending it. The engine's
  `param_format` options (`rfc3339` verbatim / `unix_seconds` / `date` (`YYYY-MM-DD`, drops
  time-of-day) / `github_date_range`) have no option that reproduces Awin's exact
  seconds-precision-no-timezone format from an RFC3339 input — `rfc3339` sends the configured value
  byte-for-byte (including any `Z`/offset suffix Awin's docs don't show in examples), and `date`
  would incorrectly truncate away the time-of-day component of a real timestamp bound (a genuine
  data-changing narrowing of the filter boundary, not cosmetic). This bundle uses `rfc3339`
  (verbatim passthrough, the least lossy of the available options) and documents that `start_date`
  must be supplied pre-formatted in Awin's exact wire shape for a byte-exact request — a config
  input-format narrowing (what values are accepted), not an emitted-record-data change for any
  value this bundle does accept. Whether Awin's API tolerates a trailing `Z`/offset it doesn't
  document is unverified; operators should supply the exact `YYYY-MM-DDTHH:MM:SS` shape to be safe.
- 5 read streams and 1 write action are now implemented, covering every genuinely advertiser-facing
  endpoint this research surfaced; see `api_surface.json`'s `excluded` entries for the full,
  corrected disposition of every publisher-facing-only endpoint this connector does not (and
  should not) cover.
- **`ENGINE_GAP` — transaction batch validation is not implemented**: Awin's only transaction
  mutation endpoint (`POST /advertisers/{id}/transactions/batch`, approve/decline/amend, up to
  40,000 transactions per request) requires a top-level JSON ARRAY request body. The engine's
  declarative write dialect (`engine/write.go`'s `executeWriteRecord`) always sends exactly one
  JSON object per record per request for `body_type: json`/`form` — there is no array-wrapping or
  multi-record-per-request mechanism in the dialect at all (§6's escape-hatch tree: not
  expressible in Tier 1, and this bundle has no other Tier-2 hook trigger to justify escalating to
  a hook package for this alone). See `api_surface.json`'s `excluded` entry for this endpoint.
- **`create_offer`'s conditional-required fields are not locally enforced**: Awin documents
  `voucherCode` as required only when `type` is `"voucher"`, and `regions` as required only when
  `appliesToAllRegions` is `false`. The engine's `record_schema` is draft-07 JSON Schema with no
  `if`/`then`/`allOf`-conditional support wired into write validation (mirroring stripe's
  `minProperties`-approximation deviation, ledger item 1) — this bundle's `record_schema` declares
  both fields optional or omits enforcing the conditional, so a caller can submit a voucher with no
  `voucherCode` and have it accepted by LOCAL validation (Awin's own API still enforces this
  server-side and will reject such a request with a 400). Strictly more permissive than Awin's
  real contract, never stricter — no valid Awin-accepted offer is ever rejected by this bundle.
- **`page_size` config override not exposed**: legacy accepts a `page_size` config value (integer,
  1-100, default 100) that directly controls the `pageSize` query param sent on every request
  (`awinPageSize`). This bundle's `base.pagination.page_size` is a fixed `100` (legacy's own
  default and max) because the engine's `page_number` paginator reads `PageSize` as a static
  literal from `streams.json`, not a templated value — there is no mechanism in the current
  dialect to wire a `spec.json` config property into `base.pagination.page_size`. Every request
  this bundle sends therefore uses legacy's default page size (100 records/page); an operator who
  relied on legacy's override to request smaller pages (e.g. for slower downstream processing)
  cannot reproduce that here. This never changes emitted record DATA (only request/response
  cadence) for any operator who left legacy's `page_size` at its default — see the parity-deviation
  ledger.

### Parity-deviation ledger (this connector)

| description | verdict |
|---|---|
| `endDate` sent as a static far-future sentinel instead of legacy's `time.Now()` | ACCEPTABLE — provably identical result set for any transaction that exists at read time (see Streams notes) |
| `advertiserId` numeric-format validation is deferred to the live Awin API instead of a local pre-flight check | ACCEPTABLE — no valid numeric advertiserId behaves differently; only the failure mode for an already-invalid config differs |
| `start_date` must be pre-formatted in Awin's exact wire shape rather than legacy's RFC3339-or-date auto-normalization | ACCEPTABLE (documented config-surface narrowing) — no engine `param_format` option reproduces legacy's exact normalization; verbatim `rfc3339` passthrough is the least-lossy available option |
| `page_size` config override (legacy: 1-100, default 100) is not exposed; `base.pagination.page_size` is fixed at legacy's own default/max (100) | ACCEPTABLE (documented config-surface narrowing) — the engine's `page_number` paginator does not support a templated/config-driven `page_size`; fixing it at legacy's default reproduces legacy's default-configuration behavior exactly and never diverges in emitted data, only in request cadence for an operator who had overridden it away from the default |
| `creative_performance`'s `startDate`/`region` fall back to fixed literals (`2020-01-01`/`GB`) instead of Awin's own required-parameter validation error when unset | ACCEPTABLE (documented config-surface behavior) — this is a NEW stream with no legacy implementation to deviate from; the fallback exists so the engine always sends an Awin-accepted request rather than a 400, and is fully documented as an explicit default an operator can override |
| `create_offer`'s `voucherCode`-required-iff-voucher and `regions`-required-iff-not-all-regions conditionals are not locally enforced | ACCEPTABLE — draft-07 `record_schema` has no conditional-required primitive wired into write validation; strictly more permissive than Awin's real contract (Awin's own API still enforces this server-side), never stricter, and no valid Awin-accepted record is ever rejected here |
| Awin's transaction batch-validation endpoint (array-body write) is not implemented | `ENGINE_GAP` — the declarative write dialect has no array-bodied/multi-record-per-request write mechanism; see Known limits and `api_surface.json`'s excluded entry |
