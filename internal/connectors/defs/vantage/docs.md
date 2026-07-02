# Overview

Vantage is a cloud cost visibility platform. This bundle reads **cost** records from the Vantage
API (`GET {base_url}/v2/costs`, default `https://api.vantage.sh`). Read-only; it migrates
`internal/connectors/vantage` (185 loc), which stays registered and unchanged until wave6's
registry flip.

## Auth setup

Vantage authenticates via a single secret, `access_token`, sent as a Bearer token
(`Authorization: Bearer <access_token>`), matching legacy's `connsdk.Bearer(token)` requester
exactly (`vantage.go:107-117`).

## Streams notes

One stream: `costs`, `GET /v2/costs`, records extracted from the `costs` array (`vantage.go:88`),
each record mapped field-for-field to `{id, service, amount, date}` (`vantage.go:133-135`). No
pagination and no query parameters at all — legacy issues exactly one unparameterized request per
read (`vantage.go:84`); no incremental cursor either (legacy's `costsStream()` declares no
`CursorFields` and `Read` never builds a query at all).

**`id` is force-cast to a string.** Legacy's `costRecord` calls `stringValue(item["id"])`
(`vantage.go:134,173-184`), converting a raw JSON value into a Go string. This is reproduced via
`{{ record.id | last_path_segment }}` rather than a bare `{{ record.id }}` reference: the bare form
would trigger the engine's typed-extraction path (preserving the raw wire type), which would
diverge from legacy's deliberate string cast. `last_path_segment`'s documented contract
guarantees a slash-free value "passes through unchanged" (`docs/migration/conventions.md` §3), so
this filter application is a pure identity transform whose only effect is forcing the value
through `Interpolate`'s stringify path — the same byte-exact reproduction technique used for
`uservoice`'s `id` field (see that bundle's `docs.md` for the fuller rationale). Schema declares
`id` as `"type": "string"` accordingly.

## Write actions & risks

None. Vantage's legacy connector is read-only: `Write` always returns
`connectors.ErrUnsupportedOperation` (`vantage.go:103-105`); `capabilities.write` is `false` and
this bundle ships no `writes.json`.

## Known limits

- Full Vantage API surface (cost reports, budgets, providers) is out of scope for this wave; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- **`Check` now dials the network; legacy's `Check` never did.** Legacy `Check`
  (`vantage.go:43-57`) validates config/secret presence offline only. This bundle's `base.check`
  issues a real `GET /v2/costs`, matching the wave's general "fail loud, not fail silent"
  preference for `Check` — a deliberate, strictly-improving behavior change with zero record-data
  impact.
- No incremental/date-range filtering is modeled — legacy's `costs` endpoint accepts none either
  (unlike `uservoice`/`veeqo`/`vercel`, `vantage.go`'s `Read` builds no query at all).
- Legacy's `mode=fixture` config value (a testing affordance that short-circuits network access
  and emits one synthetic record) is not part of this bundle; parity is instead proven against
  legacy's live read path via fixture-replay conformance.
