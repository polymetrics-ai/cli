# Overview

Reads and writes ZapSign documents, signers, templates, and webhooks.

Readable streams: `documents`, `signers`, `templates`, `webhooks`.

Write actions: `create_document_from_template`, `cancel_document`, `delete_document`, `add_signer`,
`update_signer`, `remove_signer`, `create_webhook`, `delete_webhook`.

Service API documentation: https://docs.zapsign.com.br/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); ZapSign API token, sent as an 'Authorization: Token
  <api_token>' header.
- `base_url` (optional, string); default `https://api.zapsign.com.br/api/v1`; format `uri`; ZapSign
  API base URL.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.zapsign.com.br/api/v1`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token` using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/docs/`.

## Streams notes

Default pagination: single request; no pagination.

- `documents`: GET `/docs/` - records path `results`; computed output fields `id`.
- `signers`: GET `/signers/` - records path `results`; computed output fields `id`.
- `templates`: GET `/templates/` - records path `results`; computed output fields `id`.
- `webhooks`: GET `/webhooks/` - records path `results`.

## Write actions & risks

Overall write risk: external mutation: creates documents from templates, cancels/deletes documents,
adds/updates/removes signers, and manages webhook subscriptions that receive live document-event
data.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_document_from_template`: POST `/models/create-doc/` - kind `create`; body type `json`;
  required record fields `template_id`, `signers`; accepted fields `allow_refusal`, `creator_email`,
  `custom_metadata`, `folder_path`, `send_automatic_email`, `send_automatic_whatsapp`,
  `signature_order_active`, `signers`, `template_id`; risk: creates a new signable document from an
  existing template and notifies signers by email/WhatsApp if
  send_automatic_email/send_automatic_whatsapp is set; external mutation, approval required.
- `cancel_document`: POST `/docs/{{ record.token }}/cancel/` - kind `update`; body type `none`; path
  fields `token`; required record fields `token`; accepted fields `token`; risk: irreversibly
  interrupts an in-progress signature flow for a document; any signer who has not yet signed can no
  longer complete it.
- `delete_document`: DELETE `/docs/{{ record.token }}/delete/` - kind `delete`; body type `none`;
  path fields `token`; required record fields `token`; accepted fields `token`; missing records
  treated as success for status `404`; risk: soft-deletes a document, hiding it from the ZapSign web
  interface for end users while it remains readable via the API.
- `add_signer`: POST `/docs/{{ record.doc_token }}/add-signer/` - kind `create`; body type `json`;
  path fields `doc_token`; body fields `name`, `email`, `phone_country`, `phone_number`,
  `auth_mode`, `send_automatic_email`, `send_automatic_whatsapp`; required record fields
  `doc_token`, `name`; accepted fields `auth_mode`, `doc_token`, `email`, `name`, `phone_country`,
  `phone_number`, `send_automatic_email`, `send_automatic_whatsapp`; risk: adds a new signer to an
  existing document and, if send_automatic_email/send_automatic_whatsapp is set, immediately
  notifies them with a signing link.
- `update_signer`: POST `/signers/{{ record.token }}/` - kind `update`; body type `json`; path
  fields `token`; required record fields `token`; accepted fields `auth_mode`, `email`,
  `lock_email`, `lock_name`, `lock_phone`, `name`, `phone_country`, `phone_number`, `token`; risk:
  mutates an existing signer's contact details or authentication mode; only succeeds if the signer
  has not yet signed the document (ZapSign rejects the request once the signer has already signed,
  surfaced as an ordinary per-record write failure).
- `remove_signer`: DELETE `/signer/{{ record.token }}/remove/` - kind `delete`; body type `none`;
  path fields `token`; required record fields `token`; accepted fields `token`; missing records
  treated as success for status `404`; risk: permanently removes a signer from a document; this is
  irreversible, and re-adding the same person issues a brand new signing token/link.
- `create_webhook`: POST `/webhooks/` - kind `create`; body type `json`; required record fields
  `url`, `type`; accepted fields `enabled`, `type`, `url`; risk: registers a new outbound webhook
  that will POST live document-event data to an external URL of the caller's choosing; verify the
  target endpoint before enabling.
- `delete_webhook`: DELETE `/webhooks/{{ record.id }}/` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a webhook subscription; event delivery to its target
  URL stops immediately.

## Known limits

- API coverage includes 4 stream-backed endpoint group(s), 8 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4, duplicate_of=4, requires_elevated_scope=2.
