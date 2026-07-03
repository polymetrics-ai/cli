# Overview

Jamf Pro is a read-only MDM-configuration source connector. It reads buildings, departments,
categories, and scripts through the Jamf Pro modern REST API
(`https://<subdomain>.jamfcloud.com/api`) using Basic-credential token-exchange authentication.
This bundle migrates `internal/connectors/jamf-pro` (the hand-written legacy connector); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Jamf Pro's auth is a two-step token exchange, not a single declarative auth mode: POST HTTP Basic
credentials (`username`/`password`, no request body) to `/v1/auth/token`, which returns a custom
JSON shape `{"token": "...", "expires": "..."}`; that token is then sent as `Authorization: Bearer`
on every subsequent request. This is a genuine token-exchange auth scheme — a named legitimate
Tier-2 `AuthHook` trigger (`conventions.md` §1's hook table: "token-exchange auth (GitHub App
JWT->installation token)"), mirroring github's own `AuthHook` shape exactly. Neither built-in
declarative auth mode fits: `"basic"` sends Basic credentials on every request rather than
exchanging them once for a Bearer token, and `"oauth2_client_credentials"` always POSTs a
`grant_type`/`client_id`/`client_secret`/`scope` form body — Jamf Pro's exchange sends no body at
all and uses HTTP Basic instead of form-encoded client credentials. `hooks/jamf-pro/hooks.go`'s
`AuthHook.Authenticator` builds a token-caching `connsdk.Authenticator` (refreshes 60s before the
declared `expires` timestamp, falling back to a conservative 1-hour TTL when `expires` doesn't
parse) that performs this exchange, matching legacy's `fetchToken` request/response shape exactly.
Provide `username` (plain config) and `password` (secret, never logged) plus a required `base_url`.

## Streams notes

All 4 streams (`buildings`, `departments`, `categories`, `scripts`) share the same shape: `GET`
against a Jamf Pro modern-API list endpoint that returns `{"totalCount": N, "results": [...]}`
(`records.path: "results"`), primary key `["id"]`. Pagination is genuinely 0-indexed
(`pagination.type: page_number`, `start_page: 0`, `page_param: page`, `size_param: page-size`,
`page_size: 100`) — the first request sends `page=0`, matching legacy's
`jamf_pro_test.go` assertions and `harvest`'s `for page := 0; ...` loop exactly; pagination stops
on a short/empty page (fewer than `page_size` records), identical to legacy's
`len(records) < pageSize` check. None of the 4 streams declare an `incremental` block — these are
core MDM configuration resources with no cursor field in legacy's own `jamfStreams()`.

**Documented parity deviation**: legacy also stops early when the running record count reaches the
response body's `totalCount` field (`jamfTotalCount`), in addition to the short-page stop. The
engine's `page_number` paginator only implements the short-page stop signal, not a body-field
total-count check. This can only cause, at most, one harmless extra request on the rare page where
a full-size page happens to exactly exhaust `totalCount` (the following request then returns an
empty/short page and stops normally) — it never omits, duplicates, or reorders any record for any
input legacy itself would accept, so it is ACCEPTABLE per the meta-rule (see the parity-deviation
ledger).

`check` issues a single bounded `GET /v1/buildings?page=0&page-size=1`, mirroring legacy's `Check`
implementation conceptually (a 1-record probe confirms auth and connectivity without mutating
anything) — unlike legacy's `Check`, which only performs the token exchange itself, this bundle's
declarative `check` also issues one bounded data request, since `AuthHook`-based auth has no
separate declarative "just fetch a token" check primitive; the token exchange still happens first,
as a side effect of resolving auth for that request.

## Write actions & risks

None. Jamf Pro is read-only (`capabilities.write: false`); legacy's own `Write` always returns
`connectors.ErrUnsupportedOperation` and there is no reverse-ETL write target for this API.

## Known limits

- Only the 4 legacy-parity read streams are implemented; see `api_surface.json`. Jamf Pro's full
  documented surface (computers, mobile devices, policies, configuration profiles, webhooks, etc.)
  is out of scope until Pass B.
- `base_url` is **required** in this bundle rather than derived automatically from a bare
  `subdomain` config value, as legacy's `jamfBaseURL` does at runtime
  (`https://<subdomain>.jamfcloud.com/api`) — the engine's `spec.json` `"default"` mechanism only
  materializes a FIXED literal default, not one derived from another config value (see
  `conventions.md` §3's `spec.json` `"default"` materialization note, and algolia's identical
  documented `base_url`-required narrowing for the same reason). Provide the full base URL
  directly.
- Every stream is full-refresh only; these are core MDM configuration resources with no cursor
  field, matching legacy exactly.
