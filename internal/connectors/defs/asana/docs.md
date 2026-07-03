# Overview

Asana is a pure declarative-HTTP Tier-1 migration of `internal/connectors/asana` (legacy is
connsdk-HTTP-based: a single Bearer-authenticated requester, plain JSON responses, no protocol-
native SDK, no hook-worthy auth/stream logic). It reads Asana workspaces, projects, and tasks
through the Asana v1 REST API (`GET https://app.asana.com/api/1.0/...`). Read-only — legacy's
`Write` always returns `connectors.ErrUnsupportedOperation` and this bundle ships no `writes.json`.

## Auth setup

Provide an Asana personal access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`asana.go:154`). `base_url` defaults to
`https://app.asana.com/api/1.0` and may be overridden for tests/proxies.

## Streams notes

All 3 streams (`workspaces`, `projects`, `tasks`) share the same shape: `GET` against the Asana
list endpoint, records at `data`, primary key `["gid"]`. Each stream sends its own
`opt_fields` value (matching legacy's per-endpoint `optFields`) and a static `limit=100`
(legacy's `asanaDefaultPageSize`/`asanaMaxPageSize`, both 100) — `page_size` is declared in
`spec.json` for documentation parity with legacy's config surface, but the `next_url` paginator
has no config-driven page-size override mechanism (it never reads a page-size field the way
`page_number`/`offset_limit` paginators do), so (matching bitly's identical, already-documented
limitation) the live request always sends Asana's own default rather than a runtime override.

`projects` and `tasks` optionally scope to a workspace via the `workspace` query param
(`{{ config.workspace_id }}`, `omit_when_absent: true` — left off entirely when unset, matching
legacy's `endpoint.resource != "workspaces"` guard: the `workspaces` stream never sends a
`workspace` param on itself). `tasks` additionally scopes to `project`/`assignee` via
`{{ config.project_id }}`/`{{ config.assignee }}`, both `omit_when_absent: true`, matching
legacy's `asanaQuery`'s `endpoint.resource == "tasks"` guards exactly.

Pagination follows Asana's `next_page.uri` convention (`pagination.type: next_url`,
`next_url_path: "next_page.uri"`) exactly like legacy's `harvest` loop, which follows
`resp.next_page.uri` verbatim as the next request path (with no query) until it is empty
(`asana.go:119-127`). None of the 3 streams declare an `incremental` block — legacy's Asana v1
endpoints expose no server-side "modified since" filter this connector's catalog wires up (no
`CursorFields` declared anywhere in legacy's `asanaStreams()`), so every read is full refresh,
matching legacy exactly.

## Write actions & risks

None. Legacy `asana.Write` always returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size` is not runtime-configurable.** Same limitation as bitly's `next_url`-paginated
  `bitlinks` stream (see `docs/migration/conventions.md` and bitly's own `docs.md`): the engine's
  `next_url` paginator has no analogous page-size knob, so `page_size`/`max_pages` (legacy's
  config-driven overrides, `asanaPageSize`/`asanaMaxPages`) are not wired into the live request —
  Asana's own default (`limit=100`) is sent unconditionally. `page_size` remains declared in
  `spec.json` for config-surface parity/documentation but is not consumed by any template.
- **Fixtures ship one page per stream (sanctioned `next_url` exception, conventions.md §4).** A
  `next_url` stream's next-page URL is the replay server's own address, unknown ahead of time to a
  static fixture file — a genuine harness limitation, not a fixture-authoring shortcut. All 3
  streams paginate identically in legacy; `pagination_terminates` exercises whichever stream
  `conformance` selects as its first eligible stream from this single-page shape (an empty
  `next_page` value stops pagination immediately, proving termination on an already-short page).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) synthesizes `gid`/`name`/`resource_type`/`created_at`/
  `modified_at`/`completed` values directly in Go and never talks to a real Asana API
  (`asana.go:132-143`); this bundle's own `conformance`/fixture-replay harness
  (`internal/connectors/conformance`) provides the equivalent credential-free test affordance, so
  no fixture-mode config branch is modeled here — matching SPEC's instruction to target the live
  record shape only.
