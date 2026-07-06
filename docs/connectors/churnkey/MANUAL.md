# pm connectors inspect churnkey

```text
NAME
  pm connectors inspect churnkey - Churnkey connector manual

SYNOPSIS
  pm connectors inspect churnkey
  pm connectors inspect churnkey --json
  pm credentials add <name> --connector churnkey [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Churnkey cancel-flow sessions and aggregated session counts through the Churnkey Data API, and sends usage/billing events and customer attribute updates through the Churnkey Event Tracking API.

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
  app_id
  base_url
  api_key (secret)

ETL STREAMS
  sessions:
    primary key: _id
    cursor: created_at
    fields: _id(), aborted(), abtest(), accepted_offer(), blueprint_id(), canceled(), created_at(), customer(), customer_billing_interval(), customer_email(), customer_id(), customer_plan_id(), discount_cooldown_applied(), feedback(), mode(), offer_type(), org(), provider(), segment_id(), survey_choice_id(), survey_choice_value(), survey_id(), updated_at()
  session_aggregation:
    fields: aborted(), billing_interval(), canceled(), count(), month(), offer_type(), plan_id(), save_type(), trial()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_event:
    endpoint: POST /v1/api/events/new
    risk: external mutation; records a usage/billing event against a Churnkey customer, influencing cancel-flow offer targeting; approval required
  update_customer:
    endpoint: POST /v1/api/events/customer-update
    risk: external mutation; overwrites a Churnkey customer's tracked attributes used to drive cancel-flow segmentation and offer eligibility; approval required
  set_billing_users:
    endpoint: POST /v1/api/events/customer-update/set-users
    risk: external mutation; overwrites which users on a Churnkey customer account receive Payment Recovery billing-contact emails; approval required

SECURITY
  read risk: external Churnkey API read of cancel-flow session and customer data
  write risk: external mutation of Churnkey customer event/attribute data used to drive cancel-flow targeting; approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect churnkey

  # Inspect as structured JSON
  pm connectors inspect churnkey --json

AGENT WORKFLOW
  - Run pm connectors inspect churnkey before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
