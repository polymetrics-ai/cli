# Overview

ChargeDesk is a wave2 fan-out migration. This bundle reads ChargeDesk charges, customers,
subscriptions, and products through the ChargeDesk REST API, migrating
`internal/connectors/chargedesk` (the legacy hand-written connector, which stays registered and
unchanged until wave6's registry flip) at capability parity. ChargeDesk's catalog supports only
full-refresh and exposes no safe reverse-ETL writes, so this connector is read-only.

## Auth setup

ChargeDesk authenticates with HTTP Basic auth
(https://chargedesk.com/api-docs/authentication). Provide the ChargeDesk secret API key via the
`password` secret; by default it is sent as the Basic auth USERNAME with a blank password
(`Authorization: Basic base64(<password>:)`), matching legacy's default scheme. An optional
`username` config value overrides this: when set, `username` is the Basic auth username and
`password` is its Basic auth password instead (`Authorization: Basic base64(<username>:<password>)`).

This is a **dual-auth candidate list**, evaluated first-match-wins (`base.auth`'s declaration
order is load-bearing — see `docs/migration/conventions.md` §3's dual-auth-ordering rule): the
`username`-gated candidate is declared FIRST (`when: "{{ config.username }}"`, true only when
`username` is configured), falling through to the secret-as-username default candidate when
`username` is absent — reproducing legacy's exact `switch { case username != "": ...; case secret
!= "": ...}` precedence (username override wins whenever both are present).

## Streams notes

All 4 streams (`charges`, `customers`, `subscriptions`, `products`) share the same shape: `GET`
against the ChargeDesk list endpoint, records at `data`, offset/count pagination
(`pagination.type: offset_limit`, `limit_param: count`, `offset_param: offset`, `page_size: 100` —
matches legacy's `chargedeskDefaultPageSize`), stopping on a short page (fewer than 100 records),
matching legacy's `len(records) < pageSize` rule exactly (ChargeDesk's own envelope has no
`has_more` flag). Each stream declares `incremental.cursor_field: occurred` (matching legacy's
declared `CursorFields`) with NO `request_param` and NO `client_filtered` — legacy's own `harvest`
never sends any incremental filter to the API and never client-side filters either (every sync,
incremental or full, walks every page); the bare `cursor_field` declaration exists only so the
engine derives `incremental_append` sync-mode eligibility (matching legacy's own published catalog
capability), with the actual read remaining an unfiltered full walk on every sync, exactly as
legacy behaves. Primary keys: `charges` uses `charge_id`, `customers` uses `customer_id`,
`subscriptions` uses `subscription_id`, `products` uses `product_id` — matching legacy's declared
`PrimaryKey`.

`spec.json` intentionally does NOT declare `page_size`/`max_pages` as runtime-configurable
properties (unlike legacy, which accepts config overrides for both): `PaginationSpec.PageSize`/
`MaxPages` are read exclusively from `streams.json`'s static `pagination` JSON literal, never from
a `config.*`-templated value (F6, `conventions.md`). See Known limits.

## Write actions & risks

None. This connector is read-only, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- `page_size`/`max_pages` runtime overrides are not exposed (see Streams notes above) — every
  read uses the fixed `page_size: 100`/unbounded-pages shape baked into `streams.json`. This never
  changes any single emitted record's DATA, only how many requests a sync issues and at what page
  size — parity-deviation ledger candidate, ACCEPTABLE under the meta-rule.
- Full ChargeDesk API surface (refunds, disputes, notes, mutating charge/customer/subscription
  actions) is out of scope for wave2; see `api_surface.json`'s
  `excluded: {category: out_of_scope}` entries. Only the 4 legacy-parity read streams are
  implemented.
