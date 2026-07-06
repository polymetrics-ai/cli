# pm connectors inspect zendesk-talk

```text
NAME
  pm connectors inspect zendesk-talk - Zendesk Talk connector manual

SYNOPSIS
  pm connectors inspect zendesk-talk
  pm connectors inspect zendesk-talk --json
  pm credentials add <name> --connector zendesk-talk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zendesk Talk phone numbers, greetings, greeting categories, IVRs, and agent activity statistics through the Zendesk Talk (voice) REST API.

ICON
  asset: icons/zendesk-talk.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://support.zendesk.com/hc/en-us/sections/4405298889242-Developer-updates

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  access_token (secret)
  api_token (secret)
  email (secret)

ETL STREAMS
  phone_numbers:
    primary key: id
    cursor: created_at
    fields: country_code(), created_at(), display_number(), id(), nickname(), number(), recorded(), sms_enabled(), toll_free(), voice_enabled()
  greetings:
    primary key: id
    fields: active(), audio_name(), audio_url(), category_id(), default(), has_sub_settings(), id(), name()
  greeting_categories:
    primary key: id
    fields: id(), name()
  ivrs:
    primary key: id
    fields: id(), name(), phone_number_ids(), phone_number_names()
  agents_activity:
    primary key: agent_id
    fields: agent_id(), agent_state(), available_time(), avatar_url(), away_time(), call_status(), calls_accepted(), calls_denied(), calls_missed(), forwarding_number(), name(), online_time(), total_call_duration(), total_talk_time(), total_wrap_up_time(), via()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Zendesk Talk API read of phone number, greeting, and agent activity data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zendesk-talk

  # Inspect as structured JSON
  pm connectors inspect zendesk-talk --json

AGENT WORKFLOW
  - Run pm connectors inspect zendesk-talk before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
