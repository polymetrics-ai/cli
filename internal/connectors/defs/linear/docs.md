# Overview

Linear is a project/issue-tracking tool exposed entirely through a single
GraphQL endpoint (`POST https://api.linear.app/graphql`). This migration
ports `internal/connectors/linear` (`linear.go` + `streams.go`), which reads
4 root connections (`issues`, `teams`, `projects`, `users`), each shaped
`{ <connection>: { nodes: [...], pageInfo: { hasNextPage, endCursor } } }`,
walking the cursor connection via the `after` GraphQL variable. Linear's
GraphQL POST reads with pagination state carried in the response body's
`pageInfo` (and continuation driven by a GraphQL variable, not a query
param/header) is the identical documented Tier-2 `StreamHook` trigger
monday.com's migration already established (`docs/migration/conventions.md`
§1's Tier-2 table, `internal/connectors/hooks/monday/hooks.go`) —
`internal/connectors/engine/bundle.go`'s `StreamSpec.Body` field exists but
`internal/connectors/engine/read.go`'s declarative read path never sends a
body (`read.go:142` always passes a literal `nil`). `internal/connectors/
hooks/linear/hooks.go` implements `StreamHook` (all 4 streams) and
`CheckHook`, porting `linear.go`/`streams.go`'s GraphQL query construction,
cursor pagination, and record mapping verbatim. This bundle is engine-vs-legacy
parity-tested against `internal/connectors/linear` (the hand-written
connector it migrates); the legacy package stays registered and unchanged
until wave6's registry flip.

## Auth setup

