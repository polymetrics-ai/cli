# Overview

Captain Data is a Tier-1 declarative-HTTP migration of `internal/connectors/captain-data`
(legacy Go package `captaindata`). It reads Captain Data workspace, workflows, jobs, and job
results through the Captain Data v3 REST API. The connector is read-only / full-refresh: legacy
exposes no reverse-ETL writes for Captain Data.

## Auth setup

Provide a Captain Data API key via the `api_key` secret; it is sent as the `X-API-Key` header
(`streams.json` `base.auth`'s `api_key_header` mode, no prefix), matching legacy's
`connsdk.APIKeyHeader(captainDataAPIKeyHeader, secret, "")`. Never logged.

`project_uid` is a **required** config value sent as the `X-Project-Id` header on every request
(`streams.json` `base.headers`), matching legacy's mandatory project scoping
(`captainDataProjectHeader`) — legacy's own `Check`/`requester` hard-error when it is unset, and
so does this bundle (a required-but-absent header reference is always a hard error, never
silently omitted, per `docs/migration/conventions.md` §3's header decision table).

## Streams notes

`workspace` (`GET /workspace`) and `workflows` (`GET /workflows`) are top-level collections read
directly; Captain Data returns each as a bare top-level JSON array (`records.path: ""`), matching
legacy's `connsdk.RecordsAt(resp.Body, "")`.

`jobs` and `job_results` are scoped by a parent uid supplied through config, exactly like legacy's
`resolvePath`: `jobs`'s path is templated as `/workflows/{{ config.workflow_uid }}/jobs`, and
`job_results`'s path is templated as `/jobs/{{ config.job_uid }}/results`
(`InterpolatePath`, urlencoded by default). Neither `workflow_uid` nor `job_uid` is declared in
`spec.json`'s top-level `required[]` (they only matter for their respective scoped stream, not
every read), but an absent referenced config key in a stream `path` is always a hard error in
ordinary `Interpolate`/`InterpolatePath` resolution — reproducing legacy's own
`"captain-data stream requires config %s"` hard-error exactly, just surfaced as the engine's
generic unresolved-config-key error instead of a connector-specific message.

`job_results` is Captain Data's only paginated stream: it returns
`{results:[...], paging:{next, have_next_page}}`, matching `pagination.type: cursor` with
`cursor_param: cursor`, `token_path: paging.next`, and `stop_path: paging.have_next_page` —
pagination continues only while `paging.have_next_page` is the literal string `"true"`
(`stop_path`'s falsy-stops rule, `docs/migration/conventions.md` §3), exactly reproducing
legacy's `hasNext != "true"` stop condition in `harvest` (`captain_data.go:192-197`), including
legacy's defensive stop on an empty `paging.next` token (the engine's `tokenPathCursor` stops
whenever the token itself is absent/empty, independent of `stop_path`).

`job_results`'s `data` field is a raw nested JSON object in both legacy (`item["data"]`, an
`any`-typed map) and this bundle (`"data": {"type": ["object", "null"]}` — plain schema
projection copies the raw object value unmodified, no `computed_fields` rename needed since
Captain Data's wire field names already match legacy's output field names one-for-one).

No stream is incremental: legacy declares no `CursorFields` for any Captain Data stream (the
source supports full-refresh only), and no schema here declares `x-cursor-field`.

## Write actions & risks

None. Captain Data is read-only in legacy (no approved reverse-ETL action set exists);
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- Only the 4 legacy-parity streams are implemented; any broader Captain Data v3 surface beyond
  workspace/workflows/jobs/job_results is out of scope for this wave.
- `max_pages` is not exposed as config: `PaginationSpec.MaxPages` is a static JSON int with no
  config-driven override anywhere in the engine (the same `page_size`/`max_pages`-is-dead-config
  shape documented in `auth0`'s and `searxng`'s goldens), so it is intentionally not declared in
  `spec.json` (F6, REVIEW.md) — a live `job_results` read is unbounded (matches legacy's own
  `max_pages` default of `0`/unlimited when unset).
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Captain Data, so none is added here either.
