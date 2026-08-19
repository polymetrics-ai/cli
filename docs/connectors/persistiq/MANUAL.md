# pm connectors inspect persistiq

```text
NAME
  pm connectors inspect persistiq - PersistIQ connector manual

SYNOPSIS
  pm connectors inspect persistiq
  pm connectors inspect persistiq --json
  pm credentials add <name> --connector persistiq [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PersistIQ leads, users, campaigns, mailboxes, activities, accounts, DNC domains, events, lead fields/statuses, tags, webhook plugin settings, and per-campaign leads/replies, and creates/updates leads and campaigns, adds/removes campaign leads, replies to campaign messages, and adds DNC domains, through v1 REST endpoints.

ICON
  id: persistiq
  asset: icons/persistiq.svg
  source: official
  review_status: official_verified
  review_url: https://persistiq.com/api-docs/index.html

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
  leads:
    primary key: id
    fields: email(string), id(string), name(string), status(string), updated_at(string)
  users:
    primary key: id
    fields: email(string), id(string), name(string), status(string)
  campaigns:
    primary key: id
    fields: email(string), id(string), name(string), status(string)
  mailboxes:
    primary key: id
    fields: email(string), id(string), name(string), status(string)
  activities:
    primary key: id
    fields: email(string), id(string), name(string), status(string)
  accounts:
    primary key: id
    fields: email(string), id(string), name(string), status(string)
  dnc_domains:
    primary key: id
    fields: id(string), name(string)
  events:
    primary key: id
    fields: created_at(string), data(object), event_type(string), id(string)
  lead_fields:
    primary key: id
    fields: id(string), label(string), name(string)
  lead_statuses:
    primary key: id
    fields: id(string), name(string)
  tags:
    primary key: id
    fields: id(string), name(string)
  webhook_plugin:
    fields: post_email_opened(boolean), post_email_opened_url(string), post_email_reply(boolean), post_email_reply_url(string), post_new_prospect(boolean), post_new_prospect_url(string), post_updated_prospect(boolean), post_updated_prospect_url(string), raw_events(boolean), raw_events_url(string)
  campaign_leads:
    primary key: id
    fields: campaign_id(string), id(string), lead(object), mailbox_id(string)
  campaign_replies:
    primary key: id
    fields: body(string), campaign_id(string), cc_emails(array), from_email(string), id(string), kind(string), lead_id(string), preview(string), sent_at(string), sentiment(string), step_message_id(string), subject(string), to_emails(array)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  update_lead:
    endpoint: PATCH /v1/leads/{{ record.id }}
    required fields: id
    risk: external mutation of an existing PersistIQ lead's fields; changing status/status_id/owner_id can move a lead into or out of active outbound-sequence automation depending on the target account's own campaign rules; approval required
  create_campaign:
    endpoint: POST /v1/campaigns
    required fields: campaign_name, owner_id
    risk: creates a new outbound-email campaign in the target PersistIQ account; approval required
  duplicate_campaign:
    endpoint: POST /v1/campaigns/duplicate
    required fields: campaign_id, owner_id
    risk: duplicates an existing campaign (including its steps/sequence) into a new campaign in the target account; approval required
  add_lead_to_campaign:
    endpoint: POST /v1/campaigns/{{ record.campaign_id }}/leads
    required fields: campaign_id
    risk: enrolls a lead into a live outbound-email campaign; the lead may start receiving automated outreach immediately depending on campaign schedule/state; approval required
  remove_lead_from_campaign:
    endpoint: DELETE /v1/campaigns/{{ record.campaign_id }}/leads/{{ record.id }}
    required fields: campaign_id, id
    risk: removes a lead from a live outbound-email campaign, stopping any further scheduled automated outreach to it in that sequence; approval required
  reply_to_campaign_message:
    endpoint: POST /v1/campaigns/{{ record.campaign_id }}/replies
    required fields: campaign_id, inbox_message_id, body
    risk: sends a real outbound email reply on behalf of the campaign's mailbox owner; irreversible once delivered; approval required
  add_dnc_domain:
    endpoint: POST /v1/dnc_domains
    required fields: name
    risk: adds a domain to the account's Do-Not-Contact list; blocks future outreach to that domain account-wide; approval required

SECURITY
  read risk: external PersistIQ API read of leads, users, campaigns, mailboxes, activities, accounts, DNC domains, events, lead fields/statuses, tags, webhook plugin settings, and per-campaign leads/replies
  write risk: external mutation of PersistIQ leads and campaigns: update_lead can move a lead into or out of active outbound-sequence automation; create_campaign/duplicate_campaign create new live campaigns; add_lead_to_campaign enrolls a lead into automated outreach immediately depending on campaign state; remove_lead_from_campaign stops scheduled outreach to a lead; reply_to_campaign_message sends a real outbound email on behalf of the mailbox owner; add_dnc_domain blocks future outreach to a domain account-wide
  approval: required; every write can trigger or halt outbound-email automation, send a real email, or change account-wide contact policy outside this connector's control depending on the target PersistIQ account's own configuration
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect persistiq

  # Inspect as structured JSON
  pm connectors inspect persistiq --json

AGENT WORKFLOW
  - Run pm connectors inspect persistiq before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
