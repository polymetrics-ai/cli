# pm connectors inspect customer-io

```text
NAME
  pm connectors inspect customer-io - Customer.io connector manual

SYNOPSIS
  pm connectors inspect customer-io
  pm connectors inspect customer-io --json
  pm credentials add <name> --connector customer-io [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Customer.io campaigns, newsletters, segments, broadcasts, activities, messages, exports, transactional templates, object types, reporting webhooks, sender identities, snippets, subscription channels/topics, workspaces, and collections; writes snippet/webhook/segment mutations and can send transactional email or trigger broadcasts, through the Customer.io App API.

ICON
  id: customer-io
  asset: icons/customer-io.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://customer.io/docs/api/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url (required)
  mode
  page_size
  app_api_key (secret) (required)

ETL STREAMS
  campaigns:
    primary key: id
    cursor: updated
    fields: active(boolean), created(integer), id(integer), name(string), state(string), type(string), updated(integer)
  newsletters:
    primary key: id
    cursor: updated
    fields: created(integer), id(integer), name(string), subject(string), type(string), updated(integer)
  segments:
    primary key: id
    cursor: updated
    fields: description(string), id(integer), name(string), state(string), type(string), updated(integer)
  broadcasts:
    primary key: id
    cursor: updated
    fields: active(boolean), created(integer), id(integer), name(string), state(string), type(string), updated(integer)
  activities:
    primary key: id
    cursor: timestamp
    fields: customer_id(string), customer_identifiers(object), data(object), delivery_id(string), delivery_type(string), id(string), timestamp(integer), type(string)
  messages:
    primary key: id
    fields: action_id(integer), broadcast_id(integer), campaign_id(integer), content_id(integer), created(integer), customer_id(string), failure_message(string), id(string), newsletter_id(integer), recipient(string), subject(string), type(string)
  exports:
    primary key: id
    cursor: updated_at
    fields: created_at(integer), description(string), downloads(integer), failed(boolean), id(integer), status(string), total(integer), type(string), updated_at(integer), user_email(string), user_id(integer)
  transactional:
    primary key: id
    cursor: updated_at
    fields: created_at(integer), description(string), hide_message_body(boolean), id(integer), link_tracking(boolean), name(string), open_tracking(boolean), queue_drafts(boolean), send_to_unsubscribed(boolean), updated_at(integer)
  object_types:
    primary key: id
    fields: enabled(boolean), icon(string), id(string), name(string), singular_name(string), singular_slug(string), slug(string)
  reporting_webhooks:
    primary key: id
    fields: disabled(boolean), endpoint(string), events(array), full_resolution(boolean), id(integer), name(string), with_content(boolean)
  sender_identities:
    primary key: id
    fields: address(string), auto_generated(boolean), deduplicate_id(string), email(string), id(integer), name(string), template_type(string)
  snippets:
    primary key: name
    cursor: updated_at
    fields: name(string), updated_at(integer), value(string)
  subscription_channels:
    primary key: id
    fields: description(string), id(integer), name(string), subscribed_by_default(boolean), type(string)
  subscription_topics:
    primary key: id
    fields: description(string), id(integer), identifier(string), name(string), subscribed_by_default(boolean)
  workspaces:
    primary key: id
    fields: billable_messages_sent(integer), id(integer), messages_sent(integer), name(string), object_types(integer), objects(integer), people(integer)
  collections:
    primary key: id
    cursor: updated_at
    fields: bytes(integer), created_at(integer), id(integer), name(string), rows(integer), updated_at(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_snippet:
    endpoint: POST /snippets
    required fields: name, value
    risk: external mutation; creates a reusable content snippet referenced by live messages/newsletters
  update_snippet:
    endpoint: PUT /snippets
    required fields: name, value
    risk: external mutation; overwrites the content of a live snippet, changing every message/newsletter that references it
  delete_snippet:
    endpoint: DELETE /snippets/{{ record.name }}
    required fields: name
    risk: external mutation; permanently removes a snippet; irreversible, breaks any message/newsletter still referencing it; approval required
  create_reporting_webhook:
    endpoint: POST /reporting_webhooks
    required fields: name, endpoint, events
    risk: external mutation; registers a new reporting webhook that will deliver live workspace event data to the given endpoint URL
  update_reporting_webhook:
    endpoint: PUT /reporting_webhooks/{{ record.id }}
    required fields: id, name, endpoint, events
    risk: external mutation; changes a live reporting webhook's target endpoint/event selection or enables/disables delivery
  delete_reporting_webhook:
    endpoint: DELETE /reporting_webhooks/{{ record.id }}
    required fields: id
    risk: external mutation; permanently removes a reporting webhook; event delivery to its target URL stops immediately; approval required
  create_manual_segment:
    endpoint: POST /segments
    required fields: segment
    risk: external mutation; creates a new manual segment in the live workspace
  delete_manual_segment:
    endpoint: DELETE /segments/{{ record.id }}
    required fields: id
    risk: external mutation; permanently removes a manual segment; irreversible, any campaign/newsletter targeting it loses that audience slice immediately; approval required
  send_email:
    endpoint: POST /send/email
    required fields: to
    risk: sends a live transactional email to the given recipient on the workspace's behalf; irreversible once delivered
  trigger_broadcast:
    endpoint: POST /campaigns/{{ record.broadcast_id }}/triggers
    required fields: broadcast_id
    risk: triggers a live API-triggered broadcast to its default audience; sends real messages to recipients, irreversible once delivered

SECURITY
  read risk: external Customer.io App API read of campaign/newsletter/segment/broadcast/activity/message/export/transactional/webhook/subscription/workspace metadata
  write risk: external mutation of live Customer.io workspace config (snippets/webhooks/segments) and live message sends (transactional email, broadcast triggers); irreversible once delivered; approval required
  approval: read: none; write: required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect customer-io

  # Inspect as structured JSON
  pm connectors inspect customer-io --json

AGENT WORKFLOW
  - Run pm connectors inspect customer-io before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
