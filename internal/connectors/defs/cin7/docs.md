# Overview

Cin7 is a wave2 fan-out declarative-HTTP migration. It reads Cin7 Core (DEAR Inventory) products,
customers, suppliers, sales, and purchases through the Cin7 Core External API v2
(`GET https://inventory.dearsystems.com/externalapi/v2/...`). This bundle targets capability
parity with `internal/connectors/cin7` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip. The connector is read-only.

## Auth setup

Cin7 Core authenticates every request with two headers. The application key is a secret: provide
it via the `api_key` secret, sent as the `api-auth-applicationkey` header
(`{"mode": "api_key_header", "header": "api-auth-applicationkey", "value": "{{ secrets.api_key }}"}`)
and never logged. The account id is not secret-shaped and is sent directly as a declared
`base.headers` entry, `api-auth-accountid: {{ config.accountid }}`; `accountid` is `required` in
`spec.json`, matching legacy's own hard requirement (`cin7.go:79-84`, `requester`).

## Streams notes

All 5 streams share the identical Cin7 Core envelope: `GET` against a list resource, records at a
resource-specific top-level array key (`Products`/`CustomerList`/`SupplierList`/`SaleList`/
`PurchaseList`), primary key `["id"]`. `products`/`customers`/`suppliers` additionally send
`IncludeDeprecated=true` on every request, matching legacy's `streamEndpoint.params`. Pagination is
`page_number` (`page`/`limit`, `start_page: 1`, `page_size: 100` — matches legacy's
`cin7DefaultPageSize`); the engine's short-page stop (`recordCount < page_size`) is identical to
legacy's own `harvest` stop condition. Every stream's raw uppercase Cin7 field names (`ID`, `Name`,
`SKU`, ...) are renamed to the schema's lowercase, legacy-matching field names via
`computed_fields` bare `{{ record.<Field> }}` references (typed extraction preserves each field's
native JSON type — `price_tier1`/`cost`/`invoice_amount` stay numeric, matching legacy's own
pass-through of Cin7's numeric JSON values).

None of the 5 objects expose a documented incremental filter parameter on the Cin7 Core API and
legacy's own `harvest` never applies one; matching legacy, no stream in this bundle declares an
`incremental` block — every read is a full paginated sweep.

## Write actions & risks

None. Cin7 Core is read-only in this bundle (`capabilities.write: false`, no `writes.json`),
matching legacy's `Write` unconditionally returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`accountid` as a secrets-store alias is not modeled.** Legacy's `cin7AccountID` allows the
  account id to arrive via EITHER `cfg.Config["accountid"]` OR `cfg.Secrets["accountid"]`
  (secrets checked first) — some deployments store the (non-secret) account id alongside real
  secrets for convenience. This bundle declares `accountid` as a `spec.json` `config` property
  only; a caller that previously supplied it exclusively via the secrets store must move it to
  config. This is a config-surface narrowing (ACCEPTABLE, never changes accepted-input DATA,
  conventions.md §5) — the resolved account id value once supplied is identical either way.
- **`id`-field fallback chains are narrowed to the primary field only.** Legacy's `firstField`
  helper falls back through a priority list when the primary id-shaped field is absent:
  `products.id` = `firstField(item, "ID", "SKU")`, `sale_list.id` = `firstField(item, "SaleID",
  "ID")`, `purchase_list.id` = `firstField(item, "ID", "TaskID")`. The engine's `computed_fields`
  dialect has no "first of N paths" coalesce primitive — only a single bare `{{ record.<path> }}`
  reference (or a filter chain over one reference). This bundle references only the
  higher-priority field (`ID`/`SaleID`/`ID` respectively); the documented Cin7 Core wire shape
  always populates that field for every real product/sale/purchase record (the fallback exists
  defensively in legacy for malformed/partial API responses, never observed in practice).
  ACCEPTABLE per conventions.md §5's meta-rule: this never diverges for any record the real Cin7
  API actually returns; it is a genuine `ENGINE_GAP` only for the theoretical malformed-response
  case legacy defends against.
- `page_size` (the query size param) is a fixed `streams.json` pagination value (100, matching
  legacy's default); legacy's config-driven `page_size` override (1-1000) is not modeled —
  the engine's `page_number` paginator's `PageSize` is a static value set once in `streams.json`,
  not template-resolvable (same shape as the aha/appfigures wave2 precedent). `max_pages` (legacy's
  0/all/unlimited-or-positive-integer request-count cap) is likewise not modeled — `spec.json`
  intentionally omits both `page_size` and `max_pages` (a declared-but-unwireable key is worse than
  an absent one, per conventions.md F6).
- Legacy's `base_url` SSRF-guard scheme/host validation (https/http only, host required) is
  reproduced by the engine's own base-URL handling; no bundle-level behavior change.
