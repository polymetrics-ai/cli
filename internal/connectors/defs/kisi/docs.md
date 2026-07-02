# Overview

Kisi reads Kisi physical access-control data — members, locks, groups, users, and logins — through
the Kisi REST API (`https://api.kisi.io`). This bundle migrates `internal/connectors/kisi` (the
hand-written connector) to a declarative defs bundle at capability parity; the legacy package stays
registered and unchanged until wave6's registry flip. The API is full-refresh only (no incremental
cursor) and exposes no safe reverse-ETL writes for a physical access-control system, so
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Auth setup

Provide a Kisi API key via the `api_key` secret. Kisi does not use a plain Bearer scheme: legacy
sends `Authorization: KISI-LOGIN <api_key>` (`connsdk.APIKeyHeader("Authorization", secret,
"KISI-LOGIN ")`). This bundle reproduces the identical header via `streams.json` `base.auth`'s
`api_key_header` mode: `{"mode": "api_key_header", "header": "Authorization", "prefix": "KISI-LOGIN
", "value": "{{ secrets.api_key }}"}`, which the engine's `auth.go` resolves to the exact same
`connsdk.APIKeyHeader(header, value, prefix)` call legacy makes directly. The secret only ever flows
into the constructed authenticator; it is never logged.

## Streams notes

All 5 streams (`members`, `locks`, `groups`, `users`, `logins`) share the identical shape: `GET`
against the Kisi list endpoint, records at the response body's top-level JSON array (`records.path:
""`), primary key `["id"]`. Pagination is `offset_limit` (`limit_param: limit`, `offset_param:
offset`, `page_size: 100`, matching legacy's `kisiDefaultPageSize`/`OffsetPaginator`); the engine
stops once a page returns fewer than `page_size` records. `max_pages` defaults to `0` (unlimited),
matching legacy's `kisiMaxPages` default.

No stream declares an `incremental` block or `x-cursor-field`: legacy's `kisiStreams()` catalog
itself publishes no `CursorFields` for any stream (Kisi's API is full-refresh only), so this bundle
matches that exactly — every read is a full stream scan.

## Write actions & risks

None. Kisi is read-only for pm; `capabilities.write` is `false` and no `writes.json` is shipped,
matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`). A physical door-unlock
mutation endpoint exists on the Kisi API but is deliberately excluded (`destructive_admin`) — legacy
never implemented it and it is unsafe for unattended reverse-ETL.

## Known limits

- Full Kisi API surface (events, elevator floors, lock-unlock mutations) is out of scope for this
  wave; see `api_surface.json`'s `excluded` entries. Only the 5 legacy-parity read streams are
  implemented.
- No incremental sync is available for any stream — this matches legacy's real behavior (a
  full-refresh-only API), not a scope-narrowing introduced by this migration.
