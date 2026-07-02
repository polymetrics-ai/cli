# Overview

UserVoice is a customer-feedback platform. This bundle reads customer **suggestions** from the
UserVoice API (`GET {base_url}/api/v2/suggestions`, default `https://api.uservoice.com`).
Read-only; it migrates `internal/connectors/uservoice` (195 loc), which stays registered and
unchanged until wave6's registry flip.

## Auth setup

UserVoice authenticates via a single secret, `api_key`, sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(apiKey)` requester exactly
(`uservoice.go:111-121`).

## Streams notes

One stream: `suggestions`, `GET /api/v2/suggestions`, records extracted from the `suggestions`
array (`uservoice.go:92`), each record mapped field-for-field to `{id, title, state, created_at}`
(`uservoice.go:138-145`). No pagination — legacy issues exactly one request per read (no
page/cursor param is ever sent or consumed).

**`id` is force-cast to a string.** Legacy's `suggestionRecord` calls `stringValue(item["id"])`
(`uservoice.go:140,183-194`), which converts a raw JSON number (UserVoice's real wire shape for
`id`) into a Go string — this is a deliberate legacy behavior, not an accidental side effect, so it
is reproduced exactly rather than "corrected" to a native-typed passthrough. Because the engine's
typed-`computed_fields` extraction only applies to a bare, unfiltered `{{ record.id }}` reference
(which would instead preserve the raw numeric type — see `docs/migration/conventions.md` §3's
"Typed extraction" note), this bundle deliberately uses `{{ record.id | last_path_segment }}`: the
`last_path_segment` filter's own documented contract guarantees "a value with no `/` at all passes
through unchanged, never errors" (conventions.md §3), so applied to a slash-free numeric-turned-
string id it is a pure identity transform whose only effect is routing the value through
`Interpolate`'s stringify path instead of the bare-reference typed-extraction path. This reproduces
legacy's cast byte-for-byte without inventing a new filter or a hook. Schema declares `id` as
`"type": "string"` accordingly.

**Optional `start_date` query passthrough.** Legacy only sends `?start_date=<value>` when the
config value is present and non-empty (`uservoice.go:85-87`); an absent value sends no query param
at all. This is expressed via the `stream.Query` optional-query dialect
(`{"template": "{{ config.start_date }}", "omit_when_absent": true}`), NOT as a
`streams.json` `incremental` block: legacy never persists or advances a cursor across syncs — it
re-reads `start_date` from config verbatim on every single read, with no stateful "resume from
last seen" behavior. Declaring a real `incremental` block would introduce genuine new
state-tracking behavior (the app would begin persisting and replaying a cursor value legacy never
produces or consumes), which is out of scope for a parity migration. `x-cursor-field` is therefore
intentionally NOT declared on the schema, even though `created_at` is emitted as a field, matching
legacy's full-refresh-only functional behavior exactly.

## Write actions & risks

None. UserVoice's legacy connector is read-only: `Write` always returns
`connectors.ErrUnsupportedOperation` (`uservoice.go:107-109`); `capabilities.write` is `false` and
this bundle ships no `writes.json`.

## Known limits

- Full UserVoice API surface (forums, users, comments, suggestion write actions) is out of scope
  for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- **`Check` now dials the network; legacy's `Check` never did.** Legacy `Check`
  (`uservoice.go:43-57`) validates config/secret presence offline only. This bundle's
  `base.check` issues a real `GET /api/v2/suggestions`, matching the wave's general "fail loud, not
  fail silent" preference for `Check` — a deliberate, strictly-improving behavior change with zero
  record-data impact.
- The optional `start_date` filter is a stateless, config-only passthrough (see Streams notes) —
  not a true incremental sync. A future capability-expansion pass could add real incremental
  support if UserVoice's suggestions endpoint documents a stable, monotonic cursor field.
- Legacy's `mode=fixture` config value (a testing affordance that short-circuits network access
  and emits one synthetic record) is not part of this bundle; parity is instead proven against
  legacy's live read path via fixture-replay conformance and, where applicable, a live parity test.
