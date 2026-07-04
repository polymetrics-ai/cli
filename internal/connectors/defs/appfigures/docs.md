# Overview

Appfigures is a declarative-HTTP connector for the Appfigures v2 REST API
(`https://api.appfigures.com/v2/...`). It reads app-store reviews, tracked products, analytics
report aggregates (sales, ratings, revenue, subscriptions, ads, download/revenue estimates),
release-event markers, connected external store accounts, account users, account/usage info, and
reference data (categories, countries, languages, currencies, stores, tracked SDKs), and it manages
review responses and release-event markers through Pass B's full-surface expansion. This bundle
originally targeted full capability parity with `internal/connectors/appfigures` (the hand-written
connector it migrates) across its 5 legacy streams; the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide an Appfigures Personal Access Token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://api.appfigures.com/v2` and may be overridden for tests/proxies. Every stream and write
shares this one credential.

## Streams notes

18 streams, 3 shapes:

**Paginated list** (`reviews`, `users`) — 1-based `page_number` pagination (`page_param: page`,
`size_param: count`, `page_size: 100`), stopping on a short page.
- `reviews` — `GET /reviews`, records at `reviews`, primary key `id`. Optional per-request filters
  (`search_store` -> `store`, `group_by` -> `group_by`, `start_date` -> `start`, `end_date` -> `end`)
  are wired via the opt-in optional-query dialect (`omit_when_absent: true`).
- `users` — `GET /users`, records at `results` (the `metadata.resultset` envelope is not
  projected), primary key `id`. Same `page`/`count` pagination params as the base spec, matching
  the documented request shape exactly.

**Keyed-object** (`products`, `sales`, `ratings`, `categories`, `events`, `external_accounts`,
`data_countries`, `data_languages`, `data_stores`) — a single unpaginated request
(`pagination: {"type":"none"}`) whose body is a JSON object keyed by an arbitrary id, exploded via
`records: {"path":"","keyed_object":true}` (`docs/migration/conventions.md` §3). `key_field` is set
only where the value objects don't already carry a field equal to the map key themselves
(`data_countries` stamps `iso`, `data_stores` stamps a synthetic `store_key`); it is left unset on
`products`/`sales`/`ratings`/`categories`/`events`/`external_accounts`/`data_languages` since each
value object already has its own natural id/date/code field.
- `products` — `GET /products/mine`, primary key `id`.
- `sales` / `ratings` — `GET /reports/sales` / `/reports/ratings`, no primary key (`date` alone is
  not guaranteed unique across products/stores). Optional `store`/`group_by`/`start`/`end` filters,
  same as `reviews`.
- `categories` — `GET /data/categories`, primary key `id`; reference data, no date filters.
- `events` — `GET /events/`, primary key `id`; release/marketing markers overlaid on every
  Appfigures analytics chart.
- `external_accounts` — `GET /external_accounts`, primary key `id`; connected app-store developer
  accounts.
- `data_countries` — `GET /data/countries`, primary key `iso` (also the stamped `key_field`).
- `data_languages` — `GET /data/languages`, primary key `code` (already present on each value).
- `data_stores` — `GET /data/stores`, primary key `store_key` (a stamped synthetic key — the store
  code, e.g. `apple`/`google_play`, since the value objects carry their own numeric `id` but no
  field equal to the map key itself).

**Single-object** (`revenue`, `subscriptions`, `ads`, `estimates`, `account_info`) — a single
unpaginated request whose entire response body IS one record (`records: {"path": ""}`, no
`keyed_object`). `revenue`/`subscriptions`/`ads`/`estimates` share the same optional
`store`/`start`/`end` filters as `sales`; each stamps a static-literal `report` field
(`"revenue"`/`"subscriptions"`/`"ads"`/`"estimates"`) as its primary key, since these are
account-wide aggregate totals with no natural id. **These 4 streams model the documented
ungrouped-totals response shape only** — passing a `group_by` value restructures the response into
a nested breakdown this schema does not project (see Known limits). `account_info` reads the root
`GET /` endpoint (identity + daily API usage); `computed_fields` flattens the nested
`user.{id,name,email}` and `usage.{daily_used,daily_limit}` objects onto the record's top level
(typed bare-reference extraction keeps `user_id`/`daily_used`/`daily_limit` as native numbers),
primary key `user_id`.

`data_currencies` and `data_sdks` are plain top-level JSON arrays (`records: {"path": ""}`, no
`keyed_object` — the response is already an array, not a keyed object), primary keys `Currency` and
`id` respectively.

None of the 18 streams are incremental — Appfigures' v2 API has no server-side cursor filter for
any of them.

## Write actions & risks

`capabilities.write: true`. 4 actions, all requiring reverse-ETL plan approval before executing
(`metadata.json`'s `risk.approval`):

- `reply_to_review` — `POST /reviews/{id}/response`, body `{content}`. Publishes a developer
  response visible on the review's public app-store listing; Appfigures processes this
  asynchronously (`202 Accepted`).
- `create_event` / `update_event` / `delete_event` — `POST`/`PUT`/`DELETE /events/[{id}]`. Manages
  release/marketing event markers that appear overlaid on every analytics chart for the account;
  `delete_event` returns `204 No Content` on success.

## Known limits

- `page_size`/`max_pages` config overrides legacy exposes for `reviews` are not runtime-configurable
  here: the engine's `page_number` paginator's `PageSize` is a static int set once in
  `streams.json`, not template-resolvable. `spec.json` intentionally does not declare
  `page_size`/`max_pages` (a declared-but-unwireable key is worse than an absent one).
- The keyed-object and single-object streams are read in a single request with no bound on
  response size — Appfigures' report/reference endpoints return their entire result set in one
  body, with no pagination affordance to bound it.
- `revenue`/`subscriptions`/`ads`/`estimates` model the ungrouped-totals response shape only: if an
  operator sets `group_by` (a config value shared with `sales`/`ratings`, which DOES support
  grouping), Appfigures restructures the response into a nested per-dimension breakdown these 4
  streams' schemas do not project — the emitted record would be empty/malformed. This connector
  does not currently guard against that misconfiguration; operators reading these 4 streams should
  leave `group_by` unset.
- `/reports/adspend`, `/reports/payments`, and the report-style `/reports/usage` (distinct from the
  account-level `/usage` surfaced via `account_info`) are out of scope: no fixture-verifiable
  ungrouped-totals response shape could be confirmed against the live documentation during this
  research pass, unlike `revenue`/`subscriptions`/`ads`/`estimates`, whose flat-object shape is
  explicitly documented.
- Path-templated resources requiring an already-known product id and/or explicit date-range path
  segments (`ranks`, `ranks/snapshots`, `aso`, `aso/stats`, `featured/*`, `products/{id}/sdks`) are
  out of scope for this wave — a fan_out over the `products` stream's ids could reach some of these
  in a future capability expansion, but was not prioritized here. See `api_surface.json` for the
  full per-endpoint rationale.
