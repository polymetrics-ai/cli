# Overview

FireHydrant is a PASS B declarative bundle for the FireHydrant REST API v1. It keeps the five legacy-parity projected streams (incidents, services, teams, environments, and functionalities) and adds passthrough streams for documented JSON GET resources that the engine can read directly from the official FireHydrant API reference. Direct JSON/no-body mutations are exposed as write actions.

## Auth setup

Provide `api_token` as a secret from FireHydrant API keys. The connector sends it as `Authorization: Bearer <api_token>`.

`base_url` defaults to `https://api.firehydrant.io/v1`. FireHydrant also documents `https://api-read.firehydrant.io/v1` for read-only traffic; use the default primary API root when write actions are enabled.

## Streams notes

The legacy streams retain schema projection and typed schemas so they continue emitting the same fields as `internal/connectors/firehydrant`: `incidents`, `services`, `teams`, `environments`, and `functionalities`.

Additional PASS B streams use `projection: passthrough` with an open generic object schema. This preserves the FireHydrant JSON resource shape for broad read coverage without hand-maintaining hundreds of nested response schemas. Paginated endpoints read records from top-level `data` and follow `pagination.next` with the `page` query parameter; single-resource endpoints emit the response object as one record.

Streams for detail or subresource endpoints interpolate path IDs from optional config properties named after the documented path parameter, such as `incident_id`, `service_id`, `team_id`, `id`, or `slug`. Streams with required query parameters use operation-specific config keys, for example `search_zendesk_tickets_query` and `list_mttx_metrics_start_date`.

## Write actions & risks

`writes.json` exposes documented POST, PUT, PATCH, and DELETE endpoints when they are expressible as one JSON or no-body request. Path parameters are read from the input record and all non-path fields are sent as the JSON body. DELETE actions are marked destructive and treat 404 as an idempotent missing result.

These actions can create incidents, catalog objects, tasks, runbooks, status pages, roles, schedules, notification policies, integration resources, and related FireHydrant records; they can also archive, delete, close, publish, page, trigger, or otherwise mutate live operational data. Use the reverse ETL plan, preview, approval, execute flow for every write.

## Known limits

- SCIM endpoints are excluded because they require directory-provisioning administration and SCIM-specific payload semantics outside this REST connector.
- Multipart/file endpoints, image uploads, attachments, and export/download endpoints are excluded as `binary_payload`.
- Mutations that require query parameters are excluded because `writes.json` has no query-parameter dialect.
- `GET /v1/catalogs/{catalog_id}/refresh` is excluded because it schedules a refresh rather than returning a durable data resource.
- Connectivity endpoints `/v1/ping` and `/v1/noauth/ping` are excluded from streams because the bundle's check request covers connectivity.
- Non-legacy PASS B streams intentionally use passthrough generic schemas; they preserve raw JSON shape but do not provide narrow field catalogs or per-resource primary keys.
- `page_size` range is documented by FireHydrant as max 200, but the declarative spec does not enforce numeric minimum/maximum constraints; out-of-range values are sent to FireHydrant and handled by the API.
