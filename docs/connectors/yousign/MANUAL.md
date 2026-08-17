# pm connectors inspect yousign

```text
NAME
  pm connectors inspect yousign - Yousign connector manual

SYNOPSIS
  pm connectors inspect yousign
  pm connectors inspect yousign --json
  pm credentials add <name> --connector yousign [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Yousign signature requests, contacts, documents, webhooks, templates, users, and workflow sessions through the Yousign REST API.

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
  limit
  mode
  api_key (secret)

ETL STREAMS
  signature_requests:
    primary key: id
    cursor: updated_at
    fields: id(), name(), status(), updated_at()
  contacts:
    primary key: id
    cursor: updated_at
    fields: id(), name(), status(), updated_at()
  documents:
    primary key: id
    cursor: updated_at
    fields: id(), name(), status(), updated_at()
  webhooks:
    primary key: id
    cursor: updated_at
    fields: created_at(), enabled(), endpoint(), id(), name(), subscribed_events(), updated_at()
  templates:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), workspace_id()
  users:
    primary key: id
    cursor: created_at
    fields: created_at(), email(), id(), is_active(), name(), role(), status()
  workflow_sessions:
    primary key: id
    fields: created_at(), id(), status(), workflow_template_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_signature_request:
    endpoint: POST /signature_requests
    risk: creates a new draft signature request (no documents/signers attached yet); external mutation, approval required
  activate_signature_request:
    endpoint: POST /signature_requests/{{ record.id }}/activate
    required fields: id
    risk: activates a draft signature request, taking it out of draft status; if delivery_mode is not none this immediately notifies approvers/signers/followers by email; external mutation, approval required
  cancel_signature_request:
    endpoint: POST /signature_requests/{{ record.id }}/cancel
    required fields: id
    risk: irreversibly cancels a signature request in approval or ongoing status; external mutation, approval required
  create_contact:
    endpoint: POST /contacts
    risk: creates a new saved contact profile; external mutation, approval required
  update_contact:
    endpoint: PATCH /contacts/{{ record.id }}
    required fields: id
    risk: mutates an existing contact's profile fields; external mutation, approval required
  delete_contact:
    endpoint: DELETE /contacts/{{ record.id }}
    required fields: id
    risk: irreversibly deletes a saved contact profile; external mutation, approval required
  create_webhook:
    endpoint: POST /webhooks
    risk: registers a new webhook subscription that will receive real-time event notifications at an external endpoint; external mutation, approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.id }}
    required fields: id
    risk: irreversibly deletes a registered webhook subscription, silently stopping the caller's own event delivery; external mutation, approval required

SECURITY
  read risk: external Yousign API read of signature request, contact, document, webhook, template, user, and workflow session data
  write risk: external mutation of e-signature-critical Yousign records: signature request creation/activation (may immediately notify signers by email) and cancellation (irreversible), contact create/update/delete, and webhook subscription create/delete
  approval: required for all write actions; cancel_signature_request, delete_contact, and delete_webhook require explicit destructive confirmation
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect yousign

  # Inspect as structured JSON
  pm connectors inspect yousign --json

AGENT WORKFLOW
  - Run pm connectors inspect yousign before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
