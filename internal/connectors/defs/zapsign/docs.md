# Overview

ZapSign is an electronic signature platform. This bundle reads documents, signers, templates, and
webhooks from the ZapSign REST API (`GET {base_url}/docs/`, `/signers/`, `/templates/`,
`/webhooks/`), and writes document/signer/webhook mutations that the dialect can express without a
multipart file body. It migrates `internal/connectors/zapsign` (the hand-written legacy connector)
to a declarative Tier-1 bundle at capability parity, then expands it to ZapSign's full practical
API surface (Pass B); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Requires a single secret, `api_token` (ZapSign API token), sent as `Authorization: Token
<api_token>` — an `api_key_header` auth spec with `header: Authorization` and `prefix: "Token "`,
matching legacy's `connsdk.APIKeyHeader("Authorization", token, "Token ")` exactly
(`zapsign.go:118`). `base_url` defaults to `https://api.zapsign.com.br/api/v1`
(`zapsign.go:17`'s `defaultBaseURL`), materialized via `spec.json`'s `"default"` when unset.

## Streams notes

- **`documents`** (`GET /docs/`) and **`templates`** (`GET /templates/`) are `page_number`
  paginated (`page` query param, no size override — ZapSign's own default page size, 25 for
  documents / 20 for templates, is fixed server-side and documented, not caller-controlled), with
  a short-page stop when a page returns fewer records than the configured `page_size` threshold.
- **`signers`** (`GET /signers/`) is a legacy-parity carryover: ZapSign's published API reference
  does not document a top-level, account-wide "list all signers" endpoint (only
  `GET /docs/{token}/` embeds a document's own signers, and `GET /signers/{signer_token}` fetches
  one signer by token) — this stream reproduces the ORIGINAL hand-written legacy connector's own
  behavior (`zapsign.go`'s `streams["signers"]` hitting `/signers/`) verbatim rather than inventing
  a new capability; kept for read-parity with the legacy connector this bundle is migrating, not
  because the endpoint is independently ZapSign-documented. See Known limits.
- **`webhooks`** (`GET /webhooks/`) is new in this pass: lists registered outbound webhook
  subscriptions.
- Each stream's raw API record carries its identifier under the field name `token` (`documents`/
  `signers`/`templates`) or `id` (`webhooks`); a `computed_fields` rename
  (`"id": "{{ record.token }}"`) maps the `token`-shaped streams to the schema's `id` property,
  matching legacy's `mapDocument`/`mapSigner`/`mapTemplate` output field name exactly. Legacy's
  `signers`/`templates` mappers additionally fall back to a bare `id` field via a `first(token,
  id)` helper when `token` is absent from a record; this bundle models only the primary
  `token`-present case (ZapSign's documented wire shape always includes `token` for every one of
  these object types), since the declarative dialect has no ordered-fallback-across-two-fields
  primitive — see Known limits.

## Write actions & risks

- **`create_document_from_template`** (`POST /models/create-doc/`) creates a new signable document
  from an existing template, replacing the template's dynamic fields with the submitted data.
  External mutation; sends a real signing notification to every listed signer when
  `send_automatic_email`/`send_automatic_whatsapp` is set. Approval required.
- **`cancel_document`** (`POST /docs/{token}/cancel/`) interrupts an in-progress signature flow.
  Irreversible for any signer who has not yet signed. Approval required.
- **`delete_document`** (`DELETE /docs/{token}/delete/`) soft-deletes a document (hidden from the
  ZapSign web UI, still readable via the API). `delete.missing_ok_status: [404]` treats an
  already-deleted document as a successful, idempotent delete.
- **`add_signer`** (`POST /docs/{token}/add-signer/`) adds a new signer to an existing document;
  `body_fields` restricts the request body to signer-shaped fields only (`doc_token` stays
  path-only, never leaks into the body). Immediately notifies the signer when
  `send_automatic_email`/`send_automatic_whatsapp` is set.
- **`update_signer`** (`POST /signers/{token}/`) mutates an existing signer's contact
  details/authentication mode. ZapSign rejects this once the signer has already signed the
  document — surfaced as an ordinary per-record write failure, not a special engine case.
- **`remove_signer`** (`DELETE /signer/{token}/remove/`) permanently removes a signer; re-adding
  the same person issues a brand-new signing token/link. `delete.missing_ok_status: [404]`.
- **`create_webhook`** / **`delete_webhook`** (`POST`/`DELETE /webhooks/`) manage outbound event
  subscriptions; creating one starts delivering live document-event payloads to a caller-supplied
  URL — verify the target before enabling.

All write actions carry `body_type: "json"` except the two `delete`-kind actions
(`delete_document`, `remove_signer`), which are `body_type: "none"` pure path-parameterized
mutations, and `cancel_document`, which sends no body at all (`body_type: "none"` with no
`body_fields`).

## Known limits

- **`signers`/`templates`' `token`-absent id fallback is not modeled.** Legacy's `mapSigner`/
  `mapTemplate` compute `id` as `first(item["token"], item["id"])` — an ordered fallback that only
  matters if a record ever omits `token` and instead carries a bare `id` field. ZapSign's documented
  API always returns `token` as the canonical identifier for signers and templates, so this fallback
  is defensive/unreachable on the real wire shape; the declarative dialect has no
  ordered-multi-field-fallback primitive (only a single bare `{{ record.<path> }}` reference or a
  filter chain), so only the `token`-present case is expressed. Deliberately out-of-scope
  edge case, not a defect.
- **`signers` stream has no independently-documented top-level list endpoint.** See Streams notes
  above — this stream is a verbatim legacy-parity carryover (`GET /signers/`), not a capability
  verified against ZapSign's current published API reference. Retained rather than removed to avoid
  a read-parity regression for existing callers of the legacy connector; a future pass should
  confirm against a live ZapSign account whether this path still resolves, or replace it with a
  per-document `GET /docs/{token}/` signers sub-resource read (which would require the `fan_out`
  dialect over the `documents` stream's own tokens, not attempted here to keep this pass's Blast
  radius to additive changes only).
- **Document/template creation from an uploaded file is out of scope.** `POST /docs/`, `POST
  /docs/oneclick/create/`, `POST /templates/create-docx/`, and `PUT /templates/{id}/` all require a
  multipart/form-data file body; the engine's write dialect supports `json`/`form`/`none` bodies
  only. See `api_surface.json`'s `binary_payload` exclusions.
- **Partner/reseller endpoints are out of scope.** `POST /partners/create-account/` and `POST
  /partners/update-payment-status/` require a separate Partners API credential this bundle's
  `spec.json` does not model (`requires_elevated_scope`).
