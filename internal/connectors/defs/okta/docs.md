# Overview

Okta is a read-only Tier-1 declarative migration of `internal/connectors/okta` (legacy `connsdk`-HTTP
connector). It reads Okta users, groups, and system log events from the Okta Management REST APIs.
The legacy package stays registered and unchanged until wave6's registry flip; this bundle is
engine-vs-legacy parity-tested against it.

## Auth setup

Two credential shapes are accepted, matching legacy's own first-match-wins precedence in
`okta.authenticator`: an `api_token` secret (Okta API token, sent as `Authorization: SSWS
<api_token>` via `api_key_header` mode) takes priority when configured; otherwise an `access_token`
secret (OAuth access token, sent as a standard Bearer token) is used. When both are configured,
`api_token` wins — the dual-`auth` candidate list declares `api_token`'s `api_key_header` spec
FIRST, exactly reproducing legacy's precedence (see `docs/migration/conventions.md` §3's
"Dual-auth ordering is load-bearing"). At least one of the two secrets must be configured, or auth
selection hard-errors.

## Streams notes

- `users` (`GET /api/v1/users`) and `groups` (`GET /api/v1/groups`) both use `link_header`
  pagination (RFC 5988 `Link: <url>; rel="next"`), matching legacy's `connsdk.Harvest` +
  `connsdk.LinkHeaderPaginator` call. `records.path` is `.` (body root is a bare JSON array) for
  both — legacy's `connsdk.Harvest` reads the array from the response root the same way.
  `computed_fields` rename `profile.email`/`profile.login` (users) and `profile.name`/
  `profile.description` (groups) up to top-level schema fields, matching legacy's `userRecord`/
  `groupRecord` mapping functions field-for-field; `lastLogin` is renamed to `last_login` for the
  same reason.
- `system_logs` (`GET /api/v1/logs`) sends a `since` query parameter only when an incremental lower
  bound resolves (declared via `incremental.request_param: since`, `param_format: rfc3339` — legacy
  only sets `since` when `lowerBound(req)` is non-empty, exactly matching the engine's own
  absent-lower-bound omission), and paginates identically via `link_header`. `computed_fields` rename
  the raw API's `eventType`/`displayMessage` camelCase fields to `event_type`/`display_message`.
- Every stream sends `limit=200` (matches legacy's default/max `page_size` of 200) via each
  stream's static `query: {"limit": "200"}`; `base.pagination`'s own `page_size: 200` is the
  paginator's short-page stop threshold, not a query param the `link_header` paginator itself sends
  (it reads `Link` headers only, mirroring the `cursor`/`last_record_field` "declared but not
  read by this paginator type" pattern documented for stripe's `page_size`/`limit_param`).

## Write actions & risks

None. `capabilities.write` is `false`; legacy's `Write` is an unconditional
`connectors.ErrUnsupportedOperation` stub, matched by this bundle omitting `writes.json` entirely.

## Known limits

- **`domain` config key dropped; `base_url` is now required.** Legacy derives the API host from a
  bare `domain` value (prepending `https://` when the value has no scheme) when `base_url` is
  unset (`okta.baseURL`). The engine's spec-default materialization only fills in a literal
  per-key default — it cannot express "derive `base_url` by prepending a scheme onto `domain`", a
  cross-key template (the same class chargebee's `site`/sentry's `hostname`/chargify's
  `domain`/`subdomain` hit; see `docs/migration/conventions.md`'s Known-limits pattern for
  chargify). This bundle drops `domain` entirely and requires `base_url` instead: an operator
  migrating a legacy `domain`-only config must now supply the fully-formed
  `https://<domain>` URL as `base_url`. Documented config-surface narrowing, not a data-shape
  regression — every legacy-accepted `domain` value has an operator-reachable `base_url`
  equivalent.
- **`link_header` pagination ships a single-page fixture per stream**, not the usual 2-page
  requirement (§4): a `Link: <url>; rel="next"` header's URL must point at the replay server's own
  runtime-assigned port, which a static fixture file cannot embed correctly — the identical
  sanctioned harness limitation already documented for gitlab/freshdesk's `link_header` bundles.
  `pagination_terminates` still exercises real single-page termination-on-no-Link-header behavior
  against this fixture; no live `paritytest` exists for a genuine 2-page Link-header advance in this
  wave (a future wave could add one using bitly/calendly's `next_url` live-parity pattern if Okta's
  pagination behavior needs to be proven live).
- Full Okta admin surface (apps, policies, factors, roles, user/group lifecycle writes) is out of
  scope for this migration; see `api_surface.json`'s `excluded` entries. Only the 3 legacy-parity
  read streams are implemented.
