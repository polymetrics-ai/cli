# Overview

Gmail is a wave1-pilot Tier-2 (AuthHook) migration (PLAN.md P-10, SPEC.md §5.7). It reads Gmail
messages, threads, drafts, and labels via the Google OAuth 2.0 **refresh-token grant** only — the
3-legged consent/acquisition dance is out of scope (the refresh token arrives as a pre-issued
secret; the credentials layer already owns acquisition/storage). This bundle is engine-vs-legacy
parity-tested against `internal/connectors/gmail` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip. Read-only: legacy
`gmail.go:191-192` always returns `ErrUnsupportedOperation` from `Write`, and this bundle declares
`capabilities.write: false` with no `writes.json` to match.

## Auth setup

Provide three secrets: `client_id`, `client_secret` (optional for some OAuth client types — see
below), and `client_refresh_token` (long-lived; never logged). `hooks/gmail/hooks.go` implements
`AuthHook`, mirroring legacy `gmail/auth.go`'s `oauthRefreshAuth`: it POSTs
`grant_type=refresh_token` + `refresh_token` + `client_id` [+ `client_secret`, `scope`] to
`token_url` (default `https://oauth2.googleapis.com/token`, config-overridable), caches the
resulting access token until 60 seconds before its declared expiry, and sets
`Authorization: Bearer <access_token>` on every request. `client_secret` is omitted from the
token-request form when unset (matches legacy's `if a.clientSecret != ""` guard) — some Google
OAuth client types (e.g. installed-app/native clients) issue refresh tokens that don't require a
client secret at token-refresh time.

`token_url` MUST resolve to an `https://` URL (THREAT-MODEL.md Delta 2): the hook fails closed on
a non-https or unparseable override rather than sending the refresh token/client secret to an
attacker-chosen endpoint. This is the one new SSRF-adjacent surface this phase adds, mirroring
legacy's `validatedURL` (gmail.go:339) but tightened to https-only in the hook (legacy's
`validatedURL` also accepted plain `http`; the hook narrows this specifically for the token
endpoint that receives secret material — see Known limits).

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook": "gmail", ...}`
— legacy has no alternate auth path (no static API key, no public/no-auth fallback), so there is no
`when`-gated bearer-token bypass to declare, unlike github's token-OR-app-JWT "auto" resolution.

## Streams notes

