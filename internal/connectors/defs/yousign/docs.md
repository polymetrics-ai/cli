# Overview

Yousign (also branded "Youtrust") is an e-signature and identity-verification platform. This
bundle reads Yousign signature requests, contacts, documents, webhooks, templates, users, and
workflow sessions through the Yousign REST API v3 (`GET {base_url}/signature_requests|contacts|
documents|webhooks|templates|users|workflow_sessions`), and writes signature-request lifecycle,
contact, and webhook mutations. It migrates `internal/connectors/yousign` (the hand-written
connector, which was read-only with only the 3 legacy-parity streams); the legacy package stays
registered and unchanged until wave6's registry flip. Pass B full-surface expansion
(`docs/migration/conventions.md`) added the 4 new streams and 8 write actions and flipped
`capabilities.write` to `true` — see `api_surface.json` for the full documented Yousign API v3
surface (~135 operations) and why everything not covered here is excluded.

## Auth setup

Provide `api_key` (secret) for Bearer auth (`Authorization: Bearer <api_key>`), matching legacy's
`connsdk.Bearer(token)`. Never logged.

## Streams notes

The 3 legacy-parity streams (`signature_requests`, `contacts`, `documents`) share the same shape:
`GET` against the Yousign list endpoint, records at `data`, primary key `["id"]`, cursor field
`updated_at`, and passthrough projection. No pagination is declared for these 3 — legacy issues a
single unpaginated request per stream and emits every record in the response's `data` array after
copying all raw fields, so this bundle's `streams.json` omits any `pagination` block (defaulting to
`none`) and preserves the raw records, matching legacy exactly.

An optional `limit` config value is sent as the `limit` query parameter on every legacy-parity
stream's read request when set (`stream.Query`'s `omit_when_absent: true` opt-in dialect), matching
legacy's `baseQuery` (`yousign.go:176-182`). `updated_at` uses the same fallback as legacy's
`mapRecord`: keep a raw `updated_at` when present, otherwise fill it from `created_at`. `contacts`
and `documents` also preserve legacy's `name` fallbacks (`email` and `filename`, respectively)
when the raw `name` field is absent.

**Pass B new streams**: `webhooks`, `templates`, `users`, `workflow_sessions` — the top-level list
resources with a plain GET-list shape and no legacy behavior to match (new coverage, authored fresh
from the live API reference, not ported). `webhooks` returns a bare JSON array at the response root
(no `data`/`meta` envelope, confirmed against the live reference) — `records.path: ""` selects the
whole decoded body directly; `computed_fields` renames the wire's `description` to `name` (webhook
records identify themselves by their target endpoint description, not a dedicated name field), and
the wire's own `updated_at` field survives straight through schema projection as the cursor field,
no rename needed. `templates`, `users`, and `workflow_sessions` all use the real Yousign v3
cursor-pagination convention (`GET ...?after=<token>`, response envelope
`{"meta":{"next_cursor":...},"data":[...]}` — confirmed against 3 separate live reference pages),
wired as `pagination: {type: cursor, cursor_param: after, token_path: meta.next_cursor}` with no
`stop_path` declared (a `null`/absent `next_cursor` is already the correct stop signal via the
paginator's default stop-on-empty-token behavior, §3). `users`' `name` is a `computed_fields`
concatenation of the wire's `first_name`/`last_name` fields (no single wire field carries a display
name).

## Write actions & risks

8 write actions, all requiring approval; `cancel_signature_request`, `delete_contact`, and
`delete_webhook` additionally require explicit destructive confirmation (`confirm: "destructive"`):

- **`create_signature_request`** (`POST /signature_requests`) — creates a new draft signature
  request (no documents/signers attached yet). Requires `name` and `delivery_mode`
  (`email`/`none`).
- **`activate_signature_request`** (`POST /signature_requests/{id}/activate`) — activates a draft
  signature request; if `delivery_mode` is not `none` this immediately triggers email
  notifications to approvers/signers/followers. No request body.
