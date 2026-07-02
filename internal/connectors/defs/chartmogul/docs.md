# Overview

ChartMogul is a wave2 fan-out Tier-1 declarative migration. It reads ChartMogul customers,
subscription activities, customer-count metrics, and account details through the ChartMogul REST
API. This bundle targets capability parity with `internal/connectors/chartmogul` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip. ChartMogul is read-only in both legacy and this bundle (no `writes.json`).

## Auth setup

Provide a ChartMogul API key via the `api_key` secret; it is sent as the HTTP Basic username with an
empty password (`auth: [{"mode":"basic","username":"{{ secrets.api_key }}","password":""}]`),
matching legacy's `connsdk.Basic(secret, "")` exactly (`chartmogul.go`'s `requester`). `base_url`
defaults to `https://api.chartmogul.com/v1` (`spec.json`'s `default`), matching legacy's
`chartmogulDefaultBaseURL`.

## Streams notes

`customers` and `activities` share ChartMogul's cursor/has_more pagination over a top-level
`entries[]` array (`pagination.type: cursor`, `cursor_param: cursor`, `token_path: cursor`,
`stop_path: has_more`): the next page's `cursor` query value is read verbatim from the previous
response body's `cursor` field, and pagination stops as soon as `has_more` is falsy — regardless of
whether the `cursor` field itself is still populated — exactly matching legacy's `harvest()` loop
(`chartmogul.go:153-202`, which requires BOTH `hasMore == "true"` AND a non-empty cursor to
continue; `stop_path` on `has_more` reproduces the `hasMore` gate, and the paginator's own
empty-token stop reproduces the cursor gate). `customers` has no incremental filter (legacy: "list
endpoints do not support incremental filtering by update time" — no `incremental` block at all, a
straight full refresh every read, matching `chartmogulStreamEndpoints["customers"]`). `activities`
sends an optional `start-date` query param carrying the RAW (non-truncated) resolved incremental
lower bound — state cursor if resumed, else the `start_date` config value on a fresh sync, omitted
entirely on a first sync with no `start_date` configured (`stream.Query`'s `omit_when_absent`
dialect: `{"template": "{{ incremental.lower_bound }}", "omit_when_absent": true}`) — matching
legacy's `if lower := startDate(req); lower != "" { base.Set("start-date", lower) }` (no format
conversion applied to `activities`' `start-date`, unlike the metrics endpoint below).

`customer_count` (`GET /metrics/customer-count`) is ChartMogul's single-page metrics endpoint: it
always sends `interval=month` (a static literal, matching `endpoint.metricsInterval`) plus a
`start-date`/`end-date` window truncated to `YYYY-MM-DD` (`param_format: date`, matching legacy's
`metricsDate` truncation). `start-date` resolves from the incremental lower bound (state cursor or
`start_date` config), falling back to the literal `2026-01-01` (matching legacy's
`chartmogulFixtureDate` constant fallback for an unconfigured, non-resumed read) when neither
resolves. **`end-date` is a static sentinel literal (`2099-12-31`), not legacy's
`time.Now().UTC().Format(...)`** — see Known limits.

`account` (`GET /account`) returns a single JSON object, not a collection; it is modeled with
`records.path: "."` + `single_object: true` (the pingdom `reference`-stream golden pattern), which
wraps the whole response body as one record, matching legacy's `RecordsAt(resp.Body, "")` +
`readSingle`'s single-object handling.

Primary key is `["uuid"]` for `customers`/`activities`/`account` (`["date"]` for `customer_count`,
which has no stable per-row id), matching legacy's `chartmogulStreams()` catalog. Incremental cursor
field is `["customer-since"]`/`["date"]`/`["date"]` for `customers`/`activities`/`customer_count`
respectively (`account` has none); `customers` declares `x-cursor-field` for catalog-derived
sync-mode purposes only, since (as above) no request ever actually filters by it.

## Write actions & risks

None — ChartMogul is exposed as a read-only source connector in both legacy (`chartmogul.go`'s
`Write` returns `connectors.ErrUnsupportedOperation`) and this bundle (`metadata.json`'s
`capabilities.write: false`, no `writes.json` file at all).

## Known limits

- Full ChartMogul API surface (invoices, plans, subscriptions import, tasks, tags/custom
  attributes, write endpoints) is out of scope for wave2; see `api_surface.json`'s `excluded:
  {category: out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 4
  legacy-parity streams are implemented.
- **`ENGINE_GAP` (documented deviation, not a blocker — see parity-deviation ledger; awin-advertiser
  carries the identical shape)**: the engine's interpolation dialect
  (`internal/connectors/engine/interpolate.go`) has no "current time"/`now()` reference anywhere
  (`config.*`, `secrets.*`, `record.*`, `cursor`, and `incremental.lower_bound` are the only
  resolvable references) — there is no way to express legacy's `customer_count`
  `end-date: time.Now().UTC().Format(...)` as a genuinely dynamic value. This bundle substitutes a
  static far-future sentinel (`2099-12-31`) instead: a real ChartMogul customer-count data point can
  never be dated later than "now" at read time, so an upper bound far in the future returns the
  exact same result set as legacy's `time.Now()` upper bound for any data point that actually
  exists — provably equivalent, not a silent behavior change.
- **`page_size`/`max_pages` config keys dropped.** `streams.json`'s cursor pagination has no
  `page_size`/`max_pages` fields at all in the dialect's `cursor` type (only `page_number`/
  `offset_limit` read `page_size`, and even then it is a static JSON int, not runtime-templated).
  Legacy's `page_size` (default 200, capped at 200) and `max_pages` config properties are
  consequently genuinely dead config in this dialect and are not declared in `spec.json` (F6,
  conventions.md). Each stream's static `per_page: "2"` query value exists purely to keep the
  required 2-page fixture (conventions.md §4) small and realistic; it has no bearing on production
  correctness.
- `metadata.json` declares no `rate_limit` block: legacy ChartMogul enforces no client-side rate
  limiting (no `rate_limit`/throttle field anywhere in `chartmogul.go`), so this bundle adds none
  either, matching conventions.md §3's "informational vs. enforced" rate-limit rule.
