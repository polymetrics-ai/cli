# Overview

Microsoft Entra ID is a Tier-2 quarantine-repair migration (previously quarantined
`ENGINE_GAP`, `docs/migration/quarantine.json`). It reads Microsoft Entra ID (Azure AD) directory
objects — users, groups, applications, service principals, and directory roles — through the
Microsoft Graph API (`v1.0`), read-only. This bundle is capability-parity migrated from
`internal/connectors/microsoft-entra-id` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `client_id`, `client_secret`, and `tenant_id` secrets. Auth is
`oauth2_client_credentials` with two `when`-gated candidates evaluated in declared order
(`docs/migration/conventions.md` §3's dual-auth-ordering pattern, applied to two candidates of the
SAME mode rather than two different modes — the identical mechanism `sharepoint-lists-enterprise`
and `microsoft-teams` use for the same Graph token-endpoint shape): the first candidate uses
`config.token_url` directly and is gated `when: {{ config.token_url }}` (matches only when a full
override is configured, e.g. for a test server); the second, unconditional candidate derives the
endpoint as `{{ config.login_base_url }}/{{ secrets.tenant_id }}/oauth2/v2.0/token` (defaulting
`login_base_url` to `https://login.microsoftonline.com`). This exactly reproduces legacy's own
override precedence (`microsoft-entra-id.go`'s `tokenURL`: an explicit `token_url` config override
always wins; the derived tenant-scoped endpoint is the fallback). Both candidates use the
`config.scope` value (default `https://graph.microsoft.com/.default`, matching legacy's
`graphScope` constant). None of `client_id`/`client_secret`/`tenant_id` is ever logged.

## Streams notes

Five streams, every one a flat Microsoft Graph directory collection: `users`, `groups`,
`applications`, `serviceprincipals` (`GET /servicePrincipals`), `directoryroles`
(`GET /directoryRoles`). Every endpoint returns `{"value": [...], "@odata.nextLink": "<url>"}` —
records live at `records.path: "value"`, matching Graph's real wire shape.

Every schema property is a snake_case rename of the raw Graph camelCase field (e.g.
`display_name` from `displayName`, `user_principal_name` from `userPrincipalName`), matching
legacy's own `mapRecord` functions (`microsoft-entra-id/streams.go`) field-for-field. Every
Graph directory object exposes a string `id`, so `x-primary-key: ["id"]` is uniform across all 5
streams. Graph directory collections are full-refresh only in legacy (no `CursorFields` published
anywhere in `streams()`), so no stream declares an `incremental` block or `x-cursor-field`.

