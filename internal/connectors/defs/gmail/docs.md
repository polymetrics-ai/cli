# Overview

Gmail started as a wave1-pilot Tier-2 (AuthHook) migration (PLAN.md P-10, SPEC.md §5.7) and is now
a **Pass B full-surface expansion** (measure-first probe): every documented Gmail API v1 REST
endpoint (79 total, per the live Discovery document at
`https://www.googleapis.com/discovery/v1/apis/gmail/v1/rest`) is enumerated in `api_surface.json`,
with 10 practical GET list/detail resources implemented as streams and every expressible mutation
implemented as a `writes.json` action (35 actions). Auth is still the Google OAuth 2.0
**refresh-token grant** only — the 3-legged consent/acquisition dance is out of scope (the refresh
token arrives as a pre-issued secret; the credentials layer already owns acquisition/storage). The
wave1-pilot read-only streams (`messages`/`threads`/`drafts`/`labels`) remain engine-vs-legacy
parity-tested against `internal/connectors/gmail` (the hand-written connector this bundle
originally migrated); the legacy package stays registered and unchanged until wave6's registry
flip. **This bundle is no longer read-only**: `capabilities.write` is now `true` and `writes.json`
declares 35 actions across messages/threads/drafts/labels/filters/send-as aliases/delegates/
forwarding addresses/account settings — legacy's own `gmail.go:191-192` (`ErrUnsupportedOperation`
from every `Write` call) describes the ORIGINAL pilot scope, not this connector's current surface;
`paritytest/gmail`'s write-parity assertions against that legacy behavior were updated alongside
this expansion (see Write actions & risks).

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

**`scopes` default widened for Pass B**: the wave1-pilot default was the read-only
`https://www.googleapis.com/auth/gmail.readonly` scope. Since this bundle now declares mutating
write actions, `spec.json`'s `scopes` default is now the full-mailbox
`https://mail.google.com/` scope — narrow this back to a read-only scope string in `config.scopes`
if only this bundle's streams are ever exercised for a given credential.

## Streams notes

**10 streams total.** `messages`, `threads`, `drafts` (all three paginated via Gmail's
`pageToken`/`nextPageToken` cursor convention — `pagination.type: cursor` with
`token_path: nextPageToken`, `cursor_param: pageToken`), `labels`, `filters`, `send_as`,
`delegates`, `forwarding_addresses`, and `profile` are unpaginated (labels/filters/send_as/
delegates/forwarding_addresses all return their full collection in one response; `profile` is a
single-record mailbox-identity resource, not a collection at all — `records.path: ""` selects the
whole response body as one record). `history` is paginated via the same `pageToken`/
`nextPageToken` cursor convention as messages/threads/drafts.

Every paginated stream sends `maxResults` per `config.page_size` (default 100) except `labels`,
which takes no page-size query param at all (matches legacy exactly — `applyListFilters`/
pagination params only apply to the paginated branch, gmail.go:155-177).

