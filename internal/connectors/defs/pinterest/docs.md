# Overview

Pinterest is a fresh declarative-HTTP migration. It reads Pinterest ad accounts, boards,
campaigns, ad groups, and audiences through the Pinterest API v5. This bundle is engine-vs-legacy
parity-tested against `internal/connectors/pinterest` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until the registry flip. It is read-only:
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy exactly
(legacy has no `write.go` and reports `Capabilities.Write: false`).

## Auth setup

Pinterest authenticates with OAuth 2.0's refresh-token grant. Provide `client_id`/`client_secret`/
`refresh_token` secrets; the connector exchanges the refresh token for a short-lived access token
at `token_url` (default `https://api.pinterest.com/v5/oauth/token`) and sends that access token as
`Authorization: Bearer <token>` on every data request.

The engine's declarative `auth` dialect only supports `oauth2_client_credentials`
(`grant_type=client_credentials`), never `grant_type=refresh_token` — this is a genuine Tier-2
`AuthHook` trigger (token-exchange auth, conventions.md §1), so the token exchange lives in
`internal/connectors/hooks/pinterest/hooks.go` (~240 lines, one hook interface, well under the
Tier-2 caps). The hook authenticates the CLIENT via HTTP Basic (`client_id`/`client_secret` in the
`Authorization` header on the token request), exactly matching legacy's
`refreshTokenAuth.accessToken` (`pinterest.go`) — this is a deliberate, load-bearing difference from
other refresh-token connectors in this repo (e.g. `strava`), whose token endpoints instead accept
`client_id`/`client_secret` as form fields; Pinterest's token endpoint does not. The cached access
token is refreshed 60 seconds before its declared `expires_in` expiry, falling back to a 1-hour TTL
when the token response omits `expires_in` (Pinterest's documented default access-token lifetime),
matching legacy's `decodeTokenResponse` exactly.

`base_url` defaults to `https://api.pinterest.com/v5`; `token_url` defaults to
`https://api.pinterest.com/v5/oauth/token`. Both accept an override for tests/proxies and are
validated as well-formed `http(s)://` URLs with a host before use (fails closed rather than risking
exfiltrating the refresh token/client secret to an attacker-chosen endpoint).

## Streams notes

All 5 streams share Pinterest v5's list-endpoint shape: `GET`, records at `items`, bookmark-cursor
pagination (`pagination.type: cursor` with `token_path: bookmark`/`cursor_param: bookmark` — the
next page is requested with `?bookmark=<token>`; pagination stops when the response's `bookmark` is
absent/null/empty, matching legacy's `harvest` loop exactly, including its `next == "" || next ==
"null"` check: a JSON `null` bookmark decodes to Go `nil`, which `connsdk.StringAt`'s `stringify`
renders as `""`, so the engine's stock stop-on-empty-token cursor behavior already reproduces
legacy's check with no extra `stop_path` needed).

`ad_accounts` and `boards` are unscoped (global to the authenticated user). `campaigns`,
`ad_groups`, and `audiences` are ad-account-scoped: their `path` templates
`{{ config.account_id }}` directly (e.g. `/ad_accounts/{{ config.account_id }}/campaigns`) — an
absent `account_id` for these 3 streams hard-errors at path-interpolation time (`config.account_id`
referenced but unresolved), exactly reproducing legacy's `resolveResourcePath`'s explicit "pinterest
account-scoped stream requires config account_id" error. `account_id` is declared in `spec.json`'s
properties but deliberately NOT in `required[]`, since it is only needed by 3 of the 5 streams (the
same asymmetric requirement legacy enforces at read time, not connection time).

`page_size` is sent via each stream's opt-in optional-query dialect
(`{"template": "{{ config.page_size }}", "omit_when_absent": true}`) — omitted entirely when unset
so the Pinterest API applies its own default, matching legacy's `pinterestPageSize`'s
`page_size == 0` "let API default apply" behavior byte-for-byte (legacy never sends the param at
all when unset; a materialized `spec.json` default here would have changed that, so `page_size` has
no `"default"` value).

No stream publishes an `incremental` block: legacy's Pinterest connector is full-refresh only (no
catalog `CursorFields`, no client- or server-side incremental filter of any kind) — this bundle
matches that (design §8's incremental truth table: no cursor field published, no incremental
block).

## Write actions & risks

None. Legacy `internal/connectors/pinterest` is read-only end to end (`Capabilities.Write: false`,
no `write.go`); `capabilities.write` is `false` here and this bundle ships no `writes.json`.

## Known limits

- Full Pinterest v5 API surface (pins, pin/campaign analytics, catalogs, conversion events, user
  account, ads, media uploads) is out of scope — legacy itself only ever implemented the 5 streams
  this bundle migrates; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- Legacy accepts a runtime `max_pages` cap, but the declarative engine only supports fixed
  bundle-authored `pagination.max_pages` integers. This bundle intentionally does not declare an
  ignored `max_pages` `spec.json` property.
