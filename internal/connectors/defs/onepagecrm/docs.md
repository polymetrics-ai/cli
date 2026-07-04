# Overview

OnePageCRM is a Pass B declarative-HTTP bundle for the OnePageCRM API v3. It keeps the legacy
`contacts`, `deals`, `actions`, `companies`, and `users` stream projections from
`internal/connectors/onepagecrm`, then expands the documented OpenAPI surface with additional
account/configuration streams, detail streams, nested CRM subresource streams, and declarative write
actions where the current JSON/path write dialect can model the request.

## Auth setup

Provide the OnePageCRM API user ID as the `username` config value and the OnePageCRM API key as the
`password` secret; both are required. They are sent as HTTP Basic auth (`username:password`),
matching legacy `onepagecrm.go`'s `connsdk.Basic(username, password)` exactly. The password is never
logged.

## Streams notes

The 5 legacy streams (`contacts`, `deals`, `actions`, `companies`, `users`) preserve OnePageCRM's
legacy list-endpoint projections field-for-field. They read `GET /<resource>`, records under
`data.<resource>` (`users` is the exception: records live directly under `data`), and unwrap fields
with typed `computed_fields`, matching legacy's `unwrap` plus `mapRecord` behavior.

Pass B streams use `projection: "passthrough"` plus a computed `id` so the documented OnePageCRM
envelope is preserved for new resources while the engine still has a stable primary key. New
streams cover bootstrap/account data, users/detail, lead sources, statuses, custom/deal/company
fields, predefined actions/items/groups, notes, calls, call results, meetings, relationship types,
countries, filters, detail records, company/contact nested subresources, cascade contacts,
action/team streams, notifications, webhooks, and pipelines.

Pagination is `page`/`per_page` (`pagination.type: page_number`, `page_param: page`, `size_param:
per_page`, `start_page: 1`, `page_size: 100` matching legacy's default `onepagecrmDefaultPageSize`).

Documented parity deviation (stop condition): legacy's `harvest` loop stops primarily on the
response body's `data.max_page` field (falling back to a short-page heuristic only when `max_page` is
absent or unparsable). The engine's `page_number` paginator has no body-driven stop signal at all —
it stops purely on a short/empty page (`recordCount < page_size`). For every real OnePageCRM response
this bundle's pagination terminates at the exact same page a still-fully-populated final page (one
whose record count happens to equal `page_size`) causes one extra page request versus legacy (which
would stop immediately via `max_page`); that extra request returns an empty page and the engine stops
on the following iteration. No record is ever duplicated, dropped, or reordered by this difference —
it is a request-count/efficiency deviation, never an emitted-record-data deviation, so it is
ACCEPTABLE per `docs/migration/conventions.md`'s parity-deviation meta-rule.

## Write actions & risks

`writes.json` declares the documented JSON/path POST, PUT, PATCH, and DELETE mutations that the
engine can express: account configuration records, notes/calls/meetings/deals/actions, attachments,
relationship types, company/contact updates and subresource mutations, notifications, and webhooks.
Delete actions include idempotent 404 handling and destructive confirmation where appropriate.

Every write is a live OnePageCRM mutation and must go through plan/preview/approval before
execution. The schemas validate path identifiers plus representative documented body fields; the
full endpoint-specific payload contract remains server-validated by OnePageCRM.

## Known limits

- `POST /change_auth_key` is excluded because it rotates the current API key and returns new
  credential material; credential rotation belongs in a dedicated credential-management flow, not a
  generic connector write action.
- Photo/logo upload/update endpoints are excluded as binary payloads because they accept base64
  image data. Attachment metadata/association writes that use ordinary JSON remain modeled.
- `DELETE /contacts/delete` is excluded as a broad destructive operation: the OpenAPI operation has
  no path identifier and no JSON request body/query dialect that `writes.json` can safely model.
- `GET /attachments/s3_form` is excluded as a non-data upload-helper endpoint for pre-authorized S3
  file upload, not a durable CRM object stream.
- The pagination stop-condition deviation described above (body `max_page` vs. short-page heuristic)
  is an accepted, request-count-only difference — see Streams notes.
- `contacts`/`deals`/`actions`/`companies` declare `updated_at` as `x-cursor-field` for manifest
  parity with legacy's published `CursorFields`, but neither legacy nor this bundle actually issues a
  server-side incremental filter against it (legacy's OnePageCRM API integration performs full syncs
  only); `users` has no cursor field at all, matching legacy exactly.
