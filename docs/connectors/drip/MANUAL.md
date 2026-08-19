# pm connectors inspect drip

```text
NAME
  pm connectors inspect drip - Drip connector manual

SYNOPSIS
  pm connectors inspect drip
  pm connectors inspect drip --json
  pm credentials add <name> --connector drip [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Drip subscribers, campaigns, broadcasts, accounts, workflows, forms, tags, and webhooks, and writes subscriber/tag/broadcast/workflow/event/webhook mutations through the Drip REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id (required)
  base_url
  mode
  api_key (secret) (required)

ETL STREAMS
  subscribers:
    primary key: id
    cursor: created_at
    fields: created_at(string), custom_fields(object), email(string), id(string), ip_address(string), lifetime_value(integer), status(string), tags(array), time_zone(string), user_agent(string), utc_offset(integer)
  campaigns:
    primary key: id
    cursor: created_at
    fields: created_at(string), email_count(integer), from_email(string), from_name(string), id(string), name(string), status(string), subscriber_count(integer)
  broadcasts:
    primary key: id
    cursor: created_at
    fields: created_at(string), from_email(string), from_name(string), id(string), name(string), send_at(string), status(string), subject(string)
  accounts:
    primary key: id
    cursor: created_at
    fields: created_at(string), id(string), name(string)
  workflows:
    primary key: id
    fields: id(string), name(string), status(string)
  forms:
    primary key: id
    fields: button_text(string), confirmation_heading(string), confirmation_text(string), description(string), headline(string), href(string), id(string)
  webhooks:
    primary key: id
    fields: event_types(array), id(string), post_url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_or_update_subscriber:
    endpoint: POST /{{ config.account_id }}/subscribers
    required fields: subscribers
    risk: creates a new subscriber or updates an existing one matched by email; low-risk external mutation, no approval required. Drip's real API strictly requires the one-subscriber-per-write body to still be an ARRAY under the "subscribers" key ({"subscribers": [{...}]}), not a bare object — record_schema's "subscribers" field is itself that array, and body_fields copies it verbatim so the wire body matches Drip's real shape exactly; the engine's schema dialect has no minItems/maxItems keyword to mechanically cap it to exactly one element, so callers are expected to supply exactly one (this action is not the true batch endpoint).
  delete_subscriber:
    endpoint: DELETE /{{ config.account_id }}/subscribers/{{ record.id }}
    required fields: id
    risk: permanently removes a subscriber and their event/tag history; destructive, approval required
  unsubscribe_subscriber:
    endpoint: POST /{{ config.account_id }}/unsubscribes
    required fields: subscribers
    risk: unsubscribes the named email from ALL mailings in the account; stops all future campaign/broadcast/workflow sends to them, approval required
  apply_tag:
    endpoint: POST /{{ config.account_id }}/subscribers/{{ record.subscriber_id }}/tags
    required fields: subscriber_id, tags
    risk: applies one or more tags to a subscriber, potentially triggering any workflow with a matching tag-applied trigger; low-risk external mutation, no approval required
  remove_tag:
    endpoint: DELETE /{{ config.account_id }}/subscribers/{{ record.subscriber_id }}/tags/{{ record.tag }}
    required fields: subscriber_id, tag
    risk: removes a tag from a subscriber, potentially triggering any workflow with a matching tag-removed trigger; low-risk external mutation, no approval required
  record_event:
    endpoint: POST /{{ config.account_id }}/events
    required fields: events
    risk: records a custom behavioral event on a subscriber, which can trigger any workflow with a matching event trigger; low-risk external mutation, no approval required. Drip's real API requires the body to be an ARRAY under the "events" key even for one event; record_schema's "events" field is itself that array (callers are expected to supply exactly one element; this action is not the true batch endpoint).
  create_broadcast:
    endpoint: POST /{{ config.account_id }}/broadcasts
    required fields: broadcasts
    risk: creates a new draft single-email campaign (broadcast); low-risk in draft form (not sent), no approval required. Drip's real API requires the body to be an ARRAY under the "broadcasts" key even for one broadcast; record_schema's "broadcasts" field is itself that array (callers are expected to supply exactly one element; this action is not the true batch endpoint).
  update_broadcast:
    endpoint: PATCH /{{ config.account_id }}/broadcasts/{{ record.id }}
    required fields: id, broadcasts
    risk: mutates an existing draft broadcast's subject/content; Drip only allows updating broadcasts still in draft status, so this cannot alter an already-sent email; external mutation, approval required
  delete_broadcast:
    endpoint: DELETE /{{ config.account_id }}/broadcasts/{{ record.id }}
    required fields: id
    risk: permanently removes a broadcast; destructive if it was already sent (removes historical record, though delivered emails cannot be recalled), approval required
  activate_workflow:
    endpoint: POST /{{ config.account_id }}/workflows/{{ record.id }}/activate
    required fields: id
    risk: activates a paused workflow, resuming automated sends to everyone currently enrolled and allowing new triggers to enroll people; external mutation, approval required
  pause_workflow:
    endpoint: POST /{{ config.account_id }}/workflows/{{ record.id }}/pause
    required fields: id
    risk: pauses an active workflow, stopping automated sends to everyone currently enrolled and disabling new triggers; low-risk (reversible via activate_workflow), no approval required
  create_webhook:
    endpoint: POST /{{ config.account_id }}/webhooks
    required fields: webhooks
    risk: registers a new outbound webhook that will POST live subscriber/campaign event data to an external URL of the caller's choosing; verify the target endpoint before enabling. Drip's real API requires the body to be an ARRAY under the "webhooks" key even for one webhook; record_schema's "webhooks" field is itself that array (callers are expected to supply exactly one element; this action is not the true batch endpoint).
  delete_webhook:
    endpoint: DELETE /{{ config.account_id }}/webhooks/{{ record.id }}
    required fields: id
    risk: permanently removes a webhook subscription; stops future event delivery to its post_url; destructive, approval required

SECURITY
  read risk: external Drip API read of subscriber, campaign, broadcast, account, workflow, form, tag, and webhook data
  write risk: external Drip API mutation of subscribers, tags, broadcasts, workflows, custom events, and webhooks; delete_subscriber and delete_broadcast/delete_webhook are destructive and require approval
  approval: required for delete_subscriber, delete_broadcast, delete_webhook, and unsubscribe_subscriber; tag/event/workflow-state writes are lower-risk and do not require approval
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect drip

  # Inspect as structured JSON
  pm connectors inspect drip --json

AGENT WORKFLOW
  - Run pm connectors inspect drip before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
