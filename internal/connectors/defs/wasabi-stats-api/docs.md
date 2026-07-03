# Overview

Wasabi Stats API is a Tier-2 (AuthHook + RecordHook) migration, quarantined in wave1 under
`AUTH_COMPLEX` (`docs/migration/quarantine.json`): legacy's `requester()` branches auth mode on the
**contents** of the `api_key` secret at runtime — `connsdk.APIKeyHeader("Authorization", key,
"Bearer ")` is used unless `strings.SplitN(key, ":", 2)` yields exactly two parts, in which case it
switches to `connsdk.Basic(parts[0], parts[1])`. The engine's `when` grammar (equality/membership/
truthiness over a whole resolved value) has no string-split/substring primitive to branch on the
*shape* of a secret's value, so this cannot be expressed as a declarative `when`-gated dual-auth
candidate list (contrast zendesk-support's dual-auth precedent, conventions.md §3, which branches on
which of two DIFFERENT secrets is set — not on parsing one secret's contents). This bundle resolves
the blocker via `hooks/wasabi-stats-api/hooks.go`'s `AuthHook`, porting legacy's branch verbatim.
Read-only: legacy's `Write` always returns `ErrUnsupportedOperation`, and this bundle declares
`capabilities.write: false` with no `writes.json` to match.

## Auth setup

Provide one secret, `api_key`. `hooks/wasabi-stats-api/hooks.go`'s `AuthHook` inspects its resolved
value: if it splits into exactly two non-empty-separator parts on `:` (a `"username:password"`-shaped
value), it authenticates every request with HTTP Basic auth using those two parts; otherwise it sends
`Authorization: Bearer <api_key>` verbatim — matching legacy `wasabi_stats_api.go:138-141` exactly,
including legacy's use of `strings.SplitN(key, ":", 2)` (a value with *more* than one `:` still
splits into exactly 2 parts, the second part retaining any remaining colons; a value with zero `:`
falls through to the Bearer branch).

`streams.json`'s `base.auth` declares a single `mode: custom` candidate naming the hook, carrying
`value: "{{ secrets.api_key }}"` — `AuthSpec.Value` is repurposed here as the field that carries the
raw secret to branch on (API-CONTRACT.md's documented field-mapping convention, matching gmail's
reuse of `Token` for its refresh token): wasabi's `AuthSpec` has no dedicated field for "a secret
whose contents determine the auth mode," and `Value` is otherwise unused by the custom mode.

## Streams notes

Two streams, both primary-keyed on `id`, cursor-field `date`: `bucket_stats` (`GET v1/stats`) and
`account_stats` (`GET v1/accounts`), both un-paginated (`pagination: {"type": "none"}`, matching
legacy exactly — `streamEndpoints` declares no pagination anywhere) and both reading records from
the response body's `data` array (`records.path: "data"`).

Both streams send `start_date` as a query parameter **only when `config.start_date` is configured**
(`omit_when_absent: true` — the optional-query dialect, conventions.md §3), matching legacy's Read
path exactly (`wasabi_stats_api.go:100-102`: `if start := ...; start != "" { q.Set("start_date",
start) }`). This differs deliberately from `base.check`'s `start_date` handling (below).

**`hooks/wasabi-stats-api/hooks.go`'s `RecordHook` ports legacy's per-record `id`-fallback
derivation**, which no `computed_fields` template can express (a computed field is a single
template evaluated once per output field, not a multi-field conditional with a per-stream literal
fallback): legacy's Read (`wasabi_stats_api.go:115-117`) fills a record's `id` from the raw API
response when present; when the raw record has no `id` at all, it derives one as the first
non-empty value of `bucket`, then `date`, then the literal stream name (`"bucket_stats"` /
`"account_stats"`) — `account_stats` records have no `bucket` field at all, so in practice this
falls back to `date`, then the stream-name literal, for that stream. The hook reads `raw["bucket"]`/
`raw["date"]` (the PRE-projection record MapRecord receives) exactly as legacy reads `item["bucket"]`/
`item["date"]` before its own id-fallback check, and only sets `projected["id"]` when it is
genuinely unset — a record whose raw payload already has `id` is left untouched by the hook,
identical to legacy's `if item["id"] == nil` guard.

**No true incremental sync mode**: `x-cursor-field: date` is declared per legacy's own
`CursorFields: []string{"date"}` catalog entry (`streams()`, `wasabi_stats_api.go:166`), but no
stream declares an `incremental` block — legacy's own `start_date` is a fixed, non-advancing
per-connection filter (see above), never a persisted state cursor read back via `InitialState`, so
there is nothing for the engine's `incremental.request_param`/`cursor_field` machinery to drive.

## Write actions & risks

None — Wasabi Stats API is read-only. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`wasabi_stats_api.go:125-127`).

## Known limits

- **`AUTH_COMPLEX` resolution, not a new deviation.** The AuthHook's content-based Bearer-vs-Basic
  branch is the documented resolution of the recorded quarantine blocker
  (`docs/migration/quarantine.json`), ported verbatim from legacy's `requester()`.
  `metadata.json`'s `skip_dynamic` marker exists because conformance's synthetic secret value
  (`"synthetic-conformance-secret"`, which contains no `:`) can only ever exercise the Bearer
  branch — the Basic branch is proven solely by `hooks/wasabi-stats-api/hooks_test.go`'s
  `TestAuthenticator_ColonSeparatedKeyUsesBasicAuth`-style cases, mirroring gmail's identical
  custom-auth-only skip_dynamic precedent (conventions.md's conformance section: "a custom-auth
  `AuthHook` whose real request needs a config value... conformance's synthetic non-secret config...
  can never meaningfully populate").
- **`base.check`'s `start_date` query param is always present (even as an empty string) when
  `config.start_date` is unset**, deliberately DIFFERENT from the two streams' `omit_when_absent`
  behavior: legacy's `Check` (`wasabi_stats_api.go:66`) unconditionally builds
  `url.Values{"start_date": []string{strings.TrimSpace(cfg.Config["start_date"])}}` — always
  setting the key, never omitting it, even when the trimmed value is `""` — while legacy's `Read`
  (`wasabi_stats_api.go:100-102`) omits the param entirely when empty. This bundle reproduces both
  behaviors faithfully via two different `stream.Query` dialect shapes (`check.query` uses
  `"default": ""`, always present; the streams use `"omit_when_absent": true`) rather than
  collapsing them to one shared shape, since collapsing would change accepted-input behavior for
  whichever side lost its distinct handling.
