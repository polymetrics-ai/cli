# Overview

Todoist is a wave2 fan-out declarative-HTTP migration. It reads projects, sections, tasks, and
comments from the Todoist REST API v2 (`GET https://api.todoist.com/rest/v2/...`). This bundle is
migrated at capability parity from `internal/connectors/todoist` (the hand-written connector it
replaces); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Todoist personal API token via the `token` secret; it is sent as a Bearer token
(`Authorization: Bearer <token>`) and is never logged, matching legacy's `connsdk.Bearer(token)`
(`todoist.go:118`). Legacy's `firstSecret(cfg, "token", "bearer_token")` accepts either the
`token` OR `bearer_token` secret, `token` taking precedence when both are configured
(`todoist.go:114,178-190`) — this bundle reproduces that exact precedence with a two-candidate
`base.auth` list (`when`-gated on each secret's presence in declaration order), matching
conventions.md §3's dual-auth-ordering rule. `base_url` defaults to
`https://api.todoist.com/rest/v2` and may be overridden for tests/proxies.

## Streams notes

`projects`, `sections`, and `tasks` are simple, non-paginated list endpoints (`GET /projects`,
`/sections`, `/tasks`); Todoist's REST v2 API returns the full list in one response for personal
task lists (legacy never paginates any of these three, `todoist.go:126-131`), so no `pagination`
block is declared and records are extracted from the response root (`records.path: "."`), matching
legacy's `connsdk.RecordsAt(resp.Body, ".")` (`todoist.go:90`). None of the four streams expose an
incremental cursor field in legacy, so all four are always full-refresh reads.

`comments` optionally scopes to one project or task via `project_id`/`task_id` query parameters,
sent only when configured (legacy: `todoist.go:78-85`, `strings.TrimSpace` then conditionally
`q.Set`). This bundle reproduces the identical optional behavior via the opt-in `omit_when_absent`
query dialect (conventions.md §3) — both params are left off the request entirely when their
config keys are unset, exactly like legacy's conditional `q.Set` calls.

## Write actions & risks

None. Todoist's legacy connector is read-only (package doc: "implements a read-only native Go
connector for the Todoist REST API"); `capabilities.write` is `false` and this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a
  static `fixture: true` marker and synthesizes `content`/`name`/`id` fields onto two fixture
  records per stream (`todoist.go:149-160`). None of these are part of the LIVE record shape; this
  bundle's schemas and fixtures target the live path only. The engine's own conformance/
  fixture-replay harness (`internal/connectors/conformance`) provides the credential-free test
  affordance this bundle needs, so no fixture-mode equivalent is needed here.
- **No pagination is modeled for any stream**, matching legacy exactly — none of the four Todoist
  REST v2 list endpoints legacy calls are paginated in the hand-written connector, so this bundle
  declares no `pagination` block anywhere and ships single-page fixtures for every stream.
