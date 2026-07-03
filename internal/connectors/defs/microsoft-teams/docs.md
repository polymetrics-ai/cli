# Overview

Microsoft Teams is a Tier-2 hooks migration of the quarantined `microsoft-teams` legacy connector
(`docs/migration/quarantine.json`'s `microsoft-teams` entry, `blocker_type: ENGINE_GAP`). It reads
Microsoft 365 / Entra ID users, groups, Teams channels, and device-usage reports through the
Microsoft Graph REST API (`v1.0`), read-only. This bundle is parity-tested against
`internal/connectors/microsoft-teams` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Authentication is an OAuth2 client-credentials grant against Microsoft Entra ID — this is fully
expressible by the engine's declarative `oauth2_client_credentials` auth mode, no `AuthHook`
needed. Provide `client_id` (config, matches legacy's `graphClientID` config-first resolution) plus
the `client_secret`/`tenant_id` secrets. The token endpoint is derived from `tenant_id`
(`https://login.microsoftonline.com/<tenant_id>/oauth2/v2.0/token`, matching legacy's
`graphTokenURL`) via a two-candidate `auth` list — the first candidate (an explicit `token_url`
override, used by tests) wins when set; the second derives the endpoint from `login_base_url` +
`tenant_id` (the sharepoint-lists-enterprise golden pattern, since the engine's spec-default
materialization cannot express a cross-key derived default). `scope` defaults to the static Graph
application-permission scope `https://graph.microsoft.com/.default`, matching legacy's
`graphDefaultScope`.

## Streams notes

Four streams: `users`, `groups`, `channels` (`/teams/getAllChannels`), and
`team_device_usage_report` (`/reports/getTeamsDeviceUsageUserDetail`, scoped by the `period` config
value: `D7`/`D30`/`D90`/`D180`, default `D7`). Every endpoint returns `{value:[...],
"@odata.nextLink":"<absolute-url-or-absent>"}`.

**Pagination — Tier-2 StreamHook, not declarative (matches the quarantine's ENGINE_GAP reason
verbatim)**: Microsoft Graph's `@odata.nextLink` cursor is an absolute URL carrying its own query
(`$skiptoken` etc.), read from a response-body key that contains a literal `.` (`@odata.nextLink`).
The engine's declarative `next_url` pagination type reads its cursor via `connsdk.StringAt`'s
dotted-path traversal (`engine/paginate.go`), which necessarily treats any `.` in a path as a
nesting separator — there is no way to address a literal dotted key with that parser. This is the
identical gap `docs/migration/quarantine.json` documents for `microsoft-entra-id` and
`microsoft-lists`. `hooks/microsoft-teams/hooks.go` implements `StreamHook`, porting legacy's
`harvest`/`nextLink` loop verbatim: `connsdk.RecordsAt(resp.Body, "value")` per page, followed by a
literal-key JSON decode of `@odata.nextLink` (never `connsdk.StringAt`), passing the absolute URL
back as the next request's path with no extra query (matching legacy's `path = next; query = nil`).

Every stream in this bundle carries an explicit `"conformance": {"skip_dynamic": true, "reason":
"..."}` marker (`docs/migration/conventions.md` §4/§6): `internal/connectors/conformance/dynamic.go`
honors this by Skipping every dynamic fixture-replay check for these streams, since the StreamHook
(always `handled=true`) is what every real `Read()` call actually dispatches through, and a
declarative-only fixture replay cannot exercise `@odata.nextLink` following at all. The
authoritative substitute this marker names is `paritytest/microsoft-teams`'s dedicated 2-page
`@odata.nextLink` test (`TestParityMicrosoftTeams_UsersNextLinkPagination`) and
`hooks/microsoft-teams/hooks_test.go`. `streams.json`'s own `base.pagination` stays declared
`{"type": "none"}` (a single, honest request) since it is never dynamically exercised now.

No incremental cursor is ever sent as a request parameter in legacy: `CursorFields` is `nil` on
every stream (`microsoft-teams/streams.go`'s `graphStreams`), and neither `Read` nor `harvest` ever
forwards a state cursor into a request. This bundle matches that exactly by declaring no
`incremental` block on any stream (full_refresh only).

## Write actions & risks

None. Microsoft Teams is read-only in legacy (`Write` returns `connectors.ErrUnsupportedOperation`,
`microsoft-teams.go:235-237`); `capabilities.write` is `false` and no `writes.json` is declared.

## Known limits

- Full Microsoft Graph surface (mail, calendar, drives, chat messages, planner, etc.) is out of
  scope for this migration; see `api_surface.json`'s `excluded` entries. Only the 4 legacy-parity
  read streams are implemented.
- Pagination is a Tier-2 `StreamHook` (`hooks/microsoft-teams/hooks.go`), not declarative
  `next_url` — see "Streams notes" above. Candidate future engine feature: a
  literal-dotted-key-aware body path accessor (or a dedicated Graph `@odata.nextLink` pagination
  type) would let this connector drop the hook; not implemented in this phase per the `ENGINE_GAP`
  recurrence rule (`conventions.md` §6) — tracked jointly with `microsoft-entra-id`/
  `microsoft-lists`, which hit the identical gap.
- `streams.json`'s declared `base.pagination: {"type": "none"}` is not the real production
  pagination behavior; every stream's `conformance.skip_dynamic` marker documents why (see
  "Streams notes").
- `max_pages` is consumed only by `hooks/microsoft-teams/hooks.go`'s `StreamHook` (mirroring
  legacy's `graphMaxPages`), not by a declarative `PaginationSpec.MaxPages` field, since pagination
  here is entirely hook-driven. Declared in `spec.json` so it is not dead config.
