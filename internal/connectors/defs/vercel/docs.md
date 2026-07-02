# Overview

Vercel is a frontend deployment platform. This bundle reads **deployments** from the Vercel REST
API (`GET {base_url}/v6/deployments`, default `https://api.vercel.com`). Read-only; it migrates
`internal/connectors/vercel` (132 loc), which stays registered and unchanged until wave6's
registry flip.

## Auth setup

Vercel authenticates via a single secret, `access_token`, sent as a Bearer token
(`Authorization: Bearer <access_token>`), matching legacy's `connsdk.Bearer(token)` requester
exactly (`vercel.go:95-105`).

## Streams notes

One stream: `deployments`, `GET /v6/deployments`, records extracted from the `deployments` array
(`vercel.go:79`), each record mapped to `{id, name, state, created}` (`vercel.go:83-86`). No
pagination — legacy issues exactly one request per read; no incremental cursor (legacy's
`Catalog` declares `CursorFields: ["created"]` as stream metadata only — `Read` never reads or
advances a persisted cursor, it only ever reads the raw `start_date` config value verbatim, same
as `uservoice`/`veeqo`).

**`id` renames the wire's `uid` field, preserving its native type.** Legacy assigns
`item["uid"]` directly to the record's `id` key with no stringify call
(`vercel.go:84` — contrast with `uservoice`/`vantage`/`veeqo`'s explicit `stringValue(item["id"])`
casts). This bundle's `computed_fields` entry `"id": "{{ record.uid }}"` is a bare, unfiltered
single reference, so the engine's typed-extraction path applies and copies the raw wire value
verbatim (a JSON string for Vercel's real `uid` shape) — exactly matching legacy's direct
assignment, no cast of any kind.

**`created` is preserved as a native integer (Unix milliseconds).** Legacy assigns
`item["created"]` directly with no conversion (`vercel.go:86`); Vercel's real wire shape is a
Unix-milliseconds JSON number. The bare `{{ record.created }}` computed_fields reference likewise
preserves this native numeric type — schema declares `"type": ["integer", "null"]`, never a
stringified/widened form.

**Optional `start_date` maps to the `from` query param.** Legacy only sends `?from=<value>` when
the config value is present and non-empty (`vercel.go:72-74`); an absent value sends no query
param at all. Expressed via the `stream.Query` optional-query dialect
(`{"template": "{{ config.start_date }}", "omit_when_absent": true}`), not a `streams.json`
`incremental` block, for the identical reason documented in `uservoice`'s `docs.md`: legacy never
persists or advances a cursor across syncs. `x-cursor-field` is intentionally NOT declared on the
schema even though `created` is emitted as a field, matching legacy's full-refresh-only functional
behavior.

## Write actions & risks

None. Vercel's legacy connector is read-only: `Write` always returns
`connectors.ErrUnsupportedOperation` (`vercel.go:91-93`); `capabilities.write` is `false` and this
bundle ships no `writes.json`.

## Known limits

- Full Vercel API surface (projects, domains, teams, deployment creation) is out of scope for this
  wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- **`Check` now dials the network; legacy's `Check` never did.** Legacy `Check`
  (`vercel.go:33-47`) validates config/secret presence offline only. This bundle's `base.check`
  issues a real `GET /v6/deployments`, matching the wave's general "fail loud, not fail silent"
  preference for `Check` — a deliberate, strictly-improving behavior change with zero record-data
  impact.
- The optional `start_date`/`from` filter is a stateless, config-only passthrough (see Streams
  notes) — not a true incremental sync.
- Legacy's `mode=fixture` config value (a testing affordance that short-circuits network access
  and emits one synthetic record, including a legacy-only `fixture: true` marker field never
  present on real API responses) is not part of this bundle; parity is instead proven against
  legacy's live read path via fixture-replay conformance.
