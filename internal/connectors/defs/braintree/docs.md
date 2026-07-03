# Overview

Braintree reads transactions, customers, and subscriptions through a read-only HTTP API scoped
under `merchants/{{ config.merchant_id }}/...`. This bundle migrates
`internal/connectors/braintree` (the hand-written connector) to a declarative defs bundle at
capability parity; the legacy package stays registered and unchanged until wave6's registry flip.
It is a pure Tier-1 declarative migration — legacy is a connsdk-HTTP-based connector with no
custom auth/stream/write hooks, so no `hooks/braintree/` package is needed.

## Auth setup

Provide `merchant_id`, `public_key`, and `private_key`. `public_key`/`private_key` are sent as HTTP
Basic auth (`Authorization: Basic base64(public_key:private_key)`), matching legacy's
`connsdk.Basic(cfg.Config["public_key"], secret(cfg, "private_key"))`. Both key fields are marked
`x-secret: true` in `spec.json` per this repo's x-secret discipline (credential-shaped fields are
always marked regardless of which legacy map — `Config` vs `Secrets` — happened to hold them);
legacy itself stored `public_key` in the non-secret `Config` map (it is a Braintree "public" key,
often visible client-side) but this bundle's `x-secret` marking only governs redaction from
`DryRunWrite` previews/logging, and Braintree has no write actions here, so the distinction is
purely cosmetic for this connector.

`base_url` defaults to `https://api.sandbox.braintreegateway.com`, matching legacy's default when
its `environment` config is unset or not `"production"`. Legacy also accepts an `environment` enum
(`"production"` vs anything else) that derives one of two fixed base URLs at runtime; this bundle
requires the resolved `base_url` directly instead (documented scope narrowing — see Known limits),
since the engine's `spec.json` `"default"` mechanism only materializes a single fixed literal, not
an enum-derived choice between two URLs (the same shape already ledgered for
`alpaca-broker-api`'s `environment`/`base_url` pair). Set `base_url` to
`https://api.braintreegateway.com` for production.

## Streams notes

Three streams, matching legacy's `endpoints` table exactly:

- `transactions` — `GET /merchants/{{ config.merchant_id }}/transactions`, records at
  `transactions`.
- `customers` — `GET /merchants/{{ config.merchant_id }}/customers`, records at `customers`.
- `subscriptions` — `GET /merchants/{{ config.merchant_id }}/subscriptions`, records at
  `subscriptions`.

All three send `page_size` (default `100`, matching legacy's `defaultPageSize`) as a static
per-stream query value templated from `config.page_size`. Pagination is `cursor` with `token_path:
pagination.next_page` (`cursor_param: page`) — the next page's `page` value is read verbatim from
the current page's response body at `pagination.next_page`, matching legacy's `readPaged` exactly
(`next, err := connsdk.StringAt(resp.Body, "pagination.next_page")`; an absent/empty value stops
pagination). Legacy always sends an explicit `page=1` on the FIRST request too; this bundle's
`token_path` cursor paginator issues page 1 with no `page` query param at all (the paginator only
sets `page` from the second request onward, once a `next_page` token is found) — Braintree's own
API defaults an absent `page` param to page 1, so this never changes which records are returned,
only whether the literal string `page=1` appears on the wire for the first request (see Known
limits).

Every stream declares `projection: "passthrough"`: legacy's `readPaged` emits
`connectors.Record(rec)` — a verbatim cast of the raw decoded JSON object, not a field-built
mapping — so schema-mode projection (which would drop every field not explicitly declared) would
silently narrow legacy's real output. `schemas/*.json` declare the fields legacy's own `Catalog()`
and test fixtures name explicitly (`id`, `status`, `amount`) as the verified baseline; passthrough
mode means any additional real Braintree API field survives in the emitted record regardless of
whether it is declared in the schema.

None of the three streams are incremental: legacy's `Catalog()` sets no `CursorFields` on any of
them, and `Read`/`readPaged` performs no time-based filtering at all (full extraction every sync).
This bundle therefore declares no `incremental` block and no `x-cursor-field`, matching legacy's
real behavior.

## Write actions & risks

None. This connector is `capabilities.write: false`; no `writes.json` is shipped, matching
legacy's `Write` always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- Only the 3 legacy-parity read streams are implemented. Braintree's full merchant API (disputes,
  settlement batch summaries, and any mutation surface) is out of scope for this migration; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries. Legacy itself never implemented any Braintree write action, so there is no write parity
  to port.
- **`environment`-derived `base_url` is not modeled as a two-way enum.** Legacy accepts an
  `environment` config value (`"production"` selects the production host; anything else, including
  unset, selects sandbox) and derives `base_url` from it in code. The engine's `spec.json`
  `"default"` mechanism materializes exactly one fixed literal for an absent key, not an
  enum-conditioned choice between two literals — there is no templating dialect for "pick literal A
  or literal B based on another config value" at the config-default layer (`computed_fields` is
  record-scoped only, not usable for `base_url` resolution). `environment` is still declared in
  `spec.json` for documentation/acceptance parity (a caller passing it does no harm — it is simply
  never read by any template) but this bundle requires `base_url` to be set directly for production
  use rather than deriving it from `environment`. This never changes any emitted record's DATA
  (a request against the wrong host would simply fail outright, not silently return wrong data),
  matching this repo's parity-deviation meta-rule (`ACCEPTABLE`, same class as
  `alpaca-broker-api`'s identical `environment`/`base_url` narrowing).
- **First-page request omits the literal `page=1` query param.** Legacy's `readPaged` always sends
  `page` starting at `"1"` explicitly; this bundle's `cursor`+`token_path` paginator sends no `page`
  param at all on the first request (only subsequent pages, once a `pagination.next_page` token is
  found in the response body). Braintree's documented list endpoints default an absent `page` to
  page 1, so this is a wire-shape difference only, never a difference in which records are returned
  for any input legacy itself would accept (`ACCEPTABLE`, matching this repo's parity-deviation
  meta-rule).
- **`max_pages` config override is not modeled.** Legacy exposes `max_pages` (default 100) as a
  config-driven hard cap on request count. The engine's `cursor`+`token_path` paginator has no
  config-driven request-count-cap knob — `PaginationSpec.MaxPages` is a fixed JSON integer declared
  once in `streams.json`, not a per-request-templatable value — so `max_pages` is not declared in
  `spec.json` (a declared-but-unwireable key is worse than an absent one, per this repo's dead-config
  rule). This bundle relies on the token-absence stop signal alone (matching Braintree's real
  pagination termination for any bounded result set); an operator needing a hard page-count cap for
  a specific sync can express it at the orchestration layer instead.
