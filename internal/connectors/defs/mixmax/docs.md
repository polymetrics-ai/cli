# Overview

Mixmax is a wave2 fan-out declarative-HTTP migration. It reads Mixmax code snippets, messages,
rules, sequences, and meeting types through the Mixmax REST API (`GET
https://api.mixmax.com/v1/...`). This bundle targets capability parity with
`internal/connectors/mixmax` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Mixmax API token via the `api_key` secret; it is sent as the `X-API-Token` header
(`auth.mode: api_key_header`, `header: X-API-Token`, `prefix: ""`), matching legacy's
`connsdk.APIKeyHeader(mixmaxAuthHeader, secret, "")` (`mixmax.go:246`) exactly — no `Authorization`
header is ever sent, matching legacy's own parity test asserting `Authorization` stays empty. The
secret is never logged. `base_url` defaults to `https://api.mixmax.com/v1` and may be overridden for
tests/proxies.

## Streams notes

All 5 streams (`codesnippets`, `messages`, `rules`, `sequences`, `meetingtypes`) share Mixmax's
body-cursor pagination (`pagination.type: cursor`, `token_path: next`, `cursor_param: next`,
`stop_path: hasNext`): the next page's cursor token is read from the response body's top-level
`next` field and sent back as the `next` query param, and pagination stops when `hasNext` is not the
literal `true` (the engine's `stop_path` falsy-stops rule, conventions.md §3) OR the `next` token
is empty — matching legacy's own `hasNext == "false" || token == ""` stop condition
(`mixmax.go:179-182`) exactly, including the case where `hasNext` is absent/missing (falsy) even if
a stray `next` token were present. Records live at the top-level `results` array for every stream,
keyed on `_id` (`x-primary-key: ["_id"]`), matching legacy's uniform
`{results:[...], next, hasNext}` envelope (`streams.go:17`).

Each stream's schema declares `x-cursor-field` matching legacy's own `CursorFields` declaration
(`codesnippets`/`rules`/`sequences`/`meetingtypes` -> `createdAt`, `messages` -> `updatedAt`) as
descriptive metadata only — **no stream declares an `incremental` block**, because legacy's own
`Read`/`harvest` never sends any filter parameter derived from these cursor fields, and never
filters records client-side either (confirmed by `mixmax.go`'s `Read`/`harvest`: every read is an
unconditional full list, regardless of `CursorFields`). Declaring an `incremental` block here would
therefore invent filtering behavior legacy never had; this bundle stays full-refresh-only, matching
legacy's actual behavior over its declared-but-unused metadata.

## Write actions & risks

None. Mixmax is modeled as a read-only source (legacy's own package doc: "There is no safe
reverse-ETL write surface for the streams modeled here"); `capabilities.write` is `false` and this
bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation` wrapped with a read-only-source message.

## Known limits

- **`CursorFields` are legacy-declared but never enforced — modeled as schema metadata only, not as
  `incremental` behavior.** See Streams notes above; this is not a capability loss (legacy itself
  never implemented incremental filtering for these streams), just an honest non-invention of
  behavior legacy's own code never had.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`mixmaxPageSize`/`mixmaxMaxPages`, `mixmax.go:279-307`, page_size default 50, max 100).
  The engine's `cursor` (token_path) paginator has no `PageSize`/`MaxPages`-equivalent config-driven
  knob at all (unlike `page_number`/`offset_limit`, it never reads a page-size query param —
  Mixmax's own `limit` query param that legacy sends on every page is likewise not modeled, since
  the engine's `tokenPathCursor` paginator does not carry a query-building hook comparable to
  `offset_limit`'s `LimitParam`); neither is declared in `spec.json` (F6, REVIEW.md). Pagination is
  bounded only by the `hasNext`/empty-token stop signal, matching Mixmax's own real termination
  behavior.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) synthesizes deterministic records directly in Go, including a
  `previous_cursor` field echoing `req.State["cursor"]` when set (`mixmax.go:219-223`) — a
  fixture-mode-only, non-live-wire-shape field. This bundle's schemas and fixtures target the LIVE
  `harvest`/`mapRecord` path only; the engine's own conformance/fixture-replay harness provides the
  credential-free test affordance legacy's fixture mode existed for, so no fixture-mode equivalent
  is needed here.
- **`limit` is not sent as a query param.** Legacy always sends `limit=<page_size>` on every request
  (`mixmax.go:150`). The engine's `cursor` (token_path) paginator has no static/config-driven query
  param injection point of its own for this pagination type; a per-stream static `query: {"limit":
  "50"}` value was considered but omitted because it would only ever apply to the FIRST page (the
  paginator's own `Next` builds each subsequent page's query from scratch, containing only
  `cursor_param`), diverging from legacy's every-page behavior in the opposite direction bitly's
  `size` re-send divergence took — declaring it was judged more misleading than omitting it. Mixmax's
  documented default page size (its own API-side default, unconfirmed exact value from the reachable
  docs) governs page size when `limit` is absent, so no data is lost, only the exact page-size
  request shape differs from legacy's explicit `limit=50`.
