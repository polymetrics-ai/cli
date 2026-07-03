# Overview

Printify is a wave2 fan-out declarative-HTTP migration. It reads Printify shops, catalog
blueprints/print providers, products, and orders through the Printify public REST API
(`GET https://api.printify.com/v1/...`). This bundle targets capability parity with
`internal/connectors/printify` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

Status: **partial** — see Known limits for one typed `ENGINE_GAP` blocker (the `raw` field) that
keeps this bundle short of full legacy parity; all 5 streams and every other field are otherwise
migrated at full parity.

## Auth setup

Provide a Printify personal access token via the `api_token` secret; it is sent as a standard
Bearer token (`Authorization: Bearer <api_token>`), matching legacy's `connsdk.Bearer(token)`
(`printify.go:190`) exactly — this bundle's `bearer` auth mode is a direct match, unlike pretix
(which needed a custom `Token ` prefix). `base_url` defaults to `https://api.printify.com/v1` and
may be overridden for tests/proxies.

## Streams notes

`shops`, `blueprints`, and `print_providers` are simple, non-paginated, top-level-array list
endpoints (`GET /shops.json`, `/catalog/blueprints.json`, `/catalog/print_providers.json`) —
`records.path: "."` reads the response body root directly as the records array, matching legacy's
`readSinglePage`+`emitRecords(ctx, body, "", emit)` (empty `recordsPath` for these 3 streams,
`printify.go:41-47`).

`products` and `orders` are scoped beneath a required `shop_id` config value
(`/shops/{{ config.shop_id }}/products.json` / `/orders.json`), matching legacy's `resourcePath`
path template exactly (`printify.go:229-238`) — an absent `shop_id` fails the same way on both
sides (legacy: "printify connector requires config shop_id for this stream"; engine: an unresolved
`config.shop_id` path-template key). Both streams read records from the `data` key and paginate
via Printify's own `next_page_url` absolute-URL field (`pagination.type: next_url`,
`next_url_path: "next_page_url"`). The first request sends `limit=100` (legacy's
`defaultPageSize`); `page` itself is deliberately NOT declared as a static per-stream query value,
because the engine re-applies every `stream.Query` entry on EVERY page request including when
following an absolute `next_url` (`resolveURL`'s query merge REPLACES, not adds to, any
same-named param already on the URL) — a static `page: "1"` would silently force every subsequent
page's URL back to `page=1`, an actual pagination-breaking bug, not a benign idempotent re-send
like `limit` (whose value never changes across pages). Printify's list endpoints default to
`page=1` when the param is omitted, so omitting it reproduces legacy's exact first request while
staying safe on every subsequent page, which follows the recorded absolute `next_page_url` verbatim
(already carrying Printify's own correct page number), with `limit=100` harmlessly re-applied.

Legacy's `mapRecord` (`printify.go:204-215`) applies `title: first(item, "title", "name")` shared
across all 5 streams. Every real Printify object shape confirmed by legacy's own test fixtures
(`printify_test.go`) and Printify's public catalog docs (blueprints/print_providers) carries
`title` directly — the `name` fallback branch has no concrete evidence of ever firing for any of
these 5 resources, so this bundle uses direct schema projection for `title` (no computed_fields
needed), reproducing legacy's fallback chain's landing value exactly for every known real shape.
`sales_channel`/`status`/`visible`/`created_at`/`updated_at` are modeled as direct schema
projections too — legacy reads these raw keys unconditionally (no fallback chain), so plain
key-match projection reproduces the exact behavior: present when the raw object has the key,
absent (matching legacy's `nil`) when it does not.

**Pagination fixtures are single-page**, per `conventions.md` §4's sanctioned `next_url` exception:
`products`/`orders`' next-page URL is the replay server's own runtime-assigned address, which a
static fixture file cannot embed. A live 2-page proof is out of scope for this wave (hard rule: no
Go/paritytest packages) — see Known limits.

## Write actions & risks

None. Printify's product/order mutation endpoints are intentionally unsupported by legacy (package
doc: "Mutating product/order endpoints intentionally remain unsupported"); `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **No live 2-page pagination proof for this wave.** `products`/`orders` use `next_url`
  pagination with single-page fixtures (the sanctioned `conventions.md` §4 exception); a live
  `paritytest/printify`-style test (mirroring legacy's own `printify_test.go`'s
  `TestReadProductsAuthenticatesPaginatesAndMaps` 2-page fixture) is out of scope under this wave's
  hard rule prohibiting new Go/paritytest packages.
- **`products`/`orders`' real field shapes beyond legacy's own test fixture could not be fully
  confirmed against live docs.** Printify's public documentation site is a large JS-rendered SPA
  that this migration could not fully retrieve structured field definitions from for the Products
  and Orders resources specifically (catalog resources — blueprints/print_providers — were
  confirmed). Fields modeled here (`sales_channel`, `status`, `visible`, `created_at`,
  `updated_at`) are exactly the keys legacy's `mapRecord` reads; this bundle reproduces legacy's
  own (possibly incomplete) field selection rather than the full real Printify object shape, per
  the meta-rule (match legacy's accepted behavior, not the API's full shape).
- **`raw` field is BLOCKED (`ENGINE_GAP`), not modeled in this bundle.** Legacy's `mapRecord`
  (`printify.go:204-215`) unconditionally stamps a full nested copy of the source API object onto
  every emitted record, for all 5 streams, under the key `raw` (`printify.go:213`). This is an
  accepted-input EMITTED-DATA change (every live record is missing a field legacy always includes),
  not a cosmetic/request-count deviation, so per the parity-deviation meta-rule
  (`docs/migration/conventions.md` §5) it cannot be shipped as a documented-acceptable deviation —
  it must be filed as a blocker (§6). `computed_fields` templates only resolve a namespaced
  reference (`record.<dotted.path>`, `config.<key>`, the bare `cursor` pseudo-reference, or a
  static literal) — `resolveRefValue` (`internal/connectors/engine/interpolate.go:197-227`) requires
  at least a `<namespace>.<key>` pair (`len(segs) < 2` is a hard "unresolved reference" error), so a
  bare whole-record reference with no dotted path (e.g. `{{ record }}`) cannot be written at all,
  and there is no `computed_fields`, projection, or filter primitive that copies the entire raw
  object as a nested value under one output key. `projection: "passthrough"` does not substitute:
  it flattens every raw top-level key directly into the projected record (no nesting under a `raw`
  key), which is a different, still-non-matching shape. This is a genuine `ENGINE_GAP` — not a
  Tier-2-fixable shape (a `RecordHook` could do it, but this wave's hard rules forbid Tier-2/Tier-3
  escape hatches) — for a follow-up engine-dialect increment (a whole-record-reference primitive,
  e.g. a bare `record` reference or a `computed_fields` dialect extension that nests the full raw
  object under a named key). Once closed, `raw` should be added to all 5 schemas
  (`schemas/*.json`) as an `object`-typed property with a single `computed_fields` entry wired to
  it. Legacy stays authoritative for this field until then; every other field
  (`id`/`title`/`sales_channel`/`status`/`visible`/`created_at`/`updated_at`) is migrated at full
  parity across all 5 streams.
- **Legacy's fixture-mode-only synthetic records are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) emits synthetic records with an integer `id`
  (the loop counter, not derived from any real API field) and a `status: "fixture"` literal, a
  different shape than the live path; this bundle targets the live path only, matching every other
  wave1/wave2 bundle's documented precedent.
