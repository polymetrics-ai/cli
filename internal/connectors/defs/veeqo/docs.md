# Overview

Veeqo is an inventory and order management platform. This bundle reads **orders** from the Veeqo
API (`GET {base_url}/orders`, default `https://api.veeqo.com`). Read-only; it migrates
`internal/connectors/veeqo` (190 loc), which stays registered and unchanged until wave6's registry
flip.

## Auth setup

Veeqo authenticates via a single secret, `api_key`, sent as the `x-api-key` header with an empty
prefix, matching legacy's `connsdk.APIKeyHeader("x-api-key", apiKey, "")` requester exactly
(`veeqo.go:120`). Expressed via the engine's `api_key_header` auth mode:
`{"mode":"api_key_header","header":"x-api-key","value":"{{ secrets.api_key }}","prefix":""}`.

## Streams notes

One stream: `orders`, `GET /orders`, records extracted from the **response body root** — Veeqo's
`/orders` endpoint returns a bare top-level JSON array, not an object envelope
(`veeqo.go:92`: `connsdk.RecordsAt(resp.Body, "")`) — expressed here as `"records": {"path": "."}`
(the documented root-selector convention; see `pivotal-tracker`'s identical shape). Each record is
mapped field-for-field to `{id, number, status, created_at}` (`veeqo.go:138-140`). No pagination —
legacy issues exactly one request per read.

**`id` is force-cast to a string.** Legacy's `orderRecord` calls `stringValue(item["id"])`
(`veeqo.go:139,178-189`), converting a raw JSON number (Veeqo's real wire shape for `id`) into a Go
string. Reproduced via `{{ record.id | last_path_segment }}` rather than a bare `{{ record.id }}`
reference, for the same reason documented in `uservoice`'s `docs.md`: the bare form would trigger
typed extraction and preserve the raw numeric type, diverging from legacy's deliberate string
cast; `last_path_segment` applied to a slash-free value is a documented no-op that only routes the
value through `Interpolate`'s stringify path. Schema declares `id` as `"type": "string"`
accordingly.

**Optional `start_date` query passthrough.** Legacy only sends `?start_date=<value>` when the
config value is present and non-empty (`veeqo.go:85-87`); an absent value sends no query param at
all. Expressed via the `stream.Query` optional-query dialect
(`{"template": "{{ config.start_date }}", "omit_when_absent": true}`), not a `streams.json`
`incremental` block — legacy never persists or advances a cursor across syncs (see `uservoice`'s
`docs.md` for the fuller rationale, identical here). `x-cursor-field` is intentionally NOT declared
on the schema even though `created_at` is emitted as a field, matching legacy's full-refresh-only
functional behavior.

## Write actions & risks

None. Veeqo's legacy connector is read-only: `Write` always returns
`connectors.ErrUnsupportedOperation` (`veeqo.go:107-109`); `capabilities.write` is `false` and this
bundle ships no `writes.json`.

## Known limits

- Full Veeqo API surface (products, customers, warehouses, order mutation) is out of scope for
  this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- **`Check` now dials the network; legacy's `Check` never did.** Legacy `Check`
  (`veeqo.go:43-57`) validates config/secret presence offline only. This bundle's `base.check`
  issues a real `GET /orders`, matching the wave's general "fail loud, not fail silent" preference
  for `Check` — a deliberate, strictly-improving behavior change with zero record-data impact.
- The optional `start_date` filter is a stateless, config-only passthrough (see Streams notes) —
  not a true incremental sync.
- Legacy's `mode=fixture` config value (a testing affordance that short-circuits network access
  and emits one synthetic record) is not part of this bundle; parity is instead proven against
  legacy's live read path via fixture-replay conformance.
