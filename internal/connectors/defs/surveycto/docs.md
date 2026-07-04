# Overview

SurveyCTO is a Tier-1 declarative-HTTP connector for the SurveyCTO **Server API v2**
(`https://<server>.surveycto.com/api/v2`). This is a Pass B full-surface expansion against the
real, published API v2 OpenAPI 3.0 spec (`https://developer.surveycto.com/specs/api-v2.json`,
found via `developer.surveycto.com`'s VitePress-rendered developer portal — the same
`docs.surveycto.com` product-docs site the pre-Pass-B bundle's `docs_url` pointed at only links
out to this separate Developer Portal for the actual API reference).

**The v2 API is a substantially different, richer surface than the earlier version of this
bundle assumed.** There is no plain `GET /forms` or `GET /cases` endpoint at all:
- "Forms" are enumerated only via `GET /forms/ids`, which returns a **bare JSON array of scalar
  strings** (`["my_survey", "registration_form", ...]`), not objects — see Known limits for why
  this cannot be modeled as a stream.
- "Cases" are not a separate resource — they are simply one dataset among others, distinguished by
  `discriminator: "CASES"` in the generic `/datasets` catalog (a `title: "Cases"`, `fieldNames:
  "id,label,formids,users,roles,sortby,enumerators"` dataset, per the API's own documented
  example). `dataset_records` (this bundle's generic per-dataset record stream) reads cases exactly
  like any other dataset's records once fanned out to the `cases` dataset id.

This bundle covers **6 read streams** (`datasets`, `dataset_records`, `submissions`, `groups`,
`roles`, `users`) and **7 write actions** (dataset create/update/delete, dataset record creation,
user create/update/delete) — the full real surface this dialect can express; see `api_surface.json`
for the disposition of every one of the 22 documented v2 paths, including two genuine `ENGINE_GAP`
blockers detailed below.

**Tier justification**: a plain declarative-HTTP bundle. Auth is HTTP Basic (unchanged from the
prior version), pagination is the engine's `cursor` type throughout (the real API's own
`CursorPaginatedResponse` envelope — `{total, limit, data, nextCursor}` — applies identically to
every list endpoint), and `dataset_records` is an ordinary single-level `fan_out` over `datasets`
— nothing needs a Go hook.

## Auth setup

Provide a SurveyCTO username and password/API key via the `username` and `password` secrets; both
are sent as HTTP Basic auth credentials (unchanged from the prior version of this bundle). Neither
is logged.

## Streams notes

