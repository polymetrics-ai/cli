# Overview

Jira is a wave2 fan-out declarative-HTTP migration. It reads Jira issues, projects, and users
through the Jira Cloud REST API v3 (`GET https://<site>.atlassian.net/rest/api/3/...`) using HTTP
Basic auth (account email + API token). This bundle migrates
`internal/connectors/jira` (the hand-written connector); the legacy package stays registered and
unchanged until wave6's registry flip. Read-only: Jira has no reverse-ETL write path.

## Auth setup

Provide an Atlassian account `email` (config) and a Jira `api_token` (secret); together they form
the Basic auth credential (`Authorization: Basic base64(email:api_token)`), matching legacy's
`connsdk.Basic(email, token)` (`jira.go:286`). The token is never logged.

`base_url` is **required** (e.g. `https://your-company.atlassian.net`). Legacy derives this from a
bare `domain` config value (stripping any scheme prefix and trailing slash) with `base_url` as an
optional full-URL override; this bundle requires the full `base_url` directly instead, since the
engine's `spec.json` `"default"` materialization mechanism only fills in a fixed literal default
and has no way to express "derive this URL from another config key" (conventions.md §3's
`default`-materialization note — the same narrowing documented on datadog's region-derived
`base_url` and jotform's `url_prefix`-derived one in this same wave). An operator who previously
set `domain=your-company.atlassian.net` now sets
`base_url=https://your-company.atlassian.net` instead — same reachable value space, different
config key name.

## Streams notes

All 3 streams share Jira's offset-envelope pagination shape
(`{startAt, maxResults, total, <records>[]}`): `pagination.type: offset_limit` with
`offset_param: startAt`, `limit_param: maxResults`. The engine's `offset_limit` paginator stops on
a short page (`recordCount < page_size`) exactly like legacy's own `count == 0 || count < pageSize`
check; legacy's *additional* early-stop on `startAt+count >= total` is not separately modeled (the
engine paginator has no `total`-envelope awareness), but this can only ever issue one extra,
zero-record trailing request when `total` is an exact multiple of the page size — never a data
difference, since that request's own short/empty page independently terminates pagination.
`streams.json`'s `pagination.page_size: 50` matches legacy's real default (`jiraDefaultPageSize`,
`jira.go:33`); `PaginationSpec.PageSize` is a plain JSON int with no config-driven override on
either side (`page_size`/`max_pages` are therefore not declared in `spec.json` at all — a
declared-but-unwireable key is worse than an absent one, F6).

- `issues` (`GET /rest/api/3/search`, records at `issues`): top-level `id`/`key`/`self` survive
  schema projection directly; the curated `fields.*` subset legacy's own `jiraIssueRecord` lifts to
  the record root (`summary`, `created`, `updated`, `status.name`, `issuetype.name`,
  `priority.name`, `assignee.displayName`, `reporter.displayName`, `project.key`) is expressed as
  `computed_fields` bare-path references (e.g. `"status": "{{ record.fields.status.name }}"`).
  Each is a single bare `record.<path>` reference, so the engine's typed-extraction rule applies
  (copies the raw value, a JSON string here, without stringify-wrapping) and, per the engine's
  documented absent-path-skip semantics, an absent/null nested object (`assignee: null` on an
  unassigned issue) is silently skipped for that field on that record rather than erroring —
  matching legacy's own nil-safe `nestedDisplayName`/`nestedName`/`nestedKey` helpers exactly.
  `x-cursor-field: updated` is declared (legacy: `jiraStreams()`'s `CursorFields: []string{"updated"}`)
  but no `incremental` block is declared: legacy never actually wires an incremental filter param
  into the request (its own `Read`/`harvest` always does a plain full page-walk), so this bundle
  matches that — cursor field is catalog metadata only on both sides.
- `projects` (`GET /rest/api/3/project/search`, records at `values`): flat field set, no renames
  needed.
- `users` (`GET /rest/api/3/users/search`, records at the response root — `records.path: "."`):
  the only stream whose envelope is a bare top-level array rather than the `{startAt,...}` shape;
  primary key is `accountId` (Jira users have no numeric/string `id` field), matching legacy.

## Write actions & risks

None. Jira is a read-only source connector (`capabilities.write: false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`domain`-based base URL derivation is not modeled** — see Auth setup above. `base_url` must be
  set to the full site URL directly.
- **`page_size`/`max_pages` are not exposed as config.** Legacy exposes both as config-driven
  overrides (`jiraPageSize`/`jiraMaxPages`, `jira.go:344-372`); the engine's `offset_limit`
  paginator reads `PaginationSpec.PageSize`/`MaxPages` as fixed values resolved once at bundle
  load, with no template/config-driven override mechanism. This bundle sends a fixed page size
  (`50`, matching legacy's own default `jiraDefaultPageSize`, `jira.go:33`) and does not cap
  `max_pages` (unbounded, matching legacy's own default of 0/unlimited).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (reached only
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps an
  extra `previous_cursor` field (echoing `req.State["cursor"]`) onto every fixture-mode record
  (`jira.go:256-259`). This is not part of the live record shape; this bundle's schemas and
  fixtures target the live path only. The engine's own conformance/fixture-replay harness
  provides the credential-free test affordance this bundle needs, so no fixture-mode equivalent is
  needed here.
