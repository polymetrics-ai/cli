# Overview

Tempo reads accounts, customers, worklogs, and workload schemes through the Tempo Cloud REST API v4
(`https://api.tempo.io/4`). This bundle migrates `internal/connectors/tempo` (the hand-written
legacy connector) at capability parity; the legacy package stays registered and unchanged until
wave6's registry flip. Tempo is read-only here — legacy has no write surface, so
`capabilities.write` is `false` and no `writes.json` is shipped.

## Auth setup

Provide a Tempo API token via the `api_token` secret. It is used only for
`Authorization: Bearer <api_token>` and is never logged, matching legacy's `connsdk.Bearer(secret)`
exactly.

## Streams notes

All 4 streams (`accounts`, `customers`, `worklogs`, `workload_schemes`) share Tempo v4's
`metadata.next` pagination convention: every list response carries its records under `results` plus
a `metadata` object whose `next` field is an absolute URL for the following page, absent when
exhausted (`pagination.type: next_url`, `next_url_path: metadata.next`) — identical to legacy's
`harvest` loop, which follows `metadata.next` verbatim until it is empty. The initial request on
every stream sends `limit=50` (default `page_size`) and `offset=0`, matching legacy's
`tempoDefaultPageSize` default and initial query construction exactly; every subsequent page is
requested by following the absolute `metadata.next` URL, which already carries its own limit/offset.

`worklogs`' schema uses `x-primary-key: tempo_worklog_id` (matching legacy's
`PrimaryKey: []string{"tempo_worklog_id"}`) and declares `computed_fields` renaming the raw camelCase
API fields (`tempoWorklogId`, `jiraWorklogId`, `timeSpentSeconds`, `billableSeconds`, `startDate`,
`startTime`, `createdAt`, `updatedAt`) to legacy's snake_case output names, plus
`issue_id: "{{ record.issue.id }}"` reaching into the nested `issue.id` object field (legacy's
`nestedInt(item, "issue", "id")`) — a bare single-reference `computed_fields` template, so the
engine's typed extraction preserves the field's native JSON integer type rather than stringifying
it. `accounts` renames `monthlyBudget` to `monthly_budget`; `workload_schemes` renames
`defaultScheme` to `default_scheme`. `customers` needs no renames (its raw field names already
match the schema).

Legacy's published catalog declares `CursorFields: ["updated_at"]` for `worklogs` only (mirrored via
the schema's `x-cursor-field`), but legacy's `harvest` never derives a request filter or client-side
drop from a persisted cursor for ANY stream — every read re-emits the complete result set from
offset 0. This bundle reproduces that exactly: no stream declares an `incremental` block (declaring
one — even `client_filtered` — would silently drop records legacy would still emit on a repeat
sync).

The Tempo API resource path for `workload_schemes` is `/workload-schemes` (hyphenated, per Tempo's
own URL convention); the stream name itself is declared `workload_schemes` (snake_case) per this
migration's naming convention (`docs/migration/conventions.md` §2) — only the `path` value keeps the
hyphen.

## Write actions & risks

None. Tempo is read-only in legacy (`Capabilities.Write` is `false`); `Write` always returns
`connectors.ErrUnsupportedOperation`. No `writes.json` is shipped for this bundle.

## Known limits

- `next_url` pagination ships a single-page fixture per stream (the sanctioned exception,
  `docs/migration/conventions.md` §4): the next-page URL is the replay server's own runtime address,
  unknown to a static fixture file. `pagination_terminates` exercises the bundle's first stream
  (`accounts`) against its single fixture page and asserts exactly one request is served — this is
  sufficient to prove termination since the fixture carries no `metadata.next`.
- Full Tempo v4 surface (teams, plans, approvals, schedules, roles) is out of scope for this wave;
  see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Tempo, so none is added here (matching legacy's real, lack-of, throttling behavior).
