# Overview

Outlook is a read-only Tier-2 (bundle + hooks) migration of `internal/connectors/outlook` (the
hand-written connector it replaces at capability parity, quarantined `AUTH_COMPLEX` in
`docs/migration/quarantine.json`: "OAuth2 refresh_token grant pre-request (hook needed, gmail
pattern)"). It reads messages, mail folders, and calendar events from the authenticated user's
mailbox through Microsoft Graph v1.0. The legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide the Microsoft Entra ID (Azure AD) application's `client_id`, `client_secret`, and a
long-lived `refresh_token` — all three are required secrets. `internal/connectors/hooks/outlook/
hooks.go`'s `AuthHook` exchanges them for a short-lived access token via an OAuth 2.0
`grant_type=refresh_token` POST, exactly porting legacy `outlook.go`'s `refreshTokenAuth` (gmail's
own pilot pattern, `internal/connectors/hooks/gmail/hooks.go`, ported here): the token endpoint is
`token_url` when set, otherwise derived from `tenant_id` (default `"common"`) as
`https://login.microsoftonline.com/{tenant_id}/oauth2/v2.0/token`; the resulting bearer token is
cached until 60s before its declared expiry (`expires_in`, default 3600s when absent) and applied
as `Authorization: Bearer <token>` on every Graph request. `scope` is optional and omitted from the
token-request form entirely when unset (matching legacy's `if strings.TrimSpace(a.Scope) != ""`
guard). `token_url` MUST resolve to an `https` URL with a host — the hook fails closed on anything
else (THREAT-MODEL.md Delta 2: an attacker-controlled `token_url` override could otherwise
exfiltrate `client_secret`/`refresh_token` to an arbitrary endpoint), stricter than legacy's
`validateBaseURL` (which also accepted plain `http`); documented as a parity deviation below since
it is never stricter for any real Microsoft identity-platform endpoint (always `https`). Secret
values (`client_secret`, the refresh token, cached access tokens) flow only into the outgoing
token-request form or the `Authorization` header; they are never logged and never appear in an
error string.

Bundle auth declares a single `mode: custom, hook: outlook` candidate — legacy has no alternate
auth path, so there is no roster to reorder (unlike zendesk-support's dual-candidate precedence
rule, conventions.md §3).

## Streams notes

All 3 streams (`messages`, `mail_folders`, `events`) share Microsoft Graph's list-endpoint shape:
`GET /me/<resource>`, records under `value`. Every stream is entirely `StreamHook`-driven
(`internal/connectors/hooks/outlook/hooks.go`'s `ReadStream`), NOT the declarative pagination path
— Graph's real pagination cursor is `@odata.nextLink`, an absolute next-page URL carried in a
response-body key that itself contains a literal `.` (`"@odata.nextLink"`). The engine's
declarative `next_url` pagination type reads its cursor via `connsdk.StringAt`'s dotted-path
parser (`engine/paginate.go`/`connsdk/extract.go`'s `selectPath`), which treats every `.` in a path
as a nesting separator and therefore cannot address a literal dotted key at all — a direct probe
(`connsdk.StringAt(body, "@odata.nextLink")` against a real Graph-shaped body) resolves silently to
`""` with no error, which would make declarative `next_url` pagination stop after page 1 for every
real Outlook response, silently dropping every record past the first page for any mailbox with
more items than `page_size` — a real accepted-input record-count regression versus legacy, not
cosmetic. This is the identical gap already recorded for `microsoft-entra-id`/`microsoft-lists`/
`microsoft-teams` in `docs/migration/quarantine.json`. Base-level `pagination.type` is declared
`"none"` (inert placeholder — every dispatch goes through the `StreamHook`, which never falls back
to the declarative path for a recognized stream name).

`ReadStream` ports legacy `outlook.go`'s `harvest`/`nextLink` loop verbatim: request the stream's
Graph collection path with `$top=<page_size>` on the first page only (subsequent pages are driven
entirely by the absolute `nextLink` URL, which already carries its own `$skiptoken`/query state,
sent with no additional query params), extract records at `value`, decode the literal
`"@odata.nextLink"` key directly via `encoding/json` (bypassing the dotted-path limitation), and
stop when it is absent or empty. `max_pages` (config, permissive parse — empty/`all`/`unlimited`/
malformed/negative all mean unbounded) mirrors legacy's `maxPages` exactly. Every stream's schema
properties are the exact field-for-field rename of legacy's `messageRecord`/`folderRecord`/
`eventRecord` (e.g. `receivedDateTime` → `received_date_time`, `displayName` → `display_name`).

## Write actions & risks

None. Outlook is exposed read-only in legacy (`Capabilities.Write: false`); this bundle ships no
`writes.json`.

## Known limits

- The `@odata.nextLink` pagination `ENGINE_GAP` above is resolved via a `StreamHook`, not a
  declarative dialect feature — this is the sanctioned "whole-stream override" Tier-2 trigger
  (conventions.md §1's table), not a workaround: legacy behavior is preserved exactly (every page
  is followed, no record dropped, capped only by the same `max_pages`), the deviation is purely
  about WHERE the pagination logic lives (Go hook vs. declarative JSON), never about emitted record
  data or count for any input legacy itself would accept. See `docs/migration/quarantine.json`'s
  `outlook` entry ("OAuth2 refresh_token grant pre-request (hook needed, gmail pattern)") for the
  auth side of the original quarantine reason; the pagination StreamHook was required in addition
  once the migration was underway (both hooks live in the same `internal/connectors/hooks/outlook/
  hooks.go` file, well under the Tier-2 400-line hard ceiling with exactly 2 hook interfaces).
- `token_url` is required to be `https` (never plain `http`), stricter than legacy's
  `validateBaseURL`. ACCEPTABLE: no real Microsoft identity-platform token endpoint is ever
  `http`, and this closes a genuine credential-exfiltration SSRF-adjacent risk (THREAT-MODEL.md
  Delta 2), matching gmail's identical, already-accepted deviation.
- Outlook's write surface (sending mail, replying, creating/deleting events) is out of scope;
  legacy itself never implemented it, so there is no parity gap, only an out-of-scope Pass B
  expansion (see `api_surface.json`).
- `messages`/`events` declare their respective legacy `CursorFields` (`received_date_time`/
  `last_modified_date_time`) as `x-cursor-field` for manifest parity, but neither legacy nor this
  bundle issues a server-side incremental filter against them (legacy performs full syncs only via
  its `harvest` loop); `mail_folders` has no cursor field at all, matching legacy exactly.
- Every stream carries a `conformance.skip_dynamic` marker (stream-level) and the bundle also
  carries one at `metadata.json` top level: conformance's synthetic, non-secret config
  (`"synthetic-conformance-value"`) can never populate a real `token_url`/refresh-token round trip,
  and every stream's real read path is StreamHook-driven rather than declarative — both are exactly
  the conditions conventions.md's skip-marker section names as legitimate (gmail/strava's identical
  bundle-level marker for the auth side; microsoft-teams' identical per-stream markers for the
  StreamHook side). The parity suite (`internal/connectors/paritytest/outlook`) and the hook's own
  unit tests (`internal/connectors/hooks/outlook/hooks_test.go`) are the authoritative proof of
  correctness this marker points to.
