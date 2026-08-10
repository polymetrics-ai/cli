---
name: pm-persistiq
description: PersistIQ connector knowledge and safe action guide.
---

# pm-persistiq

## Purpose

Reads PersistIQ leads, users, campaigns, mailboxes, activities, accounts, DNC domains, events, lead fields/statuses, tags, webhook plugin settings, and per-campaign leads/replies, and creates/updates leads and campaigns, adds/removes campaign leads, replies to campaign messages, and adds DNC domains, through v1 REST endpoints.

## Icon

- id: persistiq
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
  - fields: email(string), id(string), name(string), status(string), updated_at(string)
- users:
  - primary key: id
  - fields: email(string), id(string), name(string), status(string)
- campaigns:
  - primary key: id
  - fields: email(string), id(string), name(string), status(string)
- mailboxes:
  - primary key: id
  - fields: email(string), id(string), name(string), status(string)
- activities:
  - primary key: id
  - fields: email(string), id(string), name(string), status(string)
- accounts:
  - primary key: id
  - fields: email(string), id(string), name(string), status(string)
- dnc_domains:
  - primary key: id
  - fields: id(string), name(string)
- events:
  - primary key: id
  - fields: created_at(string), data(object), event_type(string), id(string)
- lead_fields:
  - primary key: id
  - fields: id(string), label(string), name(string)
- lead_statuses:
  - primary key: id
  - fields: id(string), name(string)
- tags:
  - primary key: id
  - fields: id(string), name(string)
- webhook_plugin:
  - fields: post_email_opened(boolean), post_email_opened_url(string), post_email_reply(boolean), post_email_reply_url(string), post_new_prospect(boolean), post_new_prospect_url(string), post_updated_prospect(boolean), post_updated_prospect_url(string), raw_events(boolean), raw_events_url(string)
- campaign_leads:
  - primary key: id
  - fields: campaign_id(string), id(string), lead(object), mailbox_id(string)
- campaign_replies:
  - primary key: id
  - fields: body(string), campaign_id(string), cc_emails(array), from_email(string), id(string), kind(string), lead_id(string), preview(string), sent_at(string), sentiment(string), step_message_id(string), subject(string), to_emails(array)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- update_lead:
  - endpoint: PATCH /v1/leads/{{ record.id }}
  - required fields: id
  - risk: external mutation of an existing PersistIQ lead's fields; changing status/status_id/owner_id can move a lead into or out of active outbound-sequence automation depending on the target account's own campaign rules; approval required
- create_campaign:
  - endpoint: POST /v1/campaigns
  - required fields: campaign_name, owner_id
  - risk: creates a new outbound-email campaign in the target PersistIQ account; approval required
- duplicate_campaign:
  - endpoint: POST /v1/campaigns/duplicate
  - required fields: campaign_id, owner_id
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
  - required fields: campaign_id, inbox_message_id, body
  - risk: sends a real outbound email reply on behalf of the campaign's mailbox owner; irreversible once delivered; approval required
- add_dnc_domain:
  - endpoint: POST /v1/dnc_domains
  - required fields: name
  - risk: adds a domain to the account's Do-Not-Contact list; blocks future outreach to that domain account-wide; approval required

## Security

