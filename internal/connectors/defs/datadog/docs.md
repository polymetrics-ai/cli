# Overview

Datadog is a wave2 fan-out migration. This bundle reads Datadog dashboards and scheduled downtimes
through the Datadog REST API, migrating 2 of the 5 streams in `internal/connectors/datadog` (the
legacy hand-written connector, which stays registered and unchanged until wave6's registry flip) at
capability parity. **`monitors`, `users`, and `slo` are NOT migrated in this wave — they are
blocked by an `ENGINE_GAP`, not deferred scope; see Known limits.**

## Auth setup

Provide `api_key` and `application_key` secrets; both are sent as raw header values —
`DD-API-KEY: <api_key>` and `DD-APPLICATION-KEY: <application_key>` — matching legacy's
`DefaultHeaders` map exactly (neither is a Bearer/Basic scheme, so `streams.json`'s `base.headers`
declares them directly rather than via an `auth` candidate; both secrets are required, so an absent
value is always a hard error per the engine's secrets-in-headers rule, matching legacy's own
explicit `errors.New(...)` checks). Neither value is ever logged.

## Streams notes

`dashboards` (`GET /api/v1/dashboard`, records at `dashboards`) and `downtimes`
(`GET /api/v1/downtime`, records at the response root `.`) are both unpaginated in legacy
(`pageStyle: pageNone` — a single call returns every record), so both use
`pagination.type: none` here, matching legacy's `harvest` loop returning immediately after the
first (only) page for these two endpoints. Field mapping is a direct schema projection for both —
legacy's `datadogDashboardRecord`/`datadogDowntimeRecord` copy fields straight off the raw item with
no renames, so no `computed_fields` are needed. `downtimes`' `id`/`monitor_id`/`start`/`end` are
Unix-seconds integers in Datadog's real wire shape (legacy stores them as `int64`); the schema
declares them `integer`, matching that native type. Neither stream is incremental in legacy (no
server-side cursor filter for either), so neither `streams.json` entry declares an `incremental`
block.

## Write actions & risks

None. This connector is read-only, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **Blocked: `monitors`, `users`, and `slo` streams (`ENGINE_GAP`).** All three are paginated in
  legacy starting at page **0** (`for page := 0; ...` in `datadog.go`'s `harvest`, sent as either
  `page=0` (`pageV1`: `monitors`/`slo`) or `page[number]=0` (`pageV2`: `users`) on the very first
  request) — Datadog's real v1/v2 list-endpoint convention is 0-based. The engine's `page_number`
  pagination type delegates to `connsdk.PageNumberPaginator`, whose `Start()` method
  (`p.page = p.StartPage; if p.page == 0 { p.page = 1 }`) unconditionally coerces an explicit
  `start_page: 0` to `1`, because a Go `int` zero value cannot be distinguished from "explicitly
  configured to 0" — confirmed directly against the paginator's actual `Start()` behavior, not
  merely inferred. Every live read of `monitors`/`users`/`slo` would therefore silently begin at
  Datadog's real SECOND page, permanently skipping the first page's records: an accepted-input
  EMITTED-DATA change (real records silently dropped), which fails the parity-deviation meta-rule
  (`docs/migration/conventions.md` §5) and cannot ship as a documented-acceptable deviation. This is
  the IDENTICAL gap class as this repo's `algolia` `indices` stream blocker (same root cause, same
  paginator, same coercion). No Tier-1 workaround exists (`PaginationSpec` fields are plain JSON,
  never templated — no config-driven escape hatch), and this wave's hard rules forbid a Tier-2 hook
  to patch around it. This is an `ENGINE_GAP` blocker for a follow-up engine-dialect increment
  (`PaginationSpec` needs a way to distinguish "start_page unset" from "start_page explicitly 0",
  e.g. a pointer/explicit-presence field or a documented sentinel), not a per-connector patch. Once
  closed, all three streams should follow legacy's exact shape: `monitors`/`slo` —
  `pagination.type: page_number`, `page_param: page`, `size_param: page_size`, `start_page: 0`,
  records at `.` (`monitors`) / `data` (`slo`); `users` — identical but `page_param: page[number]`,
  `size_param: page[size]`, records at `data`, with a `computed_fields` lift out of the JSON:API
  `attributes` wrapper (`name`/`email`/`handle`/`status`/`disabled`/`verified`/`created_at`),
  matching legacy's `datadogUserRecord`/`mapField` helper.
- A `site`-derived `base_url` (legacy's `datadogBaseURL` builds `https://api.<site>` from a `site`
  config value, e.g. `datadoghq.eu`, when `base_url` itself is unset) is not modeled: the engine's
  `spec.json` `"default"` mechanism only materializes a fixed literal, not one derived from another
  config field (conventions.md §3, the sentry/chargebee derived-default case). This is a
  documented, accepted config-surface narrowing — set `base_url` directly to the regional host
  (e.g. `https://api.datadoghq.eu`) instead of a bare `site` value.
- Full Datadog API surface (metrics, logs, APM traces, incidents, security signals, synthetics,
  monitor/downtime mutation) is out of scope for wave2; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