All streams share the base's `cursor` pagination (`cursor_param: cursor`, `token_path:
nextCursor`) — SurveyCTO's real v2 wire envelope for every list endpoint (`CursorPaginatedResponse`:
`{"total": N, "limit": N, "data": [...], "nextCursor": "..."}`, confirmed directly from the published
OpenAPI spec's own response examples for `datasets`/`dataset records`/`groups`/`roles`/`users`).
`limit=1000` (the documented max) is sent statically on every stream.

`datasets` (`GET /datasets`) is the generic dataset catalog — every SurveyCTO dataset (general
`DATA`, `ENUMERATORS`, and `CASES`-discriminated case-management datasets alike) as one uniform
stream, matching the API's own single-endpoint modeling of what the prior bundle's `datasets` and
`cases` streams incorrectly treated as two separate resources.

`dataset_records` (`GET /datasets/{dataset_id}/records`) `fan_out`s over every dataset id
(`ids_from.request`: `GET /datasets`, `records_path: data`, `id_field: id`), stamping `dataset_id`
onto every emitted record. This is the correct, general replacement for the prior bundle's
dedicated `cases` stream — reading the `cases` dataset's records now happens automatically as one
of many datasets this stream fans out over, with no special-casing needed.

`submissions` (`GET /forms/{form_id}/submissions`) remains scoped to a single, config-supplied
`form_id` (required config, unchanged in spirit from the prior version) rather than fanning out
over every form, because of the `forms/ids` `ENGINE_GAP` below — there is no config-free way to
discover every form id to fan out over. `submissions`' schema keys off `submissionId` (not `id` —
the real API's `orderBy` enum names `submissionId`/`reviewStatus`/`formDefinitionVersion`/
`submissionDate`/`completionDate` as the metadata fields, and the actual per-submission response
body shape is otherwise undocumented in the OpenAPI spec itself, `{"type": "object"}` with no
`$ref`); `x-cursor-field` is `completionDate` (the API's own default `orderBy` value) as an
informational marker only — no `incremental` block is declared, since the real per-submission
record shape (and therefore which field genuinely represents a monotonic lower bound for a given
form) is not confirmable without a live authenticated response to inspect.

`groups`/`roles`/`users` are plain top-level list streams, each with a matching `GET
/{resource}/{id}` (or `/roles/{roleId}`, permission-detail-shaped rather than a plain record read)
detail endpoint excluded per `api_surface.json`.

## Write actions & risks

- `create_dataset`/`update_dataset`/`delete_dataset` — dataset lifecycle. SurveyCTO's own API
  forbids changing a dataset's `discriminator` (type) after creation — `update_dataset` still
  requires `discriminator` in its record schema (matching the real `DatasetInput` request body
  shape) but the API itself rejects an attempt to change it, not this bundle. `delete_dataset` is
  irreversible; approval required. Create/update are lower-risk (no approval required).
- `create_dataset_record` — adds a record to a dataset (`POST /datasets/{dataset_id}/records`).
  SurveyCTO's own `DatasetRecordFieldMap` request body has no fixed schema (dataset field names are
  workspace-defined per dataset), so `record_schema` only requires the routing field `dataset_id`;
  every other record property is sent verbatim as the field-name/value map. Low-risk, no approval
  required.
- `create_user`/`update_user`/`delete_user` — user lifecycle. `create_user`/`update_user` can set
  the user's password in the same call — this is a credential-provisioning action, not an ordinary
  data mutation, and both require approval for that reason (not merely because they mutate data).
  `delete_user` revokes access; approval required.

## Known limits

- **`ENGINE_GAP` — `GET /forms/ids` and `GET /teams/ids` cannot be modeled as streams (or as
  `fan_out` id sources).** Both return a bare JSON array of SCALAR STRINGS
  (`["my_survey", "registration_form", ...]`), not objects. `connsdk.RecordsAt` — used by both
  ordinary stream reads and `fan_out`'s `ids_from.request` — only keeps array elements that decode
  as a JSON object (`map[string]any`); every element of a scalar-only array is silently dropped,
  yielding zero records or zero fan-out ids. There is no dialect mechanism for "array of bare
  scalars" extraction. This is the same class of gap `docs/migration/conventions.md`'s
  parity-deviation ledger item 12 (ip2whois's `nameservers` field) documents and formalizes as a
  genuine, still-open `ENGINE_GAP` rather than a workaround target. Consequence: there is no
  standalone `forms`/`teams` stream, and `submissions` cannot fan out over every form (it is
  instead scoped by a single required `config.form_id`, forcing one connection per form for full
  multi-form coverage).
- **`ENGINE_GAP` — dataset single-record CRUD (`getRecord`/`updateRecord`/`deleteRecord`/
  `upsertRecord`, all at `/datasets/{datasetId}/record` singular) cannot be modeled as write
  actions.** Every one of these endpoints takes `recordId` as a **query parameter**
  ("passed as a query parameter to support special characters", per the API's own docs), not a
  path segment. `engine.WriteAction` has no `query` field at all — `write.go`'s `ExecuteWrite`
  always issues `Requester.Do`/`DoForm` with a `nil` query — so there is no mechanism to templatize
  a query parameter on a write request, only `path_fields` (path segments) and the body. Only
  `create_dataset_record` (`POST /datasets/{dataset_id}/records`, the PLURAL endpoint, which takes
  no `recordId` at all — the API assigns/derives it from the submitted field map) is expressible.
  Updating or deleting a specific existing record by id is not possible through this bundle's
  writes today.
- **`submissions`' real per-record field shape (beyond the 5 documented metadata fields) is
  unconfirmed** — the OpenAPI spec's response schema for this endpoint is the bare `{"type":
  "object"}` placeholder, not a `$ref`'d schema like `datasets`/`groups`/`users` get. The 5 fields
  this bundle's schema declares (`submissionId`, `reviewStatus`, `formDefinitionVersion`,
  `submissionDate`, `completionDate`) are drawn from the endpoint's own documented `orderBy` enum
  (metadata fields the API can sort by, and therefore fields that certainly exist on the record),
  not from an inspected live response body.
- **Encrypted-form submission handling is out of scope** — the multipart `POST
  /forms/{formID}/submissions` variant (accepting an RSA private-key PEM file part to decrypt
  encrypted payloads) is excluded; handling a private-key file upload is both a materially
  different write shape than this dialect's JSON/form body model and security-sensitive (a private
  key is credential-shaped data that should not flow through an ordinary write action).
- **Bulk endpoints are out of scope** — `users/bulk/{file,json}` (create/update, multi-user in one
  request) and `datasets/{id}/records/upload` (bulk CSV record upload) are excluded; each is a
  materially different shape from this dialect's single-record write-body model, and the
  single-record equivalents (`create_user`/`update_user`/`create_dataset_record`) already cover the
  same underlying mutation one record at a time.
