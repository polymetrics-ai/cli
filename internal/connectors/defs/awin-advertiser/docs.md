# Overview

Awin Advertiser is a read-only declarative-HTTP bundle migrated from
`internal/connectors/awin-advertiser` (package `awinadvertiser`; the hand-written legacy connector,
which stays registered and unchanged until wave6's registry flip). It reads Awin advertiser
commission transactions, publisher-aggregated performance reports, and publisher relationships
through the Awin Advertiser REST API.

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

## Write actions & risks

None. Awin Advertiser is a read-only source in both legacy and this bundle
(`capabilities.write: false`, no `writes.json`).

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
- Only the 3 legacy-parity streams are implemented; the broader Awin Advertiser surface (creatives,
  vouchers, programme info, commission group management) is out of scope — see
  `api_surface.json`'s `excluded` entries.
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
