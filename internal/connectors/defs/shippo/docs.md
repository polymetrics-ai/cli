# Overview

Shippo is a wave2 fan-out declarative-HTTP migration. It reads Shippo addresses, parcels,
shipments, and transactions through the Shippo REST API (`GET https://api.goshippo.com/...`). This
bundle targets capability parity with `internal/connectors/shippo` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip. Shippo is
read-only (`capabilities.write` is `false`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a Shippo API token via the `api_token` secret; it is sent as an `Authorization` header with
the `ShippoToken ` prefix (`Authorization: ShippoToken <api_token>`) and never logged, matching
legacy's `connsdk.APIKeyHeader("Authorization", token, "ShippoToken ")` (`shippo.go:142`).
`base_url` defaults to `https://api.goshippo.com` (legacy's `shippoDefaultBaseURL`) and may be
overridden for tests/proxies.

## Streams notes

All four streams (`addresses`, `parcels`, `shipments`, `transactions`) share the identical shape:
`GET /<resource>?results=<page_size>&page=1` (legacy's `harvest`, `shippo.go:95-131`), records at
the top-level `results` key, followed by Shippo's own `next` absolute-URL cursor
(`pagination.type: next_url`, `next_url_path: "next"`) until it is null/empty — matching legacy's
own "stop when `next` is empty or the page has zero records" rule (`shippo.go:125-126`). `results`
is declared as a static per-stream query value (`results=100`, legacy's own
`shippoDefaultPageSize`), re-sent on every page request the engine issues; see Known limits for the
re-sent-vs-legacy's-behavior divergence already ledgered on other `next_url` bundles this wave
(bitly, aircall, sendgrid).

Shippo's real wire shape stamps every object type with generalized `object_id`/`object_owner`/
`object_created`/`object_updated` fields (confirmed against Shippo's own API-objects reference,
`docs.goshippo.com/docs/API_Concepts/api-objects`) rather than legacy's defensively-coded fallback
keys (`id`, `name`, `updated_at`) — legacy's `first(item, "object_id", "id")` /
`first(item, "updated_at", "object_updated")` helper tries the real key SECOND, after a fallback key
that Shippo's real API never actually populates for any of these four resource types. This bundle's
`computed_fields` therefore reference the real, always-present key directly: `id: "{{
record.object_id }}"` (all four streams), `updated_at: "{{ record.object_updated }}"` (all four),
and `name` mapped per-resource to whichever field legacy's fallback chain resolves to in practice —
`addresses`' own plain `name` field (a person/company name Shippo does return on address objects),
and `object_owner` for `parcels`, `shipments`, and `transactions` (Shippo's Shipment and
Transaction schemas define no top-level `name` field, so legacy's `first(item, "name",
"object_owner")` resolves to the always-present `object_owner` username for those objects —
confirmed against Shippo's OpenAPI spec, where `object_owner` is a `required` Shipment property and
a documented Transaction property). Legacy's `shippoRecord` (`shippo.go:159-161`) field-builds the
identical five-key record — `id`, `name`, `email`, `status`, `updated_at` — for ALL four streams
and emits it verbatim into the warehouse; catalog `Fields` does not filter emitted records, so
`name` must be mapped on every stream, not only the two whose catalog `Fields` happen to declare it.
`status` and `email` pass straight through via schema projection (Shippo's real field names match
legacy's direct, non-fallback `item["status"]`/`item["email"]` reads exactly); `email` is absent on
Shipment/Transaction objects, so it resolves to null on those streams exactly as it does in legacy.

## Write actions & risks

None. Shippo's address/parcel/shipment/transaction read endpoints have no obviously-safe
reverse-ETL writes in legacy (legacy's own package doc: "read-only native Shippo connector");
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's `id`/`updated_at` fallback keys are not modeled — Shippo's real API never exercises
  them.** Legacy's `first(item, "object_id", "id")` and `first(item, "updated_at",
  "object_updated")` helpers try a non-`object_`-prefixed key FIRST as a defensive fallback; Shippo's
  confirmed real wire shape (every object type stamps `object_id`/`object_updated`, never a bare
  `id`/`updated_at` alongside them) means the fallback branch is dead code for any real Shippo
  response. This bundle's `computed_fields` reference the real key directly. Per conventions.md
  §5's meta-rule, this is ACCEPTABLE: it never changes emitted record DATA for any input Shippo's
  real API would ever send; the untaken fallback branch only matters for a synthetic/malformed
  payload no live Shippo response produces.
- **`next_url` fixtures are single-page, per the sanctioned exception (conventions.md §4).** A
  `next_url` stream's next-page URL is the replay server's own runtime address, unknown until the
  harness picks a port — a static fixture file cannot embed the correct absolute URL for a second
  page. Every stream in this bundle ships a single-page fixture (satisfies `fixtures_present`/
  `read_fixture_nonempty`); `pagination_terminates` passes on the first stream (`addresses`) with
  its single page (`hits == len(pages) == 1`). Real 2-page `next_url`-following correctness for
  THIS bundle's exact request shape is proven by legacy's own existing test
  (`internal/connectors/shippo/shippo_test.go`'s `TestReadAddressesPaginatesAndAuthenticates`, which
  drives a real 2-page `httptest.Server` and asserts the second page is requested via the served
  `next` URL), plus the engine's own generic `next_url` paginator unit tests and read-path
  integration test. This wave does not add a new `paritytest/shippo` package (out of scope per this
  wave's JSON-only mandate).
- **`results`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`/`limit`
  (1-200, default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides (`pageSize`/`maxPages`, `shippo.go:229-254`). The engine's `next_url` paginator has no
  config-driven page-size or request-count-cap knob at all (mirrors bitly's/aircall's/sendgrid's
  identical, already-ledgered limitation this wave); `page_size`/`limit`/`max_pages` are therefore
  not declared in `spec.json`, and this bundle sends Shippo's own default (`results=100`) as a
  static per-stream query literal.