- read risk: external PersistIQ API read of leads, users, campaigns, mailboxes, activities, accounts, DNC domains, events, lead fields/statuses, tags, webhook plugin settings, and per-campaign leads/replies
- write risk: external mutation of PersistIQ leads and campaigns: update_lead can move a lead into or out of active outbound-sequence automation; create_campaign/duplicate_campaign create new live campaigns; add_lead_to_campaign enrolls a lead into automated outreach immediately depending on campaign state; remove_lead_from_campaign stops scheduled outreach to a lead; reply_to_campaign_message sends a real outbound email on behalf of the mailbox owner; add_dnc_domain blocks future outreach to a domain account-wide
- approval: required; every write can trigger or halt outbound-email automation, send a real email, or change account-wide contact policy outside this connector's control depending on the target PersistIQ account's own configuration
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run PersistIQ's declared streams and reverse-ETL actions.
- Usage: pm persistiq <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - accounts list - Run the accounts ETL stream [intent=etl availability=implemented stream=accounts]; notes: discrepancy=present-in-surface-absent-from-artifact
  - activities list - Run the activities ETL stream [intent=etl availability=implemented stream=activities]; notes: discrepancy=present-in-surface-absent-from-artifact
  - add dnc domain apply - Plan and execute the add dnc domain reverse-ETL action [intent=reverse_etl availability=implemented write=add_dnc_domain]; approval: requires plan, preview, approval, and execute; risk: adds a domain to the account's Do-Not-Contact list; blocks future outreach to that domain account-wide; approval required; flags: --name (required)
  - add lead to campaign apply - Plan and execute the add lead to campaign reverse-ETL action [intent=reverse_etl availability=implemented write=add_lead_to_campaign]; approval: requires plan, preview, approval, and execute; risk: enrolls a lead into a live outbound-email campaign; the lead may start receiving automated outreach immediately depending on campaign schedule/state; approval required; flags: --campaign_id (required)
  - api get v1 leads id - Documented GET /v1/leads/{id} (not implemented) [intent=direct_read availability=not_implemented operation=persistiq.get.v1-leads-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post v1 leads - Documented POST /v1/leads (not implemented) [intent=direct_write availability=not_implemented operation=persistiq.post.v1-leads]; approval: not implemented: the REST write executor lacks the provider-specific top-level body envelope required by this operation; risk: high; notes: named_dependency=engine.rest_write_body_envelope: the REST write executor lacks the provider-specific top-level body envelope required by this operation
  - api put v1 webhook-plugin - Documented PUT /v1/webhook_plugin (not implemented) [intent=direct_write availability=not_implemented operation=persistiq.put.v1-webhook-plugin]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
  - campaign leads list - Run the campaign leads ETL stream [intent=etl availability=implemented stream=campaign_leads]
  - campaign replies list - Run the campaign replies ETL stream [intent=etl availability=implemented stream=campaign_replies]
  - campaigns list - Run the campaigns ETL stream [intent=etl availability=implemented stream=campaigns]
  - create campaign apply - Plan and execute the create campaign reverse-ETL action [intent=reverse_etl availability=implemented write=create_campaign]; approval: requires plan, preview, approval, and execute; risk: creates a new outbound-email campaign in the target PersistIQ account; approval required; flags: --campaign_name (required), --owner_id (required)
  - dnc domains list - Run the dnc domains ETL stream [intent=etl availability=implemented stream=dnc_domains]
  - duplicate campaign apply - Plan and execute the duplicate campaign reverse-ETL action [intent=reverse_etl availability=implemented write=duplicate_campaign]; approval: requires plan, preview, approval, and execute; risk: duplicates an existing campaign (including its steps/sequence) into a new campaign in the target account; approval required; flags: --campaign_id (required), --owner_id (required)
  - events list - Run the events ETL stream [intent=etl availability=implemented stream=events]
  - lead fields list - Run the lead fields ETL stream [intent=etl availability=implemented stream=lead_fields]
  - lead statuses list - Run the lead statuses ETL stream [intent=etl availability=implemented stream=lead_statuses]
  - leads list - Run the leads ETL stream [intent=etl availability=implemented stream=leads]
  - mailboxes list - Run the mailboxes ETL stream [intent=etl availability=implemented stream=mailboxes]; notes: discrepancy=present-in-surface-absent-from-artifact
  - remove lead from campaign apply - Plan and execute the remove lead from campaign reverse-ETL action [intent=reverse_etl availability=implemented write=remove_lead_from_campaign]; approval: requires plan, preview, approval, and execute; risk: removes a lead from a live outbound-email campaign, stopping any further scheduled automated outreach to it in that sequence; approval required; flags: --campaign_id (required), --id (required)
  - reply to campaign message apply - Plan and execute the reply to campaign message reverse-ETL action [intent=reverse_etl availability=implemented write=reply_to_campaign_message]; approval: requires plan, preview, approval, and execute; risk: sends a real outbound email reply on behalf of the campaign's mailbox owner; irreversible once delivered; approval required; flags: --body (required), --campaign_id (required), --inbox_message_id (required)
  - tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]
  - update lead apply - Plan and execute the update lead reverse-ETL action [intent=reverse_etl availability=implemented write=update_lead]; approval: requires plan, preview, approval, and execute; risk: external mutation of an existing PersistIQ lead's fields; changing status/status_id/owner_id can move a lead into or out of active outbound-sequence automation depending on the target account's own campaign rules; approval required; flags: --id (required)
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]
  - webhook plugin list - Run the webhook plugin ETL stream [intent=etl availability=implemented stream=webhook_plugin]

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
