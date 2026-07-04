# Overview

7shifts is a restaurant and hospitality scheduling and labor-management platform. This bundle covers the documented 2026-01-01 7shifts API reference for ordinary tenant-scoped REST resources. The original legacy-parity streams keep their legacy projection shape; newly added streams use the documented response `data` shapes.

## Auth setup

Provide a 7shifts API access token as the `access_token` secret. The engine sends it as Bearer authentication. Provide `company_id` for company-scoped streams and write actions. Some detail, report, and settings streams also require the specific path or query config value named in `spec.json`, such as `location_id`, `user_id`, `time_off_id`, `date`, `from`, or `to`.

## Streams notes

List streams use 7shifts cursor pagination when the endpoint documents `cursor` and `limit`; the connector sends `limit=100` and follows `meta.cursor.next`. Single-resource, settings, identity, and report endpoints use one-page reads. The original `companies`, `locations`, `departments`, `roles`, `users`, `shifts`, and `time_punches` streams keep the legacy incremental `modified_since` behavior and legacy field projections.

## Write actions & risks

Supported write actions are single HTTP requests with JSON bodies or empty-body deletes. The action schema follows the documented request body plus path fields supplied in the record. Deletes are marked destructive and treat `404` as idempotent success where the API indicates the resource is already absent.

## Known limits

- Deprecated endpoints are not implemented; each is marked `deprecated` in `api_surface.json` and covered by the current non-deprecated endpoint where one exists.
- `POST /oauth2/token` is excluded as a non-data authentication flow because this bundle authenticates with an existing bearer token.
- `POST /v2/partner_company_creation` is excluded because it is an elevated partner account-creation flow, not an ordinary tenant-scoped reverse-ETL action.
- Mutations whose request body is a root JSON array are excluded with an engine-gap reason. The current declarative write engine sends JSON objects built from record fields, not root arrays.
- Optional query filters beyond required filters, pagination `cursor`, pagination `limit`, and legacy incremental `modified_since` are not surfaced as connection config. Add a dedicated stream if an optional filter becomes required for a production workflow.
