---
name: pm-persistiq
description: PersistIQ connector knowledge and safe action guide.
---

# pm-persistiq

## Purpose

Reads PersistIQ leads, users, campaigns, mailboxes, activities, accounts, DNC domains, events, lead fields/statuses, tags, webhook plugin settings, and per-campaign leads/replies, and creates/updates leads and campaigns, adds/removes campaign leads, replies to campaign messages, and adds DNC domains, through v1 REST endpoints.

## Icon

- asset: icons/persistiq.svg
- source: official
- review_status: official_verified
- review_url: https://persistiq.com/api-docs/index.html

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret)

## ETL Streams

- leads:
  - primary key: id
  - fields: email(), id(), name(), status(), updated_at()
- users:
  - primary key: id
  - fields: email(), id(), name(), status()
- campaigns:
  - primary key: id
  - fields: email(), id(), name(), status()
- mailboxes:
  - primary key: id
  - fields: email(), id(), name(), status()
- activities:
  - primary key: id
  - fields: email(), id(), name(), status()
- accounts:
  - primary key: id
  - fields: email(), id(), name(), status()
- dnc_domains:
  - primary key: id
  - fields: id(), name()
- events:
  - primary key: id
  - fields: created_at(), data(), event_type(), id()
- lead_fields:
  - primary key: id
  - fields: id(), label(), name()
- lead_statuses:
  - primary key: id
  - fields: id(), name()
- tags:
  - primary key: id
  - fields: id(), name()
- webhook_plugin:
  - fields: post_email_opened(), post_email_opened_url(), post_email_reply(), post_email_reply_url(), post_new_prospect(), post_new_prospect_url(), post_updated_prospect(), post_updated_prospect_url(), raw_events(), raw_events_url()
- campaign_leads:
  - primary key: id
  - fields: campaign_id(), id(), lead(), mailbox_id()
- campaign_replies:
  - primary key: id
  - fields: body(), campaign_id(), cc_emails(), from_email(), id(), kind(), lead_id(), preview(), sent_at(), sentiment(), step_message_id(), subject(), to_emails()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- update_lead:
  - endpoint: PATCH /v1/leads/{{ record.id }}
  - required fields: id
  - risk: external mutation of an existing PersistIQ lead's fields; changing status/status_id/owner_id can move a lead into or out of active outbound-sequence automation depending on the target account's own campaign rules; approval required
- create_campaign:
  - endpoint: POST /v1/campaigns
  - risk: creates a new outbound-email campaign in the target PersistIQ account; approval required
- duplicate_campaign:
  - endpoint: POST /v1/campaigns/duplicate
  - risk: duplicates an existing campaign (including its steps/sequence) into a new campaign in the target account; approval required
- add_lead_to_campaign:
  - endpoint: POST /v1/campaigns/{{ record.campaign_id }}/leads
  - required fields: campaign_id
  - risk: enrolls a lead into a live outbound-email campaign; the lead may start receiving automated outreach immediately depending on campaign schedule/state; approval required
- remove_lead_from_campaign:
  - endpoint: DELETE /v1/campaigns/{{ record.campaign_id }}/leads/{{ record.id }}
  - required fields: campaign_id, id
  - risk: removes a lead from a live outbound-email campaign, stopping any further scheduled automated outreach to it in that sequence; approval required
- reply_to_campaign_message:
  - endpoint: POST /v1/campaigns/{{ record.campaign_id }}/replies
  - required fields: campaign_id
  - risk: sends a real outbound email reply on behalf of the campaign's mailbox owner; irreversible once delivered; approval required
- add_dnc_domain:
  - endpoint: POST /v1/dnc_domains
  - risk: adds a domain to the account's Do-Not-Contact list; blocks future outreach to that domain account-wide; approval required

## Security

- read risk: external PersistIQ API read of leads, users, campaigns, mailboxes, activities, accounts, DNC domains, events, lead fields/statuses, tags, webhook plugin settings, and per-campaign leads/replies
- write risk: external mutation of PersistIQ leads and campaigns: update_lead can move a lead into or out of active outbound-sequence automation; create_campaign/duplicate_campaign create new live campaigns; add_lead_to_campaign enrolls a lead into automated outreach immediately depending on campaign state; remove_lead_from_campaign stops scheduled outreach to a lead; reply_to_campaign_message sends a real outbound email on behalf of the mailbox owner; add_dnc_domain blocks future outreach to a domain account-wide
- approval: required; every write can trigger or halt outbound-email automation, send a real email, or change account-wide contact policy outside this connector's control depending on the target PersistIQ account's own configuration
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect persistiq
```

### Inspect as structured JSON

```bash
pm connectors inspect persistiq --json
```

## Agent Rules

- Run pm connectors inspect persistiq before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
