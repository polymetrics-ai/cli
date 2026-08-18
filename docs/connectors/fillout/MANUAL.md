# pm connectors inspect fillout

```text
NAME
  pm connectors inspect fillout - Fillout connector manual

SYNOPSIS
  pm connectors inspect fillout
  pm connectors inspect fillout --json
  pm credentials add <name> --connector fillout [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Fillout forms and manages webhooks/submission deletion through the Fillout REST API. Question definitions and submissions LIST remain on the legacy connector pending an engine fan_out fallback-mode gap (see docs.md Known limits).

ICON
  id: simple-icons-fillout
  asset: icons/simple-icons/fillout.svg
  title: Fillout
  simple_icon_slug: fillout
  simple_icon_hex: FFC738
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Fillout
  match: exact-name-or-slug
  matched_by: fillout

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret) (required)

ETL STREAMS
  forms:
    primary key: id
    fields: id(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_webhook:
    endpoint: POST /webhook/create
    required fields: formId, url
    risk: registers a new outbound webhook subscription that will POST live form-submission data to an external URL; external mutation, approval required
  remove_webhook:
    endpoint: POST /webhook/delete
    required fields: webhookId
    risk: permanently removes a webhook subscription; event delivery to its target URL stops immediately
  delete_submission_by_id:
    endpoint: DELETE /forms/{{ record.form_id }}/submissions/{{ record.submission_id }}
    required fields: form_id, submission_id
    risk: permanently deletes a single form response; irreversible, approval required

SECURITY
  read risk: external Fillout API read of form metadata
  write risk: creates/removes outbound webhook subscriptions and deletes individual form submissions; external mutation, approval required
  approval: required for write actions; none for read
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect fillout

  # Inspect as structured JSON
  pm connectors inspect fillout --json

AGENT WORKFLOW
  - Run pm connectors inspect fillout before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
