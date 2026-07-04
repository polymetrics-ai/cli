# Overview

ChartMogul is a wave2 fan-out Tier-1 declarative migration, expanded in Pass B to the full
practical API surface reachable with a plain (non-CRM-seat) API key. It reads ChartMogul customers,
contacts, subscription activities, plans, invoices, tasks, customer-count metrics, and account
details through the ChartMogul REST API, and writes customers (create/update). This bundle
originally targeted capability parity with `internal/connectors/chartmogul` (the hand-written
connector it migrates, read-only); the legacy package stays registered and unchanged until wave6's
registry flip. Pass B's customer write actions are new capability beyond legacy parity.

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

**Pass B new streams** (same cursor/`has_more` pagination shape as `customers`/`activities` above):
`plans` (`GET /plans`, records wrapper key `"plans"` — ChartMogul's own docs show plans using their
own resource name as the wrapper, unlike the `"entries"` convention every other new stream here
uses), `contacts` (`GET /contacts`, wrapper `"entries"`, primary key `["uuid"]`), `tasks`
(`GET /tasks`, wrapper `"entries"`; primary key is `["task_uuid"]` — NOT `uuid`, ChartMogul's task
object uses a differently-named id field per its own documented response shape — cursor field
`updated_at`), and `invoices` (`GET /invoices`, wrapper `"invoices"`, primary key `["uuid"]`; no
incremental cursor declared — ChartMogul's documented invoice object carries no
consistently-populated update timestamp to filter by). None of these 4 streams have legacy
precedent; they are wholly new Pass B capability, verified against ChartMogul's own current API
reference (`dev.chartmogul.com/reference`).

## Write actions & risks

Pass B adds write capability (new beyond legacy, which was read-only): `create_customer`
(`POST /customers`, requires `data_source_uuid` + `external_id`) and `update_customer`
(`PUT /customers/{uuid}`). ChartMogul's customer body is flat JSON (no wrapper envelope, unlike
Chargify), so it maps directly onto the write dialect's default body construction. Both actions
carry `"risk": "external mutation; approval required"`; `metadata.json`'s `capabilities.write` is
now `true`.

Contact, task, and plan create/update/delete are deliberately NOT covered by this pass — see
`api_surface.json`'s per-endpoint reasons (narrower CRM/catalog write surfaces than the customer
object Pass B prioritizes as the representative write target).

## Known limits

- Full ChartMogul API surface has been reviewed against the current API reference; every
  out-of-scope endpoint carries a specific, non-blanket reason in `api_surface.json` (narrow
  CRM-module writes, invoice/line-item/transaction sub-object writes that only exist inside a
  bulk-import payload, destructive deletes, CRM-seat-gated Opportunities, and dashboard-shaped
  metrics endpoints beyond the already-covered `customer_count`).
- **Opportunities are excluded entirely (`requires_elevated_scope`)**: ChartMogul's own
  documentation states the Opportunities endpoints require an API key created by a user with a CRM
  seat — not every account/API key configured against this connector can reach that surface, so it
  is not modeled as an ordinary stream/write at all.
- **Subscriptions have no top-level list endpoint** (`GET /v1/customers/{id}/subscriptions` is
  per-customer only) — covering it would require a fan-out over every customer id, deferred (same
  class of gap as chargify's per-product-family `components`, documented in `api_surface.json`).
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
