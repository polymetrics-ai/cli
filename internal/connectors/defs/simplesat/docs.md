# Overview

Simplesat is a wave2 fan-out declarative-HTTP migration. It reads Simplesat surveys, answers,
questions, customers, and tickets through the Simplesat API (`GET https://api.simplesat.io/api/...`).
This bundle migrates `internal/connectors/simplesat` (the hand-written connector); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Simplesat API token via the `api_key` secret; it is sent as the `X-Simplesat-Token`
header (`api_key_header` auth mode), matching legacy's `connsdk.APIKeyHeader("X-Simplesat-Token",
key, "")` (`simplesat.go:157`). `base_url` defaults to `https://api.simplesat.io/api` and may be
overridden for tests/proxies, matching legacy's own `validatedBaseURL` default.

## Streams notes

All five streams (`answers`, `surveys`, `questions`, `customers`, `tickets`) hit their own
`GET <resource>/` endpoint with records extracted from the top-level `results` key, matching
legacy's `streamEndpoints` map exactly. None of the streams paginate in legacy (a single
`r.Do` call per read, no loop) — `pagination.type: none` is declared, one request per read.
`page_size` is sent on every request (default `100`, matching legacy's `defaultPageSize`);
`start_date`/`end_date` are optional passthrough query filters, omitted entirely when unset,
matching legacy's `copyConfig` (only set when non-empty). Records carry `id` (integer primary
key), `created_at`, `name`, and `rating`, the exact field set legacy's `streams()` catalog
declares for every stream.

## Write actions & risks

None. Simplesat's legacy connector is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`page_size`'s upper bound (1000) is not enforced by this bundle.** Legacy validates
  `page_size` is an integer between 1 and 1000 in Go code (`simplesat.go:175-185`) before sending
  it. The engine's declarative query dialect has no numeric-range validation primitive; an
  out-of-range `page_size` is sent to the API as-is rather than rejected client-side. This is a
  scope narrowing (client-side validation removed, not a data-shape change): the API itself is
  the ultimate arbiter of an invalid page size on both sides, and no fixture/parity test relies on
  the removed client-side bound.
- **The `Check` request no longer sends `page_size=1`.** Legacy's `Check` explicitly constrains
  the health-check request to one record (`url.Values{"page_size": []string{"1"}}`,
  `simplesat.go:53`); the engine's declarative `check` block is a bare method+path with no query
  support. The check still hits the same `answers/` endpoint and still proves the same
  auth/connectivity condition; only the response page size differs, which does not affect
  `Check`'s pass/fail semantics on either side.
