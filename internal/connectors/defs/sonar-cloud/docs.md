# Overview

SonarCloud is a wave2 fan-out migration from `internal/connectors/sonar-cloud` (the legacy
hand-written connector this bundle replaces at capability parity), expanded to full practical
surface coverage in Pass B. It reads SonarCloud issues, components, projects, security hotspots,
languages, metrics, rules, quality gates, measures, webhooks, and project analyses through the
SonarCloud Web API, and writes webhook lifecycle mutations, issue comment/assignment/tag/workflow
mutations, and project-tag mutations. The legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide a SonarCloud user token via the `user_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <user_token>`) and is never logged.

## Streams notes

All 11 streams (`issues`, `components`, `quality_gates`, `measures` — 4 legacy-parity — plus
`projects`, `hotspots`, `languages`, `metrics`, `rules`, `webhooks`, `project_analyses`, newly
added in Pass B) share legacy's single-page, non-paginated read shape: every SonarCloud list
action used here returns its full result set (paginated server-side only via `p`/`ps`, capped at
one page per `Read` call, matching legacy's own `readRecords`, which issues exactly one request
with no page-advance loop) — porting a loop where legacy has none would be a behavior change, not
a migration. `languages` takes no query parameters at all (SonarCloud's own supported-language
list is unpaginated and organization-independent).

Every paginated stream sends `p=1` (Sonar's page-number param, always fixed at page 1, matching
legacy) and `ps` (page size, `config.page_size`, defaulting to `100` exactly like legacy's
`defaultPageSize`) except `webhooks` (unpaginated by the live API itself — its `response_example`
returns a bare `webhooks` array with no `paging` envelope at all). `organization` is sent when
`config.organization` is set (`omit_when_absent`) on every stream that accepts it
(`projects`/`rules`/`webhooks` newly, alongside the 4 legacy streams); `quality_gates`/`languages`/
`metrics` do not accept an `organization` parameter per the live API and none is sent.
`component_keys` narrows results per-stream using each endpoint's own real parameter name: `q` on
`projects` (a name/key search filter, not a hard scope), `projectKey` on `hotspots`, and — as
before — `component` on `components`/`componentKeys` on `issues`/`measures`. `project_analyses`
requires a project key (SonarCloud's own API makes `project` mandatory on this action) and is
wired directly from `config.component_keys` as a plain (non-optional) template — this stream only
resolves when `component_keys` is configured, and hard-errors otherwise, matching the real API's
own required-parameter contract rather than silently omitting the scope. `start_date`/`end_date`
map to `createdAfter`/`createdBefore` on the 4 legacy streams (matching legacy's
`copyConfig(q, cfg, "start_date", "createdAfter")`/`"end_date"→"createdBefore"`) and to `from`/`to`
on `project_analyses` (that action's own real parameter names for the same bounding-window
concept).

Legacy has no incremental/state-cursor read mode (no persisted cursor is ever read or written) —
`start_date`/`end_date` are static per-read filters, not an `incremental` block, so no stream
(old or new) declares `incremental` or `x-cursor-field` here.

All 11 streams declare `"projection": "passthrough"`. Legacy's `Read` emits the raw API record
verbatim (`emit(connectors.Record(rec))`, `sonar_cloud.go:126`, inside `readRecords`) with no
field-building/filtering step for its own 4 streams; the 7 newly added streams follow the same
passthrough convention for consistency with every sibling stream in this bundle. Every real
SonarCloud field beyond each schema's declared properties (e.g. `issues`' `effort`/`debt`/`hash`/
`textRange`, `rules`' `htmlDesc`/`descriptionSections`, `hotspots`' `textRange`) survives to the
emitted record exactly as the live API would return it. Declaring the default `"schema"`
projection mode here would silently narrow every emitted record to the schema's declared
properties — so `passthrough` is required, matching conventions.md §8 rule 1.

## Write actions & risks

8 write actions, newly added in Pass B (SonarCloud was entirely read-only before this pass):

- `create_webhook` / `update_webhook` / `delete_webhook` — full webhook lifecycle
  (`POST /api/webhooks/{create,update,delete}`, `body_type: form`). Creating or updating a
  webhook causes SonarCloud to start/keep POSTing analysis-completion payloads to an
  externally-reachable URL the caller supplies — a real external side effect, external approval
  required. `delete_webhook` is a plain (non-idempotent-marked) delete: the live API's behavior
  for an unknown `webhook` key is not documented as a safe no-op 404, so no `delete.missing_ok_status`
  is declared.
