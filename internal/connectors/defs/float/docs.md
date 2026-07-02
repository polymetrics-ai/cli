# Overview

Float is a wave2 fan-out declarative-HTTP migration. It reads Float people, projects, clients,
tasks, and departments through the Float v3 REST API (`GET https://api.float.com/v3/...`). This
bundle migrates `internal/connectors/float` (the hand-written connector); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Float personal access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`float.go:249`). `base_url` defaults to `https://api.float.com/v3` and
may be overridden for tests/proxies.

## Streams notes

All 5 streams share the identical shape: `GET` against the Float list endpoint, a top-level JSON
array (no envelope — `records.path: "."`), matching legacy's `RecordsAt(resp.Body, "")` call.
Pagination is `pagination.type: page_number` with `page_param: page`, `size_param: per-page`,
`start_page: 1`, `page_size: 200` (legacy's default `floatDefaultPageSize`). Each stream's own
primary key is its resource-specific `<resource>_id` integer field, matching legacy's per-stream
`PrimaryKey`.

Legacy's real stop condition combines THREE signals: the `X-Pagination-Page-Count` response header
(when present, stop once `page >= totalPages`), a short/empty page (fewer than `page_size`
records), or a genuinely empty page. The engine's `page_number` paginator implements only the
short/empty-page signal (`connsdk.PageNumberPaginator.Next`: stop when `recordCount < PageSize`) —
it has no header-reading mechanism at all (no declarative field addresses a response header as a
pagination-stop input). This is DOCUMENTED, not silently divergent: see Known limits below for why
it never changes emitted data for any real Float account.

`page_number`'s only stop signal is a short page (`recordCount < page_size`); it has no
Stripe-style `stop_path`/token equivalent. The required 2-page fixture for `people` (conventions.md
§4) therefore ships a full `page_size`-sized (200-record) page 1 followed by a 1-record page 2 — a
smaller page 1 would trigger the short-page stop after a single request and never demonstrate real
pagination, and lowering the bundle's *declared* `page_size` purely to make a smaller fixture
easier to author would change production request cadence (far more roundtrips than Float's real
default), which this bundle avoids. The other four streams (`projects`, `clients`, `tasks`,
`departments`) keep the base `page_size: 200` and ship one short (`<200`-record) fixture page each
— the same natural short-page stop signal a real Float account with fewer than 200 records of that
type would produce.

## Write actions & risks

None. Float is read-only in legacy (`capabilities.write: false`, `Write` returns
`connectors.ErrUnsupportedOperation`); this bundle ships no `writes.json`.

## Known limits

- **The `X-Pagination-Page-Count` response header is not consulted; only the short/empty-page
  signal stops pagination.** This is an `ENGINE_GAP`-shaped limitation the dialect's `page_number`
  paginator does not close (no pagination field reads an arbitrary response header). In practice
  this never diverges from legacy for a real Float account: Float's own last page is, by
  definition, either short (fewer than `per-page` records) or exactly a multiple of `per-page` —
  in the rare exact-multiple case, this bundle issues ONE extra request past the header-reported
  last page (which returns an empty array, matching legacy's own defensive `len(records) == 0`
  stop condition), then stops. No record is ever duplicated, dropped, or reordered — the extra
  request is a harmless wire-shape difference (one additional round-trip), never an emitted-data
  change, so per conventions.md §5's meta-rule this is an ACCEPTABLE deviation, not an `ENGINE_GAP`
  blocker.
- **`page_size` is not runtime-configurable per this bundle's static `streams.json` pagination
  blocks.** Legacy exposes `page_size` (1-200, default 200) and `max_pages` as config-driven
  overrides (`floatPageSize`/`floatMaxPages`, `float.go:282-310`). `PaginationSpec.PageSize`/
  `MaxPages` are static bundle-level integers with no `{{ config.* }}` templating mechanism —
  matching bitly's/flexmail's identical documented gap. `spec.json` still declares `page_size` for
  operator-facing documentation parity with legacy's config surface, but it is not wired to any
  template; `max_pages` is not declared at all (F6, REVIEW.md).
- **Legacy's fixture-mode-only `connector`/`fixture` marker fields are not modeled.** Legacy's
  `readFixture` path (only reached when `config.mode == "fixture"`) stamps `record["connector"] =
  "float"` and `record["fixture"] = true` onto every fixture-mode record — a credential-free
  conformance-harness affordance, not part of the live wire shape. This bundle's schemas and
  fixtures target the live path only, consistent with bitly's/flexport's identical documented
  scope narrowing.
