# Overview

Factorial (FactorialHR) is an HR platform API. This bundle migrates all 5 legacy
`internal/connectors/factorial` read streams to a Tier-1 defs bundle at full capability parity:
`employees`, `teams`, `leaves`, `leave_types`, `locations`.

## Auth setup

Provide a Factorial API key via the `api_key` secret; it is sent as the `X-API-KEY` header on
every request (never logged). `base_url` defaults to
`https://api.factorialhr.com/api/v2/resources` and only needs overriding for tests or proxies.

## Streams notes

All 5 streams share Factorial's page-increment pagination (`pagination.type: page_number`,
`page_param: page`, `size_param: limit`, `start_page: 1`), records at `data`, matching legacy's
`connsdk.Harvest(..., "data", ...)` call.

- `employees` (`GET /employees/employees`) — primary key `id`, incremental cursor `updated_at`
  (`incremental.client_filtered: true` — Factorial has no server-side `updated_at` filter
  parameter; legacy tracks the max `updated_at` seen client-side via `connsdk.MaxCursor` and the
  engine's `client_filtered` mode reproduces the identical client-side drop-already-seen-records
  behavior).
- `teams` (`GET /teams/teams`) — primary key `id`, full refresh only.
- `leaves` (`GET /timeoff/leaves`) — primary key `id`, incremental cursor `updated_at`
  (`client_filtered: true`, same shape as `employees`).
- `leave_types` (`GET /timeoff/leave_types`) — primary key `id`, full refresh only.
- `locations` (`GET /locations/locations`) — primary key `id`, full refresh only.

`check` requests `GET /api_public/credentials` (matches legacy's `factorialCheckResource`).

## Write actions & risks

None. Factorial is read-only in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`).

## Known limits

- **`employees.full_name`'s empty-string fallback is not modeled (ACCEPTABLE, narrowly
  documented deviation).** Legacy's `factorialEmployeeRecord`
  (`internal/connectors/factorial/streams.go`) reads the API's `full_name` field but falls back
  to `"<first_name> <last_name>"` (trimmed) whenever `full_name` is empty or absent. This bundle
  emits the raw `full_name` field via plain schema projection with no fallback: `computed_fields`
  templates cannot express "use field X unless it's empty, then compute Y instead" — every
  template shape (bare reference, filter chain, or mixed literal+reference) either copies a
  single value type-preserving or unconditionally stringifies a fixed template; none can branch on
  another field's presence/emptiness. The real Factorial `employees` endpoint documents
  `full_name` as a populated, non-optional response field, so legacy's fallback branch is
  defensive code for a shape the live API does not produce — this deviation therefore never
  diverges from legacy for any real Factorial response, only for a synthetic/malformed one
  omitting `full_name` entirely (which cannot occur against the documented API). Per
  `docs/migration/conventions.md` §5's meta-rule this is ACCEPTABLE; it is not an `ENGINE_GAP`
  blocker because the branch it approximates is unreachable against the real API surface.
- `page_size`/`max_pages`/`limit` config knobs from legacy are not declared in `spec.json`:
  pagination fields (`streams.json`'s `base.pagination` block) are plain Go values, not
  template-interpolated, so there is no mechanism to wire a runtime config value into them — a
  fixed `page_size: 50` is used instead, matching legacy's own default
  (`factorialDefaultPageSize = 50`, recorded as `metadata.json`'s `batch.read_page_size` for
  operator awareness) and no `max_pages` cap (unbounded, matching legacy's own default).
- `rate_limit` is not declared on `streams.json`'s `base` block: legacy enforces no client-side
  rate limiting, so none is added here.
