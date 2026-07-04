# Overview

UserVoice is a customer-feedback platform. The original migration read customer **suggestions**
from the UserVoice API (`GET {base_url}/api/v2/suggestions`, default `https://api.uservoice.com`),
ported from `internal/connectors/uservoice` (195 loc) at capability parity. This revision is a
Pass B full-surface expansion against UserVoice's real, currently-published Admin API v2 OpenAPI
spec (`uservoice.uservoice.com/api/v2/public/doc/api-v2-reference`, rendered at
`developer.uservoice.com/docs/api/v2/reference/`): 8 new read streams (`forums`, `users`,
`categories`, `statuses`, `labels`, `comments`, `notes`, `teams`) and 8 new write actions covering
the suggestion/comment/label/note lifecycle. The legacy package stays registered and unchanged
until wave6's registry flip.

## Auth setup

UserVoice authenticates via a single secret, `api_key`, sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(apiKey)` requester exactly
(`uservoice.go:111-121`) — this also matches the real API's own documented auth shape (the
Admin API v2 getting-started guide's curl examples use `Authorization: Bearer <token>` verbatim),
so unlike uppromote, no auth-shape mismatch applies here: every Pass B addition below uses the
identical `bearer` auth this bundle already declares. `api_key` is a long-lived
OAuth2-client-credentials-issued access token in UserVoice's own model, but this bundle (matching
legacy) treats it as a plain opaque Bearer secret — it does not perform the token exchange itself.

## Streams notes

`suggestions` (`GET /api/v2/suggestions`, no `/admin/` path segment — legacy's own shape) is the
original parity stream: records extracted from the `suggestions` array (`uservoice.go:92`), each
record mapped field-for-field to `{id, title, state, created_at}` (`uservoice.go:138-145`). No
pagination — legacy issues exactly one request per read (no page/cursor param is ever sent or
consumed).

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
produces or consumes), which is out of scope for a parity migration. `x-cursor-field: created_at`
is declared for catalog parity because legacy published `CursorFields: []string{"created_at"}`,
but the stream has no `incremental` block and remains full-refresh at read time. This stream
applies only to `suggestions` itself; the 8 Pass B streams below have real incremental support
instead (see next).

**Pass B additions** (new in this revision, all real `/api/v2/admin/...` paths per the OpenAPI
reference, none constrained by legacy parity since legacy never modeled them): `forums`, `users`,
`categories`, `statuses`, `labels`, `comments`, `notes`, `teams` — each a straightforward
`page_number` list endpoint (`page`/`per_page`, default page size 30 matching the real API's own
documented default) with records at the same-named top-level plural key (`{"forums": [...],
"pagination": {...}}`, etc. — the real API's uniform side-loadable-collection envelope). Every
stream except `teams` (which the OpenAPI spec does not document an `updated_after` filter for)
declares a genuine `incremental` block (`cursor_field: updated_at`, `request_param: updated_after`,
`param_format: rfc3339`) — the real API documents `updated_after`/`updated_before` as a server-side
filter on every one of these list endpoints, so unlike `suggestions`' stateless `start_date` (a
legacy-parity constraint), these brand-new streams get genuine incremental sync support with no
accepted-input-behavior conflict to avoid (there is no legacy behavior to preserve for a stream
legacy never had).

## Write actions & risks

`capabilities.write` flips to `true` in this revision (legacy shipped none — `Write` always
returned `connectors.ErrUnsupportedOperation`, `uservoice.go:107-109`). 8 actions, all newly added:

- `create_suggestion` (`POST /api/v2/admin/suggestions`) — creates a new suggestion; body requires
  a nested `links: {forum: <id>}` object per the real API's schema (not a flat `forum_id` field);
  low-risk, no approval required.
- `update_suggestion` (`PUT /api/v2/admin/suggestions/{id}`) — updates title/body; no approval
  required.
- `approve_suggestion` (`PUT /api/v2/admin/suggestions/{id}/approve`) — publishes a pending
  suggestion; no approval required.
- `delete_suggestion` (`PUT /api/v2/admin/suggestions/{id}/delete`) — UserVoice's own delete is a
  **soft** moderation action (a matching `restore` endpoint exists, not modeled here), so this is
  reversible, not permanent data loss; still `kind: delete` with `missing_ok_status: [404]`
  (idempotent) for engine bookkeeping consistency, no approval required.
- `create_comment` (`POST /api/v2/admin/comments`) — posts a comment on a suggestion; body requires
  a nested `links: {suggestion: <id>}` object; low-risk, no approval required.
- `create_label` (`POST /api/v2/admin/labels`) / `update_label` (`PUT /api/v2/admin/labels/{id}`) —
  label lifecycle (create/rename); no approval required.
- `create_note` (`POST /api/v2/admin/notes`) — creates an internal (non-public) note on a
  suggestion; body requires a nested `links: {suggestion: <id>}` object; low-risk, no approval
  required.

None of the 8 actions require approval: every one is a create, field-update, or soft/reversible
moderation action, not an irreversible destructive delete.

## Known limits

- **Every `create_*` write's body nests a `links` object** (`{forum: <id>}` for suggestions,
  `{suggestion: <id>}` for comments/notes) rather than a flat foreign-key field, because that is the
  real API's own documented request shape (`CreateSuggestionsBody`/`CreateCommentsBody`/
  `CreateNotesBody` all declare `links` as a required nested object in the OpenAPI spec) — the
  engine's default JSON body construction (`buildJSONBody`) copies every record field verbatim
  including nested map values, so a caller must supply `record.links` as a nested object, not a
  flattened field; this is documented here since it is the one place this bundle's write shape
  looks unusual relative to a typical flat-body write action.
- **`Check` now dials the network; legacy's `Check` never did.** Legacy `Check`
  (`uservoice.go:43-57`) validates config/secret presence offline only. This bundle's
  `base.check` issues a real `GET /api/v2/suggestions`, matching the wave's general "fail loud, not
  fail silent" preference for `Check` — a deliberate, strictly-improving behavior change with zero
  record-data impact.
- **`suggestions`' `start_date` filter remains a stateless, config-only passthrough** (legacy
  parity, see Streams notes) — not a true incremental sync, unlike the 8 Pass B streams which do
  have real `incremental` blocks. A future capability-expansion pass could add real incremental
  support to `suggestions` itself if ever judged worth the accepted-input-behavior change.
- **`teams` has no `incremental` block** — the real API's `/admin/teams` endpoint accepts
  `page`/`per_page` but its OpenAPI parameter list does not document an `updated_after`/
  `updated_before` filter the way every other Pass B stream's endpoint does, so no incremental
  block is declared for it (declaring one against an unconfirmed filter would silently no-op or
  error against the real API).
- Legacy's `mode=fixture` config value (a testing affordance that short-circuits network access
  and emits one synthetic record) is not part of this bundle; parity is instead proven against
  legacy's live read path via fixture-replay conformance and, where applicable, a live parity test.
- Full UserVoice API surface beyond what is enumerated in `api_surface.json` (185 endpoint+method
  combinations reviewed) is out of scope — account-administration (permissions/team-assignment/
  subdomains/user-block/GDPR export), UI/presentation configuration (themes/views/
  translatable_strings), analytics/reporting (impact_reports/segments/importance_scores/
  suggestion_activity_entries), integration bookkeeping (external_accounts/external_users/
  feedback_connector_configurations), bulk-only operations with no single-record shape (import/
  batch/bulk_delete/merge/link), and single-object detail endpoints duplicating an already-covered
  list stream's record shape — see `api_surface.json`'s `excluded` entries for the specific reason
  on each.
