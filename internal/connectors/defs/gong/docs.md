# Overview

Gong is a read-only declarative-HTTP connector migrated from `internal/connectors/gong` (the
hand-written legacy connector, which remains registered and unchanged until wave6's registry
flip). It reads Gong users, calls, and scorecard definitions through the Gong REST API
(`https://api.gong.io/v2`).

## Auth setup

Provide a Gong-generated access key via the `access_key` secret and its paired access key secret
via the `access_key_secret` secret. Both flow only into HTTP Basic auth
(`Authorization: Basic base64(access_key:access_key_secret)`) and are never logged. Gong also
offers OAuth2, but legacy only implements Basic auth with a generated access key pair, so this
bundle matches that exact scope.

## Streams notes

- `users` (`GET /users`, records at `users`), `calls` (`GET /calls`, records at `calls`), and
  `scorecards` (`GET /settings/scorecards`, records at `scorecards`) all share the same Gong
  cursor-pagination shape: the next-page token is read from the response body at `records.cursor`
  and sent back as the `cursor` query parameter (`pagination.type: cursor` with
  `token_path: records.cursor`, no `stop_path` declared). This matches legacy's `harvest` exactly:
  legacy stops when `records.cursor` is absent or empty, which is precisely the engine's default
  stop-on-empty-token behavior for the `token_path` cursor variant — Gong's own listing responses
  omit the `records` envelope entirely on the final page, and a missing/absent path resolves to an
  empty string via `StringAt`, so no separate `stop_path` boolean is needed.
- Every stream sends `limit={{ config.page_size }}` (legacy's `gongPageSize`, default 100, capped
  at Gong's own 100-per-page maximum).
- All three streams apply Gong's `fromDateTime` lower-bound filter identically to legacy's
  `incrementalLowerBound`: the filter is populated from the sync's persisted cursor if present,
  else from the RFC3339 `start_date` config value, and is omitted entirely on a from-scratch sync
  with no `start_date` set (`incremental.request_param: fromDateTime` + `start_config_key:
  start_date`, sent only when the lower bound resolves — see `buildInitialQuery` in
  `internal/connectors/engine/read.go`). Legacy's Catalog metadata inconsistently left `users` and
  `scorecards` with a nil `CursorFields` despite applying the identical `fromDateTime` filter
  read-side; this bundle declares `incremental.cursor_field` (`created` for `users`, `started` for
  `calls`, `updated` for `scorecards`) uniformly across all three so the same request-time filter
  behavior is available through the standard incremental sync path rather than only being reachable
  via a manually-injected `start_date`. This only ADDS the `incremental_append` sync-mode
  capability for `users`/`scorecards` on top of legacy's full-refresh-only catalog entry; it never
  changes what data is emitted for any read legacy itself would perform (the same `fromDateTime`
  filter value flows to the same query parameter either way).
- Field renames follow Gong's camelCase wire shape exactly, via `computed_fields`:
  `email_address`/`first_name`/`last_name`/`phone_number`/`manager_id` (users),
  `is_private` (calls), `scorecard_id`/`scorecard_name`/`workspace_id` (scorecards). Every other
  field name matches the raw API key verbatim and needs no rename.

## Write actions & risks

None. Gong is a read-only source in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`).

## Known limits

- Only the 3 legacy-parity streams (`users`, `calls`, `scorecards`) are implemented; the broader
  Gong surface (call transcripts, extensive call stats, interaction/activity trackers, workspaces,
  library folders, webhooks) is out of scope for this wave — see `api_surface.json`'s
  `api_surface.json` concrete exclusion entries.
  Gong's transcript and extensive-call-detail endpoints in particular require a
  request-time list of call IDs (a sub-resource fan-out shape, not a plain list endpoint) and were
  never implemented by legacy either.
- `calls`/`scorecards` fixtures ship a single real-wire-shape page (`fixtures/streams/{calls,
  scorecards}/page_1.json`); the required 2-page pagination-termination fixture lives on `users`
  (`fixtures/streams/users/{page_1,page_2}.json`), which is sufficient to exercise the shared
  base-level cursor paginator (`conformance`'s `pagination_terminates` check only needs one
  eligible stream).
- No client-side rate limiting is declared (`streams.json`'s `base.rate_limit` is absent) because
  legacy enforces none either; this bundle intentionally does not introduce new throttling behavior
  under the guise of migration.
