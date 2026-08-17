---
name: pm-mailjet-mail
description: Mailjet Mail connector knowledge and safe action guide.
---

# pm-mailjet-mail

## Purpose

Reads Mailjet contacts, contact lists, messages, campaigns, and statistics through the Mailjet Email REST API (v3).

## Icon

- asset: icons/mailjetmail.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://dev.mailjet.com/email/reference/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- api_key
- base_url
- mode
- page_size
- api_key_secret (secret)

## ETL Streams

- contacts:
  - primary key: ID
  - fields: CreatedAt(), DeliveredCount(), Email(), ID(), IsExcludedFromCampaigns(), IsOptInPending(), IsSpamComplaining(), LastActivityAt(), LastUpdateAt(), Name()
- contactslists:
  - primary key: ID
  - fields: Address(), CreatedAt(), ID(), IsDeleted(), Name(), SubscriberCount()
- messages:
  - primary key: ID
  - fields: ArrivedAt(), AttemptCount(), CampaignID(), ContactID(), ID(), IsClickTracked(), IsOpenTracked(), MessageSize(), Status()
- campaigns:
  - primary key: ID
  - fields: CreatedAt(), FromEmail(), FromName(), ID(), IsDeleted(), IsStarred(), SendStartAt(), Status(), Subject()
- stats:
  - primary key: ID
  - fields: ID(), MessageBouncedCount(), MessageClickedCount(), MessageDeliveredCount(), MessageOpenedCount(), MessageSentCount(), MessageSpamCount(), MessageUnsubscribedCount()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Mailjet API read of contact, list, message, campaign, and statistics data
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mailjet-mail
```

### Inspect as structured JSON

```bash
pm connectors inspect mailjet-mail --json
```

## Agent Rules

- Run pm connectors inspect mailjet-mail before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
