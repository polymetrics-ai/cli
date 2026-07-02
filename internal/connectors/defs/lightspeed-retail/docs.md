# Overview

Lightspeed Retail (X-Series) is a point-of-sale / retail management platform. This bundle reads
its 2.0 REST API (`https://<subdomain>.retail.lightspeed.app/api/2.0`): products, customers,
sales, outlets, and registers. Read-only. This bundle migrates
`internal/connectors/lightspeed-retail` (the hand-written connector); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Lightspeed Retail personal token / OAuth access token via the `api_key` secret. It is
sent as `Authorization: Bearer <api_key>`; never logged.

`base_url` is derived from the required `subdomain` config value as
`https://<subdomain>.retail.lightspeed.app` (`streams.json`'s `base.url` templates
`{{ config.subdomain }}` directly into the host — the same subdomain-templated-host pattern as the
bamboo-hr/agilecrm goldens), matching legacy's own per-retailer hosted-subdomain convention.

## Streams notes

All 5 streams share the same shape: `GET` against the X-Series 2.0 list endpoint with
`page_size=100`, records at `data`, primary key `["id"]`, cursor field `version`. Pagination
follows Lightspeed's version-cursor envelope (`pagination.type: cursor` with
`token_path: version.max`, `cursor_param: after`): the next page is requested with
`after=<version.max>` until `version.max` is `null` (or absent), exactly matching legacy's
`harvest` stop condition — `connsdk.StringAt` resolves a JSON `null` to an empty string, which is
the engine's `tokenPathCursor` paginator's own empty-token stop signal, so no `stop_path` override
is needed here (unlike lever-hiring/zendesk's separate boolean stop-flag shape).

Every X-Series object exposes a string `id` and a numeric `version`; legacy declares `version` as
a `CursorFields` catalog hint but never uses it as a server-side incremental filter parameter — the
harvest loop is unconditional full-refresh pagination with no `updated_after`/`since`-style request
parameter at all (the `version` field only drives pagination advancement, never record filtering).
This bundle preserves that: every schema declares `x-cursor-field: version` (so a
`*_deduped`/`incremental_append` sync mode can dedupe/order by it client-side per design §B.6),
but no `streams.json` `incremental` block is declared anywhere, matching legacy's genuine absence
of an incremental request mechanism.

## Write actions & risks

None. This bundle is read-only (`capabilities.write: false`); Lightspeed Retail exposes no
reverse-ETL write surface here (matches legacy's `Write` returning
`connectors.ErrUnsupportedOperation` unconditionally).

## Known limits

- Legacy also accepted an explicit `base_url` config override (bounded to `http`/`https` with a
  host, primarily so tests/proxies could redirect requests) alongside the subdomain-derived
  default. This bundle drops that override and requires `subdomain` unconditionally: the engine's
  `HTTP.URL` template is a single string with no conditional "prefer this key if set, else derive
  from that key" mechanism (`docs/migration/conventions.md` §3's "derived default" note — the
  same shape as chargebee/sentry's own documented scope narrowing). Every real (non-test) legacy
  caller always supplied `subdomain` — the dropped override only ever existed to bound SSRF risk
  for test/proxy usage — so this is a documented config-surface narrowing, not a reachability loss
  for production traffic.
- Legacy exposed a runtime-configurable `page_size` (1-200) and `max_pages` (0/all/unlimited) knob.
  Neither is expressible in this dialect: `PaginationSpec`'s `page_size`/`max_pages` fields are
  fixed JSON literals in `streams.json`, never resolved from `RuntimeConfig.Config` at read time
  (mirrors the stripe golden's own documented `page_size`/`max_pages`-is-dead-config precedent,
  ledger item 3). This bundle fixes the page size at 100 (legacy's own default) and leaves
  pagination unbounded, matching legacy's own default behavior for a caller that never overrides
  either knob.
- Full Lightspeed Retail API surface (inventory adjustments, consignments, price books, gift
  cards, webhooks) is out of scope for this wave; see `api_surface.json`'s `excluded:
  {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
