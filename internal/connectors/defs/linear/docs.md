# Overview

Linear is a project/issue-tracking tool exposed through a single GraphQL
endpoint (`POST https://api.linear.app/graphql`). The legacy Go connector
under `internal/connectors/linear` reads 4 root connections (`issues`,
`teams`, `projects`, `users`), each shaped as `{ <connection>: { nodes: [...],
pageInfo: { hasNextPage, endCursor } } }`, walking the cursor connection via
the `after` GraphQL variable.

This bundle records the legacy-parity stream/schema surface, but full Pass B
expansion is quarantined. Linear's documented API is GraphQL-only: additional
reads and all mutations require JSON request bodies containing fixed GraphQL
documents and variables. The current declarative read path never sends
`stream.body`, write actions cannot inject static GraphQL documents safely, and
this workstream is restricted to `internal/connectors/defs/linear` plus a
typed quarantine entry. No `internal/connectors/hooks/linear` package exists in
this worktree.

## Auth setup

Linear accepts either a personal API key (`api_key`) sent as a bare
`Authorization` header (no `Bearer` prefix), or an OAuth access token
(`access_token`) sent with the `Bearer` prefix. The bundle wires the legacy
3-way precedence declaratively via `streams.json` `base.auth`'s
first-match-wins candidate list:

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
4. `mode: none` — matches neither of the above; unlike legacy, this bundle
   lets the request go out unauthenticated and relies on Linear's own 401
   response, mapped via `error_map`.

## Streams notes

Legacy defines 4 streams, each a distinct GraphQL root connection query field
on the same `POST /graphql` endpoint. The schema files preserve the same
record projection: issue records flatten nested `state`/`team`/`assignee`/
`creator` objects, and every stream keeps both snake_case
(`created_at`/`updated_at`) and raw camelCase (`createdAt`/`updatedAt`) time
fields that legacy record maps carry.

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

Every stream carries an explicit `conformance.skip_dynamic` marker because a
fixture replay cannot exercise Linear's real POST-body GraphQL request shape
through the current declarative read path. The fixtures remain as record-shape
documentation and static conformance inputs; the legacy Go connector remains
the runnable implementation until the GraphQL body/hook gap is closed.

## Write actions & risks

None. Linear remains read-only in this bundle (`capabilities.write: false`,
no `writes.json`), matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **Full Pass B surface is quarantined as `ENGINE_GAP`.** Linear's documented
  API is a GraphQL schema, not REST resources. Additional root connections
  require POST request bodies and body-cursor pagination; mutations require
  fixed GraphQL documents and variables in the JSON body.
- **`StreamSpec.Body` is unwired.** The engine's declarative read path
  (`engine/read.go:142`) never sends a request body, so a POST-body GraphQL
  read with in-body-driven cursor continuation cannot be expressed in
  `streams.json` alone.
- **No Linear hook package exists in this worktree.** Adding or editing one is
  outside this Pass B shard's allowed paths.
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
- **`page_size`/`max_pages` are legacy Go connector settings, not declarative
  pagination settings.** GraphQL cursor continuation lives in the request body
  and response `pageInfo`, so these values cannot be wired into the current
  declarative pagination dialect.
- **Legacy's `mode: fixture` credential-free affordance is NOT part of this
  bundle.** Legacy's `readFixture`/`fixtureMode` (`linear.go:243-304,425-430`)
  emit synthetic records without any network call when `config.mode ==
  "fixture"`, including a `previous_cursor` marker stamped from `req.State`
  — this is a legacy-only testing convenience, not part of the live record
  shape; parity is asserted against legacy's LIVE (httptest-driven) read
  path only. The `connector`/`fixture`/`previous_cursor` marker fields
  legacy's fixture mode stamps are correspondingly absent from this bundle's
  schemas.
