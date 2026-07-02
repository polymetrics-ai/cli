# Overview

Sage HR is a wave2-fan-out declarative-HTTP migration. It reads Sage HR employees, teams, and time
off requests through the Sage HR API (`GET https://api.sage.hr/v1/...`). This bundle targets
capability parity with `internal/connectors/sage-hr` (the hand-written connector it migrates,
package `sagehr`); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Sage HR API key via the `api_key` secret; it is sent as the `X-Auth-Token` request header
(`api_key_header` auth mode), never logged, matching legacy's
`connsdk.APIKeyHeader("X-Auth-Token", token, "")` (`sage_hr.go:143`). `base_url` defaults to
`https://api.sage.hr/v1` and may be overridden for tests/proxies.

## Streams notes

All 3 streams (`employees`, `teams`, `timeoff_requests`) are single-request, non-paginated `GET`
list endpoints — legacy's `Read` issues exactly one request per stream with no pagination loop at
all (`sage_hr.go:92`). None declare an `incremental` block, matching legacy's `Catalog` (no
`CursorFields` declared for any of the three streams).

Legacy's `Read` performs **zero field mapping**: it decodes the response body via
`recordsAtAny(resp.Body, "data", "")` (try a `"data"` envelope key first, fall back to the bare
root array if `"data"` is absent/empty) and emits each decoded object as `connectors.Record(rec)` —
a direct cast, not a filtered/renamed projection. This bundle therefore declares
`"projection": "passthrough"` for all 3 streams (every raw field survives, matching legacy's
pass-everything-through behavior) rather than `"schema"` mode, which would silently drop any raw
field not declared in `schemas/*.json`.

`records.path` is declared as `""` (root) for all 3 streams, matching the ONLY wire shape legacy's
own test suite proves (`sage_hr_test.go`'s `employees` fixture returns a **bare top-level JSON
array**, `[{"id":1,...}, ...]`, with no `"data"` envelope key at all) — `recordsAtAny`'s `"data"`-first
lookup only matters when a `"data"` key is actually present; for the proven bare-array shape it
falls through to `""` regardless of declaration order, so a fixed `""` path reproduces the exact,
tested behavior. `teams` and `timeoff_requests` share the identical `recordsAtAny(resp.Body, "data",
"")` call with no dedicated test coverage of their own; this bundle applies the same bare-root-array
assumption to them absent contrary evidence (see Known limits).

## Write actions & risks

None. Legacy's `Write` unconditionally returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`developers.sage.hr` was unreachable during this migration** (HTTP 403 on fetch); per
  conventions.md, legacy code (and its test fixtures) is ground truth over docs in this situation.
  Every stream/field/envelope shape above is derived from `sage_hr.go`/`sage_hr_test.go` only.
- **`teams`/`timeoff_requests` envelope shape is inferred, not directly tested.** Legacy's own test
  suite (`sage_hr_test.go`) exercises only the `employees` stream's real wire shape (a bare array);
  `teams` and `timeoff_requests` share the identical `recordsAtAny(resp.Body, "data", "")` call in
  legacy code but have no dedicated httptest coverage proving their actual envelope. This bundle
  assumes the same bare-root-array shape for both (the only shape with direct evidence); if Sage
  HR's real `teams`/`timeoff_requests` endpoints in fact wrap results in a `"data"` key, `RecordsAt`
  would return zero records against this bundle's `records.path: ""` declaration for those two
  streams specifically — a Pass-B/capability-expansion fix (declare `records.path: "data"` instead)
  once live-response evidence is available, not a hook-worthy gap.
- **No pagination is modeled**, matching legacy exactly (`sage_hr.go`'s `Read` issues exactly one
  request per stream, no loop, no page/per_page query params sent at all).
