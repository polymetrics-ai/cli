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

**RESOLVED — `labels`' 4 numeric count fields now preserve their native JSON type.**
`labels`' `messagesTotal`/`messagesUnread`/`threadsTotal`/`threadsUnread` are real Gmail API JSON
integers (legacy's `labelRecord` copies them straight off the raw decoded envelope, so legacy emits
a `json.Number` for each, matching `connsdk`'s `UseNumber`-based decode). Renaming them to their
schema's snake_case names requires `computed_fields` (plain schema projection matches by exact key
name only), and each rename template is a BARE single `{{ record.messagesTotal }}`-shaped
reference (no filter, no surrounding literal text) — the gap-loop cycle-1 engine mini-wave's typed
`computed_fields` extraction (REVIEW-A.md adjudication A1) now copies the raw JSON value straight
through for exactly this template shape instead of stringifying it via `Interpolate`, so this
bundle emits these 4 fields as native `json.Number` integers, matching legacy exactly. (This
bundle independently discovered the identical engine limitation chargebee's migration hit for its
full per-field envelope unwrap — both are now resolved by the same engine increment,
`docs/migration/conventions.md` §5.) Schema types are declared `["integer", "null"]` for these 4
fields (their real wire type) — see `paritytest/gmail`'s
`TestParityGmail_ComputedFieldsPreserveLabelCountFieldsNativeType` (formerly
`TestParityGmail_ComputedFieldsStringifyLabelCountFields`, renamed and flipped to assert native-type
equality now that the engine gap is closed).

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
- **`TestConformance/gmail`'s dynamic (fixture-replay) checks are genuinely `skip_dynamic`'d, NOT
  because conformance is hook-blind.** (Stale claim fixed: this bullet previously said conformance
  "always call[s] engine.Check/engine.Read with a nil Hooks argument (by design — conformance has
  no per-connector hook wiring mechanism)". That was true when gmail was first migrated, but R3
  (an earlier wave1-pilot repair round) made `internal/connectors/conformance/dynamic.go`
  hook-aware — it now calls `engine.HooksFor(b.Name)` at every dynamic-check call site
  (`Check`/`Read`/`Write`), and `hooks/hookset`'s blank-import wires every registered connector's
  real hooks in. github, whose bundle carries 2 hook interfaces (AuthHook + WriteHook), gets FULL
  dynamic coverage today with zero skip markers — the opposite of "conformance cannot wire hooks".)
  gmail's `metadata.json` still carries a genuine, honest `skip_dynamic` marker for a DIFFERENT
  reason: its *sole* auth candidate is `mode: custom` (legacy has no fallback auth path to declare a
  `when`-gated alternative), and conformance's synthetic config can never carry a real `https`
  `token_url` — the AuthHook's own https-only guard (see above) means no synthetic secret value can
  ever satisfy it, so every auth-resolving dynamic check would fail identically and uninformatively
  regardless of hook wiring. (github's `custom` candidate never hits this: its declarative
  bearer-token candidate is tried FIRST and conformance's synthetic token secret satisfies it before
  `custom` is ever reached — not evidence of a repo-wide hook-blindness gap, just a different
  candidate-ordering outcome.) The marker's own `reason` text (in `metadata.json`) has always been
  accurate; only this surrounding docs.md prose was stale. `paritytest/gmail` (which wires the real
  `AuthHook` via `engine.HooksFor("gmail")`, matching monday's precedent) remains the authoritative
  parity/correctness bar for this connector's auth path — `TestConformance/gmail` still passes today
  (the marker-skip path, not a bypassed/expected-fail path).
