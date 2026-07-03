# Overview

Datadog reads Datadog monitors, dashboards, users, SLOs, and scheduled downtimes through the
Datadog REST API — the full 5-stream legacy-parity surface of `internal/connectors/datadog` (the
legacy hand-written connector, which stays registered and unchanged until wave6's registry flip).
`monitors`/`users`/`slo` were originally blocked by an `ENGINE_GAP` (0-based `page_number`
pagination the engine could not express); that gap is now closed (S4 engine mini-wave item 1 —
`PaginationSpec.StartPage` is a `*int`, so an explicit `start_page: 0` is honored rather than
coerced to `1`) and all 5 streams are migrated at full capability parity in this pass.

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
declares them `integer`, matching that native type.

`monitors` (`GET /api/v1/monitor`, `pageV1`) and `slo` (`GET /api/v1/slo`, `pageV1`) both paginate
with `pagination.type: page_number`, `page_param: page`, `size_param: page_size`, `start_page: 0`
(Datadog's real v1 list convention: `page=0` is genuinely the first page — legacy's own
`harvest` loop is `for page := 0; ...`), `page_size: 100` as both the sent `page_size` query value
and the client-side short-page stop threshold, matching legacy's `datadogDefaultPageSize`.
`monitors`' records live at the response root (`records.path: "."`, a top-level JSON array);
`slo`'s live at `data`. `monitors` is the only Datadog stream legacy publishes a `CursorFields`
entry for (`modified`) — but legacy sends no server-side filter parameter for it (`harvest` never
varies the request by any stored cursor), so per conventions.md §8 rule 2 this bundle declares a
bare `incremental.cursor_field: modified` with no `request_param`, matching legacy's own
catalog-only (non-filtering) cursor publication exactly. `slo`/`dashboards`/`downtimes` publish no
`CursorFields` in legacy, so none of their schemas/streams declare an `x-cursor-field`/
`incremental` block.

`users` (`GET /api/v2/users`, `pageV2`) paginates with `page_param: page[number]`,
`size_param: page[size]`, `start_page: 0`, `page_size: 100` — Datadog's v2 JSON:API convention,
also genuinely 0-based (legacy's identical `for page := 0; ...` loop, just with the v2 query-key
shape). Records live at `data`, a JSON:API `{id, type, attributes: {...}}` envelope per user;
`computed_fields` lifts `name`/`email`/`handle`/`status`/`disabled`/`verified`/`created_at` out of
the nested `attributes` object onto the flat record (bare single-reference `computed_fields`
entries copy the raw typed value, so `disabled`/`verified` stay JSON booleans), matching legacy's
`datadogUserRecord`/`mapField` helper exactly. `id`/`type` are copied by ordinary schema projection
since they already live at the JSON:API envelope's top level.

All 4 paginated/unpaginated stream shapes share the identical `pagination`/`records` combination
legacy itself uses per endpoint — no stream in this bundle diverges from legacy's per-endpoint
pagination style.

## Write actions & risks

None. This connector is read-only, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- A `site`-derived `base_url` (legacy's `datadogBaseURL` builds `https://api.<site>` from a `site`
  config value, e.g. `datadoghq.eu`, when `base_url` itself is unset) is not modeled: the engine's
  `spec.json` `"default"` mechanism only materializes a fixed literal, not one derived from another
  config field (conventions.md §3, the sentry/chargebee derived-default case). This is a
  documented, accepted config-surface narrowing — set `base_url` directly to the regional host
  (e.g. `https://api.datadoghq.eu`) instead of a bare `site` value.
- Full Datadog API surface (metrics, logs, APM traces, incidents, security signals, synthetics,
  monitor/downtime mutation) is out of scope; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
