---
name: pm-brevo
description: Brevo connector knowledge and safe action guide.
---

# pm-brevo

## Purpose

Reads and writes Brevo (formerly Sendinblue) contacts, email campaigns, contact lists, segments, senders, sender domains, CRM companies/deals, and webhooks through the Brevo REST API.

## Icon

- id: simple-icons-brevo
- asset: icons/simple-icons/brevo.svg
- title: Brevo
- simple_icon_slug: brevo
- simple_icon_hex: 0B996E
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Brevo
- match: exact-name-or-slug
- matched_by: brevo

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- start_date
- api_key (secret) (required)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: modifiedAt
  - fields: attributes(object), createdAt(string), email(string), emailBlacklisted(boolean), id(integer), listIds(array), modifiedAt(string), smsBlacklisted(boolean)
- email_campaigns:
  - primary key: id
  - cursor: modifiedAt
  - fields: createdAt(string), id(integer), modifiedAt(string), name(string), status(string), subject(string), type(string)
- contacts_lists:
  - primary key: id
  - fields: folderId(integer), id(integer), name(string), totalBlacklisted(integer), totalSubscribers(integer), uniqueSubscribers(integer)
- senders:
  - primary key: id
  - fields: active(boolean), email(string), id(integer), name(string)
- senders_domains:
  - primary key: id
  - fields: authenticated(boolean), domain_name(string), id(integer), ip(string), verified(boolean)
- contacts_segments:
  - primary key: id
  - fields: categoryName(string), id(integer), segmentName(string), updatedAt(string)
- companies:
  - primary key: id
  - cursor: last_updated_at
  - fields: attributes(object), id(string), last_updated_at(string), linkedContactsIds(array), linkedDealsIds(array)
- crm_deals:
  - primary key: id
  - cursor: last_updated_date
  - fields: attributes(object), id(string), last_updated_date(string), linkedCompaniesIds(array), linkedContactsIds(array)
- webhooks:
  - primary key: id
  - cursor: modifiedAt
  - fields: channel(string), createdAt(string), description(string), events(array), id(integer), modifiedAt(string), type(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_contact:
  - endpoint: POST /contacts
  - risk: creates a new marketing contact; low-risk external mutation, no approval required
- update_contact:
  - endpoint: PUT /contacts/{{ record.identifier }}
  - required fields: identifier
  - risk: mutates an existing contact's attributes, list membership, or blacklist status; changing emailBlacklisted/smsBlacklisted affects real send eligibility
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.identifier }}
  - required fields: identifier
  - risk: permanently removes a contact and its engagement history; irreversible
- create_contacts_list:
  - endpoint: POST /contacts/lists
  - required fields: name, folderId
  - risk: creates a new contact list under an existing folder; low-risk external mutation, no approval required
- create_sender:
  - endpoint: POST /senders
  - required fields: name, email
  - risk: registers a new verified-sending identity; Brevo emails a verification link to the address before it can send
- update_sender:
  - endpoint: PUT /senders/{{ record.senderId }}
  - required fields: senderId
  - risk: mutates an existing sender's from-name, email, or dedicated-IP pool; affects all campaigns using this sender going forward
- delete_sender:
  - endpoint: DELETE /senders/{{ record.senderId }}
  - required fields: senderId
  - risk: permanently removes a sending identity; any scheduled campaign still referencing it will fail to send
- create_company:
  - endpoint: POST /companies
  - required fields: name
  - risk: creates a new CRM company record; low-risk external mutation, no approval required
- update_company:
  - endpoint: PATCH /companies/{{ record.id }}
  - required fields: id
  - risk: mutates an existing CRM company's name, attributes, or linked contact/deal set
- create_deal:
  - endpoint: POST /crm/deals
  - required fields: name
  - risk: creates a new CRM deal record; low-risk external mutation, no approval required
- update_deal:
  - endpoint: PATCH /crm/deals/{{ record.id }}
  - required fields: id
  - risk: mutates an existing CRM deal's stage, amount, or linked contact/company set
- create_webhook:
  - endpoint: POST /webhooks
  - required fields: url, events
  - risk: registers live event delivery (opens/clicks/bounces/unsubscribes) to an external endpoint of the caller's choosing; review the target before enabling, per metadata.json risk.write
- update_webhook:
  - endpoint: PUT /webhooks/{{ record.webhookId }}
  - required fields: webhookId
  - risk: re-points an already-registered webhook's delivery URL or event set; redirects live event delivery immediately
- delete_webhook:
  - endpoint: DELETE /webhooks/{{ record.webhookId }}
  - required fields: webhookId
  - risk: permanently removes a webhook subscription; irreversible

## Security

- read risk: external Brevo API read of contact, campaign, CRM, and sender data
- write risk: external mutation of contacts, contact lists, senders, CRM companies/deals, and webhooks; webhook writes register live event delivery to a caller-chosen URL
- approval: required for all write actions; each action's per-record risk string in writes.json is the authoritative summary
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect brevo
```

### Inspect as structured JSON

```bash
pm connectors inspect brevo --json
```

## Agent Rules

- Run pm connectors inspect brevo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
