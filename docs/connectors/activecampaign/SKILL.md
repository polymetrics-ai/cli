---
name: pm-activecampaign
description: ActiveCampaign connector knowledge and safe action guide.
---

# pm-activecampaign

## Purpose

Reads ActiveCampaign contacts, lists, deals, campaigns, tags, automations, custom fields, accounts, users, deal stages, and deal tasks through the ActiveCampaign v3 REST API.

## Icon

- asset: icons/activecampaign.svg
- source: official
- review_status: official_verified
- review_url: https://developers.activecampaign.com/reference/overview

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: udate
  - fields: cdate(), deleted(), email(), firstName(), id(), lastName(), orgid(), phone(), udate()
- lists:
  - primary key: id
  - cursor: cdate
  - fields: cdate(), id(), name(), sender_url(), stringid(), subscriber_count(), userid()
- deals:
  - primary key: id
  - cursor: mdate
  - fields: cdate(), contact(), currency(), id(), mdate(), owner(), stage(), status(), title(), value()
- campaigns:
  - primary key: id
  - cursor: cdate
  - fields: cdate(), id(), linkclicks(), mdate(), name(), opens(), send_amt(), status(), subject(), type(), uniqueopens()
- tags:
  - primary key: id
  - cursor: cdate
  - fields: cdate(), description(), id(), subscriber_count(), tag(), tagType()
- automations:
  - primary key: id
  - cursor: mdate
  - fields: cdate(), entered(), exited(), hidden(), id(), mdate(), name(), status(), userid()
- fields:
  - primary key: id
  - cursor: udate
  - fields: cdate(), descript(), id(), isrequired(), perstag(), title(), type(), udate(), visible()
- accounts:
  - primary key: id
  - cursor: updatedTimestamp
  - fields: accountUrl(), contactCount(), createdTimestamp(), dealCount(), id(), name(), updatedTimestamp()
- users:
  - primary key: id
  - fields: email(), firstName(), id(), lastName(), phone(), username()
- deal_stages:
  - primary key: id
  - cursor: udate
  - fields: cdate(), color(), group(), id(), order(), title(), udate()
- deal_tasks:
  - primary key: id
  - cursor: udate
  - fields: cdate(), duedate(), id(), note(), relid(), reltype(), status(), title(), udate()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external ActiveCampaign API read of contacts, lists, deals, campaigns, tags, automations, custom fields, accounts, users, deal stages, and deal tasks
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect activecampaign
```

### Inspect as structured JSON

```bash
pm connectors inspect activecampaign --json
```

## Agent Rules

- Run pm connectors inspect activecampaign before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
