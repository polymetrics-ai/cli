# Overview

Omnisend is a wave2 fan-out declarative-HTTP migration. It reads Omnisend contacts, campaigns,
carts, orders, and products through the Omnisend REST API (`GET https://api.omnisend.com/v3/...`).
This bundle targets capability parity with `internal/connectors/omnisend` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide an Omnisend API key via the `api_key` secret; it is sent as the `X-API-KEY` header
(`pagination.type: next_url` covers pagination separately), matching legacy's
`connsdk.APIKeyHeader("X-API-KEY", secret, "")` exactly, and is never logged.

## Streams notes

All 5 streams (`contacts`, `campaigns`, `carts`, `orders`, `products`) share the identical request
shape: `GET` against the Omnisend list endpoint with `limit={{ config.page_size }}` (default
`100`, matching legacy's `omnisendDefaultPageSize`). Each stream's records live at a
resource-specific JSON key (`contacts`/`campaign`/`carts`/`orders`/`products` — note `campaign` is
singular, matching legacy's `omnisendStreamEndpoints` table exactly). Pagination follows
Omnisend's `paging.next` absolute-URL convention (`pagination.type: next_url`, `next_url_path:
"paging.next"`): the next page is requested at the exact URL the API returns (which already
carries the cursor and limit), matching legacy's `harvest` function, which passes `paging.next`
straight through via `connsdk.Requester.resolveURL`. Pagination stops when `paging.next` is
null/absent, exactly as legacy stops on an empty string.

None of the 5 endpoints expose a server-side incremental filter parameter (legacy's `Read` never
sends a date-scoped query param); `createdAt` is carried as a cursor field on every schema for
catalog purposes only, matching legacy's own comment ("Omnisend only supports full_refresh
upstream, but each resource carries a createdAt/updatedAt timestamp that we surface as a cursor
field"). No `incremental` block is declared on any stream, matching legacy's full-refresh-only
behavior exactly.

## Write actions & risks

None. Omnisend is a read-only source connector (`capabilities.write: false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **The fixture-replay harness cannot exercise `next_url`'s real 2-page continuation for any
  stream here (structural, not connector-specific).** A `next_url` stream's next-page URL is the
  replay server's own address, unknown until the harness picks a port at runtime — a static
  fixture file cannot embed the correct absolute URL for a second page (`docs/migration/
  conventions.md` §4's sanctioned `next_url` single-page-fixture exception). Every one of
  omnisend's 5 streams uses `next_url` pagination (unlike bitly, which has 3 non-paginated
  streams available to satisfy `pagination_terminates` against instead), so — unlike bitly —
  there is no alternate non-paginated stream in this bundle for that check to target; it runs
  against `contacts` (the first-declared stream) and passes trivially against the single-page
  fixture (one request, no `paging.next` present, matching a genuine last page). Real 2-page
  `next_url` correctness is not proven by this wave's fixtures for any omnisend stream. Per
  conventions.md's guidance, the authoritative live proof would be a `paritytest/omnisend` test
  driving a real `httptest.Server`; per this migration wave's hard rule (JSON + docs.md only, no
  Go), no such test was authored here — a follow-up wave with Go authoring in scope should add
  one (mirroring bitly's `TestParityBitly_BitlinksStreamPaginates`).
- **`max_pages` is not runtime-enforced beyond the engine's generic hard cap.** Legacy exposes
  `max_pages` as a config-driven request-count override read fresh on every `Read` call
  (`omnisendMaxPages`). The `next_url` paginator type has no per-request page-count field of its
  own in `PaginationSpec` beyond the generic `MaxPages` hard cap (`read.go`'s `readDeclarative`
  loop enforces it when set to a positive integer) — this bundle declares `max_pages` in
  `spec.json` (default `"0"`, matching stripe's precedent) but does not wire it into
  `pagination.max_pages`, matching the exact gap class bitly/guru document for their own
  non-`page_number` paginators.
- **Legacy's fixture-mode-only fields (`fixture: true`, `previous_cursor`) are not modeled.**
  Legacy's `readFixture` path (only reached when `config.mode == "fixture"`) stamps synthetic
  fields not present in the live API shape. This bundle's schemas and fixtures target the live
  record shape only; the engine's own conformance/fixture-replay harness provides the
  credential-free test affordance legacy's fixture mode was built for.
