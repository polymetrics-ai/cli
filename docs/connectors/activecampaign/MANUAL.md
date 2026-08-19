# pm connectors inspect activecampaign

```text
NAME
  pm connectors inspect activecampaign - ActiveCampaign connector manual

SYNOPSIS
  pm connectors inspect activecampaign
  pm connectors inspect activecampaign --json
  pm credentials add <name> --connector activecampaign [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads ActiveCampaign contacts, lists, deals, campaigns, tags, automations, custom fields, accounts, users, deal stages, and deal tasks through the ActiveCampaign v3 REST API.

ICON
  id: activecampaign
  asset: icons/activecampaign.svg
  source: official
  review_status: official_verified
  review_url: https://developers.activecampaign.com/reference/overview

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url (required)
  mode
  api_key (secret) (required)

ETL STREAMS
  contacts:
    primary key: id
    cursor: udate
    fields: cdate(string), deleted(string), email(string), firstName(string), id(string), lastName(string), orgid(string), phone(string), udate(string)
  lists:
    primary key: id
    cursor: cdate
    fields: cdate(string), id(string), name(string), sender_url(string), stringid(string), subscriber_count(string), userid(string)
  deals:
    primary key: id
    cursor: mdate
    fields: cdate(string), contact(string), currency(string), id(string), mdate(string), owner(string), stage(string), status(string), title(string), value(string)
  campaigns:
    primary key: id
    cursor: cdate
    fields: cdate(string), id(string), linkclicks(string), mdate(string), name(string), opens(string), send_amt(string), status(string), subject(string), type(string), uniqueopens(string)
  tags:
    primary key: id
    cursor: cdate
    fields: cdate(string), description(string), id(string), subscriber_count(string), tag(string), tagType(string)
  automations:
    primary key: id
    cursor: mdate
    fields: cdate(string), entered(string), exited(string), hidden(string), id(string), mdate(string), name(string), status(string), userid(string)
  fields:
    primary key: id
    cursor: udate
    fields: cdate(string), descript(string), id(string), isrequired(string), perstag(string), title(string), type(string), udate(string), visible(string)
  accounts:
    primary key: id
    cursor: updatedTimestamp
    fields: accountUrl(string), contactCount(string), createdTimestamp(string), dealCount(string), id(string), name(string), updatedTimestamp(string)
  users:
    primary key: id
    fields: email(string), firstName(string), id(string), lastName(string), phone(string), username(string)
  deal_stages:
    primary key: id
    cursor: udate
    fields: cdate(string), color(string), group(string), id(string), order(string), title(string), udate(string)
  deal_tasks:
    primary key: id
    cursor: udate
    fields: cdate(string), duedate(string), id(string), note(string), relid(string), reltype(string), status(integer), title(string), udate(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external ActiveCampaign API read of contacts, lists, deals, campaigns, tags, automations, custom fields, accounts, users, deal stages, and deal tasks
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect activecampaign

  # Inspect as structured JSON
  pm connectors inspect activecampaign --json

AGENT WORKFLOW
  - Run pm connectors inspect activecampaign before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
