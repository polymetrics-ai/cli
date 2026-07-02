# Overview

Pendo is a read-only declarative-HTTP connector for the Pendo Engage REST API v1. It reads
visitors, accounts, pages, and features through simple REST list endpoints, without aggregation
writes — matching legacy's own conservative scope. This bundle migrates
`internal/connectors/pendo` (the hand-written connector); the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide a Pendo integration key via the `integration_key` secret; it is sent as the
`x-pendo-integration-key` request header and is never logged.

## Streams notes

All 4 streams (`visitors`, `accounts`, `pages`, `features`) share the identical shape: `GET`
against the Pendo list endpoint (`/visitor`, `/account`, `/page`, `/feature`), records at `data`,
primary key `["id"]`. Pagination follows Pendo's body-token convention (`pagination.type: cursor`
with `cursor_param: page` and `token_path: next`): every request (including the first) sends a
`page` query param, starting at `"1"` (each stream's static `query.page: "1"`), and the next page's
`page` value is read verbatim from the response body's top-level `next` field — identical to
legacy's `harvestPages` loop, which also always sends `page` starting at `"1"` and follows the
opaque `next` token rather than incrementing a counter itself. Pagination stops when `next` is
absent or empty. Every request sends the configured `limit` (`config.limit`, default `100`,
matching legacy's `pendoDefaultLimit`; legacy caps this at 500 as `pendoMaxLimit`, documented here
as an operator-enforced bound since the engine does not itself validate config value ranges).
Every stream carries its legacy cursor field (`lastVisit` for `visitors`/`accounts`, `lastUpdated`
for `pages`/`features`) as its schema `x-cursor-field`, but — matching legacy exactly — no stream
declares a server-side incremental filter: legacy's `Read` never sends a date-range filter
regardless of `req.State`, so no `incremental` block is declared in `streams.json` either.

## Write actions & risks

Not applicable — this connector is read-only (`capabilities.write: false`), matching legacy
exactly.

## Known limits

- Full Pendo API surface (events, reports, guides, polls, and the POST-body aggregation endpoints)
  is out of scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope,
  reason: "Pass B capability expansion"}` entries. Pendo's aggregation reporting API in particular
  uses POST-body query payloads incompatible with this bundle's simple GET-list streams and would
  need Tier-2/3 treatment if added later.
- No incremental sync mode is wired for any stream (see Streams notes) — this mirrors legacy's own
  full-refresh-only behavior, not a capability gap introduced by migration.
- Legacy validates `limit` is between 1 and 500 and `max_pages` is a non-negative integer or
  `all`/`unlimited` at read time; the engine does not perform bundle-declared numeric-range
  validation on config values, so an out-of-range `limit` would be sent to the API as-is. This does
  not change accepted-input behavior for any value legacy itself would accept.
