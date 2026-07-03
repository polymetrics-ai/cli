# Overview

Toggl is a wave2 fan-out declarative-HTTP migration. It reads time entries, projects, clients, and
users from the Toggl Track API v9 (`GET https://api.track.toggl.com/api/v9/...`). This bundle is
migrated at capability parity from `internal/connectors/toggl` (the hand-written connector it
replaces); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Toggl API token via the `api_token` secret. Toggl authenticates API-token requests as
HTTP Basic auth with the token as the username and the literal string `api_token` as the password
(`connsdk.Basic(token, "api_token")`, `toggl.go:122`) — this bundle reproduces that exact shape via
`base.auth`'s `{"mode":"basic","username":"{{ secrets.api_token }}","password":"api_token"}`, where
`password` is a static literal (no `{{ }}` markers), matching chargebee's identical
token-as-username Basic-auth precedent (`docs/migration/conventions.md`). `base_url` defaults to
`https://api.track.toggl.com/api/v9` and may be overridden for tests/proxies.

## Streams notes

All five streams declare `"projection": "passthrough"` (conventions.md §3/§5 defect class 1):
legacy's `Read` (`toggl.go:94-103`) does `emit(connectors.Record(rec))` on every record read from
the raw API response with no field-building/filtering step at all — `streamSpecs[...].fields` is
Catalog-only decoration (`toggl.go:125-137`, consumed solely by `Catalog()`'s `connectors.Stream`
construction), never applied to the emitted record itself. Default `"schema"` projection mode would
silently drop every real Toggl field not named in each stream's declared schema properties (e.g.
`time_entries`' real `tags`/`tag_ids`/`billable`/`at`/`server_deleted_at`/`duronly`/`created_with`,
`projects`' real `color`/`template`/`auto_estimates`/`is_private`/`billable`/`rate`/`currency`), an
undocumented silent data-shape change relative to legacy's raw passthrough. Each schema still
declares the real Toggl Track API v9 wire-shape properties it knows about (both the current
`workspace_id`/`client_id`/`user_id`-style names and the API's legacy-compat aliases
`wid`/`cid`/`uid`/`pid`/`tid` — Toggl's v9 API emits both on every record) for `x-primary-key`
typing and `records_match_schema` coverage, but passthrough mode means ANY other real field Toggl
adds or returns still survives unfiltered, matching legacy exactly.

`time_entries` reads `GET /me/time_entries` for the authenticated user, optionally filtered by
`start_date`/`end_date` config values sent only when configured (legacy: `toggl.go:82-89`,
conditional `q.Set` calls) — reproduced here via the opt-in `omit_when_absent` query dialect
(conventions.md §3).

`projects`, `clients`, and `workspace_users` are workspace-scoped
(`GET /workspaces/{{ config.workspace_id }}/{projects,clients,users}`), matching legacy's
`workspacePath` helper (`toggl.go:158-166`, `url.PathEscape(id)`) — path interpolation's per-segment
`urlencode` default (conventions.md §3) reproduces the identical escaping. `organization_users` is
organization-scoped (`GET /organizations/{{ config.organization_id }}/users`), matching legacy's
`organizationPath` helper (`toggl.go:167-175`). Neither `workspace_id` nor `organization_id` is
globally `required` in `spec.json` (only the streams that need them fail without one) — an absent
value hard-errors at path-interpolation time with an unresolved-`config.*`-key error, the same
failure classification legacy produces (`"toggl connector requires config workspace_id"` /
`"...organization_id"`) via a differently-worded message, per conventions.md §5's config-validation
parity precedent. None of the five streams expose an incremental cursor field in legacy, so all
five are always full-refresh reads. No stream is paginated — Toggl's `/me/time_entries` and
workspace/organization list endpoints return their full result set in one response in legacy's own
implementation (`toggl.go` never follows a next-page link for any of them).

## Write actions & risks

None. Toggl's legacy connector is read-only (package doc: "implements a read-only native Go
connector for the Toggl Track API"); `capabilities.write` is `false` and this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a
  static `fixture: true` marker and a hardcoded `workspace_id: "fixture_workspace"` string onto two
  synthesized records per stream (`toggl.go:176-187`). None of these are part of the LIVE record
  shape (where `workspace_id` is a real integer, not the fixture-mode string literal); this
  bundle's schemas and fixtures target the live path only. The engine's own conformance/
  fixture-replay harness (`internal/connectors/conformance`) provides the credential-free test
  affordance this bundle needs, so no fixture-mode equivalent is needed here.
- **No pagination is modeled for any stream**, matching legacy exactly — none of the five Toggl
  Track v9 list endpoints legacy calls are paginated in the hand-written connector, so this bundle
  declares no `pagination` block anywhere and ships single-page fixtures for every stream.
