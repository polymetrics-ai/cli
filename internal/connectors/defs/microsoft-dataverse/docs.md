# Overview

Microsoft Dataverse is a Tier-2 declarative-bundle-plus-hook migration. It reads Microsoft
Dataverse accounts, contacts, leads, opportunities, and system users through the Dataverse Web API
(`v9.2`), read-only. This bundle is capability-parity migrated from
`internal/connectors/microsoft-dataverse` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `client_id`, `client_secret`, and `tenant_id` secrets, plus `base_url` and `scope` config
values (`base_url`: the organization's Dataverse Web API root, e.g.
`https://org.crm.dynamics.com/api/data/v9.2`; `scope`: the matching OAuth2 resource scope, e.g.
`https://org.crm.dynamics.com/.default` — no defaults for either; see Known limits for why both are
required directly rather than derived). Auth is `oauth2_client_credentials` with two `when`-gated
candidates evaluated in declared order (`docs/migration/conventions.md` §3's dual-auth-ordering
pattern, applied to two candidates of the SAME mode — the identical mechanism
`microsoft-entra-id`/`sharepoint-lists-enterprise`/`microsoft-teams` use for the same Azure AD
token-endpoint shape): the first candidate uses `config.token_url` directly and is gated
`when: {{ config.token_url }}` (matches only when a full override is configured, e.g. for a test
server); the second, unconditional candidate derives the endpoint as
`{{ config.login_base_url }}/{{ secrets.tenant_id }}/oauth2/v2.0/token` (defaulting
`login_base_url` to `https://login.microsoftonline.com`) — this exactly reproduces legacy's own
override precedence (`microsoft-dataverse.go`'s `tokenURL`: an explicit `token_url` config
override always wins; the derived tenant-scoped endpoint is the fallback). Both candidates use the
`config.scope` value verbatim. None of `client_id`/`client_secret`/`tenant_id` is ever logged.

## Streams notes

Five streams, every one a flat Dataverse Web API entity set: `accounts`, `contacts`, `leads`,
`opportunities`, `systemusers`. Every endpoint returns
`{"value": [...], "@odata.nextLink": "<url>"}` — records live at `records.path: "value"`, matching
Dataverse's real OData wire shape (identical envelope shape to Microsoft Graph).

Every stream emits the same 5-field shape legacy's `baseRecord` helper produces: `id` (the
entity's own GUID primary-key column — `accountid`/`contactid`/`leadid`/`opportunityid`/
`systemuserid`, whichever the current stream's entity set uses), `name` (the entity's display-name
column, with a legacy-matching per-entity fallback chain: contacts/leads/systemusers try
`fullname` first, leads then also try `subject`, everything falls back to a bare `name` field
last), `email` (`emailaddress1`, falling back to `internalemailaddress` for systemusers, matching
legacy exactly), `created_on` (`createdon`), `modified_on` (`modifiedon`). `x-primary-key: ["id"]`
is uniform across all 5 streams. Dataverse entity sets are full-refresh only in legacy (no
`CursorFields` published anywhere in `streams()`), so no stream declares an `incremental` block or
`x-cursor-field`.

