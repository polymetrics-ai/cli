# Overview

BoldSign is a declarative-HTTP connector reading and writing through the BoldSign REST API
(`https://api.boldsign.com`, OpenAPI 3.0.4 spec at `https://api.boldsign.com/swagger/v1/swagger.json`).
This bundle originally migrated `internal/connectors/boldsign` (the hand-written legacy connector,
read-only); this Pass B revision expands to the full documented API surface (see `api_surface.json`
for every one of the 85 endpoints, each `covered_by` a stream/write or `excluded` with a reason).
The legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a BoldSign API key via the `api_key` secret; it is sent as the `X-API-KEY` header
(`streams.json` `base.auth`'s `api_key_header` mode), matching legacy's
`connsdk.APIKeyHeader(boldsignAPIKeyHeader, secret, "")` (`boldsign.go:241`). Never logged.
`base_url` defaults to `https://api.boldsign.com` and may be overridden for tests/proxies.

## Streams notes

8 read streams, all page-numbered list endpoints (`Page`/`PageSize` query params, `page_size: 50`
— legacy's own `boldsignDefaultPageSize`); the engine stops on a short/empty page.

- `documents` (`/v1/document/list`), `templates` (`/v1/template/list`), `teams`
  (`/v1/teams/list`), `contacts` (`/v1/contacts/list`), `brands` (`/v1/brand/list`) are the 5
  original legacy-parity streams. Four wrap records in a `result` envelope; `teams` is the
  documented exception using `results`.
- `users` (`/v1/users/list`), `contact_groups` (`/v1/contactGroups/list`), and
  `sender_identities` (`/v1/senderIdentities/list`) are new Pass B streams with no legacy
  counterpart; their schemas are derived directly from BoldSign's published OpenAPI response
  schemas (`UsersDetails`, `GroupContact`, `SenderIdentityViewModel`).

Every stream's record mapper renames camelCase raw API fields to this bundle's snake_case schema
fields via `computed_fields`; bare single `{{ record.<path> }}` entries and `coalesce` entries both
preserve the raw JSON type. ID fields that legacy resolved with alternate mixed-case fallbacks now
use `coalesce` with the same fallback order, e.g.
`document_id` from `documentId` then `documentID`, `team_id` from `teamId` then `teamID`,
`brand_id` from `brandId` then `brandID`, and contact `id` from `id` then `contactId`.
**Corrected in this revision**:
`documents.created_date`/`expiry_date` and `templates.created_date`/`teams.created_date` are
BoldSign's real wire type — Unix-seconds **integers** (confirmed against the published OpenAPI
spec and `developers.boldsign.com`'s documented sample response), not the RFC3339 strings the
prior fixtures happened to use (an artifact of legacy's own test stub, which never asserted a real
wire shape since legacy's mapper just passes `item["createdDate"]` through verbatim regardless of
type). Schemas and fixtures are now `"integer"`-typed end to end, matching the Stripe `created`
cursor precedent (§5 of `docs/migration/conventions.md`). Likewise `contacts.phone_number` is now
schema-typed `object` (BoldSign's real `{countryCode, number}` shape), not the bare string the
prior fixture used.

`documents` and `templates` declare `created_date` as `x-cursor-field`, matching legacy's
`CursorFields: []string{"created_date"}`; `teams` and `users` likewise declare `created_date`.
None of these actually filter server-side or client-side (BoldSign's list endpoints have no
incremental filter parameter), so no `incremental` block is declared on any stream — full refresh
only. `contacts`, `brands`, `contact_groups`, and `sender_identities` declare no cursor field.

## Write actions & risks

8 write actions (`capabilities.write` is now `true`):

- `create_team`/`update_team` — create/rename a BoldSign team (`POST /v1/teams/create`,
  `PUT /v1/teams/update`).
- `update_contact`/`delete_contact` — update or permanently delete a contact
  (`PUT /v1/contacts/update?id=...`, `DELETE /v1/contacts/delete?id=...`; delete is idempotent,
  404 counts as written).
- `create_contact_group`/`update_contact_group`/`delete_contact_group` — manage a BoldSign
  contact group (`POST`/`PUT`/`DELETE /v1/contactGroups/*`).
- `revoke_document` — permanently revokes a document's signature request
  (`POST /v1/document/revoke?documentId=...`); destructive (`confirm: "destructive"`).
- `remind_document` — sends an email/SMS reminder to pending signers
  (`POST /v1/document/remind?documentId=...`).
- `delete_document` — trashes or (when `deletePermanently: true`) permanently deletes a document
  (`DELETE /v1/document/delete?documentId=...&deletePermanently=...`); destructive, idempotent.
- `add_document_tags`/`delete_document_tags` — manage a document's label tags
  (`PATCH`/`DELETE /v1/document/addTags` `/deleteTags`).
- `update_user`/`change_user_team` — change a user's role/active-status, or move them to a
  different team (`PUT /v1/users/update`, `PUT /v1/users/changeTeam?userId=...`).

Every write action requires operator approval (`metadata.json`'s `risk.approval`); `revoke_document`,
`delete_document`, `delete_contact`, and `delete_contact_group` additionally set
`confirm: "destructive"`.

**Not implemented — `create_contact` and `create_user` (`ENGINE_GAP`)**: BoldSign's real wire
request body for both `POST /v1/contacts/create` and `POST /v1/users/create` is a JSON **array**
of objects (`[{...}]`), confirmed against the published OpenAPI spec and
`developers.boldsign.com/contacts/create-contact/`'s documented sample request. The engine's write
dialect (`write.go`'s `executeWriteRecord`) only ever sends a single JSON object (or form-encoded
body) per record — there is no array-wrapping body option in `writes.json` today. Modeling either
write would either send a malformed single-object body BoldSign's API would reject outright, or
require inventing an undeclared wrapping behavior the dialect doesn't support. See
`api_surface.json`'s excluded entries for both endpoints.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** Same as pre-expansion: fixed values
  baked into `streams.json`'s `base.pagination` block at bundle-author time.
- **`create_contact`/`create_user` are not migrated** — see Write actions & risks above;
  `ENGINE_GAP`, not a silently-dropped write.
- **Document/template sending, embedded-URL generation, and binary downloads are out of scope**
  (see `api_surface.json`'s 63 excluded entries): multipart file-upload flows
  (`document/send`, `document/create`, `template/create`, `template/send`, `document/edit`,
  `template/edit`), embedded interactive-session URL generators (Designer, editor, sign, preview),
  raw byte-stream downloads (document/template PDF, audit trail, ID-verification image/report),
  brand asset management, custom fields, per-signer authentication controls, and sender-identity
  verification workflows. None of these were part of legacy's read-only surface either.
- `documents`/`templates`/`teams`/`users` are full-refresh only (no server-side incremental filter
  parameter on any BoldSign list endpoint), even though each declares a schema cursor candidate.
