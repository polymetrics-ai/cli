# Overview

Captain Data is a Tier-1 declarative-HTTP migration of `internal/connectors/captain-data`
(legacy Go package `captaindata`). It reads Captain Data workspace, workflows, jobs, and job
results through the Captain Data v3 REST API, and (Pass B) writes a `launch_workflow` action
that triggers a new workflow run.

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
directly; Captain Data returns each as a bare top-level JSON array (or object, for `workspace`)
(`records.path: ""`), matching legacy's `connsdk.RecordsAt(resp.Body, "")`.

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
legacy's `hasNext != "true"` stop condition in `harvest` (`captain_data.go`'s cursor loop),
including legacy's defensive stop on an empty `paging.next` token (the engine's `tokenPathCursor`
stops whenever the token itself is absent/empty, independent of `stop_path`).

`job_results`'s `data` field is a raw nested JSON object in both legacy (`item["data"]`, an
`any`-typed map) and this bundle (`"data": {"type": ["object", "null"]}` — plain schema
projection copies the raw object value unmodified, no `computed_fields` rename needed since
Captain Data's wire field names already match legacy's output field names one-for-one).

No stream is incremental: legacy declares no `CursorFields` for any Captain Data stream (the
source supports full-refresh only), and no schema here declares `x-cursor-field`.

## Write actions & risks

**`launch_workflow`** (Pass B addition): `POST /workflows/{{ record.workflow_uid }}/launch`
(`path_fields: ["workflow_uid"]`) triggers a new run of an existing Captain Data workflow,
matching Captain Data's own documented launch endpoint. The record's `workflow_uid` is
path-only and excluded from the JSON body (the engine's default body construction: every
record field except those named in `path_fields`); the body carries whichever of `accounts`
(the connected third-party accounts/cookies the workflow's steps run as),
`accounts_rotation_enabled`, `parameters` (the workflow's own input parameters — e.g. a search
URL or a list of LinkedIn profile URLs, entirely workflow-defined), and `output_column` the
caller supplies, matching the request shape Captain Data's own quickstart/Zapier-integration
docs show. A successful launch returns a `job_uid` for the newly created job, retrievable via
the existing `jobs`/`job_results` read streams. This is read-only Captain Data's *only*
dialect-expressible mutation: every other documented write (job retry, webhook retry, user
management) is excluded in `api_surface.json` as an administrative/notification action, not a
workflow/record write, or (webhook creation) is documented by Captain Data itself as
unavailable via the API at all.

`capabilities.write` is now `true`; `metadata.json.risk.write` documents that a launch consumes
account credits and may perform real external side effects (scraping, enrichment, outreach)
entirely dependent on how the target workflow's own steps are configured — this connector has
no visibility into or control over what a given `workflow_uid` actually does.

## Known limits

- Only the 4 legacy-parity read streams are implemented (workspace, workflows, jobs,
  job_results) plus the one write action (`launch_workflow`); any broader Captain Data v3
  surface beyond these (webhooks, users/accounts, integrations, collections) is out of scope
  for this wave — see `api_surface.json`'s `excluded` entries.
- `max_pages` is not exposed as config: `PaginationSpec.MaxPages` is a static JSON int with no
  config-driven override anywhere in the engine (the same `page_size`/`max_pages`-is-dead-config
  shape documented in `auth0`'s and `searxng`'s goldens), so it is intentionally not declared in
  `spec.json` (F6, REVIEW.md) — a live `job_results` read is unbounded (matches legacy's own
  `max_pages` default of `0`/unlimited when unset).
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Captain Data, so none is added here either.
- Captain Data's official API documentation portal (`docs.captaindata.com`) and its interactive
  Apiary/API-reference pages return 403/404 to unauthenticated fetches during this Pass B review;
  the `launch_workflow` request/response shape and the excluded-endpoint list in
  `api_surface.json` are sourced from Captain Data's own published support-center articles and
  ecosystem quickstart guides (Zapier/n8n launch-workflow guides, the "How to Use the Captain
  Data API v3" concise guide, and the job-status/errors-and-retries articles), cross-checked
  against legacy's implemented request shape — not from a raw OpenAPI spec, since none is
  publicly published.
