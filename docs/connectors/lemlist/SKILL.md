---
name: pm-lemlist
description: Lemlist connector knowledge and safe action guide.
---

# pm-lemlist

## Purpose

Reads lemlist campaigns, activities, team metadata, CRM contacts/companies, schedules, tasks, webhooks, unsubscribes, field definitions, and signal-agent data through the lemlist REST API.

## Icon

- id: lemlist
- asset: icons/lemlist.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.lemlist.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret) (required)

## ETL Streams

- campaigns:
  - primary key: _id
  - fields: _id(string), labels(array), name(string)
- team:
  - primary key: _id
  - fields: _id(string), _updatedAt(string), beta(array), billing(object), createdAt(string), createdBy(string), name(string), revenueVisualization(object), userIds(array)
- team_senders:
  - primary key: userId
  - fields: campaigns(array), userId(string)
- team_credits:
  - fields: credits(integer), details(object)
- team_crm_users:
  - primary key: userId, crm
  - fields: crm(string), userId(string)
- activities:
  - primary key: _id
  - fields: _id(string), campaignId(string), campaignName(string), companyName(string), createdAt(string), createdBy(string), email(string), emailTemplateId(string), emailTemplateName(string), extra(object), firstName(string), icebreaker(string), isFirst(boolean), lastName(string), leadEmail(string), leadFirstName(string), leadId(string), leadLastName(string), linkedinUrl(string), phone(string), sequenceId(string), sequenceStep(integer), teamId(string), type(string)
- unsubscribes:
  - primary key: _id
  - fields: _id(string), email(string)
- schedules:
  - primary key: _id
  - fields: _id(string), createdAt(string), createdBy(string), deletedAt(string), deletedBy(string), end(string), name(string), public(boolean), secondsToWait(integer), start(string), teamId(string), timezone(string), weekdays(array)
- database_filters:
  - primary key: _id
  - fields: _id(string), criteria(object), name(string)
- tasks:
  - primary key: _id
  - fields: _id(string), campaignId(string), completedAt(string), createdAt(string), dueDate(string), leadId(string), status(string), type(string), userId(string)
- inbox_labels:
  - primary key: _id
  - fields: _id(string), color(string), createdAt(string), createdBy(string), name(string)
- contacts:
  - primary key: _id
  - fields: _id(string), campaigns(array), createdAt(string), createdBy(string), email(string), fields(object), fullName(string), ownerId(string), teamId(string), unsubscribed(boolean)
- contact_lists:
  - primary key: _id
  - fields: _id(string), dynamic(boolean), name(string)
- companies:
  - primary key: _id
  - fields: _id(string), createdAt(string), createdBy(string), crmSync(object), domain(string), fields(object), industry(string), location(string), name(string), ownerId(string), size(string)
- webhooks:
  - primary key: _id
  - fields: _id(string), campaignId(string), createdAt(string), targetUrl(string), type(string), zapId(integer)
- unsubscribed_variables:
  - primary key: _id
  - fields: _id(string), createdAt(string), source(string), value(string)
- watchlist_signals:
  - primary key: _id
  - fields: _id(string), company(object), contact(object), createdAt(string), receivedAt(string), signalData(object), status(string), teamId(string), type(string), watchListId(string), watchListName(string)
- user_channels:
  - fields: email(object), linkedin(object), plan(string), whatsapp(object)
- fields_contact:
  - primary key: name
  - fields: crmField(string), label(string), name(string), source(string), type(string)
- fields_company:
  - primary key: name
  - fields: crmField(string), label(string), name(string), source(string), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external lemlist API read of campaign, outreach, CRM, inbox metadata, unsubscribe, webhook, and signal-agent data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect lemlist
```

### Inspect as structured JSON

```bash
pm connectors inspect lemlist --json
```

## Agent Rules

- Run pm connectors inspect lemlist before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
