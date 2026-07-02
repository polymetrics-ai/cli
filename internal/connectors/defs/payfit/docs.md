# Overview

PayFit is a read-only declarative-HTTP connector for the PayFit Partner REST API v1. It reads
employees, contracts, and companies. This bundle migrates `internal/connectors/payfit` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a PayFit partner API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged.

## Streams notes

All 3 streams (`employees`, `contracts`, `companies`) share the identical shape: `GET` against the
PayFit list endpoint, records at `data`, primary key `["id"]`. Pagination follows PayFit's
offset-cursor convention (`pagination.type: cursor` with `cursor_param: offset` and
`token_path: meta.next_offset`): the next page's `offset` query param is read verbatim from the
response body's `meta.next_offset` field, and pagination stops when that field is absent or empty
— identical to legacy's `harvestOffset` loop (legacy treats `meta.next_offset` as an opaque token,
never an integer it increments itself). Every request sends the configured `limit`
(`config.limit`, default `100`, matching legacy's `payfitDefaultLimit`/`payfitMaxLimit`, which are
both 100). Every stream carries `updated_at` as its catalog cursor field (matching legacy's
`CursorFields`), but — matching legacy exactly — no stream declares a server-side incremental
filter: legacy's `Read` never sends a date-range filter regardless of `req.State`, so no
`incremental` block is declared in `streams.json` either.

## Write actions & risks

Not applicable — this connector is read-only (`capabilities.write: false`), matching legacy
exactly.

## Known limits

- Full PayFit API surface (payslips, absences, collective agreements) is out of scope for this
  wave; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries.
- No incremental sync mode is wired for any stream (see Streams notes) — this mirrors legacy's own
  full-refresh-only behavior, not a capability gap introduced by migration.
- Legacy validates `limit` is between 1 and 100 and `max_pages` is a non-negative integer or
  `all`/`unlimited` at read time; the engine does not perform bundle-declared numeric-range
  validation on config values, so an out-of-range `limit` would be sent to the API as-is. This does
  not change accepted-input behavior for any value legacy itself would accept.
