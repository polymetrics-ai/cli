# Overview

Adobe Commerce (Magento) is a wave2 fan-out declarative-HTTP migration. It reads Magento products,
orders, customers, categories, and invoices through the Magento REST API
(`GET https://<store>/rest/V1/...`). This bundle targets capability parity with
`internal/connectors/adobe-commerce-magento` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Magento Integration Access Token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and never logged, matching legacy's `connsdk.Bearer(secret)`
(`adobe_commerce_magento.go:277`). `base_url` is required and must be the fully composed Magento
REST base URL including the API version prefix (e.g. `https://magento.mystore.com/rest/V1`) — see
Known limits for why this diverges from legacy's `store_host`+`api_version` derivation.

## Streams notes

All five streams (`products`, `orders`, `customers`, `categories`, `invoices`) share the identical
Magento `searchCriteria` list shape: `GET /<resource>` returns
`{"items":[...],"total_count":N,"search_criteria":{...}}`, records live at the `items` array.
`customers` and `categories` read from Magento's dedicated search/list endpoints
(`/customers/search`, `/categories/list`) exactly like legacy's `magentoStreamEndpoints` routing
table.

Pagination is `page_number`: `searchCriteria[currentPage]` (1-based, `start_page: 1`) and
`searchCriteria[pageSize]` (static `page_size: 100`, matching legacy's `magentoDefaultPageSize`).
The engine's `page_number` paginator stops on a short page (`recordCount < page_size`); legacy
additionally checks `total_count` and stops as soon as the accumulated count reaches it. These two
stop conditions are equivalent for every dataset except one whose total record count is an exact
multiple of the page size, where legacy stops immediately via the `total_count` comparison and the
engine would issue one additional request that returns an empty page before stopping — no different
records are ever emitted either way, so this is a request-count-only, non-data-affecting divergence
(documented here per the conventions.md §5 meta-rule, since it is a real behavioral difference even
though not a data one).

Each stream is incremental on `updated_at`, matching legacy's `incrementalLowerBound` (state cursor,
falling back to `start_date`). Legacy expresses the `updated_at > lower_bound` filter as Magento's
three-part `searchCriteria[filter_groups][0][filters][0][field/value/condition_type]` query
convention; this bundle reproduces it via three `stream.Query` entries, each gated on
`{{ incremental.lower_bound }}` with `omit_when_absent: true` so all three are present together on
an incremental read and absent together on a full sync (`field`/`condition_type` use the `const:`
filter to send their fixed literals — `updated_at`/`gt` — only when the lower bound itself resolves;
`value` sends the lower bound's own formatted value). No `incremental.request_param` is declared
since Magento's filter has no single param name to hold — the three `stream.Query` entries carry the
whole filter instead.

## Write actions & risks

None. Magento is a read-only source in legacy (`Capabilities.Write` is `false`); this bundle ships
no `writes.json`.

## Known limits

- **`base_url` requires the fully composed REST base URL; the `store_host`+`api_version` derivation
  is not modeled.** Legacy derives its base URL from a `store_host` config value plus an optional
  `api_version` override, defaulting to `V1`, at request time (`adobe_commerce_magento.go:304-338`,
  `magentoBaseURL`). The engine's `spec.json` `"default"` materialization mechanism
  (conventions.md §3) only fills in a FIXED literal default, not one derived from another config
  value — there is no declarative way to express "concatenate `store_host` + `/rest/` +
  `api_version`" without inventing ad hoc Go (a Tier-2 escalation this bundle does not need for
  anything else). This bundle therefore requires `base_url` directly, already including the
  `/rest/V1` (or other version) suffix; an operator who previously configured `store_host` must now
  configure the fully composed URL instead. This is a documented config-surface narrowing per
  conventions.md §3's "derived default" guidance, not a silent behavior change to any request this
  bundle actually sends once configured.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-300,
  default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides (`magentoPageSize`/`magentoMaxPages`). The engine's `page_number` paginator reads its
  page size only from `PaginationSpec.PageSize`, a bundle-declared static integer with no
  template/config wiring (conventions.md's rate_limit/dead-config precedent applies identically
  here — see stripe's resolved ledger item 3), and there is no config-driven request-count cap
  mechanism at all for the `page_number` paginator. `page_size`/`max_pages` are therefore not
  declared in `spec.json` (a declared-but-unwireable key is worse than an absent one, F6); this
  bundle sends Magento's own effective default (`searchCriteria[pageSize]=100`) as a static
  pagination-block value, matching stripe's `limit=100` precedent.
- **`total_count`-based early stop is approximated by short-page stop only** — see Streams notes
  above; the only observable difference is one extra empty-page request when a stream's total
  record count is an exact multiple of 100, never a difference in which records are emitted.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (reached only
  when `config.mode == "fixture"`) stamps deterministic synthetic records with a `previous_cursor`
  field echoing `req.State["cursor"]` when set (`adobe_commerce_magento.go:253-255`). This is a
  credential-free conformance-harness affordance with no live-path equivalent; this bundle's
  schemas and fixtures target the live record shape only, and the engine's own
  `internal/connectors/conformance` fixture-replay harness provides the credential-free test
  affordance this bundle needs.
