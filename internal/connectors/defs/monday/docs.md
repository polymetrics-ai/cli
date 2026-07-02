# Overview

Monday.com exposes ALL of its data through a single GraphQL endpoint (`POST https://api.monday.com/v2`)
whose body carries a GraphQL query and whose response wraps records under a top-level `data` object.
This migration is the wave1-pilot **StreamHook pilot** (PLAN.md P-8, SPEC §5.5): monday's GraphQL
POST reads, with pagination state carried INSIDE the request body (page-number arguments for
`boards`/`users`/`teams`/`tags`; a `next_items_page` cursor envelope for `items`), are a documented
Tier-2 trigger — `internal/connectors/engine/bundle.go`'s `StreamSpec.Body` field exists but
`internal/connectors/engine/read.go`'s declarative read path never sends it (confirmed at
`read.go:142`: `rt.Requester.Do(ctx, methodOrDefault(stream.Method), reqPath, query, nil)` always
passes a literal `nil` body). `internal/connectors/hooks/monday/hooks.go` implements `StreamHook`
(all 5 streams) and `CheckHook`, porting `internal/connectors/monday/monday.go` +
`streams.go`'s GraphQL query construction, in-body pagination, and record mapping verbatim. This
bundle is engine-vs-legacy parity-tested against `internal/connectors/monday` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

