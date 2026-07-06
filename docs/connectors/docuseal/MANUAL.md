# pm connectors inspect docuseal

```text
NAME
  pm connectors inspect docuseal - DocuSeal connector manual

SYNOPSIS
  pm connectors inspect docuseal
  pm connectors inspect docuseal --json
  pm credentials add <name> --connector docuseal [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads DocuSeal templates, submissions, and submitters, and writes submission/submitter/template mutations through the DocuSeal REST API.

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
  page_size
  template_id
  api_key (secret)

ETL STREAMS
  templates:
    primary key: id
    cursor: updated_at
    fields: archived_at(), author_id(), created_at(), external_id(), folder_name(), id(), name(), slug(), updated_at()
  submissions:
    primary key: id
    cursor: updated_at
    fields: archived_at(), audit_log_url(), combined_document_url(), completed_at(), created_at(), expire_at(), id(), name(), slug(), source(), status(), template_id(), template_name(), updated_at()
  submitters:
    primary key: id
    cursor: updated_at
    fields: completed_at(), created_at(), email(), external_id(), id(), name(), opened_at(), phone(), role(), sent_at(), slug(), status(), submission_id(), updated_at(), uuid()
  template_detail:
    primary key: id
    cursor: updated_at
    fields: archived_at(), author(), author_id(), created_at(), documents(), external_id(), fields(), folder_id(), folder_name(), id(), name(), preferences(), schema(), slug(), source(), submitters(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_submission:
    endpoint: POST /submissions
    risk: external mutation; dispatches a live signature-request email/SMS to every listed submitter unless send_email/send_sms are explicitly set false; approval required
  archive_submission:
    endpoint: DELETE /submissions/{{ record.id }}
    required fields: id
    risk: external mutation; archives a live DocuSeal submission (soft-delete, still recoverable via the DocuSeal UI); approval required
  update_submitter:
    endpoint: PUT /submitters/{{ record.id }}
    required fields: id
    risk: external mutation; overwrites a live DocuSeal submitter's pre-filled values/contact info, can re-send signature request notifications, and can force-mark the submitter completed/auto-signed; approval required
  update_template:
    endpoint: PUT /templates/{{ record.id }}
    required fields: id
    risk: external mutation; renames/moves/relabels a live DocuSeal template and can unarchive it; approval required
  archive_template:
    endpoint: DELETE /templates/{{ record.id }}
    required fields: id
    risk: external mutation; archives a live DocuSeal template (soft-delete, recoverable by unarchiving via update_template); approval required
  clone_template:
    endpoint: POST /templates/{{ record.id }}/clone
    required fields: id
    risk: external mutation; creates a new live DocuSeal template by cloning an existing one; approval required

SECURITY
  read risk: external DocuSeal API read of document template, submission, and submitter data
  write risk: external mutation; sends live signature requests, archives submissions/templates, and edits submitter/template records in DocuSeal
  approval: required for every write action; create_submission dispatches real signature-request emails/SMS to submitters unless send_email/send_sms are explicitly disabled
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect docuseal

  # Inspect as structured JSON
  pm connectors inspect docuseal --json

AGENT WORKFLOW
  - Run pm connectors inspect docuseal before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