Linear accepts either a personal API key (`api_key`) sent as a BARE
`Authorization` header (no `Bearer` prefix), or an OAuth access token
(`access_token`) sent with the `Bearer` prefix — matching legacy's
`linearAuthenticator`/`linearSecret`/`isOAuth` precedence exactly
(`linear.go:306-353`). This bundle wires the identical 3-way precedence
declaratively (no AuthHook needed) via `streams.json` `base.auth`'s
first-match-wins candidate list (conventions.md §3 "Dual-auth ordering is
load-bearing"):

1. `access_token` set → `Bearer <access_token>` (legacy: `isOAuth` is true
   whenever `access_token` is non-empty, and `linearSecret` prefers
   `access_token` over `api_key`).
2. `access_token` unset AND `config.auth_type` is `oauth`/`oauth2.0` →
   `Bearer <api_key>` (legacy: `isOAuth` is also true when
   `config["auth_type"]` matches, and `linearSecret` falls back to `api_key`
   since `access_token` is empty).
3. Otherwise, `api_key` set → bare `Authorization: <api_key>` header, no
   prefix (legacy's default, non-OAuth path —
   `connsdk.APIKeyHeader("Authorization", secret, "")`, `linear.go:337`).
4. `mode: none` — matches neither of the above; unlike legacy (which hard
   errors in `requester`/`Check` before ever issuing a request when no
   secret resolves), this bundle lets the request go out unauthenticated and
   relies on Linear's own 401 response, mapped via `error_map` (the same
   accepted deviation searxng/monday's own `mode: none` fallback documents).

## Streams notes

Legacy defines 4 streams, each a distinct GraphQL root connection query field
on the same `POST /graphql` endpoint (`streams.go:21-26`,
`linearStreamEndpoints`). ALL reads use the identical cursor-connection
pagination shape (`linear.go:157-207`'s `harvest`): the `after` GraphQL
variable, `pageInfo.hasNextPage`/`pageInfo.endCursor` from the response body,
continuing while `hasNextPage` is true and `endCursor` is non-empty.
`hooks/linear/hooks.go`'s `ReadStream` ports this verbatim — the GraphQL
query text construction (`buildConnectionQuery`), the `data.<connection>.
nodes` / `pageInfo` extraction, and the identical field-mapping functions
(`linearIssueRecord`/`linearTeamRecord`/`linearProjectRecord`/
`linearUserRecord` from `streams.go`) — including `linearIssueRecord`'s
hoisting of the nested `state`/`team`/`assignee`/`creator` objects into flat
`state_id`/`state_name`/`state_type`/`team_id`/`team_key`/`assignee_id`/
`assignee_email`/`creator_id` columns, and every record preserving BOTH the
snake_case (`created_at`/`updated_at`) AND raw camelCase (`createdAt`/
`updatedAt`) keys legacy's record maps carry (used by legacy's own cursor
logic and test assertions — kept for schema-as-projection parity, not
dropped).

Every stream declares `updated_at` as `x-cursor-field` for schema-manifest
parity with legacy's published `CursorFields: []string{"updatedAt"}`
(`streams.go:90,97,104,111` all set the SAME cursor field name string used
here) — matching legacy, no stream actually filters server-side by it: the
GraphQL query legacy builds (`buildConnectionQuery`) never sends an
`updatedAt`-based filter argument, and `InitialState`/`Read` never consult
`req.State` for anything beyond stamping a `previous_cursor` marker in
FIXTURE mode only (`linear.go:253`, not part of the live read path at all).
This bundle's `streams.json` therefore declares `incremental.cursor_field`
with NO `request_param`, matching monday's identical documented pattern for
the same reason — the engine only attempts a server-side incremental filter
when `request_param` is set, so omitting it publishes the cursor field for
manifest/derived-sync-mode purposes while staying behaviorally identical to
legacy's real "always full sync" behavior.

### Declarative path (`streams.json`) vs. the live StreamHook path

`streams.json` still declares complete stream/schema metadata for all 4
streams (the bundle-still-declares-everything requirement conventions.md §1
sets for Tier-2 bundles) — this is what backs the catalog/manifest surface
regardless of which path a read actually takes. Because `hooks/linear/
hooks.go`'s `StreamHook.ReadStream` recognizes and handles every one of
these 4 stream names (returning `handled=true` unconditionally for each),
the declarative fallback in `streams.json` is **never exercised by
production traffic** — `engine.Read` only falls through to it when a
`StreamHook` returns `handled=false` or no hooks are registered at all.

Every stream in this bundle carries an explicit `"conformance":
{"skip_dynamic": true, "reason": "..."}` marker (conventions.md §4): a
declarative GET-shaped replay can never faithfully exercise a GraphQL POST +
in-body cursor-pagination StreamHook (no fixture shape satisfies both "the
engine consumes each page exactly once" and "the hook, not the declarative
path, is what a real sync actually calls"). The authoritative substitute
this marker names is `internal/connectors/paritytest/linear/parity_test.go`
(drives the real, hook-dispatched connector via
`engine.HooksFor("linear")`) and `hooks/linear/hooks_test.go` — both assert
Linear's real GraphQL wire format (query text, cursor variable, record
mapping) byte-for-byte against legacy. `fixtures/streams/<stream>/page_N.json`
are retained purely as documentation of the real record shapes each stream
emits (and to satisfy `fixtures_present`'s static "first stream ships a
fixture" requirement, plus the 2-page requirement on `issues`, the first
declared stream), not as a load-bearing replay contract.

## Write actions & risks

None. Linear is a read-only source connector in this migration
(`capabilities.write: false`, no `writes.json`), matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation` unconditionally
(`linear.go:141-143`).

## Known limits

- **`StreamSpec.Body` is unwired (ENGINE_GAP, documented, non-blocking,
  recurrence #2 after monday).** The engine's declarative read path
  (`engine/read.go:142`) never sends a request body, so a POST-body GraphQL
  read with in-body-driven cursor continuation cannot be expressed in
  `streams.json` alone. `hooks/linear/hooks.go`'s `StreamHook` implements the
  real GraphQL POST + cursor pagination entirely within the sanctioned hook
  seam, reusing `rt.Requester` exactly as the declarative path itself would.
  This is the second connector in the fleet hitting this gap (monday was
  first) — a 3rd occurrence would trigger the ENGINE_GAP recurrence
  promotion (conventions.md §6).
- **The declarative `streams.json` path is never live-dispatched** (see
  "Declarative path" above) — every stream carries a
  `conformance.skip_dynamic` marker naming `paritytest/linear`/
  `hooks/linear/hooks_test.go` as the authoritative substitute.
- **No incremental filtering, matching legacy exactly.** `updated_at` is
  published as `x-cursor-field` for manifest-surface parity, but neither
  connector filters or advances reads by it (see "Streams notes" above);
  every read is a full stream read.
- **`base_url` override must be the FULL GraphQL endpoint URL (documented
  scope narrowing, ACCEPTABLE per conventions.md §5's meta-rule).** Legacy's
  `linearBaseURL` (`linear.go:370-393`) auto-appends `/graphql` to a
  bare-host override with no path (so `https://example.com` becomes
  `https://example.com/graphql`), while leaving an explicit-path override
  intact. This dialect's `base.url` is a single `{{ config.base_url }}`
  template with no conditional path-suffixing mechanism — a bare-host
  override must include the full `/graphql` path itself with this bundle.
  The documented default (`https://api.linear.app/graphql`) and every
  override that already includes a path are unaffected; only the
  bare-host-with-auto-suffix convenience is narrowed, matching the exact
  class of gap `docs/migration/conventions.md`'s searxng/employment-hero
  worked examples document for other config-shape conveniences this dialect
  cannot express.
- **`page_size`/`max_pages` are consumed by `hooks/linear/hooks.go`'s
  `StreamHook`, not the declarative pagination dialect** (there is no
  declarative pagination block at all here, matching monday's identical
  shape — GraphQL cursor continuation lives entirely in the hook). Both are
  declared in `spec.json` since the hook genuinely reads
  `config.page_size`/`config.max_pages` (mirroring monday's `max_pages`
  carried-minor fix, applied here from the start rather than as a
  retrofit).
- **Legacy's `mode: fixture` credential-free affordance is NOT part of this
  bundle.** Legacy's `readFixture`/`fixtureMode` (`linear.go:243-304,425-430`)
  emit synthetic records without any network call when `config.mode ==
  "fixture"`, including a `previous_cursor` marker stamped from `req.State`
  — this is a legacy-only testing convenience, not part of the live record
  shape; parity is asserted against legacy's LIVE (httptest-driven) read
  path only. The `connector`/`fixture`/`previous_cursor` marker fields
  legacy's fixture mode stamps are correspondingly absent from this bundle's
  schemas.