- `add_issue_comment` — `POST /api/issues/add_comment`. Adds a permanent, user-visible comment to
  an issue; external approval required.
- `assign_issue` — `POST /api/issues/assign`. Assigns or unassigns (empty `assignee`) an issue;
  external approval required.
- `set_issue_tags` — `POST /api/issues/set_tags`. REPLACES an issue's entire tag set (not an
  additive merge — matching the live API's own `set_tags` semantics); external approval required.
- `do_issue_transition` — `POST /api/issues/do_transition`. Moves an issue through its workflow
  (`confirm`/`resolve`/`wontfix`/`falsepositive`/etc., the live API's own closed `transition` enum
  is reproduced verbatim in `record_schema`); some transitions require elevated per-project
  permissions on the live API that this bundle cannot itself verify ahead of the request — a
  transition the caller's token lacks rights for surfaces as the live API's own error response,
  not a local validation failure. External approval required.
- `set_project_tags` — `POST /api/project_tags/set`. REPLACES a project's entire tag set (not an
  additive merge, matching the live API); external approval required.

All 8 actions use `body_type: "form"` (SonarCloud's Web API POST actions are classic
`application/x-www-form-urlencoded` endpoints, matching the `stripe`/goldens' form-body pattern —
see conventions.md §3).

## Known limits

- **`GET /api/measures/search` (this bundle's pre-existing `measures` stream) is absent from
  SonarCloud's own machine-readable `api/webservices/list` catalog** (that catalog's `api/measures`
  service lists only `component`/`component_tree`/`search_history`) but is live-verified working:
  an underspecified request returns a real structured `400` validation error ("Project keys xor
  Branch ids must be provided"), a valid request returns `200 {"measures": []}`, and a genuinely
  unknown action returns `404` — this is a catalog omission, not a broken endpoint, and the stream
  is kept as-is. See `api_surface.json`'s scope note.
- **`api/project_tags/search` is not implemented as a stream**: it returns a bare JSON array of
  tag-name strings (`["official", "offshore", ...]`), not an array of objects.
  `connsdk.RecordsAt`'s array-of-objects extraction silently yields zero records for a
  scalar-string array (it only explodes elements that decode as a JSON object) — there is no
  declarative mechanism in this dialect to wrap each string into a syncable record. See
  `api_surface.json`'s `out_of_scope` entry for the precise mechanism.
- **`project_analyses`, and the deliberately-excluded `project_branches`/`project_pull_requests`,
  are all per-project-scoped resources with no organization-wide list endpoint.** This bundle
  wires `project_analyses` from the single `config.component_keys` value (a static, one-project
  scope, matching how `component_keys` is already used elsewhere in this bundle) rather than
  adding a `fan_out` block to iterate every project in an organization — `project_branches`/
  `project_pull_requests` are excluded entirely (`api_surface.json`) rather than half-wired the
  same way, as a breadth-vs-cost triage call for this pass; revisiting with `fan_out` (driven by
  the `projects` stream's own key list) is a natural follow-up if per-project branch/PR data is
  needed.
- **Bulk/admin/permission-management, quality-profile administration, and settings endpoints are
  intentionally out of scope** (`destructive_admin`/`requires_elevated_scope` in
  `api_surface.json`): they require organization- or instance-administrator rights well beyond a
  plain `user_token`'s typical grant, are irreversible (project/analysis/permission-template
  deletion), or are account/access-control administration rather than analysis data sync. See
  `api_surface.json` for the full per-endpoint reasoning.
- `page_size`'s legacy bound (1-500, `maxPageSize`) is not separately re-validated by this bundle
  (the engine has no per-config-value numeric-range validation primitive) — an out-of-range value
  is passed through to SonarCloud, which will itself reject or clamp it. This is a config-surface
  narrowing, not an emitted-record-data change.
- Legacy performs no request pagination beyond the single fixed `p=1` page — matching that
  behavior exactly means this bundle's fixtures ship a single page per stream (no 2-page fixture
  required; `pagination_terminates` is exercised against a stream elsewhere in this wave's sibling
  bundles that declare real pagination).