Four streams, all primary-keyed on `id`: `messages`, `threads`, `drafts` (all three paginated via
Gmail's `pageToken`/`nextPageToken` cursor convention — `pagination.type: cursor` with
`token_path: nextPageToken`, `cursor_param: pageToken`) and `labels` (unpaginated — a stream-level
`pagination: {"type": "none"}` override, since Gmail's `labels.list` returns every label in one
response, matching legacy's `paginated: false` routing table entry, streams.go:28). Every stream
sends `maxResults` per `config.page_size` (default 100) except `labels`, which takes no page-size
query param at all (matches legacy exactly — `applyListFilters`/pagination params only apply to the
paginated branch, gmail.go:155-177).

`computed_fields` rename each stream's camelCase raw fields to the schema's snake_case names
(`threadId` -> `thread_id`, `historyId` -> `history_id`, `messageListVisibility` ->
`message_list_visibility`, etc.) and, for `drafts`, reach into the nested raw `message` object
(`record.message.id` -> `message_id`, `record.message.threadId` -> `thread_id`) exactly like legacy
`draftRecord`'s type-asserted nested read (streams.go:116-127); a draft with no nested `message`
object silently omits those two fields for that record (computed_fields' documented
absent-source-path tolerance), matching legacy's own `nil`-default behavior for a missing/malformed
`message` object.

**Documented parity deviation — computed_fields rename stringifies `labels`' 4 numeric count
fields.** `labels`' `messagesTotal`/`messagesUnread`/`threadsTotal`/`threadsUnread` are real Gmail
API JSON integers (legacy's `labelRecord` copies them straight off the raw decoded envelope, so
legacy emits a `json.Number` for each, matching `connsdk`'s `UseNumber`-based decode). Renaming
them to their schema's snake_case names is only expressible via `computed_fields` (plain schema
projection matches by exact key name only), and `computed_fields` resolves every template through
`engine.Interpolate`, which always returns a Go `string` regardless of the raw JSON value's real
type (`engine/interpolate.go`'s `resolveExpr`/`stringify`) — so this bundle emits these 4 fields as
strings, not integers. This is the identical engine limitation chargebee's migration independently
discovered for its full per-field envelope unwrap (`docs/migration/conventions.md` §5 candidate
entry); it never changes the numeric VALUE a consumer would read (`"10"` carries the same
information as `json.Number("10")`), so it is an ACCEPTABLE, ledgered deviation, not silently
absorbed by a coercing test helper — see `paritytest/gmail`'s
`TestParityGmail_ComputedFieldsStringifyLabelCountFields`, which asserts the type change and the
value-equality explicitly. Schema types are declared `["string", "null"]` for these 4 fields
(reflecting what the bundle ACTUALLY emits) rather than `"integer"` (conventions.md §4: declare the
field's real emitted type, don't widen to paper over a limitation).

**No incremental sync mode**: legacy's own doc comment (streams.go:31-34) states the Gmail list
endpoints are newest-first with no publishable cursor field, so no stream declares an
`incremental` block — full_refresh only, matching legacy (`InitialState` always seeds an empty
cursor). The `start_date` config value is NOT modeled as pagination/incremental in this bundle: it
was legacy's own client-side Gmail search-query filter (`after:<unix-seconds>`, applied via the
`q` query param, gmail.go:282-296) rather than the engine's `incremental.request_param` machinery
— an equivalent `q` search-query template is not implemented in this pass (Known limits) since the
engine's `stream.Query` templating has no optional/absent-tolerant substitution for an
unset-by-default value the way `auth`'s `when` grammar does (conventions.md §3), and a
static `q` query key would send `after:` unconditionally, which is not what legacy does when
`start_date` is unset.

## Write actions & risks

None — Gmail is read-only. `capabilities.write: false`, no `writes.json` file, matching legacy's
`ErrUnsupportedOperation` (`gmail.go:191-192`) and every other read-only pilot in this wave
(TEST-PLAN.md §1: "All other pilots return `ErrUnsupportedOperation`... no write parity tests,
`capabilities.write: false`, no writes.json").

## Known limits

- **`start_date`/`include_spam_and_trash` search-query filters are declared in `spec.json` but not
  wired into a stream `query` template this pass** — legacy's `q=after:<unix-seconds>` and
  `includeSpamTrash=true` filters (gmail.go:264-296) have no absent-key-falsy-tolerant home in the
  engine's `stream.Query` templating (conventions.md §3: "no such omission tolerance at all...
  every `{{ }}` reference in a `query` map value is resolved unconditionally"), so wiring either as
  a plain query template would send `after:` / `includeSpamTrash=true` on every request instead of
  only when configured, which is not parity. Left undeclared in `streams.json` (not silently wrong;
  see conventions.md F6 — a spec property with no wired template is dead config, and both fields
  are retained in `spec.json` only as forward-compatible surface for a future conditional-query
  engine feature, documented here rather than silently dropped). Deferred to Pass B.
- **`token_url` https-only enforcement is stricter than legacy's `validatedURL`** (which accepted
  plain `http` too, gmail.go:342-357): the hook only accepts `https://` overrides. This is
  documented as a parity deviation (never stricter for any *production* Google OAuth endpoint,
  which is always https; strictly safer for the one new SSRF-adjacent secret-bearing surface this
  phase introduces). See the parity-deviation ledger in `docs/migration/conventions.md` §5.
- **`TestConformance/gmail`'s dynamic (fixture-replay) checks cannot fully pass**: `conformance`'s
  dynamic checks (`internal/connectors/conformance/dynamic.go`) always call `engine.Check`/
  `engine.Read` with a `nil` `Hooks` argument (by design — conformance has no per-connector hook
  wiring mechanism), so any bundle whose *sole* auth candidate is `mode: custom` (this bundle; no
  fallback auth path exists in legacy to declare a `when`-gated alternative) fails
  `buildCustomAuth`'s `hook %q not registered (no hooks provided)` error before any HTTP request is
  even attempted. This reproduces identically for any Tier-2 AuthHook-only connector (github's
  `custom` candidate is masked only because its declarative bearer-token candidate is tried first
  and conformance's synthetic secret values satisfy it — the same underlying gap, not a
  counter-example). This is an `ENGINE_GAP` (conventions.md §6) recorded in
  `.planning/phases/wave1-pilot/traces/p10-gmail-ledger.md` and this connector's `result.schema.json`
  entry, not worked around here (conventions.md §5's meta-rule: never invent Go/weaken an assertion
  to paper over an engine limitation). The `paritytest/gmail` suite (which DOES wire the real
  `AuthHook` via `engine.HooksFor("gmail")`, matching monday's precedent) is therefore the
  authoritative parity/correctness bar for this connector's auth path, not `TestConformance/gmail`.
