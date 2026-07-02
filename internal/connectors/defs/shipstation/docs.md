# Overview

ShipStation is a wave2 fan-out declarative-HTTP migration. It reads ShipStation orders, shipments,
products, and customers through the ShipStation REST API v1 (`GET
https://ssapi.shipstation.com/...`). This bundle targets capability parity with
`internal/connectors/shipstation` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip. ShipStation is read-only
(`capabilities.write` is `false`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`).

## Auth setup

Provide `api_key` and `api_secret` secrets; they are sent as HTTP Basic auth
(`api_key:api_secret`, base64-encoded) and never logged, matching legacy's
`connsdk.Basic(key, apiSecret)` (`shipstation.go:131`). `base_url` defaults to
`https://ssapi.shipstation.com` (legacy's `shipstationDefaultBaseURL`) and may be overridden for
tests/proxies.

## Streams notes

All four streams (`orders`, `shipments`, `products`, `customers`) share the identical shape: `GET
/<resource>?page=<n>&pageSize=<size>` (legacy's `harvest`, `shipstation.go:95-119`), records at the
resource-named top-level key (`orders`/`shipments`/`products`/`customers`), 1-based page-number
pagination (`pagination.type: page_number`, `page_param: page`, `size_param: pageSize`,
`start_page: 1`, `page_size: 100` — legacy's own `shipstationDefaultPageSize`); a page shorter than
100 records signals the end, matching legacy's `len(records) < pageSize` stop condition exactly —
the engine's `PageNumberPaginator.Next` uses the identical `recordCount < PageSize` rule.

ShipStation's V1 API's real wire field names (`orderId`/`orderNumber`/`orderStatus`/`modifyDate`
for orders; `shipmentId`/`shipmentStatus` for shipments; `productId` for products; `customerId` for
customers — confirmed against ShipStation's own V1 API reference) are the FIRST key legacy's
`first(item, <realKey>, <fallbackKey>)` helper tries for each field on every stream; this bundle's
`computed_fields` reference those real keys directly (`id: "{{ record.orderId }}"`, `status: "{{
record.orderStatus }}"`, `modified_at: "{{ record.modifyDate }}"`, etc. — see Known limits for the
untaken second-choice fallback keys legacy defensively also tries). `sku` (products) and `email`
(customers) pass straight through via schema projection since legacy reads them directly
(`item["sku"]`/`item["email"]`, no fallback).

None of the four streams expose a request-time incremental filter parameter in legacy — legacy's
own `Stream.CursorFields` catalog metadata (`modified_at` on all four) is purely descriptive
cataloging, never wired into an actual request-narrowing filter anywhere in `harvest`. This bundle
mirrors that exactly: `x-cursor-field` is declared on each schema (matching legacy's catalog
metadata) but no stream declares an `incremental` block, so every read is full refresh, identical to
legacy.

## Write actions & risks

None. ShipStation's order/shipment/product/customer read endpoints have no obviously-safe
reverse-ETL writes in legacy (legacy's own package doc: "read-only native ShipStation connector");
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's `id`/name/status/modified_at fallback keys are not modeled — they are the SECOND
  choice in legacy's own fallback chain, tried only when the real V1 key is absent.** Legacy's
  `first(item, "orderId", "id")` (and the analogous calls for `shipmentId`/`productId`/
  `customerId`, and `first(item, "modifyDate", "updated_at"/"shipDate")`) try the real,
  confirmed-present V1 field FIRST; this bundle's `computed_fields` reference that same real key
  directly. The dialect has no coalesce/fallback filter (only a single template per computed
  field), so the untaken SECOND-choice fallback key cannot be expressed without a
  `RecordHook` (a Tier-2 escalation this wave's JSON-only mandate forbids). Per conventions.md §5's
  meta-rule, this is ACCEPTABLE: it never changes emitted record DATA for any input ShipStation's
  real V1 API would ever send (the fallback key is dead code for a genuine V1 response); it would
  only matter for a synthetic/malformed payload no live ShipStation response produces.
- **`page`/`pageSize`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`/
  `limit` (1-500, default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as
  config-driven overrides (`pageSize`/`maxPages`, `shipstation.go:230-255`). The engine's
  `page_number` paginator's `PageSize` is a fixed bundle-declared value (`streams.json`'s
  `base.pagination.page_size: 100`, matching legacy's own default) with no per-request config
  override wired in this bundle (`stream.Query` templating a `{{ config.page_size }}` value would
  need `omit_when_absent`/`default` to compose with the paginator's own `PageSize` field, which the
  paginator does not read from `stream.Query` at all — `PaginationSpec.PageSize` is the only input);
  `page_size`/`limit`/`max_pages` are therefore not declared in `spec.json`. `max_pages` is likewise
  not modeled; pagination is bounded only by the short-page stop signal.
