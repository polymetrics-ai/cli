---
name: pm-smartreach
description: SmartReach connector knowledge and safe action guide.
---

# pm-smartreach

## Purpose

Reads SmartReach teams, campaigns, prospects, email settings, do-not-contact records, users, and accounts; creates/updates prospects and accounts, manages campaign membership/status, do-not-contact entries, and task status.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- newer_than
- older_than
- team_id
- api_key (secret) (required)

## ETL Streams

- campaigns:
  - primary key: id
  - fields: created_at(string), id(string), name(string)
- prospects:
  - primary key: id
  - fields: created_at(string), id(string), name(string)
- teams:
  - primary key: id
  - fields: created_at(string), id(string), name(string)
- email_settings:
  - primary key: id
  - fields: created_at(string), id(string), name(string)
- do_not_contact:
  - primary key: id
  - fields: created_at(string), id(string), name(string)
- users:
  - primary key: id
  - cursor: created_at
  - fields: created_at(integer), email(string), first_name(string), id(string), last_name(string), object(string), org_id(string), status(string), timezone(string)
- accounts:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(number), custom_fields(object), custom_id(string), description(string), id(string), industry(string), linkedin_url(string), name(string), object(string), source(string), updated_at(number), website(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- add_or_update_prospects:
  - endpoint: POST prospects?team_id={{ config.team_id }}
  - risk: external mutation; creates or updates a prospect (deduped by email/phone/linkedin_url/company+name per SmartReach's unique_identifier_columns rule) on the connected SmartReach account; approval required
- add_prospects_to_campaign:
  - endpoint: POST campaigns/{{ record.campaign_id }}/prospects?team_id={{ config.team_id }}
  - required fields: campaign_id, prospect_ids
  - risk: external mutation; enrolls prospects into an outbound campaign, triggering scheduled sequence messages to real recipients; approval required
- unassign_prospects_from_campaign:
  - endpoint: PUT campaigns/{{ record.campaign_id }}/prospects?team_id={{ config.team_id }}
  - required fields: campaign_id, prospect_ids
  - risk: external mutation; removes prospects from an outbound campaign, stopping any further scheduled sequence messages to them; approval required
- update_prospect_campaign_status:
  - endpoint: PUT prospects/prospect_status_change?team_id={{ config.team_id }}
  - required fields: prospect_ids, prospect_status, campaign_ids
  - risk: external mutation; changes a prospect's engagement status for a campaign (e.g. pausing/resuming/marking replied), affecting whether further outreach messages are sent; approval required
- update_campaign_status:
  - endpoint: PUT campaigns/{{ record.campaign_id }}/status?team_id={{ config.team_id }}
  - required fields: campaign_id, status
  - risk: external mutation; starts, schedules, or stops an entire outbound campaign, directly controlling whether real outreach messages are sent to prospects; approval required
- remove_from_do_not_contact:
  - endpoint: DELETE do-not-contact?team_id={{ config.team_id }}
  - required fields: dnc_ids
  - risk: external mutation; removes entries from the do-not-contact suppression list, which re-enables outreach to those emails/domains; approval required
- add_emails_to_do_not_contact:
  - endpoint: POST do-not-contact/email?team_id={{ config.team_id }}
  - required fields: emails
  - risk: external mutation; blacklists emails from all future outreach on the connected SmartReach account; approval required
- add_domains_to_do_not_contact:
  - endpoint: POST do-not-contact/domain?team_id={{ config.team_id }}
  - required fields: domains
  - risk: external mutation; blacklists entire domains from all future outreach on the connected SmartReach account; approval required
- create_or_update_account:
  - endpoint: POST accounts?team_id={{ config.team_id }}
  - required fields: name
  - risk: external mutation; creates or updates a CRM-style account (company) record on the connected SmartReach account; approval required
- update_task_status:
  - endpoint: PUT tasks/{{ record.task_id }}/status
  - required fields: task_id, status_type
  - optional fields: due_at, snoozed_till
  - risk: external mutation; changes a sales task's status (due/snoozed/done/skipped) on the connected SmartReach account; approval required

## Security

- read risk: read-only team/campaign/prospect/email-setting/do-not-contact/user/account data from a connected SmartReach account
- write risk: creates/updates prospects and CRM accounts, enrolls/unenrolls prospects in outbound campaigns, starts/stops entire campaigns, changes prospect engagement status, blacklists emails/domains from outreach, and changes sales task status
- approval: required for all 10 write actions; read is unapproved
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect smartreach
```

### Inspect as structured JSON

```bash
pm connectors inspect smartreach --json
```

## Agent Rules

- Run pm connectors inspect smartreach before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