**Pagination — Tier-2 StreamHook, not declarative (the actual blocker behind this connector's prior
quarantine)**: Microsoft Graph's list pagination is `@odata.nextLink` — a JSON key containing a
literal dot — carrying the NEXT PAGE'S FULL ABSOLUTE URL (with its own `$skiptoken` cursor already
embedded). Legacy hand-rolls this exactly (`microsoft-entra-id.go`'s `harvest`/`nextLink`): GET the
resource, extract `value[]`, decode the top-level `@odata.nextLink` string directly (NOT via a
dotted-path helper — the key's own literal dot makes dotted-path addressing ambiguous), and if
non-empty, re-request that exact URL verbatim (dropping the original `$top` query, since the
`nextLink` already carries every parameter it needs). The engine's declarative `next_url` pagination
type reads its cursor via `connsdk.StringAt`'s dotted-path parser (`interpolate.go`/`extract.go`),
which splits on `.` and therefore CANNOT address a literal key containing a dot — `@odata.nextLink`
is read back as "field `nextLink` nested under an object at key `@odata`", which does not exist. This
is a genuine, confirmed `ENGINE_GAP` (see `docs/migration/quarantine.json`'s `microsoft-entra-id`
entry), not a config or fixture issue, and it recurs identically for `microsoft-lists` and
`microsoft-teams` (all three share the exact same Graph `@odata.nextLink` shape) — below the ≥3
recurrence bar for an engine mini-wave to justify a general dotted-path-escape mechanism at the time
of this migration, and Tier-2 fully resolves it without an engine change, so the hook path is taken
here rather than filing a fresh engine increment.

`hooks/microsoft-entra-id/hooks.go` implements `StreamHook`, porting legacy's `harvest`/`nextLink`
logic exactly (same request shape — `$top={{ config.page_size }}` on the first request only — same
absolute-URL follow, same stop condition: an empty/absent `@odata.nextLink`). Every stream in this
bundle carries an explicit `"conformance": {"skip_dynamic": true, "reason": "..."}` marker
(`internal/connectors/engine/bundle.go`'s `StreamSpec.Conformance`, `docs/migration/conventions.md`
§4/§6): `internal/connectors/conformance/dynamic.go` honors this marker by Skipping every dynamic
fixture-replay check for these streams, since the StreamHook (always `handled=true`) is what every
real `Read()` call actually dispatches through, and a declarative-only fixture replay cannot
exercise an absolute-URL-follow loop at all. The authoritative substitute this marker names is
`paritytest/microsoft-entra-id`'s dedicated 2-page `@odata.nextLink` test
(`TestParityMicrosoftEntraID_UsersNextLinkPagination`) and `hooks/microsoft-entra-id/hooks_test.go`.
`streams.json`'s own `base.pagination` stays declared `{"type": "none"}` (a single, honest request)
since it is never dynamically exercised now — no shaping needed to satisfy a replay harness that no
longer runs against these streams.

`max_pages` is a hook-consumed `spec.json` config value (permissive parse: empty/`all`/`unlimited`
means unbounded), matching legacy's own `maxPages` parsing exactly — it is NOT a declarative
`streams.json` field, since pagination itself is entirely hook-driven.

## Write actions & risks

None. Microsoft Entra ID is read-only in legacy (`Write` returns
`connectors.ErrUnsupportedOperation`, `microsoft-entra-id.go`); `capabilities.write` is `false` and
no `writes.json` is declared. Mutating directory objects (users, groups, applications) remains out
of scope for reverse ETL by design, matching legacy's own doc comment.

## Known limits

- Full Microsoft Graph directory surface (devices, administrative units, conditional access
  policies, directory audit logs, etc.) is out of scope for this migration; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries. Only the 5 legacy-parity read streams are implemented.
- **Dynamic conformance checks are skipped, stream-by-stream (every stream carries the marker) and
  also at the bundle level in `metadata.json`** — pagination is hook-driven for every stream, and
  the StreamHook itself needs a real, resolvable Graph endpoint (or at minimum a live/replay HTTP
  server) rather than conformance's synthetic non-secret config values, so no stream's dynamic
  fixture-replay check can usefully exercise the real code path. Static checks (spec/schema
  validity, `interpolations_resolve`, docs/fixtures presence, secret redaction) still run and pass.
  Parity for the pagination/schema-projection shape is proven by `paritytest/microsoft-entra-id`
  (`TestParityMicrosoftEntraID_UsersNextLinkPagination`, a live `httptest.Server`-backed 2-page
  `@odata.nextLink` follow test) and `hooks/microsoft-entra-id/hooks_test.go`'s unit coverage — this
  mirrors the identical, already-accepted `sentry`/`microsoft-teams`/`sharepoint-lists-enterprise`
  `skip_dynamic` precedents for hook-covered or token-endpoint-derivation-blocked bundles.
- `page_size` is runtime-configurable (`config.page_size`, default 100, matching legacy's
  `defaultPageSize`) and forwarded as the `$top` query parameter on the FIRST request of each
  stream's sub-sequence only — subsequent pages follow `@odata.nextLink` verbatim (which already
  encodes the effective page size), exactly matching legacy's `harvest` loop.
- Candidate future engine feature: a `next_url_path` "literal key" escape (e.g. a
  `next_url_literal_key` field naming an exact top-level JSON key to read verbatim, bypassing dotted
  -path splitting) would let this connector, `microsoft-lists`, and `microsoft-teams` all drop their
  StreamHooks. Not implemented in this migration per the ENGINE_GAP recurrence rule
  (`conventions.md` §6) scoping decision recorded at the time each of these three connectors was
  migrated — revisit if a 4th connector hits the identical shape.
