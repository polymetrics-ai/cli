# Overview

WaiterAid is a restaurant reservation and table-management platform. This bundle reads restaurant
reservations from the WaiterAid API (`GET {base_url}/reservations`). It migrates
`internal/connectors/waiteraid` (the hand-written legacy connector), which stays registered and
unchanged until wave6's registry flip. Read-only; a single `reservations` stream, no pagination.

## Auth setup

Two secrets are required: `auth_hash` and `restid`, WaiterAid's own header-pair authentication
scheme. Both are sent as static request headers (`X-Auth-Hash` / `X-Restaurant-ID`) on every
request via `streams.json`'s `base.headers` — WaiterAid does not use a Bearer/Basic/API-key-query
scheme, so `base.auth` is declared as a single unconditional `{"mode": "none"}` (the credentials
flow entirely through the two headers, not through the `auth` dispatch), matching legacy's
`requester` which builds a plain `DefaultHeaders` map with no `connsdk.Auth` set. Both secrets are
required in `spec.json`; an absent header-templated secret is always a hard validate/runtime error,
matching legacy's own `Check`/`requester` validation.

`base_url` defaults to `https://api.waiteraid.com`, matching legacy's `defaultBaseURL` constant,
materialized via `spec.json`'s `"default"` value.

## Streams notes

`reservations` reads `GET /reservations` and extracts records from the response's top-level
`reservations` array, matching legacy's `connsdk.RecordsAt(resp.Body, "reservations")`. Legacy maps
`id`/`guest_name`/`date`/`status` straight through with no renaming or stringification
(`item["id"]`, not `text(item["id"])`) — the wire shape already carries `id` as a string, so plain
schema projection (exact key match) reproduces this without any `computed_fields`.

`start_date` is an OPTIONAL config value legacy passes straight through as a `start_date` query
parameter whenever configured (`q.Set("start_date", start)` — no incremental cursor tracking, no
state persistence, just a static request-time filter). This bundle wires the identical behavior via
`stream.Query`'s opt-in optional-query dialect: `{"template": "{{ config.start_date }}",
"omit_when_absent": true}` — sent only when `start_date` is configured, omitted entirely otherwise,
exactly like legacy's `if start := ...; start != ""` guard. `date` is declared as `x-cursor-field`
for catalog/manifest parity (legacy's own `Catalog` declares `CursorFields: []string{"date"}`), but
— matching legacy exactly — no `incremental` block is declared on the stream: legacy never advances
a persisted cursor or filters by `date` on repeat syncs, so this bundle does not either.

## Write actions & risks

None. `capabilities.write` is `false`; legacy's `Write` is an unconditional
`connectors.ErrUnsupportedOperation` stub, and this bundle ships no `writes.json`.

## Known limits

- Only the `reservations` stream is migrated (legacy's own only stream). WaiterAid's broader API
  surface (tables, guests, waitlists, availability windows, etc.) is intentionally out of scope for
  this wave — see `api_surface.json`.
- No pagination: legacy's `Read` issues exactly one unpaginated request per sync, so this bundle
  declares no `pagination` block, matching legacy's actual behavior exactly.
- `date` is catalog-declared as a cursor field (parity with legacy's `Catalog`) but is not a
  functioning incremental cursor in either legacy or this bundle — `start_date` is a plain
  passthrough filter, not a state-tracked lower bound. This is an honest reflection of legacy's own
  behavior, not a scope narrowing introduced by migration.
