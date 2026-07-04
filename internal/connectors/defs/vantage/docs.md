# Overview

Vantage is a cloud cost visibility and FinOps governance platform. This bundle reads cost,
budget, folder, dashboard, business-metric, resource-report, recommendation, team,
saved-filter, workspace, virtual-tag-config, tag, cost-alert, budget-alert, anomaly-alert,
managed-account, financial-commitment, segment, report-notification, recommendation-view,
network-flow-report, Kubernetes-efficiency-report, anomaly-notification, canvas, invoice, and
integration data from the Vantage API v2 (`{base_url}`, default `https://api.vantage.sh`), and
writes create/update/delete mutations for budgets, folders, dashboards, cost reports, resource
reports, saved filters, workspaces, teams, cost alerts, budget alerts, anomaly alerts, business
metrics, virtual tag configs, segments, report notifications, recommendation views, network flow
reports, Kubernetes efficiency reports, anomaly notifications, and canvases. It originally
migrated `internal/connectors/vantage` (185 loc, a read-only single-stream connector), which
stays registered and unchanged until wave6's registry flip; this Pass B pass expands the bundle to
Vantage's full documented API surface (`docs.vantage.sh/llms-full.txt`, 193 real endpoints across
every `/api/*` doc page, reviewed 2026-07-04) with every endpoint covered or excluded under a
specific, real reason — no blanket bucket.

## Auth setup

Vantage authenticates via a single secret, `access_token`, sent as a Bearer token
(`Authorization: Bearer <access_token>`), matching legacy's `connsdk.Bearer(token)` requester
exactly (`vantage.go:107-117`) and Vantage's own documented Bearer-token scheme
(`docs.vantage.sh/api/authentication`).

## Streams notes

`costs` is the original legacy-parity stream: `GET /v2/costs`, records from the `costs` array, no
pagination, `id` force-cast to a string via `last_path_segment` (see legacy rationale preserved
from the original migration).

Every other stream follows Vantage's documented `page`/`limit` query-param pagination surfaced
through a `links.next` URL in the response body (`docs.vantage.sh/api/pagination`) — expressed via
the engine's `next_url` pagination type (`next_url_path: links.next`). Each stream's
`records.path` selects the top-level array key that matches its route name (Vantage's own
documented convention: "The array of items will be available as a key that matches the name of
the route"):

- **Original Pass B set**: `cost_reports`, `budgets`, `folders`, `dashboards`, `business_metrics`,
  `resource_reports`, `recommendations`, `teams`, `saved_filters`, `workspaces`,
  `virtual_tag_configs`, `tags`, `cost_alerts`, `budget_alerts`, `anomaly_alerts`,
  `managed_accounts`, `financial_commitments`.
- **This pass's additions** (converting the prior breadth-vs-cost triage bucket into real
  coverage): `segments` (`GET /v2/segments`), `report_notifications`
  (`GET /v2/report_notifications`), `recommendation_views` (`GET /v2/recommendation_views`),
  `network_flow_reports` (`GET /v2/network_flow_reports`), `kubernetes_efficiency_reports`
  (`GET /v2/kubernetes_efficiency_reports`), `anomaly_notifications`
  (`GET /v2/anomaly_notifications`), `canvases` (`GET /v2/canvases`), `invoices`
  (`GET /v2/invoices` — the bare list endpoint has no MSP-tier gate per Vantage's own docs, only
  the invoice creation/regeneration/send actions are MSP-only, see Write actions & risks),
  `integrations` (`GET /v2/integrations`).

No stream declares an `incremental` block: Vantage's list endpoints document no server-side
updated-since filter parameter on any of these routes, so every read is a full-refresh sync,
matching legacy's own no-incremental behavior on `costs`.

## Write actions & risks

58 write actions covering create/update/delete lifecycles for every resource Vantage's REST API
supports mutating via a plain JSON body (all `body_type: json`, matching the dialect's default
flat-body construction):

- **budgets**: `create_budget`, `update_budget`, `delete_budget` (idempotent, 404 tolerated).
- **folders**: `create_folder`, `update_folder`, `delete_folder`.
- **dashboards**: `create_dashboard`, `update_dashboard`, `delete_dashboard`.
- **cost_reports**: `create_cost_report`, `update_cost_report`, `delete_cost_report`.
- **resource_reports**: `create_resource_report`, `update_resource_report`,
  `delete_resource_report`.
- **saved_filters**: `create_saved_filter`, `update_saved_filter`, `delete_saved_filter`.
- **workspaces**: `create_workspace`, `update_workspace`, `delete_workspace`.
- **teams**: `create_team`, `update_team`, `delete_team`.
- **cost_alerts**: `create_cost_alert`, `update_cost_alert`, `delete_cost_alert`.
- **budget_alerts**: `create_budget_alert`, `update_budget_alert`, `delete_budget_alert`.
- **anomaly_alerts**: `update_anomaly_alert` (Vantage documents no `POST`/`DELETE
  /v2/anomaly_alerts` — alerts are system-generated from anomaly detection, only their
  acknowledgment/threshold fields are user-editable via `PUT`).
