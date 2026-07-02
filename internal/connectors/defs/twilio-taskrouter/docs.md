# Overview

Twilio TaskRouter is a wave2 fan-out declarative-HTTP migration. It reads Twilio TaskRouter
workers, tasks, activities, task queues, and workflows for a workspace through the TaskRouter REST
API (`GET https://taskrouter.twilio.com/v1/Workspaces/<workspace_sid>/<resource>`). This bundle is
capability-parity migrated from `internal/connectors/twilio-taskrouter` (the hand-written connector
it migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `account_sid` and `auth_token` secrets; they are sent as HTTP Basic auth
(`account_sid` as username, `auth_token` as password), matching legacy's
`connsdk.Basic(sid, token)` (`twilio_taskrouter.go:154`) and are never logged. `workspace_sid` is
required and is substituted into every stream's workspace-scoped path, matching legacy's
`workspaceResource` requirement.

## Streams notes

All 5 streams (`workers`, `tasks`, `activities`, `task_queues`, `workflows`) share the identical
shape: `GET /v1/Workspaces/<workspace_sid>/<Resource>`, records at the resource's snake_case
top-level key (e.g. `{"workers":[...]}`), primary key `["sid"]`. `workers` additionally carries
`activity_name`/`available`; `tasks` carries `assignment_status`/`workflow_sid`; the other three
share a bare `sid`/`friendly_name` shape (legacy's shared `namedRecord` mapper).

Pagination is `page_number` with **no page-number query parameter ever sent** (`page_param: ""`),
matching legacy's `harvest` loop exactly: legacy increments an internal `page` counter but never
puts it on the wire (`twilio_taskrouter.go:104-124` builds the request query from only
`PageSize`, never a page-index param) — Twilio's real TaskRouter list endpoints use
opaque cursor pagination (`Page`/`PageToken`/`AfterSid`) that legacy never wired up at all; it
relies purely on `PageSize` plus the short-page stop condition (`len(records) < pageSize`) and the
`max_pages` request-count cap to bound how many times the identical first-page request repeats.
This bundle reproduces that exact (if unusual) behavior rather than "fixing" it to real
Twilio-style cursor pagination, since the legacy behavior is the migration's parity contract.
`page_size: 50` matches legacy's `defaultPageSize`; `max_pages: 1` matches legacy's
`defaultMaxPages`, so by default exactly one (identical, never-advancing) request is issued per
stream — see Known limits for the per-request override this narrows.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly.

## Known limits

- **Per-request `max_pages`/`page_size` overrides are not modeled.** Legacy accepts `max_pages`
  (`"all"`/`"unlimited"` or a non-negative integer, `twilio_taskrouter.go:183-196`) and `page_size`
  (bounded `[1,1000]`, `twilio_taskrouter.go:96-99`) config values read per-request. The engine's
  `base.pagination` block is a static, bundle-load-time spec with no per-request config
  indirection for either field, so this bundle fixes `max_pages: 1`/`page_size: 50` — matching
  legacy's own defaults exactly, but not legacy's opt-in override to raise either bound at read
  time. Declaring either as a `spec.json` property would be dead config (no template consumes
  it), so neither is declared (F6, `docs/migration/conventions.md`).
- **No real next-page cursor is modeled**, because legacy itself never implemented one (see
  Streams notes) — this is a faithful migration of legacy's own limitation, not a narrowing
  introduced by this bundle. A caller who needs TaskRouter's real full result set behind
  `max_pages: 1` and a never-advancing page is equally limited on both legacy and this bundle.
- Fixture pagination is single-page (`max_pages: 1` matches legacy's default, and no page-number
  parameter is ever sent regardless, so there is no second, distinctly-addressable page for a
  fixture to represent); `pagination_terminates` exercises the real short-circuit (exactly 1
  request made per stream).
