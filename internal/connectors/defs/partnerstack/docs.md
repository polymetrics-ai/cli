# Overview

PartnerStack is a read-only declarative-HTTP connector for the PartnerStack REST API v2. It reads
partnerships, customers, transactions, and groups. This bundle migrates
`internal/connectors/partnerstack` (the hand-written connector); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a PartnerStack API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged.

## Streams notes

All 4 streams (`partnerships`, `customers`, `transactions`, `groups`) share the identical shape:
`GET` against the PartnerStack list endpoint, records at `data`, primary key `["id"]`. Pagination
follows PartnerStack's cursor-token convention (`pagination.type: cursor` with
`cursor_param: cursor` and `token_path: pagination.next`): the next page's `cursor` query param is
read from the response body's `pagination.next` field, and pagination stops when that field is
absent or empty — identical to legacy's `harvestCursor` loop. Every request sends the configured
`limit` (`config.limit`, default `100`, matching legacy's `partnerStackDefaultLimit`; legacy caps
this at 250, documented here as an operator-enforced bound since the engine does not itself
validate config value ranges). Every stream carries `created_at` as its catalog cursor field
(matching legacy's `CursorFields`), but — matching legacy exactly — no stream declares a
server-side incremental filter: legacy's `Read` never sends a date-range filter regardless of
`req.State`, so no `incremental` block is declared in `streams.json` either (declaring one would
add filtering behavior legacy never had).

## Write actions & risks

Not applicable — this connector is read-only (`capabilities.write: false`), matching legacy
exactly.

## Known limits

- Full PartnerStack API surface (rewards, leads, deals, links, custom fields) is out of scope for
  this wave; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries.
- No incremental sync mode is wired for any stream (see Streams notes) — this mirrors legacy's own
  full-refresh-only behavior, not a capability gap introduced by migration.
- Legacy validates `limit` is between 1 and 250 and `max_pages` is a non-negative integer or
  `all`/`unlimited` at read time; this bundle's `spec.json` declares the same defaults but the
  engine does not perform bundle-declared numeric-range validation, so an out-of-range `limit`
  would be sent to the API as-is rather than rejected client-side before the request. This does not
  change accepted-input behavior for any value legacy itself would accept.
