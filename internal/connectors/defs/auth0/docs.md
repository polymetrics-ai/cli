# Overview

Auth0 is a declarative-HTTP bundle migrated from `internal/connectors/auth0` (the hand-written
legacy connector, which stays registered and unchanged until wave6's registry flip). It reads
Auth0 users, clients, connections, roles, and organizations from the Auth0 Management API v2. It is
a read-only source in both legacy and this bundle (`capabilities.write: false`, no `writes.json`).
Status: **partial** — see Known limits for two typed `ENGINE_GAP` blockers that keep this bundle
short of full legacy parity.

## Auth setup

Provide a Management API access token via the `access_token` secret; it is used directly as a
Bearer token (`{"mode": "bearer", "token": "{{ secrets.access_token }}"}`). `base_url` (your tenant
domain, e.g. `https://your-tenant.us.auth0.com`) is required with no default, matching legacy
(`auth0BaseURL` requires it explicitly — the base is tenant-specific, never a fixed literal).

**Not implemented in this bundle**: legacy also accepts M2M `client_id`/`client_secret` credentials
and performs an OAuth2 client-credentials exchange against `<base_url>/oauth/token`, always
including an `audience` form parameter (explicit `audience` config, or derived as
`<base_url>/api/v2/` when unset) so the issued token is scoped to the Management API. The engine's
`oauth2_client_credentials` `AuthSpec` (`internal/connectors/engine/bundle.go`) exposes only
`token_url`/`client_id`/`client_secret`/`scopes` — there is no `audience`/extra-token-param field,
and `buildOAuth2ClientCredentials` (`internal/connectors/engine/auth.go`) never sets
`connsdk.OAuth2ClientCredentials.ExtraParams` from any bundle-declared value. Auth0's token endpoint
requires `audience` to scope the returned token to the Management API; requesting a token without
it would either fail outright or (if the tenant has an unrelated default audience configured) issue
a token that gets 401s on every Management API call — a real, request-level divergence from legacy,
not a cosmetic one, so this is not approximated: the M2M auth candidate is simply not declared. Only
the `access_token` bearer path is implemented. See the `AUTH_COMPLEX`/`ENGINE_GAP` blocker below.

## Streams notes

All 5 streams (`users`, `clients`, `connections`, `roles`, `organizations`) share the identical
`include_totals=true` envelope shape (`{start,limit,length,total,<resource>:[...]}`) and
`page`/`per_page` page-number pagination — matching legacy's `harvest`, which advanced `page` until
a short page (fewer than `per_page`) or the reported `total` was reached. Each stream's `records`
path names its own resource key (`users`, `clients`, `connections`, `roles`, `organizations`),
matching legacy's per-endpoint `arrayKey`. `users` is the only stream with a cursor field
(`x-cursor-field: updated_at`, matching legacy's `CursorFields: []string{"updated_at"}`); no
`incremental` block is declared since legacy never actually filters `users` requests by
`updated_at` (it is catalog metadata only in `auth0Streams()`, not a wired request-param filter).

**`page_size`/`max_pages` are not exposed as config** for the same static-`PaginationSpec`-field
reason documented in `aviationstack`'s and searxng's goldens (`docs/migration/conventions.md`):
`PaginationSpec.PageSize`/`MaxPages` are plain JSON ints, resolved once at bundle load, with no
template/config-driven override. `streams.json`'s `pagination.page_size: 50` matches legacy's own
`auth0DefaultPageSize` (`auth0.go:31`) — every request, fixture-replayed or live, is driven by this
same static value. The `users` fixture (the only stream needing the required 2-page conformance
proof) ships a full 50-record page 1 (triggering the paginator's short-page continuation) plus a
1-record page 2, matching this page size exactly rather than an arbitrary fixture-convenience
number.

## Write actions & risks

None. Auth0 is a read-only source.

## Known limits

- **`ENGINE_GAP` (blocking, M2M client-credentials auth)**: see Auth setup above — the
  `oauth2_client_credentials` `AuthSpec` has no way to express Auth0's mandatory `audience` token
  request parameter. Only the `access_token` bearer credential path is implemented; the M2M
  `client_id`+`client_secret` path from legacy (`auth0.go`'s `authenticator`, taking priority when
  no `access_token` is set) is not portable to this bundle without an engine change (adding an
  `audience`/`extra_params`-style field to `AuthSpec` and wiring it into
  `buildOAuth2ClientCredentials`'s `connsdk.OAuth2ClientCredentials.ExtraParams`).
- **`ENGINE_GAP` (blocking, zero-indexed pagination)**: Auth0's Management API list endpoints are
  0-indexed (`page=0` is the first page; legacy's `harvest` starts its loop at `page := 0`). The
  engine's `page_number` paginator forces `start := 1` whenever `PaginationSpec.StartPage == 0`
  (both `internal/connectors/engine/paginate.go`'s `newPaginator` and
  `internal/connectors/connsdk/paginate.go`'s `PageNumberPaginator.Start()` independently treat an
  explicit `"start_page": 0` identically to "unset," since Go's JSON-unmarshaled zero value for an
  `int` cannot be distinguished from an omitted field) — there is no way to declare a genuinely
  0-indexed first page. Requesting `page=1` first instead of `page=0` would silently skip Auth0's
  actual first page of every resource on every sync, a real data-loss divergence, not an
  approximation; this bundle therefore declares the honest `start_page: 1` (the engine's real
  runtime behavior) rather than a misleading `start_page: 0` that the engine does not honor, and
  files this as a blocker instead of silently dropping the first page. (Two other migrated bundles,
  `algolia` and `beamer`, currently declare `"start_page": 0`; per this same code-read that value is
  inert there too and both should be flagged/repaired in a follow-up pass — not fixed here, out of
  this connector's assigned scope.)
- Only the 5 legacy-parity streams are implemented; the broader Management API surface (rules,
  actions, hooks, logs, tenant settings, device credentials) is out of scope for this wave — see
  `api_surface.json`'s `excluded` entries.