- **`cancel_signature_request`** (`POST /signature_requests/{id}/cancel`) — irreversibly cancels a
  signature request in `approval`/`ongoing` status. Requires `reason` (Yousign's closed
  cancellation-reason enum). Destructive.
- **`create_contact`** (`POST /contacts`) — creates a new saved contact profile. Requires
  `first_name`, `last_name`, `email`, `locale`.
- **`update_contact`** (`PATCH /contacts/{id}`) — mutates an existing contact's profile fields.
- **`delete_contact`** (`DELETE /contacts/{id}`) — irreversibly deletes a saved contact profile.
  Destructive. Idempotent (`missing_ok_status: [404]`).
- **`create_webhook`** (`POST /webhooks`) — registers a new webhook subscription. Requires
  `endpoint`, `subscribed_events`, `scopes`, `sandbox`, `auto_retry`, `enabled`.
- **`delete_webhook`** (`DELETE /webhooks/{id}`) — irreversibly deletes a registered webhook
  subscription, silently stopping the caller's own event delivery. Destructive. Idempotent.

Every signer/approver/document/field/follower/consent-request/document-request nested mutation
under a signature request, every Verification-family create, and every Electronic Seal/Archiving
mutation are excluded this pass — see `api_surface.json` for the full per-endpoint reasoning; the 8
actions above were chosen as the highest-value, cleanly-single-request mutations expressible
without a fan_out-scoped nested-resource create or a multipart/binary body.

## Known limits

- **Check does not send legacy's `limit=1` query parameter.** Legacy's `Check` sends a literal
  `limit=1` on its underlying `GET /signature_requests` probe (`yousign.go:63`, distinct from
  Read's config-driven `limit`), to keep the connectivity check cheap. The engine's declarative
  `check` block (`RequestSpec`) supports only `method`+`path`, no query parameters — this bundle's
  check therefore requests `/signature_requests` with no `limit` param at all. This does not change
  accepted-input behavior (the same endpoint, same auth, same 200/401/403 outcomes); it may return
  a marginally larger response body than legacy's probe, which is immaterial since `Check` only
  inspects the status code / error, never the body. Acceptable per the parity-deviation meta-rule
  (`docs/migration/conventions.md` §5): never changes emitted record data, only Check's own request
  shape.
- **The legacy-parity `documents` stream targets a deprecated/removed top-level endpoint.** Legacy
  reads `GET /documents` directly, and this bundle's `documents` stream reproduces that exact
  request for parity. Live research against the current Yousign v3 API reference found no
  top-level `GET /documents` list endpoint at all — documents now live exclusively under
  `GET /signature_requests/{signatureRequestId}/documents` (a per-signature-request sub-resource),
  and the one remaining top-level `/documents` reference is `POST /documents`, explicitly marked
  `[DEPRECATED] ... do not use` by Yousign's own docs. This is a pre-existing legacy behavior this
  migration ports unchanged (parity meta-rule: reproduce legacy's own request shape, never silently
  "fix" it mid-migration) — a real Yousign v3 tenant likely receives an error or empty response
  from this stream's live request today. Migrating to the real per-signature-request sub-resource
  requires the `fan_out` dialect (list every `signature_requests` id, then repeat the documents
  request once per id) — deferred past this pass as a materially larger scope increase than a plain
  list stream; tracked in `api_surface.json`'s `excluded` entry for
  `GET /signature_requests/{signatureRequestId}/documents`.
- Full Yousign API surface (~135 operations; every signer/approver/field/follower/consent-request
  sub-resource mutation, the 8 Verification-family resources, Electronic Seal, Archiving,
  Consumption reporting, Workspaces/Users tenant-administration) is out of scope for this pass; see
  `api_surface.json`'s per-endpoint `excluded` entries for the specific reason each was left out.
- `create_signature_request`/`activate_signature_request`/`cancel_signature_request`/
  `create_contact`/`update_contact`/`delete_contact`/`create_webhook`/`delete_webhook` are new Pass
  B write actions with no legacy Go counterpart to match against (legacy was read-only) — their
  `record_schema` and risk classification are authored fresh from the live Yousign API reference,
  not ported from an existing implementation.
