# pm connectors inspect lemlist

```text
NAME
  pm connectors inspect lemlist - Lemlist connector manual

SYNOPSIS
  pm connectors inspect lemlist
  pm connectors inspect lemlist --json
  pm credentials add <name> --connector lemlist [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads lemlist campaigns, activities, team metadata, CRM contacts/companies, schedules, tasks, webhooks, unsubscribes, field definitions, and signal-agent data through the lemlist REST API.

ICON
  asset: icons/lemlist.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.lemlist.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret)

ETL STREAMS
  campaigns:
    primary key: _id
    fields: _id(), labels(), name()
  team:
    primary key: _id
    fields: _id(), _updatedAt(), beta(), billing(), createdAt(), createdBy(), name(), revenueVisualization(), userIds()
  team_senders:
    primary key: userId
    fields: campaigns(), userId()
  team_credits:
    fields: credits(), details()
  team_crm_users:
    primary key: userId, crm
    fields: crm(), userId()
  activities:
    primary key: _id
    fields: _id(), campaignId(), campaignName(), companyName(), createdAt(), createdBy(), email(), emailTemplateId(), emailTemplateName(), extra(), firstName(), icebreaker(), isFirst(), lastName(), leadEmail(), leadFirstName(), leadId(), leadLastName(), linkedinUrl(), phone(), sequenceId(), sequenceStep(), teamId(), type()
  unsubscribes:
    primary key: _id
    fields: _id(), email()
  schedules:
    primary key: _id
    fields: _id(), createdAt(), createdBy(), deletedAt(), deletedBy(), end(), name(), public(), secondsToWait(), start(), teamId(), timezone(), weekdays()
  database_filters:
    primary key: _id
    fields: _id(), criteria(), name()
  tasks:
    primary key: _id
    fields: _id(), campaignId(), completedAt(), createdAt(), dueDate(), leadId(), status(), type(), userId()
  inbox_labels:
    primary key: _id
    fields: _id(), color(), createdAt(), createdBy(), name()
  contacts:
    primary key: _id
    fields: _id(), campaigns(), createdAt(), createdBy(), email(), fields(), fullName(), ownerId(), teamId(), unsubscribed()
  contact_lists:
    primary key: _id
    fields: _id(), dynamic(), name()
  companies:
    primary key: _id
    fields: _id(), createdAt(), createdBy(), crmSync(), domain(), fields(), industry(), location(), name(), ownerId(), size()
  webhooks:
    primary key: _id
    fields: _id(), campaignId(), createdAt(), targetUrl(), type(), zapId()
  unsubscribed_variables:
    primary key: _id
    fields: _id(), createdAt(), source(), value()
  watchlist_signals:
    primary key: _id
    fields: _id(), company(), contact(), createdAt(), receivedAt(), signalData(), status(), teamId(), type(), watchListId(), watchListName()
  user_channels:
    fields: email(), linkedin(), plan(), whatsapp()
  fields_contact:
    primary key: name
    fields: crmField(), label(), name(), source(), type()
  fields_company:
    primary key: name
    fields: crmField(), label(), name(), source(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external lemlist API read of campaign, outreach, CRM, inbox metadata, unsubscribe, webhook, and signal-agent data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect lemlist

  # Inspect as structured JSON
  pm connectors inspect lemlist --json

AGENT WORKFLOW
  - Run pm connectors inspect lemlist before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
