# Overview

ShopWired is a fresh Tier-1 (pure declarative) migration at legacy capability parity, porting
`internal/connectors/shopwired` (`shopwired.go`). It reads ShopWired products, orders, customers,
and categories through the ShopWired REST API. Read-only: legacy's `Write` always returns
`ErrUnsupportedOperation` (`shopwired.go:91-93`), and this bundle declares `capabilities.write:
false` with no `writes.json` to match. The legacy package stays registered and unchanged until the
wave6 registry flip.

## Auth setup

Provide one secret: `api_key`, sent as the `X-API-Key` request header on every call
(`streams.json`'s `base.auth`: `{"mode": "api_key_header", "header": "X-API-Key", "value": "{{
secrets.api_key }}"}`), matching legacy's `connsdk.APIKeyHeader("X-API-Key", token, "")`
(`shopwired.go:130`). `base_url` defaults to `https://api.shopwired.co.uk` and is overridable for
tests/proxies, matching legacy's `shopwiredDefaultBaseURL` fallback.

## Streams notes

Four streams, all primary-keyed on `id`, all sharing the identical shape: `GET`, a top-level JSON
array response (`records.path: "."`), `page`/`per_page` `page_number` pagination starting at page 1
with a 100-record page size (legacy's `shopwiredDefaultPageSize`), stopping on a short/empty final
page. `products`, `orders`, `customers`, `categories` map to `/products`, `/orders`, `/customers`,
`/categories` respectively (`shopwiredEndpoints`, `shopwired.go:139-144`).

Legacy's single shared `shopwiredRecord` mapper (`shopwired.go:146-148`) is applied uniformly across
all four endpoints and always emits the SAME six keys on every record regardless of stream —
`id`, `name`, `sku`, `email`, `status`, `updated_at` — leaving whichever fields don't apply to that
endpoint as `nil`. This bundle's four schemas each declare all six properties (not just the fields
that logically belong to that resource) specifically to reproduce that legacy shape field-for-field;
omitting the endpoint-irrelevant fields from a stream's schema would silently narrow the emitted
record compared to legacy under schema-mode projection.

`updated_at` is the incremental cursor field on every stream (`x-cursor-field`), matching legacy's
uniform `CursorFields: []string{"updated_at"}` (`shopwired.go:169`); no `incremental` block is
declared on any stream because legacy's `Read` never sends a server-side incremental filter
parameter and publishes no lower-bound-aware query behavior — `updated_at` is catalog-published
metadata only, not a wired incremental read (matches conventions.md §8 rule 2's truth table:
`x-cursor-field` in schema, no `incremental` block, since legacy sends no server-side filter).

## Write actions & risks

None — ShopWired is read-only. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`shopwired.go:91-93`).

## Known limits

- **Legacy's `first(item, "id", "order_id")` / `first(item, "name", "title")` /
  `first(item, "updated_at", "modified_at")` defensive-fallback field aliasing
  (`shopwiredRecord`, `shopwired.go:146-148, 180-187`) is not reproduced.** The engine's
  `computed_fields` dialect has no coalesce/fallback filter (conventions.md §3: a field can be
  renamed, filtered, or joined, but two `computed_fields` entries cannot target the same output key
  as an "or" — the JSON map itself only allows one value per key), and no evidence anywhere in the
  legacy test suite, fixtures, or the ShopWired API's own documented response shape
  (https://www.shopwired.com/api) shows `order_id`/`title`/`modified_at` ever actually appearing in
  place of `id`/`name`/`updated_at` on any of the four endpoints — every real and fixture-recorded
  response uses `id`/`name`/`updated_at` uniformly. This bundle emits `{{ record.id }}`/
  `{{ record.name }}`/`{{ record.updated_at }}` directly (typed bare-reference extraction, so `id`
  keeps its native numeric wire type instead of legacy's implicit `any`-typed passthrough). This is
  an ACCEPTABLE deviation per conventions.md §5's meta-rule: it never diverges from legacy for any
  input the real API is documented to send (the fallback names have no confirmed real-world
  trigger), and only a currently-unobserved wire shape (an `order_id`-only or `title`-only response
  with no `id`/`name` at all) would produce a different result than legacy's fallback — see the
  parity-deviation ledger entry below.
- **No config-driven `page_size`/`max_pages` runtime override.** Legacy accepted `page_size`
  (falling back to `limit`) and `max_pages` (`0`/`all`/`unlimited` meaning uncapped) as
  caller-supplied config overrides (`pageSize`/`maxPages`, `shopwired.go:215-239`). The engine's
  `page_number` pagination spec's `page_size`/`max_pages` fields are static integers on
  `streams.json`'s `base.pagination` block, not `{{ }}`-templated — there is no per-request
  config-driven override mechanism for either (identical to searxng's documented
  `page_size`/`max_pages` gap, conventions.md §1's read-only/no-auth worked example). Both
  properties are therefore NOT declared in `spec.json` (a declared-but-unwireable key is dead config
  worse than an absent one, conventions.md F6); the bundle hard-codes legacy's own default
  (`page_size: 100`, uncapped `max_pages`), matching legacy's behavior whenever the caller does not
  override either config key.
