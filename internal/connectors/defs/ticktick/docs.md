# Overview

TickTick is a wave2 fan-out declarative-HTTP migration. It reads TickTick projects and
project-scoped tasks through the TickTick Open API (`GET https://api.ticktick.com/open/v1/...`).
This bundle is capability-parity migrated from `internal/connectors/ticktick` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

TickTick issues an OAuth access token. Legacy accepts it under any of 3 aliased secret keys, in a
first-non-empty-wins fallback chain (`firstSecret(cfg, "bearer_token", "client_access_token",
"access_token")`, `ticktick.go:109,178-190`): `bearer_token` first, `client_access_token` second,
`access_token` third. This bundle reproduces the exact same precedence as 3 ordered `bearer`
auth candidates, each gated by a `when` clause on its own secret's truthiness — `base.auth`'s
first-match-wins `selectAuth` evaluation (conventions.md §3, "Dual-auth ordering is
load-bearing") reproduces legacy's fallback chain exactly: `bearer_token` wins if set (regardless
of the other two), otherwise `client_access_token` wins if set, otherwise `access_token` is used.
Whichever one resolves is sent as `Authorization: Bearer <token>`; none is ever logged.
`base_url` defaults to `https://api.ticktick.com/open/v1`, matching legacy's `defaultBaseURL`
fallback.

## Streams notes

`projects` (`GET /project`, records at the JSON response root `.`) has no pagination — TickTick's
project-list endpoint returns every project in one call, matching legacy's single unpaginated
`Do` request (`ticktick.go:81`). `tasks` (`GET /project/{project_id}/data`, records at the
`tasks` envelope key) requires `project_id` (the TickTick project id to scope reads to), matching
legacy's own hard requirement (`ticktick.go:125-129`: "ticktick tasks stream requires config
project_id") — the engine's path-interpolation hard-errors identically when `project_id` is
unset, with no special-casing needed. Both streams declare primary key `["id"]`.

## Write actions & risks

None. TickTick is read-only (`capabilities.write: false`, no `writes.json`), matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`tasks` has no incremental/cursor support**, matching legacy exactly — legacy's
  `project/{id}/data` endpoint returns the full project snapshot (tasks + columns) on every call
  with no server-side filter parameter; this bundle declares no `incremental` block for either
  stream, an honest 1:1 match of legacy's own behavior, not a scope narrowing.
- **Only one `project_id` can be read per sync.** This matches legacy's own single-project-per-call
  design (`ticktick.go:124-130`) — TickTick's Open API has no "all tasks across all projects"
  endpoint; a caller wanting every project's tasks must run one sync per `project_id`, exactly as
  legacy required.