- **business_metrics**: `create_business_metric`, `update_business_metric`,
  `delete_business_metric`.
- **virtual_tag_configs**: `create_virtual_tag_config`, `update_virtual_tag_config`,
  `delete_virtual_tag_config`. The synchronous `POST`/`PUT /v2/virtual_tag_configs[/{token}]`
  endpoints (a nested `values[]` array of tag-value definitions, each with its own `filter`/
  `aggregation` sub-object) are covered as ordinary JSON-body writes; Vantage's separate
  asynchronous variants (`PUT /v2/virtual_tag_configs/{token}/async`, polled via
  `GET /v2/virtual_tag_configs/async/{request_id}` and `GET
  /v2/virtual_tag_configs/{token}/status`) are excluded (see `api_surface.json`) since the
  synchronous path already covers the same mutation.
- **segments**: `create_segment`, `update_segment`, `delete_segment`.
- **report_notifications**: `create_report_notification`, `update_report_notification`,
  `delete_report_notification`.
- **recommendation_views**: `create_recommendation_view`, `update_recommendation_view`,
  `delete_recommendation_view`.
- **network_flow_reports**: `create_network_flow_report`, `update_network_flow_report`,
  `delete_network_flow_report`.
- **kubernetes_efficiency_reports**: `create_kubernetes_efficiency_report`,
  `update_kubernetes_efficiency_report`, `delete_kubernetes_efficiency_report`.
- **anomaly_notifications**: `create_anomaly_notification`, `update_anomaly_notification`,
  `delete_anomaly_notification`.
- **canvases**: `create_canvas`, `update_canvas`, `delete_canvas`.

**Not migrated**: invoice mutations (`POST /v2/invoices` create, `.../regenerate`, `.../send`,
`.../send_and_approve`) are all documented "MSP accounts only" — an account-tier gate this
connector's declared scope does not assume every caller has (see `api_surface.json`'s
`requires_elevated_scope` entries). Per-provider cloud integration onboarding (`POST
/v2/integrations/azure|gcp|custom_provider`) each require a distinct live-credential body shape
(Azure Service Principal secret, GCP billing/BigQuery identifiers, a 403-capable custom-provider
gate) that this pass does not collect into a single declarative write action; the `integrations`
read stream and its detail/cost sub-endpoints are still covered.

All actions carry `"risk": "external mutation; approval required"` (destructive deletes add
`"confirm": "destructive"`), matching this program's default write-risk posture.

## Known limits

- Full API-surface classification lives in `api_surface.json` (193 endpoints reviewed
  2026-07-04): 27 read streams, 58 write actions, and every remaining endpoint excluded with a
  specific, real per-endpoint reason under a closed-vocabulary category (`duplicate_of` for
  single-item detail GETs already covered by their list stream, 34 endpoints;
  `requires_elevated_scope` for account/billing/access-grant/SSO/MSP-invoicing/per-provider-cloud-
  credential administration, 30 endpoints; `non_data_endpoint` for health-check/profile/
  static-enum/forecast-projection/async-job-status-poller shapes, 28 endpoints; `binary_payload`
  for PDF/CSV file-transfer endpoints, 3 endpoints; `out_of_scope` for a small remaining set of
  narrower-priority admin/report-configuration/no-standalone-collection sub-resource mutations,
  13 endpoints, each individually reasoned).
- **`next_url` pagination ships single-page fixtures only** (the sanctioned exception,
  `docs/migration/conventions.md` §4): the next-page URL is the fixture-replay server's own
  address, unknown until test run time, so a static fixture cannot embed it. Live 2-page
  correctness for these streams is provable only via a future `paritytest/vantage` addition
  driving a real `httptest.Server` (not yet written in this pass) — `pagination_terminates`
  exercises the non-paginated `costs` stream instead, and every stream's `conformance` block is
  left unmarked (no `skip_dynamic`) since `read_fixture_nonempty`/`records_match_schema` still run
  and pass against the single-page fixture; only the specific 2-page-advance proof is deferred.
- `resources` (`GET /v2/resources`) and `unit_costs` (`GET /v2/unit_costs`) are intentionally NOT
  modeled as streams despite being plausible list endpoints: both require a mandatory
  parent-scoping token (`resource_report_token`/`workspace_token`, `cost_report_token`
  respectively) with no bare top-level collection semantics of their own, and this pass does not
  declare a `fan_out` block iterating every already-read report token to enumerate them — see
  `api_surface.json`'s `requires_elevated_scope` entries for both.
- **`id` on `costs` is force-cast to a string** via `last_path_segment` (preserved unchanged from
  the original migration) — see the original legacy rationale: `stringValue(item["id"])`
  (`vantage.go:134,173-184`).
- **`Check` dials the network; legacy's `Check` never did** — unchanged from the original
  migration, a deliberate fail-loud improvement with zero record-data impact.
- No incremental/date-range filtering is modeled on any stream — Vantage documents no server-side
  updated-since filter on any of the covered list endpoints.
