# Overview

Salesloft is a fresh declarative-HTTP migration. It reads Salesloft people, accounts, cadences,
users, and emails through the Salesloft REST API v2. This bundle is engine-vs-legacy
parity-tested against `internal/connectors/salesloft` (the hand-written connector it migrates);
the legacy package stays registered and unchanged until the registry flip. It is read-only:
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy exactly
(legacy has no `write.go` and reports `Capabilities.Write: false`).

## Auth setup

Salesloft accepts three credential shapes, and legacy tries them in a fixed precedence: an API key
wins when configured, regardless of what else is present. This bundle's `streams.json` `base.auth`
candidate list reproduces that exact precedence (declaration order is load-bearing, per
conventions.md §3's dual-auth rule):

1. `api_key` secret, sent as `Authorization: Bearer <api_key>` (`mode: bearer`, `when:
   {{ secrets.api_key }}`) — tried first, matching legacy's `authenticator()`'s first check.
2. A full OAuth2 refresh-token-grant triple (`client_id`+`client_secret`+`refresh_token`) — tried
   second (`mode: custom`, `hook: salesloft`, `when: {{ secrets.refresh_token }}`). The engine's
   declarative `oauth2_client_credentials` auth mode only performs a `grant_type=client_credentials`
   exchange, never `grant_type=refresh_token`, so this is a genuine Tier-2 `AuthHook` trigger
   (token-exchange auth, conventions.md §1); the exchange lives in
   `internal/connectors/hooks/salesloft/hooks.go` (~290 lines, one hook interface). Unlike
   Pinterest's token endpoint, Salesloft's authenticates the client via FORM-ENCODED
   `client_id`/`client_secret` fields, not HTTP Basic — this bundle's hook preserves that exact wire
   shape. The hook also rotates the refresh token whenever a token response carries a new one
   (legacy `auth.go:113-116`), and honors a pre-existing `access_token` secret exactly once before
   any network refresh call when a full OAuth triple is ALSO configured (legacy's `seedToken`
   behavior, `auth.go:63-69`) — the seed value is read directly from `secrets.access_token`.
3. A standalone `access_token` secret (no refresh triple configured), sent directly as
   `Authorization: Bearer <access_token>` (`mode: bearer`, `when: {{ secrets.access_token }}`) —
   the fallback candidate.

`base_url` defaults to `https://api.salesloft.com/v2`; `token_url` defaults to
`https://accounts.salesloft.com/oauth/token`. Both accept an override for tests/proxies and are
validated as well-formed `http(s)://` URLs with a host before use.

## Streams notes

All 5 streams share Salesloft's list-endpoint shape: `GET`, records at `data`, page-number
pagination read from the response body (`pagination.type: cursor` with `token_path:
metadata.paging.next_page`/`cursor_param: page` — the next page is requested with `?page=<N>`;
pagination stops when `metadata.paging.next_page` is absent/null, matching legacy's `harvest` loop
via the SAME `connsdk.StringAt`-based stop-on-empty-token mechanism every `token_path` cursor
stream uses: a JSON `null` decodes to Go `nil`, which `stringify` renders as `""`). Legacy also
defensively treats a literal `"0"` next_page as a stop signal; the engine's stock `token_path`
cursor paginator does not special-case `"0"` as additionally falsy (only empty string stops it) —
see Known limits.

`per_page` is sent via each stream's static `{{ config.page_size }}` query template (default `100`,
matching legacy's `salesloftDefaultPageSize`). Every stream is incremental on `updated_at`
(`incremental.request_param: updated_at[gte]`, `param_format: rfc3339`, `start_config_key:
start_date`) — the resolved lower bound (persisted cursor, or `start_date` on a fresh sync) is sent
verbatim as an RFC3339 string, matching legacy's `incrementalLowerBound`. `sort_by=updated_at` and
`sort_direction=ASC` are sent in the SAME branch as the incremental filter, via the opt-in
optional-query dialect's `{{ incremental.lower_bound | const:... }}` pattern (`omit_when_absent:
true` on both) — present only when the lower bound resolves, matching legacy's `harvest`, which
sets `updated_at[gte]`/`sort_by`/`sort_direction` together in one `if updatedSince != ""` branch and
never partially.

`people` and `accounts` additionally declare `computed_fields` (`account_id`/`owner_id` for
`people`, `owner_id` for `accounts`) that extract the nested Salesloft relationship object's `id`
(a bare `{{ record.account.id }}`/`{{ record.owner.id }}` reference — the typed-extraction rule,
conventions.md §3, preserves the raw integer type), matching legacy's `relationID` helper
byte-for-byte. `cadences`/`users`/`emails` have no relationship fields to flatten.

## Write actions & risks

None. Legacy `internal/connectors/salesloft` is read-only end to end (`Capabilities.Write: false`,
no `write.go`); `capabilities.write` is `false` here and this bundle ships no `writes.json`.

## Known limits

- **`next_page: "0"` as an additional stop signal is not modeled.** Legacy's `harvest` loop stops
  on `next == "" || next == "0"`; the engine's stock `token_path` cursor paginator stops only on an
  empty/absent token, not a literal `"0"`. Salesloft's real API returns `next_page: null` (never a
  literal `0`) on the terminal page in practice — a `0` value would be a nonsensical 0-indexed page
  number for a 1-based pagination scheme legacy itself never expected to see — so this is a
  defensive-only divergence with no observed real-world trigger; documented as a parity deviation
  rather than silently dropped. ACCEPTABLE per conventions.md §5's meta-rule: it never changes
  emitted record data for any input the real Salesloft API would send.
- Full Salesloft v2 API surface (calls, meetings, notes, tasks, action items, imports, etc.) is out
  of scope — legacy itself only ever implemented the 5 streams this bundle migrates; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- `max_pages` (config) is enforced as a hard request-count cap by the engine's declarative read
  path independent of page fullness, matching legacy's own `salesloftMaxPages`-driven loop bound
  (`0`/`all`/`unlimited` all mean unbounded on both sides).
