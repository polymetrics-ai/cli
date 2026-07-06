# pm connectors inspect zapsign

```text
NAME
  pm connectors inspect zapsign - ZapSign connector manual

SYNOPSIS
  pm connectors inspect zapsign
  pm connectors inspect zapsign --json
  pm credentials add <name> --connector zapsign [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes ZapSign documents, signers, templates, and webhooks.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_token (secret)

ETL STREAMS
  documents:
    primary key: id
    fields: created_at(), id(), name(), status()
  signers:
    primary key: id
    fields: email(), id(), name()
  templates:
    primary key: id
    fields: created_at(), id(), name()
  webhooks:
    primary key: id
    fields: enabled(), id(), type(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_document_from_template:
    endpoint: POST /models/create-doc/
    risk: creates a new signable document from an existing template and notifies signers by email/WhatsApp if send_automatic_email/send_automatic_whatsapp is set; external mutation, approval required
  cancel_document:
    endpoint: POST /docs/{{ record.token }}/cancel/
    required fields: token
    risk: irreversibly interrupts an in-progress signature flow for a document; any signer who has not yet signed can no longer complete it
  delete_document:
    endpoint: DELETE /docs/{{ record.token }}/delete/
    required fields: token
    risk: soft-deletes a document, hiding it from the ZapSign web interface for end users while it remains readable via the API
  add_signer:
    endpoint: POST /docs/{{ record.doc_token }}/add-signer/
    required fields: doc_token
    optional fields: name, email, phone_country, phone_number, auth_mode, send_automatic_email, send_automatic_whatsapp
    risk: adds a new signer to an existing document and, if send_automatic_email/send_automatic_whatsapp is set, immediately notifies them with a signing link
  update_signer:
    endpoint: POST /signers/{{ record.token }}/
    required fields: token
    risk: mutates an existing signer's contact details or authentication mode; only succeeds if the signer has not yet signed the document (ZapSign rejects the request once the signer has already signed, surfaced as an ordinary per-record write failure)
  remove_signer:
    endpoint: DELETE /signer/{{ record.token }}/remove/
    required fields: token
    risk: permanently removes a signer from a document; this is irreversible, and re-adding the same person issues a brand new signing token/link
  create_webhook:
    endpoint: POST /webhooks/
    risk: registers a new outbound webhook that will POST live document-event data to an external URL of the caller's choosing; verify the target endpoint before enabling
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.id }}/
    required fields: id
    risk: permanently removes a webhook subscription; event delivery to its target URL stops immediately

SECURITY
  read risk: external ZapSign account read of documents, signers, templates, and webhooks
  write risk: external mutation: creates documents from templates, cancels/deletes documents, adds/updates/removes signers, and manages webhook subscriptions that receive live document-event data
  approval: required for all write actions; read access uses a read-only API token with no approval needed
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zapsign

  # Inspect as structured JSON
  pm connectors inspect zapsign --json

AGENT WORKFLOW
  - Run pm connectors inspect zapsign before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
