# Overview

Slack is a Tier-2 (CheckHook + StreamHook) migration of `internal/connectors/slack` (read-only
reference, quarantined `ENGINE_GAP` in `docs/migration/quarantine.json`). It reads Slack workspace
users, channels, and channel messages through the Slack Web API. This bundle is engine-vs-legacy
parity-tested against `internal/connectors/slack`; the legacy package stays registered and unchanged
until wave6's registry flip. Read-only: legacy `slack.go:93-95` always returns
`ErrUnsupportedOperation` from `Write`, and this bundle declares `capabilities.write: false` with no
`writes.json` to match.

## Auth setup

Provide one of two secrets: `api_token` or `access_token` (both `x-secret`, sent verbatim — no
`Bearer` prefix confusion, the engine's `bearer` auth mode adds it) as the Slack bot/user/API token
(`xoxb-`/`xoxp-`/direct API token). `streams.json`'s `base.auth` declares a first-match-wins
candidate list matching legacy's own precedence exactly (`slackSecret`, slack.go:273-283:
`api_token` checked before `access_token`):

```json
"auth": [
  { "mode": "bearer", "token": "{{ secrets.api_token }}", "when": "{{ secrets.api_token }}" },
  { "mode": "bearer", "token": "{{ secrets.access_token }}", "when": "{{ secrets.access_token }}" },
  { "mode": "none" }
]
```

**This is Slack's REAL, catalog-documented auth shape** (`website/.enrich/enr/source-slack.json`):
a static bot/user/API token, not an OAuth 2.0 refresh-token grant. Slack's classic OAuth v2 install
flow issues a `client_secret`-gated one-time `oauth.v2.access` authorization-code exchange to obtain
that token (analogous to gmail/quickbooks' 3-legged consent dance) — the resulting token itself does
not expire/refresh in this flow, so there is no repeatable refresh grant for an `AuthHook` to
implement, unlike gmail/quickbooks. `client_secret` (present in the real catalog's
`credentials` shape) is therefore intentionally NOT modeled in this bundle's `spec.json`: it is
consumed only by the one-time install-time exchange, which is out of scope here exactly like
gmail's 3-legged dance is (the credentials layer already owns acquisition/storage of the resulting
token).

## Streams notes

Three streams, matching legacy's `slackStreamEndpoints` routing table exactly: `users`
(`users.list` -> `members[]`, primary key `id`), `channels` (`conversations.list` -> `channels[]`,
primary key `id`), and `channel_messages` (`conversations.history` -> `messages[]`, primary key
`ts`, requires `config.channel_id`). **This is entirely `StreamHook`+`CheckHook`-handled, not
declarative**: Slack's Web API always returns HTTP 200 even for logical failures, signaling errors
solely via a JSON body field (`{"ok":false,"error":"<code>"}`) — auth failures, `invalid_auth`,
`token_revoked`, `missing_scope`, and every other Slack API error surfaces this way, never as a
non-2xx HTTP status. The engine's declarative read/check paths only ever treat a non-2xx HTTP status
as a failure; there is no declarative mechanism to inspect a response BODY field as a stop/error
condition (`docs/migration/quarantine.json`'s original `slack` blocker; still unresolved by the
S3/S4 engine mini-waves). `hooks/slack/hooks.go`'s `ReadStream`/`Check` port legacy's
`harvest`/`slackOK` verbatim: every page (and the `Check`'s `auth.test` call) is checked for
`ok:true`, converting an `ok:false` body into a Go error carrying Slack's own error code, even
though the HTTP status was 200.

Pagination is Slack's `cursor`/`response_metadata.next_cursor` convention, driven entirely inside
the hook (ordinary declarative `cursor` pagination COULD express the token-following mechanics
alone, but not the per-page `ok:false` check that must run before any page's records/cursor are
trusted — so the whole read stays hook-side for one consistent code path, matching legacy's own
single `harvest` helper). `channel_messages` additionally requires `config.channel_id` (validated
inside the hook, matching legacy's own `errors.New("slack stream channel_messages requires config
channel_id")` guard) — a runtime-conditional-required field the declarative dialect has no
mechanism to express per-stream.

Record mapping is hook-side Go (matching legacy's `slackUserRecord`/`slackChannelRecord`/
`slackMessageRecord` field-for-field, including `profile.email`/`profile.display_name` and
`topic.value`/`purpose.value` nested-object unwrapping), not declarative `computed_fields`.
`streams.json`'s `records.path` is documented for schema/tooling consistency but is never actually
read by the declarative fallback path in this bundle (StreamHook always returns `handled=true`).

## Write actions & risks

None — Slack is read-only here. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`slack.go:93-95`).

## Known limits

- **No incremental sync mode**: legacy has no reliable incremental cursor across the Web API list
  methods (`slackStreams`' doc comment, streams.go:29) — full_refresh only, matching legacy exactly.
- **`base.check`'s declarative block (`GET auth.test`) is a documentation/fallback shape only**: the
  bundle registers a `CheckHook` that actually inspects the response body's `ok` field (Slack's
  primary error-signaling mechanism) before Check can report success, exactly like `ReadStream`'s
  per-page check — a bare declarative check would incorrectly treat an `ok:false`/HTTP-200 auth
  failure as a passing check.
- **`TestConformance/slack`'s stream-level dynamic (fixture-replay) checks carry `skip_dynamic`
  markers on all 3 streams** (not a bundle-level marker — `base.auth`'s declarative bearer candidate
  is tried and satisfied by conformance's synthetic secret exactly like monday/github, so
  `check_fixture` and every other non-stream-scoped dynamic check still runs normally): each
  stream's real read is `StreamHook`-handled with an `ok:false`-in-HTTP-200 check the declarative
  replay path cannot independently drive. `internal/connectors/paritytest/slack` (live
  `httptest.Server` tests proving cursor pagination, the `ok:false` error surfacing, and
  `channel_messages`' `channel_id` requirement) and `internal/connectors/hooks/slack/hooks_test.go`
  are the authoritative correctness bar these markers name.
- **`client_secret`/the one-time `oauth.v2.access` authorization-code exchange is out of scope**,
  identically to gmail/quickbooks' 3-legged OAuth dance — see Auth setup above.
