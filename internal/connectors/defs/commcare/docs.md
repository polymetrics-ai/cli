# Overview

CommCare is a Dimagi mobile data-collection and case-management platform. This
bundle covers the official CommCare HQ API documentation at
https://commcare-hq.readthedocs.io/api/index.html plus the legacy connector's
v0.5 form and case list endpoints. It reads application structure, form and
case data, users, groups, reports, locations, lookup tables, DET exports, and
messaging events. It also exposes declarative JSON write actions for the
documented case v2, mobile worker, web-user invitation/access, group, location
v2, lookup table, and lookup table row mutations.

The legacy `internal/connectors/commcare` package emitted records verbatim for
forms and cases. Every stream here therefore uses passthrough projection so
project-defined fields, nested form payloads, case properties, permissions,
report columns, lookup-table fields, and messaging payloads are preserved.

## Auth setup

Provide `api_key` as a secret and `project_space` as config. The API key is sent
as `Authorization: ApiKey <value>`, matching the legacy connector's
`APIKeyHeader` behavior. `base_url` defaults to `https://www.commcarehq.org`.

Several detail streams and writes require id-like config or record fields such
as `app_id`, `case_id`, `mobile_worker_id`, `web_user_id`, `group_id`,
`location_id`, `lookup_table_id`, and `lookup_table_item_id`. These are ordinary
resource identifiers, not credentials.

## Streams notes

Most list endpoints use CommCare's `objects` envelope with offset/limit
pagination. The legacy v0.5 `forms` and `cases` streams are kept for parity.
The documented API v1/v2 endpoints are exposed separately (`forms_v1`,
`cases_v1`, `cases_v2`, and detail variants).

`cases_v2` and `messaging_events` use documented next-link pagination. `report_data`
uses offset/limit pagination and extracts rows from the `data` field; its row
shape is report-specific, so the schema intentionally allows arbitrary columns.
Form attachments and OTA restore are excluded because they return binary/XML
payloads rather than JSON records.

## Write actions & risks

Write actions are live external mutations and require the normal reverse-ETL
plan, preview, approval, and execute flow. Covered actions include case v2
create/update/upsert, mobile worker create/update/delete/reset email, web-user
invitation/update/enable/disable, group create/bulk-create/update/delete,
location v2 create/update/bulk upsert, lookup table create/update/delete, and
lookup table row create/update/delete.

The mobile-worker create/update actions intentionally model account/profile
fields without password-bearing fixture examples. Password-reset is represented
as its own empty-body action.

## Known limits

- Multipart uploads are excluded: application import, multimedia upload, Excel
  case upload, and Excel lookup-table upload require file bodies the declarative
  JSON write dialect cannot construct.
- OpenRosa form submission endpoints are excluded because they require XForm XML
  body construction.
- Case v2 `POST /bulk_fetch/` is a read-style POST with an arbitrary body; stream
  reads cannot send request bodies. Case v2 bulk create/update with a root JSON
  array is also excluded because declarative writes send one JSON object per
  record.
- OTA restore is excluded because it returns XML and uses mobile-worker Basic
  authentication rather than the project API-key JSON path.
- Dynamic report row schemas are report-specific. The `report_data` stream uses
  passthrough projection and an open schema to avoid dropping columns.