**Pagination — Tier-2 StreamHook, not declarative (the actual reason this is Tier-2, not Tier-1)**:
Dataverse's Web API list pagination is `@odata.nextLink` — a JSON key containing a literal dot —
carrying the NEXT PAGE'S FULL ABSOLUTE URL (with its own `$skiptoken` cursor already embedded).
Legacy hand-rolls this exactly (`microsoft-dataverse.go`'s `harvest`/`nextLink`): GET the resource,
extract `value[]`, decode the top-level `@odata.nextLink` string directly (NOT via a dotted-path
helper — the key's own literal dot makes dotted-path addressing ambiguous), and if non-empty,
re-request that exact URL verbatim (dropping the original `$top` query, since `nextLink` already
carries every parameter it needs). The engine's declarative `next_url` pagination type reads its
cursor via `connsdk.StringAt`'s dotted-path parser, which splits on `.` and therefore CANNOT
address a literal key containing a dot — `@odata.nextLink` is read back as "field `nextLink`
nested under an object at key `@odata`", which does not exist. This is the exact, already-confirmed
`ENGINE_GAP` behind `microsoft-entra-id`/`microsoft-lists`/`microsoft-teams`'s identical Tier-2
shape (4th occurrence of this exact recurrence at the time of this migration) — resolved here the
same way, via a Tier-2 `StreamHook`, not an engine change.

`hooks/microsoft-dataverse/hooks.go` implements `StreamHook`, porting legacy's `harvest`/`nextLink`
logic exactly (same request shape — `$top={{ config.page_size }}` on the first request only — same
absolute-URL follow, same stop condition: an empty/absent `@odata.nextLink`). Every stream in this
bundle carries an explicit `"conformance": {"skip_dynamic": true, "reason": "..."}` marker
(`internal/connectors/engine/bundle.go`'s `StreamSpec.Conformance`, `docs/migration/conventions.md`
§4/§6): `internal/connectors/conformance/dynamic.go` honors this marker by Skipping every dynamic
fixture-replay check for these streams, since the StreamHook (always `handled=true`) is what every
real `Read()` call actually dispatches through, and a declarative-only fixture replay cannot
exercise an absolute-URL-follow loop at all. The authoritative substitute this marker names is
`paritytest/microsoft-dataverse`'s dedicated 2-page `@odata.nextLink` test
(`TestParityMicrosoftDataverse_AccountsNextLinkPagination`) and
`hooks/microsoft-dataverse/hooks_test.go`.

`max_pages` is a hook-consumed `spec.json` config value (permissive parse: empty/`all`/`unlimited`
means unbounded), matching legacy's own `maxPages` parsing exactly — it is NOT a declarative
`streams.json` field, since pagination itself is entirely hook-driven.

## Write actions & risks

None. Microsoft Dataverse is read-only in legacy (`Write` returns
`connectors.ErrUnsupportedOperation`, `microsoft-dataverse.go`); `capabilities.write` is `false`
and no `writes.json` is declared.

## Known limits

- Full Dataverse Web API surface (custom entities, `$expand` relationship traversal, actions/
  functions, `$batch` requests, entity metadata) is out of scope for this migration; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries. Only the 5 legacy-parity read streams are implemented.
- **Config surface narrowing, documented parity deviation (base_url)**: legacy accepts EITHER a
  full `base_url` config value OR a bare `org_url` (with `base_url` derived as
  `org_url + "/api/data/v9.2"`). This bundle requires `base_url` directly and does not model the
  `org_url`-derivation branch: the engine's template dialect has no string-concatenation
  primitive that could express "append a fixed literal path segment onto a config value" inside
  `streams.json`'s `url` field the way legacy's Go code does (`strings.TrimRight(org, "/") +
  "/api/data/v9.2"`). This never changes emitted record DATA for any input that supplies
  `base_url` directly (the common case for any already-configured org), and is a strictly narrower
  accepted-config-shape, not a data-emission divergence, per the §5 parity-deviation meta-rule.
  Callers currently supplying only `org_url` must be migrated to supply the equivalent full
  `base_url` instead.
- **Config surface narrowing, documented parity deviation (scope)**: legacy always DERIVES the
  OAuth2 scope from `base_url`'s own scheme+host (`scope()`: `<scheme>://<host>/.default`) rather
  than accepting it as separate config. The engine's `{{ }}` template dialect has no mechanism to
  strip a URL's path component from an already-resolved config value (no string-manipulation
  filter exists for this), so this bundle requires `scope` explicitly instead of deriving it: never
  a data-emission divergence for any input that supplies both `base_url` and the matching `scope`
  (the overwhelmingly common real case, since Dataverse's scope is a deterministic function of the
  org URL any caller already knows).
- **Dynamic conformance checks are skipped, both stream-by-stream (every stream carries the
  marker) and at the bundle level (`metadata.json`'s `conformance.skip_dynamic`)** — pagination is
  hook-driven for every stream (the stream-level markers), and separately, `check_fixture` itself
  would otherwise attempt the declarative `oauth2_client_credentials` dual-candidate auth against
  conformance's synthetic non-secret config: EVERY spec property (including `token_url`) receives a
  synthetic non-empty value, which makes the first `when: {{ config.token_url }}`-gated candidate
  match and attempt a real OAuth2 token-request POST to the literal string
  `"synthetic-conformance-value"` — an always-failing, uninformative request that has nothing to do
  with this bundle's actual pagination/schema-projection correctness. The bundle-level marker
  Skips `check_fixture` (and every other auth-resolving dynamic check) outright instead of
  reporting that predictable, non-actionable failure — the IDENTICAL shape and justification as
  `microsoft-entra-id`'s own bundle-level marker (same dual when-gated
  `oauth2_client_credentials` auth candidates, same synthetic-config false-positive trigger).
  Static checks (spec/schema validity, `interpolations_resolve`, docs/fixtures presence, secret
  redaction) still run and pass. Parity for both auth-candidate-ordering and the
  pagination/schema-projection shape is proven by `paritytest/microsoft-dataverse` (a live
  `httptest.Server`-backed 2-page `@odata.nextLink` follow test,
  `TestParityMicrosoftDataverse_AccountsNextLinkPagination`) and
  `hooks/microsoft-dataverse/hooks_test.go`'s unit coverage — this mirrors the identical,
  already-accepted `microsoft-entra-id`/`microsoft-teams`/`sharepoint-lists-enterprise`
  `skip_dynamic` precedents for hook-covered or token-endpoint-derivation-blocked bundles.
- `page_size` is runtime-configurable (`config.page_size`, default 100, matching legacy's
  `defaultPageSize`, max 5000 matching legacy's `maxPageSize`) and forwarded as the `$top` query
  parameter on the FIRST request of each stream's sub-sequence only — subsequent pages follow
  `@odata.nextLink` verbatim (which already encodes the effective page size), exactly matching
  legacy's `harvest` loop.
- `Check` sends `GET /accounts?$top=1`, matching legacy's `Check` implementation exactly (a bounded
  read of the accounts list confirms auth and connectivity without mutating anything).
