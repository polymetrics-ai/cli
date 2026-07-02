# Overview

WooCommerce is a wave2 fan-out declarative-HTTP migration. It reads WooCommerce orders, products,
customers, and coupons through the WordPress-hosted WooCommerce REST API (`wc/v3`,
`GET {base_url}/orders|products|customers|coupons`). This bundle migrates
`internal/connectors/woocommerce` (the hand-written connector it replaces); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide the WooCommerce REST API consumer key/secret as `api_key`/`api_secret` secrets; they are
sent as HTTP Basic auth (consumer key as username, consumer secret as password), matching legacy's
`connsdk.Basic(key, secret)` (`woocommerce.go:281`), and are never logged. `base_url` must be the
full `wc/v3` base URL (e.g. `https://example.com/wp-json/wc/v3`) — see Known limits below for why
legacy's `shop`-only shorthand is not modeled.

## Streams notes

All four streams (`orders`, `products`, `customers`, `coupons`) hit their respective WooCommerce
list endpoints and return a top-level JSON array (`records.path: ""`), matching legacy's
`connsdk.RecordsAt(resp.Body, "")` call. Pagination is WordPress page-number pagination
(`page`/`per_page`, default page size 10 — legacy's `woocommerceDefaultPage`), stopping on a short
page (fewer records than `per_page`), matching legacy's own fallback stop condition exactly
(`harvest`'s `if len(records) < pageSize`). Legacy additionally inspects the `X-WP-TotalPages`
response header to stop early when the reported total page count is reached; the engine's
`page_number` paginator has no header-inspection stop signal, so this bundle relies solely on the
short-page fallback — see Known limits. Every stream is sorted `order=asc&orderby=id` for stable
pagination, matching legacy's static query base.

Incremental reads filter by `date_modified_gmt` via the `modified_after` query param
(`incremental.request_param`), matching legacy's `modified_after` filter derived from the resolved
cursor/`start_date`. Legacy also sends the identical lower-bound value a second time as `after` "as
a fallback for resources that ignore [modified_after]" (`harvest`'s own comment,
`woocommerce.go:165-168`); this bundle reproduces that exact redundant-param behavior via the
optional-query dialect (`"after": {"template": "{{ incremental.lower_bound }}", "omit_when_absent":
true}`), sent in the identical branch as `modified_after` (both present together, or both absent on
a full sync) since `incremental.lower_bound` resolves to the exact same formatted value
`request_param` itself receives.

## Write actions & risks

None. This connector is read-only in both legacy and this bundle (`capabilities.write: false`); no
`writes.json` is shipped.

## Known limits

- **`shop`-derived base URL shorthand is not modeled.** Legacy accepts either an explicit
  `base_url` override or a bare `shop` host (e.g. `example.com`), deriving
  `https://<shop>/wp-json/wc/v3` in code (`woocommerceBaseURL`). The engine's `spec.json` `default`
  materialization only fills a fixed literal for a genuinely-absent key; it has no mechanism to
  derive one config value's default from another config value's runtime content. This bundle
  therefore requires `base_url` directly (the full `wc/v3` URL) and does not declare a `shop`
  config key at all (a declared-but-unwireable key is worse than an absent one, per
  `docs/migration/conventions.md`'s F6 precedent). This is a documented config-surface narrowing:
  a caller who previously configured only `shop` must now supply the equivalent full `base_url`;
  the emitted request and data are byte-identical once they do.
- **`X-WP-TotalPages` header-based early stop is not modeled.** Legacy inspects this response
  header to stop pagination as soon as the reported total page count is reached, independent of
  whether the final page happens to be short. The engine's `page_number` paginator has no
  response-header inspection hook — only the short-page (`len(records) < page_size`) and
  `MaxPages` stop signals are available. In practice this converges to the identical stop point:
  WooCommerce's last page of any list is, by construction, either exactly full (continues to an
  empty next page, itself short) or short (stops immediately) — the header is a redundant
  optimization legacy uses to skip one possible trailing empty-page request in the exactly-full
  case, never a source of different DATA emitted. `max_pages`'s config override is retained as a
  hard request-count cap via the engine's own `MaxPages` enforcement, matching legacy's
  `woocommerceMaxPages` semantics (0/all/unlimited = unbounded).
- **Legacy's fixture-mode-only fields (`readFixture`, reached only when `config.mode ==
  "fixture"`) are not modeled.** That credential-free path stamps extra fields not present on any
  real WooCommerce API response (a synthetic `previous_cursor` echo, among others). This bundle's
  schemas and conformance fixtures target the live wire shape only; the engine's own
  fixture-replay conformance harness supplies the credential-free test affordance legacy's fixture
  mode existed for, matching the precedent in `docs/migration/conventions.md`'s bitly ledger entry.
