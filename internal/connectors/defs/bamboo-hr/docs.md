# Overview

BambooHR is an HR source and reverse-ETL target backed by the documented BambooHR REST API. This Pass B bundle keeps the legacy employee-directory and metadata streams, then expands the Basic-auth JSON API surface to 84 streams and 101 write actions. The coverage manifest also lists OAuth-only, binary/multipart, deprecated, and read-like POST endpoints with explicit exclusions.

## Auth setup

Provide the BambooHR account `subdomain` from `https://<subdomain>.bamboohr.com` and an `api_key` secret. The API key is sent as the HTTP Basic username with a literal `x` password, matching BambooHR's API-key convention. The bundle base URL is `https://{ config.subdomain }.bamboohr.com`; stream paths include `/api/v1`, `/api/v1_1`, `/api/v1_2`, or `/api/v2` as documented.

## Streams notes

The first four streams preserve legacy record DATA: `employees`, `meta_fields`, `meta_lists`, and `time_off_types` keep their snake_case fields and stringified primary keys. New Pass B streams use BambooHR's documented JSON field names directly and are intentionally permissive for account-specific fields. Detail streams with path parameters use optional config values such as `employee_id`, `report_id`, or stream-specific `*_id` fields; those values are only required when reading that stream.

Cursor and next-link list endpoints use `next_url` pagination when BambooHR returns a next URL. Legacy `employees` directory pagination keeps the `page`/`limit` short-page behavior and a full first fixture page.

## Write actions & risks

Write actions cover Basic-auth JSON/path POST, PUT, PATCH, and DELETE operations that the declarative write dialect can express. Delete actions are marked destructive and treat 404 as idempotent missing-ok where BambooHR reports not-found. All writes are one HTTP request per input record and require the normal reverse-ETL plan, preview, approval, and execute flow.

## Known limits

- OAuth-only public API operations are excluded as `requires_elevated_scope`; adding OAuth would change the legacy connector's API-key credential model and needs a separate auth design.
- File, photo, CSV/PDF export, import, and multipart upload endpoints are excluded as `binary_payload` because this bundle is JSON-only.
- Read-like POST endpoints, including dataset data queries and field-option POSTs, are not exposed as writes. The current declarative read path does not send stream request bodies, so treating those endpoints as reverse-ETL writes would misrepresent their behavior.
- `api_surface.json` contains 340 documented method/path entries; 155 are explicit exclusions with closed-vocabulary categories.
