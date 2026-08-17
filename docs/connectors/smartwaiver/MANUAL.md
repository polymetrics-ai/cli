# pm connectors inspect smartwaiver

```text
NAME
  pm connectors inspect smartwaiver - Smartwaiver connector manual

SYNOPSIS
  pm connectors inspect smartwaiver
  pm connectors inspect smartwaiver --json
  pm credentials add <name> --connector smartwaiver [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Smartwaiver waivers, checkins, templates, published keys, user info, and account settings; sends prefill/SMS/webhook mutations through the Smartwaiver API.

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
  end_date
  page_size
  start_date
  api_key (secret)

ETL STREAMS
  waivers:
    primary key: waiverId
    fields: createdOn(), email(), expirationDate(), expired(), firstName(), lastName(), templateId(), title(), verified(), waiverId()
  checkins:
    primary key: checkinId
    fields: checkinId(), date(), dateSigned(), firstName(), lastName(), templateId(), waiverId()
  templates:
    primary key: templateId
    fields: kioskUrl(), publishedOn(), publishedVersion(), templateId(), title(), webUrl()
  published_keys:
    primary key: key
    fields: createdAt(), key(), label()
  user_info:
    primary key: username
    fields: email(), ipAddress(), signupDate(), username()
  settings:
    primary key: type
    fields: settings(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  set_webhook_config:
    endpoint: PUT /v4/webhooks/configure
    risk: changes where the account's near-real-time waiver-signed webhook notifications are delivered; approval required
  resend_webhook:
    endpoint: PUT /v4/webhooks/resend/{{ record.waiver_id }}
    required fields: waiver_id
    risk: re-triggers the new-waiver webhook delivery for a specific waiver (testing aid, heavily rate limited by Smartwaiver at 2/minute); approval required
  send_sms:
    endpoint: POST /v4/sms
    risk: sends an outbound SMS with a waiver-signing link to a real phone number (rate limited daily by Smartwaiver for anti-spam); approval required
  prefill_template:
    endpoint: POST /v4/templates/{{ record.template_id }}/prefill
    required fields: template_id
    risk: generates a prefilled waiver-signing link carrying real participant PII (name/DOB/address/custom fields); approval required

SECURITY
  read risk: read-only waiver/checkin/template/published-key/user/settings data from a connected Smartwaiver account
  write risk: configures the account's webhook delivery endpoint, resends a waiver's webhook notification, sends an outbound SMS waiver-signing link to a real phone number, and generates a prefilled waiver-signing link carrying participant PII
  approval: required for all 4 write actions (set_webhook_config, resend_webhook, send_sms, prefill_template); read is unapproved
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect smartwaiver

  # Inspect as structured JSON
  pm connectors inspect smartwaiver --json

AGENT WORKFLOW
  - Run pm connectors inspect smartwaiver before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
