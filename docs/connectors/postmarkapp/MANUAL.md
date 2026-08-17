# pm connectors inspect postmarkapp

```text
NAME
  pm connectors inspect postmarkapp - Postmark App connector manual

SYNOPSIS
  pm connectors inspect postmarkapp
  pm connectors inspect postmarkapp --json
  pm credentials add <name> --connector postmarkapp [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Postmark server-token API resources including messages, bounces, templates, message streams, stats, webhooks, suppressions, and inbound rules; exposes server-token write actions for sends and resource mutations.

ICON
  asset: icons/postmark.svg
  source: official
  review_status: official_verified
  review_url: https://postmarkapp.com/developer

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  bounce_id
  bulk_request_id
  message_id
  message_stream_id
  mode
  template_id_or_alias
  webhook_id
  X-Postmark-Server-Token (secret)

ETL STREAMS
  outbound_messages:
    primary key: id
    cursor: received_at
    fields: from(), id(), received_at(), status(), subject(), to()
  inbound_messages:
    primary key: id
    fields: from(), id(), status(), subject(), to()
  current_server:
    primary key: ID
    fields: ID()
  bulk_email_status:
    primary key: Id
    fields: Id()
  delivery_stats:
  bounces:
    primary key: ID
    fields: ID()
  bounce:
    primary key: ID
    fields: ID()
  bounce_dump:
  templates:
    primary key: TemplateId
    fields: TemplateId()
  template:
    primary key: TemplateId
    fields: TemplateId()
  message_streams:
    primary key: ID
    fields: ID()
  message_stream:
    primary key: ID
    fields: ID()
  outbound_message_details:
    primary key: MessageID
    fields: MessageID()
  outbound_message_dump:
  inbound_message_details:
    primary key: MessageID
    fields: MessageID()
  outbound_message_opens:
  outbound_message_opens_by_message:
  outbound_message_clicks:
  outbound_message_clicks_by_message:
  stats_outbound:
  stats_outbound_sends:
  stats_outbound_bounces:
  stats_outbound_spam:
  stats_outbound_tracked:
  stats_outbound_opens:
  stats_outbound_open_platforms:
  stats_outbound_email_clients:
  stats_outbound_clicks:
  stats_outbound_click_browser_families:
  stats_outbound_click_platforms:
  stats_outbound_click_location:
  inbound_rule_triggers:
    primary key: ID
    fields: ID()
  webhooks:
    primary key: ID
    fields: ID()
  webhook:
    primary key: ID
    fields: ID()
  suppressions:

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  send_email:
    endpoint: POST /email
    risk: sends a live Postmark email; approval required
  send_bulk_email:
    endpoint: POST /email/bulk
    risk: submits a live Postmark bulk email request; approval required
  send_email_with_template:
    endpoint: POST /email/withTemplate
    risk: sends a live Postmark template email; approval required
  edit_current_server:
    endpoint: PUT /server
    risk: mutates the current Postmark server settings; approval required
  activate_bounce:
    endpoint: PUT /bounces/{{ record.bounce_id }}/activate
    required fields: bounce_id
    risk: reactivates a bounced email address in Postmark; approval required
  create_template:
    endpoint: POST /templates
    risk: creates a Postmark template; approval required
  edit_template:
    endpoint: PUT /templates/{{ record.template_id_or_alias }}
    required fields: template_id_or_alias
    risk: updates a Postmark template; approval required
  delete_template:
    endpoint: DELETE /templates/{{ record.template_id_or_alias }}
    required fields: template_id_or_alias
    risk: deletes a Postmark template; destructive external mutation
  validate_template:
    endpoint: POST /templates/validate
    risk: validates Postmark template content; no persistent mutation expected but still invokes the external API
  create_message_stream:
    endpoint: POST /message-streams
    risk: creates a Postmark message stream; approval required
  edit_message_stream:
    endpoint: PATCH /message-streams/{{ record.message_stream_id }}
    required fields: message_stream_id
    risk: updates a Postmark message stream; approval required
  archive_message_stream:
    endpoint: POST /message-streams/{{ record.message_stream_id }}/archive
    required fields: message_stream_id
    risk: archives a Postmark message stream; approval required
  unarchive_message_stream:
    endpoint: POST /message-streams/{{ record.message_stream_id }}/unarchive
    required fields: message_stream_id
    risk: unarchives a Postmark message stream; approval required
  bypass_inbound_message:
    endpoint: PUT /messages/inbound/{{ record.message_id }}/bypass
    required fields: message_id
    risk: bypasses inbound blocking for one Postmark message; approval required
  retry_inbound_message:
    endpoint: PUT /messages/inbound/{{ record.message_id }}/retry
    required fields: message_id
    risk: retries processing for one inbound Postmark message; approval required
  create_inbound_rule_trigger:
    endpoint: POST /triggers/inboundrules
    risk: creates an inbound rule trigger; approval required
  delete_inbound_rule_trigger:
    endpoint: DELETE /triggers/inboundrules/{{ record.trigger_id }}
    required fields: trigger_id
    risk: deletes an inbound rule trigger; destructive external mutation
  create_webhook:
    endpoint: POST /webhooks
    risk: creates a Postmark webhook endpoint; approval required
  edit_webhook:
    endpoint: PUT /webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: updates a Postmark webhook endpoint; approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: deletes a Postmark webhook endpoint; destructive external mutation
  create_suppression:
    endpoint: POST /message-streams/{{ record.message_stream_id }}/suppressions
    required fields: message_stream_id
    risk: adds one or more suppressions to a Postmark message stream; approval required
  delete_suppression:
    endpoint: POST /message-streams/{{ record.message_stream_id }}/suppressions/delete
    required fields: message_stream_id
    risk: removes one or more suppressions from a Postmark message stream; approval required

SECURITY
  read risk: external Postmark API read of message, bounce, template, stream, stats, webhook, suppression, and inbound-rule data
  write risk: sends emails and mutates Postmark templates, message streams, server settings, webhooks, inbound rules, suppressions, and inbound message processing state
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect postmarkapp

  # Inspect as structured JSON
  pm connectors inspect postmarkapp --json

AGENT WORKFLOW
  - Run pm connectors inspect postmarkapp before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
