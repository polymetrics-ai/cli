# pm connectors inspect brevo

```text
NAME
  pm connectors inspect brevo - Brevo connector manual

SYNOPSIS
  pm connectors inspect brevo
  pm connectors inspect brevo --json
  pm credentials add <name> --connector brevo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Brevo (formerly Sendinblue) contacts, email campaigns, contact lists, segments, senders, sender domains, CRM companies/deals, and webhooks through the Brevo REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  start_date
  api_key (secret)

ETL STREAMS
  contacts:
    primary key: id
    cursor: modifiedAt
    fields: attributes(), createdAt(), email(), emailBlacklisted(), id(), listIds(), modifiedAt(), smsBlacklisted()
  email_campaigns:
    primary key: id
    cursor: modifiedAt
    fields: createdAt(), id(), modifiedAt(), name(), status(), subject(), type()
  contacts_lists:
    primary key: id
    fields: folderId(), id(), name(), totalBlacklisted(), totalSubscribers(), uniqueSubscribers()
  senders:
    primary key: id
    fields: active(), email(), id(), name()
  senders_domains:
    primary key: id
    fields: authenticated(), domain_name(), id(), ip(), verified()
  contacts_segments:
    primary key: id
    fields: categoryName(), id(), segmentName(), updatedAt()
  companies:
    primary key: id
    cursor: last_updated_at
    fields: attributes(), id(), last_updated_at(), linkedContactsIds(), linkedDealsIds()
  crm_deals:
    primary key: id
    cursor: last_updated_date
    fields: attributes(), id(), last_updated_date(), linkedCompaniesIds(), linkedContactsIds()
  webhooks:
    primary key: id
    cursor: modifiedAt
    fields: channel(), createdAt(), description(), events(), id(), modifiedAt(), type(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_contact:
    endpoint: POST /contacts
    risk: creates a new marketing contact; low-risk external mutation, no approval required
  update_contact:
    endpoint: PUT /contacts/{{ record.identifier }}
    required fields: identifier
    risk: mutates an existing contact's attributes, list membership, or blacklist status; changing emailBlacklisted/smsBlacklisted affects real send eligibility
  delete_contact:
    endpoint: DELETE /contacts/{{ record.identifier }}
    required fields: identifier
    risk: permanently removes a contact and its engagement history; irreversible
  create_contacts_list:
    endpoint: POST /contacts/lists
    risk: creates a new contact list under an existing folder; low-risk external mutation, no approval required
  create_sender:
    endpoint: POST /senders
    risk: registers a new verified-sending identity; Brevo emails a verification link to the address before it can send
  update_sender:
    endpoint: PUT /senders/{{ record.senderId }}
    required fields: senderId
    risk: mutates an existing sender's from-name, email, or dedicated-IP pool; affects all campaigns using this sender going forward
  delete_sender:
    endpoint: DELETE /senders/{{ record.senderId }}
    required fields: senderId
    risk: permanently removes a sending identity; any scheduled campaign still referencing it will fail to send
  create_company:
    endpoint: POST /companies
    risk: creates a new CRM company record; low-risk external mutation, no approval required
  update_company:
    endpoint: PATCH /companies/{{ record.id }}
    required fields: id
    risk: mutates an existing CRM company's name, attributes, or linked contact/deal set
  create_deal:
    endpoint: POST /crm/deals
    risk: creates a new CRM deal record; low-risk external mutation, no approval required
  update_deal:
    endpoint: PATCH /crm/deals/{{ record.id }}
    required fields: id
    risk: mutates an existing CRM deal's stage, amount, or linked contact/company set
  create_webhook:
    endpoint: POST /webhooks
    risk: registers live event delivery (opens/clicks/bounces/unsubscribes) to an external endpoint of the caller's choosing; review the target before enabling, per metadata.json risk.write
  update_webhook:
    endpoint: PUT /webhooks/{{ record.webhookId }}
    required fields: webhookId
    risk: re-points an already-registered webhook's delivery URL or event set; redirects live event delivery immediately
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.webhookId }}
    required fields: webhookId
    risk: permanently removes a webhook subscription; irreversible

SECURITY
  read risk: external Brevo API read of contact, campaign, CRM, and sender data
  write risk: external mutation of contacts, contact lists, senders, CRM companies/deals, and webhooks; webhook writes register live event delivery to a caller-chosen URL
  approval: required for all write actions; each action's per-record risk string in writes.json is the authoritative summary
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect brevo

  # Inspect as structured JSON
  pm connectors inspect brevo --json

AGENT WORKFLOW
  - Run pm connectors inspect brevo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
