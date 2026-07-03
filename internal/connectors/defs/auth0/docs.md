# Overview

Auth0 is a declarative-HTTP bundle migrated from `internal/connectors/auth0` (the hand-written
legacy connector, which stays registered and unchanged until wave6's registry flip). It reads
Auth0 users, clients, connections, roles, and organizations from the Auth0 Management API v2. It is
a read-only source in both legacy and this bundle (`capabilities.write: false`, no `writes.json`).
Both of legacy's previously-blocking `ENGINE_GAP`s are now closed: 0-indexed pagination
(`start_page: 0`, S4 engine mini-wave item 1) and the M2M `audience` token-request parameter
(`auth[].extra_params`, S4 engine mini-wave item 4) are both expressible in this dialect now. This
bundle has full stream and auth-mode parity with legacy.

## Auth setup

Legacy accepts two credential shapes with `access_token` taking priority whenever both are
configured (dual-auth ordering is load-bearing, conventions.md §3) — this bundle's `base.auth`
candidate list reproduces that exact precedence:

1. **`access_token` secret** (bearer, first candidate): used directly as a Bearer token
   (`{"mode": "bearer", "token": "{{ secrets.access_token }}", "when": "{{ secrets.access_token }}"}`).
2. **M2M `client_id`/`client_secret` secrets** (`oauth2_client_credentials`, fallback candidate,
   `when: "{{ secrets.client_id }}"`): performs an OAuth2 client-credentials exchange against
   `{{ config.base_url }}/oauth/token`, always including the `audience` config value as the
   `extra_params.audience` token-request form parameter (Auth0 requires this to scope the issued
   token to the Management API) — matching legacy's `authenticator`'s M2M branch and its
   `<base_url>/api/v2/`-by-convention audience exactly, modulo the one narrowing noted below.

`base_url` (your tenant domain, e.g. `https://your-tenant.us.auth0.com`) is required with no
default, matching legacy (`auth0BaseURL` requires it explicitly — the base is tenant-specific,
never a fixed literal).

**Narrowed from legacy: `audience` has no derived default.** Legacy defaults `audience` to
`<base_url>/api/v2/` when the config value is unset (`authenticator`'s `if audience == ""`
branch). The engine's `extra_params` dialect resolves each value via ordinary `Interpolate` with NO
`omit_when_absent`/`default` tolerance (conventions.md §3: "a misconfigured audience/subject param
should fail loudly rather than silently omit a value a real OAuth2 provider may require"), and
`audience` is a function of ANOTHER config value (`base_url`), not a fixed literal — exactly the
"DERIVED default" case conventions.md's `spec.json` `"default"` materialization section says to
either require explicitly or express via a computed-field-style mechanism the dialect doesn't yet
have for base-URL-shaped construction. This bundle takes the documented "require it explicitly"
path: `audience` must be set when using the M2M auth candidate (typically `<base_url>/api/v2/`,
matching legacy's own default value — operators just set it explicitly instead of relying on
derivation). This never changes emitted record data for any configuration legacy itself would
accept; it only requires one more explicit config value than legacy did. Documented parity
deviation, ACCEPTABLE.

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

- **`audience` has no derived default** — see Auth setup above. Documented parity deviation,
  ACCEPTABLE: operators using the M2M auth candidate must set `audience` explicitly (typically
  `<base_url>/api/v2/`, legacy's own default value) rather than relying on derivation from
  `base_url`; never changes emitted data for any configuration legacy itself would accept.
- Only the 5 legacy-parity streams are implemented; the broader Management API surface (rules,
  actions, hooks, logs, tenant settings, device credentials) is out of scope for this wave — see
  `api_surface.json`'s `excluded` entries.