monday.com accepts either a personal API token (`api_token`) or an OAuth access token
(`access_token`), sent VERBATIM (no `Bearer` prefix) as the `Authorization` header — matching
legacy's `connsdk.APIKeyHeader("Authorization", secret, "")` (`monday.go:396`). This bundle wires
the identical shape declaratively (no AuthHook needed) via two `when`-gated `api_key_header` auth
candidates in `streams.json` `base.auth`: the first candidate applies when `api_token` is set, the
second when `access_token` is set (using the engine's `when` absent-key-falsy tolerance, the same
pattern searxng's optional bearer-proxy `api_key` uses), falling back to `mode: none` when neither
is configured (matching legacy's own `mondaySecret` fallback chain, `monday.go:407-422` — though
unlike legacy, an unauthenticated request is only reached if BOTH secrets are absent; legacy hard
errors before ever issuing a request in that case, this bundle instead lets the request go out and
rely on monday's own 401 response, mapped via `error_map`).

An optional `API-Version` header (`config.api_version`) is sent when configured, omitted entirely
otherwise — the same declared-optional-config-header pattern Stripe's `Stripe-Account` uses
(conventions.md §3 "Conditional headers").

## Streams notes

Legacy defines 5 streams, each a distinct GraphQL root query field on the same `POST /v2` endpoint:
`boards`, `items`, `users`, `teams`, `tags` (`streams.go:33-58`, `mondayStreams()`). ALL reads are
GraphQL POSTs whose pagination state lives INSIDE the query text:

- `boards`/`users`/`teams`/`tags` use page-number pagination — a GraphQL `(limit: N, page: M)`
  argument pair (`monday.go:151-182 readPaged`), continuing while a page returns exactly
  `page_size` records and stopping on the first short page.
- `items` is special: the first request fetches `boards { items_page (limit: N) { cursor items {
  ... } } }`, then continues via the top-level `next_items_page(limit: N, cursor: "...")` field
  until monday returns a null cursor (`monday.go:184-268`).

`hooks/monday/hooks.go`'s `ReadStream` ports every one of these shapes verbatim: the GraphQL query
text construction, the `data.<root>` / `data.boards[].items_page` / `data.next_items_page` record
extraction, and the identical field-mapping functions (`boardRecord`/`itemRecord`/`userRecord`/
`teamRecord`/`tagRecord` from `streams.go`) — including `itemRecord`'s hoisting of the nested
`group`/`board` objects into flat `group_id`/`group_title`/`board_id`/`board_name` columns, and
every record's `id` field being coerced to a string via legacy's `stringField` helper (monday
returns numeric ids as JSON numbers in some GraphQL contexts and strings in others).

`boards`/`items` declare `updated_at` as `x-cursor-field` for schema-manifest parity with legacy's
published `CursorFields: []string{"updated_at"}` (`streams.go:71,78`) — but, matching legacy exactly,
NEITHER connector actually filters or advances reads by it: legacy's `Read` never consults `req.State`
at all (grep confirms no state/cursor read anywhere in `monday.go`), so both connectors always
perform a full stream read regardless of any incremental state passed in. This bundle's
`streams.json` therefore declares `incremental.cursor_field` with NO `request_param` — the engine
only ever attempts a server-side incremental filter when `request_param` is set
(`read.go`'s `buildInitialQuery`), so omitting it is how a stream can publish a cursor field for
manifest/derived-sync-mode purposes (conventions.md's schema-as-projection sync-mode derivation,
`incremental_append` requires only an `incremental` block, not a wired `request_param`) while
staying byte-for-byte behaviorally identical to legacy's real "always full sync" behavior.

### Declarative path (`streams.json`) vs. the live StreamHook path

`streams.json` still declares complete stream/schema metadata for all 5 streams (SPEC §5.5's
requirement: "bundle still declares every stream/schema/fixture even though reads dispatch through
the hook") — this is what backs the catalog/manifest surface (stream names, schemas, PK/cursor
fields) regardless of which path a read actually takes. Because `hooks/monday/hooks.go`'s
`StreamHook.ReadStream` recognizes and handles every one of these 5 stream names (returning
`handled=true` unconditionally for each), the declarative fallback in `streams.json` is **never
exercised by production traffic** — `engine.Read` only falls through to it when a `StreamHook`
returns `handled=false` or no hooks are registered at all.

Every stream in this bundle carries an explicit `"conformance": {"skip_dynamic": true, "reason":
"..."}` marker (`internal/connectors/engine/bundle.go`'s `StreamSpec.Conformance`,
`docs/migration/conventions.md` §4/§6): `internal/connectors/conformance/dynamic.go` honors this
marker by Skipping (not attempting) every dynamic fixture-replay check for these streams, since a
declarative GET-shaped replay can never faithfully exercise a GraphQL POST + in-body-pagination
StreamHook (no fixture shape satisfies both "the engine consumes each page exactly once" and "the
hook, not the declarative path, is what a real sync actually calls"). The authoritative substitute
this marker names is `internal/connectors/paritytest/monday/parity_test.go` (drives the real,
hook-dispatched connector via `engine.HooksFor("monday")`) and `hooks/monday/hooks_test.go` — both
assert monday's real GraphQL wire format (query text, in-body pagination, record mapping)
byte-for-byte against legacy. `streams.json`'s remaining `path`/`method`/`records` fields are kept
minimal and honest (no fictional per-stream `pagination`/`query` shaping to satisfy a replay
harness that no longer runs against them) — `fixtures/streams/<stream>/page_N.json` are retained
purely as documentation of the real record shapes each stream emits (and to satisfy
`fixtures_present`'s static "first stream ships a fixture" requirement), not as a load-bearing
replay contract.

## Write actions & risks

None. monday is a read-only source connector (`capabilities.write: false`, no `writes.json`),
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation` unconditionally
(`monday.go:101-103`).

## Known limits

- **`StreamSpec.Body` is unwired (ENGINE_GAP, documented, non-blocking).** The engine's declarative
  read path (`engine/read.go:142`) never sends a request body, so a POST-body GraphQL read with
  in-body pagination state cannot be expressed in `streams.json` alone. This was pre-identified in
  SPEC §5.5 as a sanctioned Tier-2 trigger for this wave, not a blocker: `hooks/monday/hooks.go`'s
  `StreamHook` implements the real GraphQL POST + in-body pagination entirely within the sanctioned
  hook seam, reusing `rt.Requester` (the engine's already-built HTTP client/auth/base-URL plumbing)
  exactly as the declarative path itself would. If 3+ wave2+ connectors need POST-body reads, the
  ENGINE_GAP recurrence rule (conventions.md §6) promotes this to a real engine feature; this is
  occurrence #1 for the pilot.
- **The declarative `streams.json` path is never live-dispatched** (see "Declarative path" above)
  — every stream carries a `conformance.skip_dynamic` marker naming `paritytest/monday`/
  `hooks/monday/hooks_test.go` as the authoritative substitute; conformance's dynamic (fixture
  replay) checks Skip these streams outright rather than exercising a declarative shape that would
  never match monday's real GraphQL wire format.
- **No incremental filtering, matching legacy exactly.** `updated_at` is published as
  `x-cursor-field` for manifest-surface parity, but neither connector filters or advances reads by
  it (see "Streams notes"); every read is a full stream read.
- **Legacy's `mode: fixture` credential-free affordance is NOT part of this bundle.** Legacy's
  `readFixture`/`fixtureMode` (`monday.go:270-336`) emit synthetic records without any network call
  when `config.mode == "fixture"` — this is a legacy-only testing convenience, not part of the live
  record shape, and parity is asserted against legacy's LIVE (httptest-driven) read path only,
  matching the wave1-pilot convention (SPEC §5.1's identical note for xkcd).
- **`connector`/`fixture` marker fields are NOT modeled.** Legacy's fixture-mode records stamp
  `connector: "monday"` and `fixture: true` (`monday.go:329-330`) — these are fixture-mode-only
  fields, never emitted by legacy's live GraphQL read path, so they are correctly absent from this
  bundle's schemas.
