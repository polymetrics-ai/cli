# Overview

100ms is a wave2 fan-out declarative-HTTP migration. It reads 100ms rooms, sessions, recordings,
and templates through the 100ms server-side REST API (`GET https://api.100ms.live/v2/...`). This
bundle targets capability parity with `internal/connectors/100ms` (package `onehms`, the
hand-written connector it migrates); the legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide a 100ms management token via the `management_token` secret; it is used only for Bearer
auth (`Authorization: Bearer <management_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`onehms.go:245`).

## Streams notes

All 4 streams (`rooms`, `sessions`, `recordings`, `templates`) share the same shape: `GET` against
the 100ms list endpoint, records at `data`, primary key `["id"]`. Every 100ms object exposes a
string `id` and an RFC3339 `created_at`/`updated_at` pair; `created_at` is declared as the soft
cursor field on every schema (matching legacy's `streams.go` catalog), but **no stream declares an
`incremental` block** — legacy's `harvest` loop has no server-side incremental filter parameter at
all (`onehms.go`'s `Read`/`harvest` always walks the full result set every sync; `InitialState`
always seeds an empty cursor), so this bundle reproduces the identical full-refresh-only behavior
rather than inventing an incremental filter legacy never had.

Pagination follows 100ms's own `data`/`last` body-cursor convention (`pagination.type: cursor` with
`token_path: last`, `cursor_param: start`): the next page is requested with `start=<last>`, and
pagination stops when `last` is empty, the page returns no records, or (the engine's
`tokenPathCursor` loop guard) the same token repeats twice in a row — this is an exact match for
legacy's own three-way stop condition (`onehms.go:183`: `last == "" || len(records) == 0 || last ==
start`). No `stop_path` is declared: 100ms has no separate boolean "has more" signal the way
Stripe/Zendesk do, only the token's own emptiness, so the engine's default stop-on-empty-token
behavior for the `token_path` cursor variant is already exact parity. Every request sends
`limit=100` (matches legacy's default `page_size`) via each stream's static `query: {"limit":
"100"}`.

`page_size`/`max_pages` are **not modeled as config-driven overrides** here even though legacy
exposes them (`onehms.go`'s `pageSize`/`maxPages` config helpers, defaults 100/0-unlimited): the
`cursor`+`token_path` paginator never reads `PaginationSpec.PageSize`, and `PaginationSpec.MaxPages`
is a fixed integer, not a templated field — there is no mechanism to wire a `config.*` value into
either at read time. Declaring `page_size`/`max_pages` in `spec.json` with no template anywhere
consuming them would be dead config (F6, REVIEW.md's rule, matching bitly's identical documented
narrowing for its own `next_url` paginator). This bundle bakes legacy's own defaults (`limit=100`,
unbounded pages) as fixed values instead.

## Write actions & risks

None. 100ms is read-only both in legacy and here: legacy's own package doc states "the server-side
API exposes no safe reverse-ETL surface for pm, so writes are unsupported" (`onehms.go:250-255`).
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- Full 100ms management API surface (room creation/updates, room-codes, active-rooms, webhooks,
  polls, recording-asset downloads) is out of scope for wave2; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 4
  legacy-parity read streams are implemented.
- `page_size`/`max_pages` are fixed (not config-driven); see Streams notes above for why the engine
  dialect cannot wire them for this paginator shape without changing behavior.
- Legacy's fixture-mode-only fields (`onehms.go`'s `readFixture`, e.g. `previous_cursor` echoing a
  prior sync cursor) are not modeled — they only ever appeared in legacy's own credential-free
  fixture path, never in a live record; this bundle's schemas target the live record shape only,
  and the engine's own conformance/fixture-replay harness is the credential-free test affordance
  for this bundle.
