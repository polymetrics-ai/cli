# Mixpanel

## Overview

Reads and writes the full documented Mixpanel API surface: annotations, A/B experiments, feature
flags (management and evaluation), GDPR/CCPA data-subject jobs, identity linking, event/profile/group
ingestion, Lexicon governance schemas, the Query API's report and lookup endpoints, service accounts,
warehouse connector imports, and warehouse export pipelines.

The documented surface is **104 operations**, re-derived on 2026-08-07 from the 13 real OpenAPI 3.x
documents Mixpanel publishes at `https://docs.mixpanel.com/openapi/<name>.openapi.yaml` (annotations,
data-pipelines, experiments, export, feature-flags-management, feature-flags, gdpr, identity,
ingestion, lexicon-schemas, query, service-accounts, warehouse-connectors — 392,835 combined bytes).
Raw operationId count summed across the 13 files is 105, an exact match to the provider-artifact
ledger this connector previously recorded; the -1 delta is one genuine cross-file path collision:
`POST /import` is documented twice (`identity-merge` in identity.yaml, `import-events` in
ingestion.yaml), which the counting policy's method+path dedup rule collapses into the single
`import_events` write below (its `event` field accepts the literal `$merge` for identity-merge
semantics, exactly as Mixpanel's own batch-import endpoint does).

No OAS 3.1 top-level `webhooks` object exists in any of the 13 files, and the string `webhook` does
not appear anywhere in them either — Mixpanel documents no webhook surface at all here.

Every documented operation is partitioned exactly once in `api_surface.json` and carries exactly one
disposition — executable (`covered_by`), or blocked with a named dependency and a source citation
(`operation`). None is blank, and none uses the legacy `excluded` category.

This bundle predates the 13-file ledger with 10 read-only streams covering a mix of legacy Query API
v2.0 endpoints (not part of the 13-file ledger at all) and current-API list/detail GETs. Of those 10:
7 already matched a documented current-API path exactly and are unchanged; the `cohorts` and `engage`
streams keep their legacy `/api/2.0` wire calls but now also serve as this bundle's implementation of
the documented `POST /cohorts/list` and `POST /engage` operations (the original bundle's own
`duplicate_of` classification already established these as the same record family — see Known
limits); and the legacy, non-project-scoped `annotations` stream was retired as redundant with the
project-scoped `project_annotations` stream, which already implements the current, documented
`GET /projects/{projectId}/annotations` operation.

## Auth setup

Reads and admin-surface writes (annotations, experiments, feature flags, GDPR, Lexicon schemas,
Query API, service accounts, warehouse connectors, warehouse pipelines) use the existing Basic-auth
`mixpanel` hook — a Mixpanel service-account username plus secret (password or legacy `api_secret`),
config-then-secrets precedence, unchanged by this sweep:

```
pm credentials add mixpanel --from-env MIXPANEL_USERNAME=username MIXPANEL_PASSWORD=password
```

Two new endpoint families need their own credential, both `x-secret`, never logged:

- `project_token` — Mixpanel project token, required by ingestion (`/track`, `/engage`, `/groups`,
  `/import`) and feature-flag evaluation (`/flags`, `/flags/definitions`). Ingestion writes also
  accept `$token`/`token` as an ordinary record field per Mixpanel's own token-in-body ingestion
  design (the transport-level Basic auth above is sent regardless and is harmless alongside it, per
  Mixpanel's documented `security: []` on these operations).
- `gdpr_token` — a separate GDPR API token from project settings, required by
  `POST /data-retrievals/v3.0` and `POST /data-deletions/v3.0`.

A further ~18 non-secret config properties (`experiment_id`, `flag_id`, `organization_id`,
`import_id`, `pipeline_name`, `entity_type`, `schema_name`, and the Query API report filters such as
`from_date`/`to_date`/`event_name`/`funnel_id`) name path/query parameters for the new detail-lookup
and report streams; each is also exposed as an optional per-invocation CLI flag that overrides the
stored connection default.

## Streams notes

42 streams (9 pre-existing, 33 new); all non-paginated (`pagination: {"type": "none"}` — the base
bundle's inherited cursor pagination does not apply to any of the endpoints below, since none of
them return a `next`/`page` cursor). Grouped by family:

- **Annotations** (unchanged): `project_annotations`, `project_annotation`, `annotation_tags`.
- **Legacy-compatible Query API** (unchanged wire calls, now also covering documented rows):
  `cohorts`, `engage`, `saved_funnels`, `activity_stream`, `top_events`, `event_property_names`.
- **Experiments**: `experiments` (list), `experiment` (detail, requires `experiment_id`).
- **Feature flag administration**: `feature_flags_admin` (list), `feature_flag` (detail, requires
  `flag_id`).
- **Feature flag evaluation**: `feature_flag_assignments`, `feature_flag_definitions` — hit
  `api.mixpanel.com`, authenticated by `secrets.project_token`.
- **GDPR jobs**: `gdpr_retrieval`, `gdpr_deletion` — job status lookups by `gdpr_tracking_id`.
- **Lexicon governance schemas**: `lexicon_schemas`, `lexicon_schemas_for_entity`,
  `lexicon_schema_by_name`.
- **Query API reports** (single-object responses; primary key is a static `report` marker field
  stamped via `computed_fields`, since Mixpanel's aggregate report bodies have no natural per-row
  id): `insights_report`, `funnels_report`, `retention_report`, `retention_frequency_report`,
  `segmentation_report`, `segmentation_numeric_report`, `segmentation_sum_report`,
  `segmentation_average_report`, `events_report`, `event_names_report`, `event_properties_report`,
  `event_property_values_report`.
- **Service accounts**: `service_accounts_org`, `service_account`, `project_service_accounts`.
- **Warehouse connector imports**: `warehouse_imports`, `warehouse_import`,
  `warehouse_import_history`.
- **Warehouse export pipelines**: `warehouse_pipeline_jobs`, `warehouse_pipeline_status`,
  `warehouse_pipeline_timeline` — hit `data.mixpanel.com`.
- **Ingestion**: `lookup_tables` (list only; replacing a table is blocked — see Known limits).

Every stream targets an explicit absolute URL (bypassing `config.base_url`) for its real Mixpanel
host — `mixpanel.com/api/app` (admin), `mixpanel.com/api/query` (Query API),
`data.mixpanel.com/api/2.0` (warehouse pipelines), or `api.mixpanel.com` (ingestion, flag
evaluation) — exactly matching the pattern 7 of the 10 pre-existing streams already used, because the
declarative `operations.json` direct-read/direct-write/binary-download executors require a
connector-relative path resolved against the ONE configured `base_url` and this bundle's documented
surface spans 5 real hosts (see Known limits).

## Write actions & risks

57 reverse-ETL write actions (33 `implemented`, 24 `partial` — flat CLI flags cannot express a
required object/array-of-object field; supply those records from a typed reverse-ETL source table).
Grouped by risk:

- **critical**: `create_gdpr_deletion` (starts a GDPR/CCPA deletion job), `create_service_account`,
  `delete_service_account`, `delete_warehouse_import` (optionally also deletes imported data).
- **high**: destructive deletes (`delete_annotation`, `delete_experiment`, `delete_feature_flag`,
  `delete_profile`, `delete_group`, `delete_all_schemas`, `delete_schemas_for_entity`,
  `delete_schema_by_name`, `cancel_gdpr_deletion`, `cancel_warehouse_pipeline`), lifecycle actions
  that change live behavior (`force_conclude_experiment`, `launch_experiment`, `update_feature_flag`),
  identity/ingestion writes that mutate canonical identity or historical event data
  (`create_identity`, `create_identity_alias`, `import_events`), and admin-scope creates
  (`create_feature_flag`, `upload_schemas`, `upload_schema_by_name`,
  `add_service_accounts_to_projects`, `remove_service_accounts_from_projects`,
  `create_event_stream_import`, `create_people_import`, `create_groups_import`,
  `create_lookup_table_import`, `update_warehouse_import`).
- **medium**: ordinary create/update actions (annotations, experiment lifecycle, GDPR retrieval,
  ingestion track/profile/group writes, warehouse pipeline pause/resume, manual import run).
- **low**: `create_annotation_tag`.

All 24 ingestion writes (`track_event`, `import_events`, the 9 `profile_*` actions, and the 7
`group_*` actions) send Mixpanel's real wire shape — a JSON **array** of one or more event/update
objects — via `body_type: "json_array"` with `body_field`/`body_schema` naming the wrapping array
property (`events` or `updates`); this is why they are `partial`: no CLI flag type expresses an array
of typed objects, so they are reachable only through a typed reverse-ETL source record, exactly like
`profile_batch_update`/`group_batch_update` (which instead take the batch as a single
`x-www-form-urlencoded` `data` string, per Mixpanel's own documented shape for those two operations).

## Known limits

- **`operation_ledger_version: 1`**; `api_surface.json` carries all 104 documented rows (99
  `covered_by`, 5 blocked). Blocked rows, each with a reason, a source citation, and a
  `named_dependency=` note:
  - `POST /jql` — custom JQL executes caller-provided JavaScript; this is the query-API equivalent
    of the generic script/SQL-write tools this repository forbids outright. Permanently disallowed.
  - `GET /export` — raw event export lives on `data.mixpanel.com`, a different host than every other
    still-in-scope operation's chosen base, and returns newline-delimited JSON rather than one JSON
    document or a paginated list; neither the stream reader nor the connector-relative
    `binary_download` executor fits without a new engine capability (a per-operation base-host
    override, or an NDJSON-aware reader).
  - `POST /nessie/pipeline/create` and `POST /nessie/pipeline/edit` — both request bodies are
    `oneOf`/`allOf` discriminated unions (create has 8 alternate pipeline-type shapes); no single
    fixed `record_schema` represents either operation, so each stays non-implemented rather than
    silently only supporting one arm (AGENTS.md's "Command Surface Must Stay Executable").
  - `PUT /lookup-tables/{id}` — replaces a lookup table via a raw `text/csv` body; the engine's write
    `body_type` dialect supports only json/form/none/graphql/json_array/multipart/base64_upload, none
    of which reproduce a raw non-JSON content type.
- `cohorts`/`engage` streams keep their pre-existing `GET /api/2.0/cohorts/list`/`GET /api/2.0/engage`
  wire calls rather than switching to the documented `POST` current-API equivalents, matching this
  bundle's own prior `duplicate_of` judgment that they are the same record family; a future pass could
  additionally implement the POST shapes as their own streams if that judgment needs re-verifying
  against live traffic.
- Batch defaults carried over unchanged: `read_page_size=1000`.
- `capabilities.query` remains `false` — Mixpanel has no ad hoc/custom-query capability this connector
  exposes (JQL is the closest fit and is permanently disallowed above).