`computed_fields` rename each stream's camelCase raw fields to the schema's snake_case names
(`threadId` -> `thread_id`, `historyId` -> `history_id`, `messageListVisibility` ->
`message_list_visibility`, `delegateEmail` -> `delegate_email`, `forwardingEmail` ->
`forwarding_email`, `sendAsEmail` -> `send_as_email`, `emailAddress` -> `email_address`, etc.) and,
for `drafts`, reach into the nested raw `message` object (`record.message.id` -> `message_id`,
`record.message.threadId` -> `thread_id`) exactly like legacy `draftRecord`'s type-asserted nested
read (streams.go:116-127); a draft with no nested `message` object silently omits those two fields
for that record (computed_fields' documented absent-source-path tolerance), matching legacy's own
`nil`-default behavior for a missing/malformed `message` object. `filters`' `action`/`criteria`
fields require no rename (the raw API keys already match the schema's field names exactly).

**`labels`' 4 numeric count fields, `history`'s array fields, and `profile`'s counts preserve their
native JSON type.** `labels`' `messagesTotal`/`messagesUnread`/`threadsTotal`/`threadsUnread`,
`profile`'s `messagesTotal`/`threadsTotal`, and `history`'s `messagesAdded`/`messagesDeleted`/
`labelsAdded`/`labelsRemoved` are all sourced via a BARE single `{{ record.<camelCaseField> }}`
reference (no filter, no surrounding literal text) — the engine's typed `computed_fields`
extraction (gap-loop cycle-1, REVIEW-A.md adjudication A1) copies the raw JSON value straight
through for exactly this template shape instead of stringifying it via `Interpolate`, so these
fields emit as native `json.Number`/array types, matching Gmail's real wire shape. Schema types are
declared `["integer", "null"]`/`["array", "null"]` accordingly (their real wire type) — see
`paritytest/gmail`'s `TestParityGmail_ComputedFieldsPreserveLabelCountFieldsNativeType`.

**`history` is genuinely incremental** (unlike messages/threads/drafts): `incremental.cursor_field:
id`, `incremental.request_param: startHistoryId`, `incremental.start_config_key:
start_history_id`. Gmail's `users.history.list` requires a `startHistoryId` obtained from any
message/thread/label/profile record's `historyId` field and returns only changes since that point
— a genuine server-recognized cursor, unlike the newest-first, no-cursor `messages`/`threads`/
`drafts` list endpoints. `config.start_history_id` is the seed value for a connector's first
incremental sync (obtain it from an initial `profile` read's `history_id`, or any message/thread's
own `historyId`); the engine's own state-cursor persistence drives every subsequent incremental
read from there. A stale/expired `startHistoryId` (Google documents these as valid "typically at
least a week") surfaces as an HTTP 404 from Gmail — the connector does not special-case this into
an automatic full-resync; the operator must re-seed `start_history_id` (or trigger a fresh full
sync of `messages`/`threads` instead) per Google's own documented recovery guidance.

**No incremental sync mode for `messages`/`threads`/`drafts`/`labels`**: legacy's own doc comment
(streams.go:31-34) states these list endpoints are newest-first with no publishable cursor field,
so none of these 4 streams declare an `incremental` block — full_refresh only, matching legacy
(`InitialState` always seeds an empty cursor for them). The `start_date` config value is NOT
modeled as pagination/incremental for these 4 streams: it was legacy's own client-side Gmail
search-query filter (`after:<unix-seconds>`, applied via the `q` query param, gmail.go:282-296)
rather than the engine's `incremental.request_param` machinery. The bundle wires the same
`q=after:<unix-seconds>` query template via the engine's `unix_seconds` filter and also wires
legacy's `includeSpamTrash` filter for `messages`/`threads`/`drafts`/`labels`.

## Write actions & risks

**35 write actions across 8 resources.** Every action executes exactly one HTTP request per
record (design §B.5); none require a Tier-2 `WriteHook` (all are single, direct REST calls — no
compound multi-request sequences, unlike github's `create_pull_request`).

- **Messages** (7): `send_message` (send a new RFC 2822 message), `insert_message` (insert without
  sending/no SMTP delivery), `import_message` (bulk-migration insert, bypasses default spam
  classification), `modify_message` (add/remove label IDs), `trash_message`/`untrash_message`,
  `delete_message` (**permanent, bypasses Trash** — `confirm: "destructive"`,
  `delete.missing_ok_status: [404]` for idempotent re-delete).
- **Threads** (4): `modify_thread`, `trash_thread`/`untrash_thread`, `delete_thread` (**permanent,
  bypasses Trash** — `confirm: "destructive"`, idempotent 404).
- **Drafts** (4): `create_draft`, `update_draft` (PUT, full replace), `send_draft` (sends an
  existing draft; body is `{"id": "<draftId>"}` per Gmail's `drafts.send` contract), `delete_draft`
  (idempotent 404).
- **Labels** (4): `create_label`, `update_label` (PUT, full replace), `patch_label` (PATCH, partial
  update), `delete_label` (idempotent 404; Gmail itself rejects deleting a system label).
- **Filters** (2): `create_filter`, `delete_filter` (idempotent 404). Filters can auto-forward mail
  externally (`action.forward`) — a filter that does so is a standing, unattended mail-exfiltration
  risk if misconfigured.
- **Send-as aliases** (6): `create_send_as` (triggers a verification email from Google before the
  new alias can send), `update_send_as` (PUT), `patch_send_as` (PATCH), `delete_send_as` (idempotent
  404; the account's primary address cannot be deleted — Gmail rejects that request), `verify_send_as`
  (re-sends the pending verification email).
- **Delegates** (2): `create_delegate` (grants another account read/send/delete access to this
  mailbox — Google Workspace accounts only; a significant access-control change), `delete_delegate`
  (idempotent 404).
- **Forwarding addresses** (2): `create_forwarding_address` (proposes a new address; requires
  owner-side email verification before use), `delete_forwarding_address` (idempotent 404).
- **Account settings singletons** (5, none path-parameterized — one PUT per mailbox):
  `update_auto_forwarding` (enabling this silently copies all future incoming mail externally),
  `update_vacation`, `update_language`, `update_imap` (disabling breaks any connected external IMAP
  client), `update_pop`.

None of these actions were present in the original wave1-pilot legacy `gmail.Connector` (which is
entirely read-only, `gmail.go:191-192`); they are new, Pass-B-only surface with no legacy Go
counterpart to port from, authored directly from the Gmail API v1 Discovery document. Every action
was chosen because it maps to exactly one documented REST endpoint with a well-defined non-batch
request/response shape — see `api_surface.json` for the 15 `duplicate_of`-excluded batch/
detail-GET endpoints this reasoning does not extend to.

## Known limits

- **`start_date` invalid-value handling is stricter than legacy's silent ignore**: legacy parsed
  `config.start_date` as RFC3339 and, if parsing failed, simply omitted the `q=after:<unix-seconds>`
  filter (gmail.go:288-296). The declarative bundle uses the engine's `unix_seconds` interpolation
  filter, so a malformed `start_date` fails the read instead of degrading to an unfiltered full read.
  For valid RFC3339 values, the emitted-data filter matches legacy.
- **Explicit `include_spam_and_trash=false` is sent as `includeSpamTrash=false` instead of being
  omitted**: when unset, the param is omitted exactly like legacy's false default; when set to
  `"true"`, it sends legacy's `includeSpamTrash=true`. The query dialect has no value-equality gate,
  so an explicitly configured `"false"` is sent as the API-equivalent false value rather than being
  omitted.
- **`token_url` https-only enforcement is stricter than legacy's `validatedURL`** (which accepted
  plain `http` too, gmail.go:342-357): the hook only accepts `https://` overrides. This is
  documented as a parity deviation (never stricter for any *production* Google OAuth endpoint,
  which is always https; strictly safer for the one new SSRF-adjacent secret-bearing surface this
  phase introduces). See the parity-deviation ledger in `docs/migration/conventions.md` §5.
- **Batch endpoints (`messages.batchModify`, `messages.batchDelete`) are excluded, not
  implemented** — the engine's write dialect is one-request-per-record; the identical outcome is
  reached by calling `modify_message`/`delete_message` once per id. See `api_surface.json`'s
  `duplicate_of` exclusions.
- **Attachments, S/MIME certificates, and Client-Side Encryption (CSE) key management are
  excluded** — attachment bytes are a `binary_payload`, not a syncable record; S/MIME and CSE
  require the elevated `https://mail.google.com/` scope plus (for CSE) a Google Workspace
  Enterprise/Education Plus add-on with organization-admin enablement, well outside an ordinary
  OAuth refresh-token grant's reach. See `api_surface.json`'s `requires_elevated_scope`/
  `binary_payload`/`destructive_admin` exclusions for the full per-endpoint reasoning (14 + 2 + 1
  endpoints respectively).
- **`watch`/`stop` (Cloud Pub/Sub push notification control) are excluded as `non_data_endpoint`** —
  they register/cancel a push subscription, a control-plane side effect with no record data of its
  own to read or write.
- **`TestConformance/gmail`'s dynamic READ (fixture-replay) checks are genuinely `skip_dynamic`'d,
  NOT because conformance is hook-blind; WRITE checks (`write_request_shape`/`delete_semantics`)
  are NOT skipped and run for real against every one of the 35 write-action fixtures.**
  `internal/connectors/conformance/dynamic.go`'s bundle-level skip-marker branch
  (`runDynamicChecks`) explicitly still calls `checkWriteRequestShape`/`checkDeleteSemantics` even
  when every read-side dynamic check is skipped — write checks are never gated by the read-side
  auth-resolution problem described below, since `fixtures/writes/*.json` supply the record/expect
  shape directly rather than depending on a live/replay auth handshake. gmail's `metadata.json`
  still carries a genuine, honest `skip_dynamic` marker for READS: its *sole* auth candidate is
  `mode: custom` (legacy has no fallback auth path to declare a `when`-gated alternative), and
  conformance's synthetic config can never carry a real `https` `token_url` — the AuthHook's own
  https-only guard means no synthetic secret value can ever satisfy it, so every auth-resolving
  dynamic READ check would fail identically and uninformatively regardless of hook wiring. (github's
  `custom` candidate never hits this: its declarative bearer-token candidate is tried FIRST and
  conformance's synthetic token secret satisfies it before `custom` is ever reached — not evidence
  of a repo-wide hook-blindness gap, just a different candidate-ordering outcome.) `paritytest/gmail`
  (which wires the real `AuthHook` via `engine.HooksFor("gmail")`, matching monday's precedent)
  remains the authoritative parity/correctness bar for this connector's auth path — `TestConformance/gmail`
  still passes today (the marker-skip path for reads, combined with real dynamic write checks, not a
  bypassed/expected-fail path for either).
