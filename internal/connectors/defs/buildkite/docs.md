# Overview

Buildkite is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the full practical
read/write surface. It reads Buildkite organizations, pipelines, builds, agents, teams, and
clusters through the Buildkite REST API v2 (`https://api.buildkite.com/v2/...`), and writes
pipeline/build/job/agent/team lifecycle mutations. This bundle originally migrated
`internal/connectors/buildkite` (the hand-written connector, read-only); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Buildkite API access token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`, matching legacy's `connsdk.Bearer(secret)`) and is never
logged. `base_url` defaults to `https://api.buildkite.com/v2` and may be overridden for
tests/proxies.

## Streams notes

`organizations` (`GET /organizations`) is top-level and needs no organization slug, matching
legacy's `scopeTopLevel`. `pipelines`, `builds`, and `agents` are organization-scoped
(`/organizations/{{ config.organization }}/...`, matching legacy's `scopeOrganization` +
`organizationSlug` validation) and require the `organization` config value; it is declared in
`spec.json` but intentionally NOT in `required[]` since the `organizations` stream never
references it — an absent `organization` hard-errors only when an org-scoped stream's path is
resolved, identical to legacy's own per-stream (not global) validation.

Every stream's list response is a top-level JSON array (`records.path: ""`, root), matching
legacy's "Buildkite list endpoints return a top-level JSON array" comment. Pagination follows
Buildkite's own RFC 5988 `Link: <url>; rel="next"` header convention (`pagination.type:
link_header`) — the byte-accurate parity choice, since legacy's own `connsdk.LinkHeaderPaginator`
IS Link-header following (unlike, e.g., this repo's `github` bundle, which deliberately chose
`page_number` because ITS legacy implementation used `page`/`per_page` query params instead of
Link headers despite GitHub's API also supporting them). Every request sends `per_page=100`
(matches legacy's `buildkiteDefaultPageSize`) via each stream's static query.

`builds` supports Buildkite's `created_from` incremental lower bound (legacy: `if stream ==
"builds" { base.Set("created_from", createdGTE) }`) — expressed via the opt-in optional-query
dialect referencing `{{ incremental.lower_bound }}` with `omit_when_absent: true`, so
`created_from` is sent ONLY when the incremental lower bound resolves (persisted cursor, or the
RFC3339 `start_date` config value on a fresh sync), exactly matching legacy's own conditional
branch. `organizations`, `pipelines`, and `agents` retain the legacy catalog's `created_at` cursor
metadata in their schemas but declare no `incremental` block, matching legacy request behavior
(the `created_from` param is attached to `builds` only; legacy's own comment: "for other streams
the param is ignored harmlessly by the API but we only attach it to builds").

**Pass B additions.** `teams` (`GET /organizations/{{ config.organization }}/teams`) and `clusters`
(`GET /organizations/{{ config.organization }}/clusters`) follow the identical org-scoped,
`link_header`-paginated, top-level-JSON-array shape as `pipelines`/`agents` — no incremental
cursor marker (neither endpoint accepts a server-side modified-since filter, and there is no legacy
catalog cursor for these Pass B-only streams).

## Write actions & risks

Seventeen write actions, none present in legacy (legacy shipped `capabilities.write: false`):

- **`create_pipeline`** / **`update_pipeline`** / **`archive_pipeline`** / **`unarchive_pipeline`**
  / **`delete_pipeline`** — pipeline lifecycle. `create_pipeline` requires `cluster_id` (Buildkite's
  own API marks it required regardless of whether the org uses multiple clusters); changing
  `configuration`/`repository` via `update_pipeline` affects every future build; `delete_pipeline`
  is irreversible.
- **`create_build`** / **`cancel_build`** / **`rebuild_build`** — build lifecycle. `create_build`
  and `rebuild_build` both run arbitrary pipeline-defined commands on real agent capacity the
  instant they're accepted — the highest-consequence writes in this bundle; `cancel_build`
  terminates any in-progress jobs immediately.
- **`create_annotation`** — posts an HTML/Markdown annotation onto a build's detail page (no
  delete/list write modeled; see Known limits' fan_out ENGINE_GAP below for why list/delete by
  uuid aren't reachable).
- **`retry_job`** / **`unblock_job`** — job control. Both require a `job_id` the caller already has
  in hand (e.g. from a build record's embedded `jobs[]` array read externally) since no stream
  enumerates jobs directly (Buildkite has no top-level jobs-list endpoint).
- **`stop_agent`** / **`pause_agent`** / **`resume_agent`** — agent lifecycle. `stop_agent` with
  `force: true` cancels whatever job the agent is currently processing.
- **`create_team`** / **`update_team`** / **`delete_team`** — team lifecycle. `delete_team` is
  irreversible and detaches every pipeline/member association the team held.

Every action's per-record `risk` string in `writes.json` is the authoritative, reviewable summary;
`metadata.json`'s `risk.write`/`risk.approval` roll these up for the connector as a whole.

## Known limits

- **The fixture-replay harness cannot exercise `link_header`'s real 2-page continuation.**
  `fixtures/streams/**` (this repo's fixture-replay JSON shape, `conformance/replay.go`'s
  `fixtureResponse`) has no field for declaring HTTP RESPONSE headers — only `status` and `body` —
  so a fixture page can never carry the `Link: <url>; rel="next"` header the real Buildkite API
  sends, and `pagination_terminates`/`records_match_schema` can only ever observe the paginator's
  natural single-page stop (no Link header present = no next page, exactly like a real
  last-page response). This is a structural fixture-format limitation affecting ANY `link_header`
  bundle in this repo, not a buildkite-specific shortcut — the same gap conventions.md §4
  documents for `next_url`'s single-page exception, but `link_header` has no declared harness
  exception of its own (yet) since no fixture-response-header field exists for either pagination
  type to exploit even if one wanted to. Every stream fixture here is therefore a single,
  representative page; the 2-page Link-header-following codepath itself
  (`internal/connectors/engine/paginate.go`'s `linkHeaderPaginator`) is exercised by the shared
  engine's own `paginate_test.go` coverage, not by this bundle's fixtures. Per hard-rule scope for
  this migration wave, no Go (hooks/paritytest) was authored to work around this — a future wave
  extending the fixture-replay format to carry response headers, or reusing github's `page_number`
  substitution IF Buildkite's real API accepted it (it is a legitimate Link-header-only API in
  production for pagination continuation, so that substitution would NOT be byte-accurate parity
  here), would close this gap properly.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (1-100,
  default 100) and `max_pages` as config-driven overrides read fresh on every `Read` call. This
  bundle sends the legacy default `per_page=100` statically, and pagination is bounded by the
  absence of a `Link: rel="next"` header, matching Buildkite's normal termination behavior. These
  keys are intentionally not declared in `spec.json` because no runtime config template consumes
  them.
- Legacy's fixture-mode-only fields (`connector`, `fixture`, `previous_cursor` static/echoed
  markers stamped only under `config.mode == "fixture"`) are not modeled; this bundle's schemas and
  parity target the live wire shape only, matching this repo's established convention for a legacy
  in-code fixture path now superseded by the engine's own conformance/fixture-replay harness.
- **`ENGINE_GAP`: annotations and artifacts cannot be read as streams because their path needs a
  two-level fan-out the dialect only supports one level of.** Both resources live at
  `/organizations/{org.slug}/pipelines/{pipeline.slug}/builds/{build.number}/annotations` (or
  `/artifacts`) — three path variables deep, where `pipeline.slug` and `build.number` both vary per
  record and neither is a fixed config value. `engine.FanOutSpec` (bundle.go) resolves EXACTLY ONE
  parent id list per stream declaration (`ids_from` is either one `config_key` or one preliminary
  `request`, substituted into ONE `path_var`/`query_param`) — there is no way to declare "first
  enumerate pipelines, then for each pipeline enumerate its builds, then for each build read
  annotations" in a single `fan_out` block; that would require chaining two independent fan-outs,
  which the dialect does not support. `create_annotation` IS modeled as a write action despite this,
  since a write only needs the caller to SUPPLY a `pipeline_slug`/`build_number` pair on the record
  (no read-side enumeration required) — only the READ (list) side and any write requiring a
  discovered id (delete-by-`annotation.uuid`) are blocked. Closing this properly needs either a
  chained/nested `fan_out` dialect addition or a dedicated multi-level `StreamHook` — out of scope
  for a Pass B connector-only expansion per this task's own instructions (no engine changes, no new
  hook packages).
