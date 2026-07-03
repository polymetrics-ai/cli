# Overview

ConfigCat is a feature flag management platform. This bundle reads ConfigCat organizations,
products, configs, environments, and tags through the ConfigCat Public Management API
(`https://api.configcat.com`). It migrates `internal/connectors/configcat` (the hand-written
legacy connector), which stays registered and unchanged until wave6's registry flip. Read-only:
the ConfigCat Public Management API exposes no reverse-ETL write target used by legacy, matching
legacy's own `Capabilities.Write = false`.

This bundle was UNBLOCKED from `docs/migration/quarantine.json` once the engine gained the
`stream.fan_out` dialect (S4 engine mini-wave item 2) — legacy's `readNested` first lists
`/v1/products`, then issues one request per product id (`/v1/products/{id}/configs` etc.),
stamping `product_id` onto every nested record, which the pre-increment declarative dialect had
no mechanism to express short of a Tier-2 `StreamHook`.

## Auth setup

Provide a ConfigCat Public Management API password via the `password` secret; it is used only for
HTTP Basic auth and is never logged. The Basic auth username is resolved with the same precedence
as legacy's `configcatUsername`: the (non-secret) `username` config value if set, else a
defensively-checked `username` secret if set, else an empty username — expressed as three ordered
`basic` auth candidates gated by `when` (first-match-wins, matching legacy's own
config-then-secrets fallback order exactly). `base_url` defaults to
`https://api.configcat.com` and may be overridden for tests/proxies.

## Streams notes

`organizations` (`GET /v1/organizations`) and `products` (`GET /v1/products`) are flat, top-level
JSON array endpoints (`records.path: ""`), matching legacy's `readList` exactly; ConfigCat's
Public Management API paginates neither (legacy declares no pagination for either), so this
bundle declares no `pagination` block (`type: none`, the default).

`configs`, `environments`, and `tags` are nested-under-product resources: legacy's `readNested`
first lists every accessible product (`GET /v1/products`), then reads the sub-resource once per
product id, stamping `product_id` onto every record. This bundle reproduces that exact pattern
with `stream.fan_out`: `ids_from.request` issues a preliminary `GET /v1/products` listing
(the SAME endpoint the `products` stream itself reads, extracting `productId` off each record);
`into.path_var` makes the resolved product id referenceable in the stream's own `path` as
`{{ fanout.id }}` (e.g. `/v1/products/{{ fanout.id }}/configs`); `stamp_field: product_id` writes
the current product id onto every emitted record of that stream, after projection/computed_fields
— exactly matching legacy's `readList`'s conditional `rec["product_id"] = productID` stamp (the
stamped id and legacy's own nested-`product.productId`-derived fallback are always the same value
for records returned under that product's own endpoint, so the two approaches are behaviorally
identical for every record legacy itself would emit).

Every stream's `product_id`/`organization_id` cross-reference field is a renamed camelCase→snake_case
copy of the raw API field (`{{ record.organizationId }}`, `{{ record.product.productId }}`, etc.),
matching legacy's per-stream `mapRecord` functions field-for-field.

None of the 5 streams exposes a legacy-recognized incremental cursor field — ConfigCat's Public
Management API surfaces configuration metadata, not an event stream; legacy's own catalog
publishes no `CursorFields` for any stream. All 5 streams are full-refresh only.

`check` issues a single bounded `GET /v1/organizations`, mirroring legacy's `Check` implementation
exactly (a bounded read of the organizations list confirms auth and connectivity without
mutating anything).

## Write actions & risks

None. `capabilities.write` is `false`; no `writes.json` is shipped. Legacy itself implements no
write path for ConfigCat (`Write` is a stub returning `ErrUnsupportedOperation`).

## Known limits

- Only the 5 legacy-parity read streams are implemented; see `api_surface.json`. ConfigCat's
  broader documented API surface (webhooks, permission groups, members, etc.) is out of scope
  until Pass B.
- `configs`/`environments`/`tags` fan out across every accessible product; a workspace with many
  products issues one request per product per stream per sync, matching legacy's own `readNested`
  cost profile exactly (no new request-count regression introduced by this migration).
- `fixtures/streams/{configs,environments,tags}/page_1.json` each record the preliminary
  `/v1/products` listing (2 fixture product ids); `page_2.json`/`page_3.json` record the
  per-product sub-resource response for each of those 2 ids, exercising the fan-out path under
  `conformance`'s replay harness end to end (mirrors `cisco-meraki`'s identical fan-out fixture
  shape).
